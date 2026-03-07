package store

import (
	"context"
	"errors"

	"aimtp/internal/aimtp_controller/model"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/store/where"

	"gorm.io/gorm"
)

type PodStore interface {
	Create(ctx context.Context, obj *model.PodStatusM) error
	Update(ctx context.Context, obj *model.PodStatusM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.PodStatusM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.PodStatusM, error)
}

// podStore 是 PodStore 接口的实现.
type podStore struct {
	store *datastore
}

// 确保 podStore 实现了 PodStore 接口.
var _ PodStore = (*podStore)(nil)

func newPodStore(store *datastore) *podStore {
	return &podStore{
		store: store,
	}
}

func (s *podStore) Create(ctx context.Context, obj *model.PodStatusM) error {
	if err := s.store.DB(ctx).Create(obj).Error; err != nil {
		log.Errorw("Failed to insert pod_status", "err", err, "pod_status", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *podStore) Update(ctx context.Context, obj *model.PodStatusM) error {
	if err := s.store.DB(ctx).Save(obj).Error; err != nil {
		log.Errorw("Failed to update pod_status", "err", err, "pod_status", obj)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *podStore) Delete(ctx context.Context, opts *where.Options) error {
	if err := s.store.DB(ctx, opts).Delete(&model.PodStatusM{}).Error; err != nil {
		log.Errorw("Failed to delete pod_status", "err", err, "conditions", opts)
		return errno.ErrDBWrite.WithMessage("%s", err.Error())
	}
	return nil
}

func (s *podStore) Get(ctx context.Context, opts *where.Options) (*model.PodStatusM, error) {
	var obj model.PodStatusM
	if err := s.store.DB(ctx, opts).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrPodNotFound // 需要定义 ErrPodNotFound
		}
		log.Errorw("Failed to get pod_status", "err", err, "conditions", opts)
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}
	return &obj, nil
}

func (s *podStore) List(ctx context.Context, opts *where.Options) (count int64, ret []*model.PodStatusM, err error) {
	err = s.store.DB(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	if err != nil {
		log.Errorw("Failed to list pod_status", "err", err, "conditions", opts)
		err = errno.ErrDBRead.WithMessage("%s", err.Error())
	}
	return
}
