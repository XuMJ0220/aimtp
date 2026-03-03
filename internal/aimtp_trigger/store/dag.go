package store

import (
	"context"
	"fmt"

	"aimtp/internal/aimtp_server/model"
)

type DAGStore interface {
	TryClaimCreation(ctx context.Context, dagName string) (bool, error)
	UpdateCreationStatus(ctx context.Context, dagName string, status string, errorMsg *string) error
}

type dagStore struct {
	store *datastore
}

var _ DAGStore = (*dagStore)(nil)

func newDAGStore(store *datastore) *dagStore {
	return &dagStore{
		store: store,
	}
}

func (s *dagStore) TryClaimCreation(ctx context.Context, dagName string) (bool, error) {
	if dagName == "" {
		return false, fmt.Errorf("dag name is empty")
	}
	result := s.store.DB(ctx).Model(&model.DagStatusSummaryM{}).
		Where("dag_name = ? AND creation_status = ?", dagName, "pending").
		Updates(map[string]any{
			"creation_status": "creating",
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func (s *dagStore) UpdateCreationStatus(ctx context.Context, dagName string, status string, errorMsg *string) error {
	if dagName == "" {
		return fmt.Errorf("dag name is empty")
	}
	result := s.store.DB(ctx).Model(&model.DagStatusSummaryM{}).
		Where("dag_name = ?", dagName).
		Updates(map[string]any{
			"creation_status": status,
			"error_msg":       errorMsg,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected for dag %s", dagName)
	}
	return nil
}
