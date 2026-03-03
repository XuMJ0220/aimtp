package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"aimtp/internal/pkg/client"
	"aimtp/internal/pkg/log"
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
)

type TriggerBiz interface {
	HandleCreateDAG(ctx context.Context, payload []byte) error
}

type triggerBiz struct {
	controllerClients map[string]*client.WorkerClient
}

var _ TriggerBiz = (*triggerBiz)(nil)

func New(controllerClients map[string]*client.WorkerClient) *triggerBiz {
	return &triggerBiz{
		controllerClients: controllerClients,
	}
}

func (b *triggerBiz) HandleCreateDAG(ctx context.Context, payload []byte) error {
	var rq apiv1.CreateDAGRequest
	if err := json.Unmarshal(payload, &rq); err != nil {
		log.Errorw("Failed to unmarshal create dag request", "err", err)
		return err
	}

	

	cluster := rq.GetCluster()
	if cluster == "" {
		return fmt.Errorf("cluster is empty")
	}
	controller, ok := b.controllerClients[cluster]
	if !ok {
		return fmt.Errorf("controller not found for cluster %s", cluster)
	}

	if err := createDAG(ctx, controller, payload); err != nil {
		log.Errorw("Failed to call controller for dag", "err", err, "dag_name", rq.GetDagName(), "cluster", cluster)
		return err
	}

	log.Infow("Trigger handled dag", "dag_name", rq.GetDagName(), "cluster", cluster)
	return nil
}

func createDAG(ctx context.Context, controller *client.WorkerClient, payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("payload is empty")
	}
	url := fmt.Sprintf("%s/v1/dags", controller.BaseURL())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := controller.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call controller: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("controller returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
