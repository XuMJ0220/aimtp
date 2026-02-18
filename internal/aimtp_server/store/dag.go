package store

import (
	"aimtp/internal/aimtp_server/model"
	"aimtp/internal/pkg/errno"
	"aimtp/pkg/log"
	"aimtp/pkg/store/where"
	"context"
	"errors"

	"gorm.io/gorm"
)

type DAGStore interface {
	Create(ctx context.Context, obj *model.DagStatusSummaryM) error
	Update(ctx context.Context, obj *model.DagStatusSummaryM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.DagStatusSummaryM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.DagStatusSummaryM, error)
	DAGExpansion
}

type DAGExpansion interface{}

// dagStore 是 DAGStore 接口的实现.
type dagStore struct {
	// 为了调用 DX 和 TX 等方法，需要持有对 datastore 的引用
	store *datastore
}

// 确保 dagStore 实现了 DAGStore 接口.
var _ DAGStore = (*dagStore)(nil)

// newDAGStore 创建 dagStore 的实例.
func newDAGStore(store *datastore) *dagStore {
	return &dagStore{
		store: store,
	}
}

func (s *dagStore) Create(ctx context.Context, obj *model.DagStatusSummaryM) error {
	if err := s.store.DB(ctx).Create(&obj).Error; err != nil {
		log.Errorw(err, "Failed to insert dag_status_summary", "err", err, "dag_status_summary", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *dagStore) Update(ctx context.Context, obj *model.DagStatusSummaryM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		log.Errorw(err, "Failed to update dag_status_summary", "err", err, "dag_status_summary", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *dagStore) Delete(ctx context.Context, opts *where.Options) error {
	if err := s.store.DB(ctx, opts).Delete(new(model.DagStatusSummaryM)).Error; err != nil {
		log.Errorw(err, "Failed to delete dag_status_summary", "err", err, "conditions", opts)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *dagStore) Get(ctx context.Context, opts *where.Options) (*model.DagStatusSummaryM, error) {
	var obj model.DagStatusSummaryM
	if err := s.store.DB(ctx, opts).First(&obj).Error; err != nil {
		log.Errorw(err, "Failed to get dag_status_summary", "err", err, "conditions", opts)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrDAGNotFound
		}
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return &obj, nil
}

func (s *dagStore) List(ctx context.Context, opts *where.Options) (count int64, ret []*model.DagStatusSummaryM, err error) {
	err = s.store.DB(ctx, opts).Order("dag_id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	if err != nil {
		log.Errorw(err, "Failed to list dag_status_summary", "err", err, "conditions", opts)
		err = errno.ErrDBRead.WithMessage("%s", err.Error())
	}
	return
}
