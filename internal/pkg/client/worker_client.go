package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WorkerClient Worker 客户端
type WorkerClient struct {
	baseURL    string
	httpClient *http.Client
	cluster    string
}

// sharedHTTPClient 共享的 HTTP 客户端
// 可以被多个 Worker 客户端共享，避免重复创建
var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// NewWorkerClient 创建 Worker 客户端
func NewWorkerClient(baseURL, cluster string) *WorkerClient {
	return &WorkerClient{
		baseURL:    baseURL,
		cluster:    cluster,
		httpClient: sharedHTTPClient,
	}
}

// HealthCheck 健康检查
func (c *WorkerClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call worker: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Worker returned %d", resp.StatusCode)
	}

	return nil
}

func NewControllerClients(cc map[string]string) (map[string]*WorkerClient, error) {
	controllerClients := map[string]*WorkerClient{}
	if len(cc) == 0 {
		return controllerClients, nil
	}
	for cluster, baseURL := range cc {
		controllerClients[cluster] = NewWorkerClient(baseURL, cluster)
	}
	return controllerClients, nil
}
