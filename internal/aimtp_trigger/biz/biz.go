package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"aimtp/internal/aimtp_trigger/store"
	"aimtp/internal/pkg/client"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	apiv1 "aimtp/pkg/api/aimtp_server/v1"
)

type TriggerBiz interface {
	HandleCreateDAG(ctx context.Context, payload []byte) error
}

type triggerBiz struct {
	controllerClients map[string]*client.WorkerClient
	store             store.IStore
}

type controllerError struct {
	status int
	body   string
}

type errorResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (e *controllerError) Error() string {
	return fmt.Sprintf("controller returned %d: %s", e.status, e.body)
}

var _ TriggerBiz = (*triggerBiz)(nil)

func New(controllerClients map[string]*client.WorkerClient, store store.IStore) *triggerBiz {
	return &triggerBiz{
		controllerClients: controllerClients,
		store:             store,
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

	claimed, err := b.store.DAG().TryClaimCreation(ctx, rq.GetDagName())
	if err != nil {
		log.Errorw("Failed to claim dag creation", "err", err, "dag_name", rq.GetDagName())
		return err
	}
	if !claimed {
		log.Infow("Trigger skipped dag because it is already claimed", "dag_name", rq.GetDagName(), "cluster", cluster)
		return nil
	}

	if err := createDAG(ctx, controller, payload); err != nil {
		if isAlreadyExistsError(err) {
			if updateErr := b.store.DAG().UpdateCreationStatus(ctx, rq.GetDagName(), "created", nil); updateErr != nil {
				log.Errorw("Failed to update dag creation status after idempotent success", "err", updateErr, "dag_name", rq.GetDagName())
			}
			log.Infow("Trigger treated controller response as idempotent success", "dag_name", rq.GetDagName(), "cluster", cluster)
			return nil
		}
		log.Errorw("Failed to call controller for dag", "err", err, "dag_name", rq.GetDagName(), "cluster", cluster)
		errMsg := err.Error()
		if updateErr := b.store.DAG().UpdateCreationStatus(ctx, rq.GetDagName(), "pending", &errMsg); updateErr != nil {
			log.Errorw("Failed to rollback dag creation status", "err", updateErr, "dag_name", rq.GetDagName())
		}
		return err
	}

	if err := b.store.DAG().UpdateCreationStatus(ctx, rq.GetDagName(), "created", nil); err != nil {
		log.Errorw("Failed to update dag creation status", "err", err, "dag_name", rq.GetDagName())
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
		return &controllerError{status: resp.StatusCode, body: string(body)}
	}

	return nil
}

func isAlreadyExistsError(err error) bool {
	ce := &controllerError{}
	if !errors.As(err, &ce) {
		return false
	}
	if ce.status == http.StatusConflict {
		var resp errorResponse
		if json.Unmarshal([]byte(ce.body), &resp) == nil {
			if resp.Reason == errno.ErrDAGAlreadyExist.Reason || resp.Message == errno.ErrDAGAlreadyExist.Message {
				return true
			}
		}
	}
	body := strings.ToLower(ce.body)
	return strings.Contains(body, "already exists") || strings.Contains(body, "already exist") || strings.Contains(body, "exists")
}
