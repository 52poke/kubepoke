package haproxy

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/mudkipme/kubepoke/internal/interfaces"
)

type HAProxyHelper struct {
	config *HAProxyHelperConfig
}

type configData struct {
	Nodes             []*interfaces.NodeInfo
	HTTPInternalPort  int
	HTTPSInternalPort int
}

func NewHAProxyHelper(config *HAProxyHelperConfig) *HAProxyHelper {
	if config == nil {
		config = defaultHAProxyHelperConfig
	}
	if config.HAProxyConfigTemplate == "" {
		config.HAProxyConfigTemplate = defaultHAProxyHelperConfig.HAProxyConfigTemplate
	}
	if config.HAProxyConfigPath == "" {
		config.HAProxyConfigPath = defaultHAProxyHelperConfig.HAProxyConfigPath
	}
	if config.ReloadCommand == "" {
		config.ReloadCommand = defaultHAProxyHelperConfig.ReloadCommand
	}
	if config.HTTPInternalPort == 0 {
		config.HTTPInternalPort = defaultHAProxyHelperConfig.HTTPInternalPort
	}
	if config.HTTPSInternalPort == 0 {
		config.HTTPSInternalPort = defaultHAProxyHelperConfig.HTTPSInternalPort
	}
	return &HAProxyHelper{config: config}
}

func (h *HAProxyHelper) ClusterNodesUpdated(ctx context.Context, nodes []*interfaces.NodeInfo) error {
	configData := configData{Nodes: nodes, HTTPInternalPort: h.config.HTTPInternalPort, HTTPSInternalPort: h.config.HTTPSInternalPort}
	t := template.Must(template.New("haproxy").Parse(h.config.HAProxyConfigTemplate))
	var configOutput strings.Builder
	if err := t.Execute(&configOutput, configData); err != nil {
		slog.ErrorContext(ctx, "Failed to generate HAProxy configuration", "error", err)
		return err
	}

	if err := os.WriteFile(h.config.HAProxyConfigPath, []byte(configOutput.String()), 0644); err != nil {
		slog.ErrorContext(ctx, "Failed to write HAProxy configuration", "error", err)
		return err
	}

	fields := strings.Fields(h.config.ReloadCommand)
	cmd := exec.Command(fields[0], fields[1:]...)
	if err := cmd.Run(); err != nil {
		slog.ErrorContext(ctx, "Failed to reload HAProxy", "error", err)
		return err
	}

	slog.InfoContext(ctx, "HAProxy configuration updated")
	return nil
}
