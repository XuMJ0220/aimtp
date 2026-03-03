package biz

import (
	"context"
	"time"

	"aimtp/internal/aimtp_watcher/store"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/kafka"
)

type WatcherBiz interface {
	RunOnce(ctx context.Context) error
}

type watcherBiz struct {
	store      store.IStore
	producer   *kafka.Producer
	topic      *kafka.TopicConfig
	staleAfter time.Duration
	batchSize  int
}

var _ WatcherBiz = (*watcherBiz)(nil)

func New(store store.IStore, producer *kafka.Producer, topic *kafka.TopicConfig, staleAfter time.Duration, batchSize int) *watcherBiz {
	return &watcherBiz{
		store:      store,
		producer:   producer,
		topic:      topic,
		staleAfter: staleAfter,
		batchSize:  batchSize,
	}
}

func (b *watcherBiz) RunOnce(ctx context.Context) error {
	cutoff := time.Now().Add(-b.staleAfter)
	if err := b.markRetryExceeded(ctx, cutoff); err != nil {
		return err
	}
	items, err := b.store.DAG().ListStalePending(ctx, cutoff, b.batchSize)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if item.Payload == nil || *item.Payload == "" {
			log.W(ctx).Warnw("Watcher skipped dag due to empty payload", "dag_name", item.DagName)
			continue
		}
		if err := b.producer.SendMessage(ctx, b.topic.Topic, []byte(item.DagName), []byte(*item.Payload)); err != nil {
			log.W(ctx).Errorw("Watcher failed to requeue dag", "err", err, "dag_name", item.DagName)
			continue
		}
		if err := b.store.DAG().IncrementRetry(ctx, item.DagName); err != nil {
			log.W(ctx).Errorw("Watcher failed to increment retry count", "err", err, "dag_name", item.DagName)
		}
		log.W(ctx).Infow("Watcher requeued dag", "dag_name", item.DagName, "retry_count", item.RetryCount)
	}

	return nil
}

func (b *watcherBiz) markRetryExceeded(ctx context.Context, cutoff time.Time) error {
	rows, err := b.store.DAG().MarkRetryExceeded(ctx, cutoff, b.batchSize)
	if err != nil {
		return err
	}
	if rows > 0 {
		log.W(ctx).Infow("Watcher marked dag creation as failed due to retry limit", "count", rows)
	}
	return nil
}
