package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigOptions struct {
	Kubeconfig string         `json:"kubeconfig" mapstructure:"kubeconfig"`
	InCluster  bool           `json:"in-cluster" mapstructure:"in-cluster"`
	QPS        float32        `json:"qps" mapstructure:"qps"`
	Burst      int            `json:"burst" mapstructure:"burst"`
	Timeout    time.Duration  `json:"timeout" mapstructure:"timeout"`
	UserAgent  string         `json:"user-agent" mapstructure:"user-agent"`
	Insecure   bool           `json:"insecure" mapstructure:"insecure"`
	Storage    StorageOptions `json:"storage" mapstructure:"storage"`
}

type StorageOptions struct {
	Type           string `json:"type" mapstructure:"type"`                         // hostPath, pvc, emptyDir
	HostPathPrefix string `json:"host-path-prefix" mapstructure:"host-path-prefix"` // e.g. /run/desktop/mnt/host/d/aimtp/data/ or /data/
	PVCName        string `json:"pvc-name" mapstructure:"pvc-name"`                 // if Type == pvc
}

func NewRestConfig(opts ConfigOptions) (*rest.Config, error) {
	var inClusterErr error
	if opts.InCluster {
		cfg, err := rest.InClusterConfig()
		if err == nil {
			applyConfigOptions(cfg, opts)
			return cfg, nil
		}
		inClusterErr = err
	}

	kubeconfig := opts.Kubeconfig
	if kubeconfig == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfig = env
		} else if home, err := os.UserHomeDir(); err == nil {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		if inClusterErr != nil {
			return nil, fmt.Errorf("in-cluster config failed: %w; kubeconfig config failed: %w", inClusterErr, err)
		}
		return nil, err
	}

	applyConfigOptions(cfg, opts)
	return cfg, nil
}

func applyConfigOptions(cfg *rest.Config, opts ConfigOptions) {
	if opts.QPS > 0 {
		cfg.QPS = opts.QPS
	}
	if opts.Burst > 0 {
		cfg.Burst = opts.Burst
	}
	if opts.Timeout > 0 {
		cfg.Timeout = opts.Timeout
	}
	if opts.UserAgent != "" {
		cfg.UserAgent = opts.UserAgent
	}
	if opts.Insecure {
		cfg.Insecure = true
		cfg.TLSClientConfig.Insecure = true
		cfg.TLSClientConfig.CAData = nil
		cfg.TLSClientConfig.CAFile = ""
	}
}
