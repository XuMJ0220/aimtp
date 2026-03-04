package store

import (
	"aimtp/internal/aimtp_controller/model"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/store/where"
	"context"
	"errors"

	"gorm.io/gorm"
)

type JobStore interface {
	Create(ctx context.Context, obj *model.JobStatusM) error
	Update(ctx context.Context, obj *model.JobStatusM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.JobStatusM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.JobStatusM, error)
	JobExpansion
}

type JobExpansion interface{}

// jobStore 是 JobStore 接口的实现.
type jobStore struct {
	store *datastore
}

// 确保 jobStore 实现了 JobStore 接口.
var _ JobStore = (*jobStore)(nil)

func newJobStore(store *datastore) *jobStore {
	return &jobStore{
		store: store,
	}
}

func (s *jobStore) Create(ctx context.Context, obj *model.JobStatusM) error {
	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		log.Errorw( "Failed to insert job_status", "err", err, "job_status", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *jobStore) Update(ctx context.Context, obj *model.JobStatusM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		log.Errorw( "Failed to update job_status", "err", err, "job_status", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *jobStore) Delete(ctx context.Context, opts *where.Options) error {
	if err := s.store.DB(ctx, opts).Delete(&model.JobStatusM{}).Error; err != nil {
		log.Errorw( "Failed to delete job_status", "err", err, "conditions", opts)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *jobStore) Get(ctx context.Context, opts *where.Options) (*model.JobStatusM, error) {
	var obj model.JobStatusM
	if err := s.store.DB(ctx, opts).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrJobNotFound // Use generic not found or define specific ErrJobNotFound
		}
		log.Errorw( "Failed to get job_status", "err", err, "conditions", opts)
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}
	return &obj, nil
}

func (s *jobStore) List(ctx context.Context, opts *where.Options) (count int64, ret []*model.JobStatusM, err error) {
	err = s.store.DB(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	if err != nil {
		log.Errorw( "Failed to list job_status", "err", err, "conditions", opts)
		err = errno.ErrDBRead.WithMessage("%s", err.Error())
	}
	return
}
