package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/mudkipme/kubepoke/internal/haproxy"
	"github.com/mudkipme/kubepoke/internal/interfaces"
	"github.com/mudkipme/kubepoke/internal/kube"
	"github.com/mudkipme/kubepoke/internal/s3policy"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Config struct {
	Cron       string   `json:"cron"`
	Helpers    []string `json:"helpers"`
	Kubernetes struct {
		User    clientcmdapi.AuthInfo `json:"user"`
		Cluster clientcmdapi.Cluster  `json:"cluster"`
		Context clientcmdapi.Context  `json:"context"`
	} `json:"kubernetes"`
	HAProxy *haproxy.HAProxyHelperConfig   `json:"haproxy"`
	S3      *s3policy.S3PolicyHelperConfig `json:"s3"`
}

var config Config

func main() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/kubepoke/")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	if err = viper.Unmarshal(&config, viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	kubeService, err := kube.NewKubeService(&clientcmd.ConfigOverrides{
		AuthInfo:    config.Kubernetes.User,
		ClusterInfo: config.Kubernetes.Cluster,
		Context:     config.Kubernetes.Context,
	})
	if err != nil {
		log.Fatalf("Failed to create kube service: %v", err)
	}

	var helpers []interfaces.KubeHelper
	for _, helperName := range config.Helpers {
		switch helperName {
		case "haproxy":
			helpers = append(helpers, haproxy.NewHAProxyHelper(config.HAProxy))
		case "s3":
			s3policyHelper, err := s3policy.NewS3PolicyHelper(config.S3)
			if err != nil {
				log.Fatalf("Failed to create s3 policy helper: %v", err)
			}
			helpers = append(helpers, s3policyHelper)
		}
	}

	var latestNodeList []*interfaces.NodeInfo

	c := cron.New()
	c.AddFunc(config.Cron, func() {
		ctx := context.TODO()
		nodeList, err := kubeService.GetNodeInfo(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get node info", "error", err)
			return
		}

		if reflect.DeepEqual(nodeList, latestNodeList) {
			slog.InfoContext(ctx, "No changes in cluster nodes")
			return
		}

		for _, helper := range helpers {
			if err := helper.ClusterNodesUpdated(ctx, nodeList); err != nil {
				slog.ErrorContext(ctx, "Failed to update cluster nodes", "error", err)
				return
			}
		}

		latestNodeList = nodeList
	})

	c.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	c.Stop()
}
