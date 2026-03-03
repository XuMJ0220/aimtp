package store

import (
	"context"
	"fmt"
	"time"

	"aimtp/internal/aimtp_server/model"
	"aimtp/internal/pkg/known"

	"gorm.io/gorm"
)

type DAGStore interface {
	ListStalePending(ctx context.Context, cutoff time.Time, limit int) ([]*model.DagStatusSummaryM, error)
	MarkRetryExceeded(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	IncrementRetry(ctx context.Context, dagName string) error
}

type dagStore struct {
	store *datastore
}

var _ DAGStore = (*dagStore)(nil)

func newDAGStore(store *datastore) *dagStore {
	return &dagStore{store: store}
}

func (s *dagStore) ListStalePending(ctx context.Context, cutoff time.Time, limit int) ([]*model.DagStatusSummaryM, error) {
	query := s.store.DB(ctx).Model(&model.DagStatusSummaryM{}).
		Where("creation_status = ?", "pending").
		Where("updated_at <= ?", cutoff).
		Where("COALESCE(retry_count, 0) < COALESCE(max_retries, ?)", known.DAGMaxRetries).
		Order("updated_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	var items []*model.DagStatusSummaryM
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *dagStore) MarkRetryExceeded(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	query := s.store.DB(ctx).Model(&model.DagStatusSummaryM{}).
		Where("creation_status = ?", "pending").
		Where("updated_at <= ?", cutoff).
		Where("COALESCE(retry_count, 0) >= COALESCE(max_retries, ?)", known.DAGMaxRetries).
		Order("updated_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	result := query.Updates(map[string]any{
		"creation_status": "failed",
		"error_msg":       "retry limit exceeded",
	})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (s *dagStore) IncrementRetry(ctx context.Context, dagName string) error {
	if dagName == "" {
		return fmt.Errorf("dag name is empty")
	}
	result := s.store.DB(ctx).Model(&model.DagStatusSummaryM{}).
		Where("dag_name = ? AND creation_status = ?", dagName, "pending").
		Updates(map[string]any{
			"retry_count": gorm.Expr("COALESCE(retry_count, 0) + 1"),
			"error_msg":   nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected for dag %s", dagName)
	}
	return nil
}
