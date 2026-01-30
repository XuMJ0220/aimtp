package store

import (
	"context"
	"aimtp/pkg/store/where"
	"sync"

	"gorm.io/gorm"
)

var (
	once sync.Once
	// 全局变量，方便其它包直接调用已初始化好的 datastore 实例.
	S *datastore
)

// IStore 定义了 Store 层需要实现的方法.
type IStore interface {
	// 返回 Store 层的 *gorm.DB 实例，在少数场景下会被用到.
	DB(tx context.Context, wheres ...where.Where) *gorm.DB
	// 在事务中执行 fn 函数，fn 内部的所有操作都将在同一个事务中完成，fn 中最好利用 DB() 来实现对数据库的访问.
	TX(tx context.Context, fn func(ctx context.Context) error) error

	// 得到各张表的接口
	User() UserStore
	Post() PostStore
}

// transactionKey 用于在 context.Context 中存储事务上下文的键.
type transactionKey struct{}

// datastore 是 IStore 的具体实现.
type datastore struct {
	core *gorm.DB

	// 可以根据需要添加其他数据库实例
	// fake *gorm.DB
}

// 确保 datastore 实现了 IStore 接口.
var _ IStore = (*datastore)(nil)

func NewStore(db *gorm.DB) IStore {
	// 确保 S 只被初始化一次
	once.Do(func() {
		S = &datastore{
			core: db,
		}
	})

	return S
}

// 可以传入 wheres 参数来构造查询条件.
// 当传入的 tx 是经过 TX() 方法处理过的 context.Context 时，最终得到的是事务内的 *gorm.DB 实例.
func (store *datastore) DB(tx context.Context, wheres ...where.Where) *gorm.DB {
	db := store.core
	// 从上下文中提取事物实例
	if tx, ok := tx.Value(transactionKey{}).(*gorm.DB); ok {
		db = tx
	}

	// 遍历所有条件，依次应用到 db 这个 *gorm.DB 实例上
	for _, whr := range wheres {
		db = whr.Where(db)
	}

	return db
}

// TX 返回一个新的事物实例
// 在使用的使用一般第二个参数的 fn 函数要调用 DB()
// 例如 store.TX(ctx, func(ctx context.Context)error{ store.BD(ctx).Create(...) ... })
func (store *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return store.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// Users 返回一个实现了 UserStore 接口的实例.
func (store *datastore) User() UserStore {
	return newUserStore(store)
}

// Posts 返回一个实现了 PostStore 接口的实例.
func (store *datastore) Post() PostStore {
	return newPostStore(store)
}
