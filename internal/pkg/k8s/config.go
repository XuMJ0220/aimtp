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
	Kubeconfig string
	InCluster  bool
	QPS        float32
	Burst      int
	Timeout    time.Duration
	UserAgent  string
	Insecure   bool
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
