package store

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

type IStore interface {
	DB(ctx context.Context) *gorm.DB
	TX(ctx context.Context, fn func(ctx context.Context) error) error
	DAG() DAGStore
}

type transactionKey struct{}

type datastore struct {
	core *gorm.DB
}

var (
	once sync.Once
	S    *datastore
)

func NewStore(db *gorm.DB) *datastore {
	once.Do(func() {
		S = &datastore{core: db}
	})
	return S
}

func (store *datastore) DB(ctx context.Context) *gorm.DB {
	db := store.core
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

func (store *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return store.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

func (store *datastore) DAG() DAGStore {
	return newDAGStore(store)
}
