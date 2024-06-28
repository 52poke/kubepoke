package kube

import (
	"context"
	"sort"

	"github.com/mudkipme/kubepoke/internal/interfaces"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeService struct {
	client *kubernetes.Clientset
}

func NewKubeService(config *clientcmd.ConfigOverrides) (*KubeService, error) {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		config,
	)

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &KubeService{
		client: clientset,
	}, nil
}

func (s *KubeService) GetNodeInfo(ctx context.Context) ([]*interfaces.NodeInfo, error) {
	// Get the list of nodes
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Extract internal IPs
	var nodeList []*interfaces.NodeInfo
	for _, node := range nodes.Items {
		var internalIP string
		var externalIP string
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case v1.NodeInternalIP:
				internalIP = address.Address
			case v1.NodeExternalIP:
				externalIP = address.Address
			}
		}
		if internalIP != "" {
			nodeList = append(nodeList, &interfaces.NodeInfo{
				Name:       node.Name,
				InternalIP: internalIP,
				ExternalIP: externalIP,
			})
		}
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].Name < nodeList[j].Name
	})

	return nodeList, nil
}
