//go:build windows
// +build windows

package containerd

import (
	"context"
	"os"

	"github.com/containerd/containerd"
	"github.com/rancher/wharfie/pkg/registries"
	"github.com/sirupsen/logrus"
	"github.com/xiaods/k8e/pkg/agent/templates"
	util2 "github.com/xiaods/k8e/pkg/agent/util"
	"github.com/xiaods/k8e/pkg/daemons/config"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

func getContainerdArgs(cfg *config.Node) []string {
	args := []string{
		"containerd",
		"-c", cfg.Containerd.Config,
	}
	return args
}

// setupContainerdConfig generates the containerd.toml, using a template combined with various
// runtime configurations and registry mirror settings provided by the administrator.
func setupContainerdConfig(ctx context.Context, cfg *config.Node) error {
	privRegistries, err := registries.GetPrivateRegistries(cfg.AgentConfig.PrivateRegistry)
	if err != nil {
		return err
	}

	if cfg.SELinux {
		logrus.Warn("SELinux isn't supported on windows")
	}

	var containerdTemplate string

	containerdConfig := templates.ContainerdConfig{
		NodeConfig:            cfg,
		DisableCgroup:         true,
		SystemdCgroup:         false,
		IsRunningInUserNS:     false,
		PrivateRegistryConfig: privRegistries.Registry,
	}

	containerdTemplateBytes, err := os.ReadFile(cfg.Containerd.Template)
	if err == nil {
		logrus.Infof("Using containerd template at %s", cfg.Containerd.Template)
		containerdTemplate = string(containerdTemplateBytes)
	} else if os.IsNotExist(err) {
		containerdTemplate = templates.ContainerdConfigTemplate
	} else {
		return err
	}
	parsedTemplate, err := templates.ParseTemplateFromConfig(containerdTemplate, containerdConfig)
	if err != nil {
		return err
	}

	return util2.WriteFile(cfg.Containerd.Config, parsedTemplate)
}

func Client(address string) (*containerd.Client, error) {
	addr, _, err := util.GetAddressAndDialer(address)
	if err != nil {
		return nil, err
	}

	return containerd.New(addr)
}
