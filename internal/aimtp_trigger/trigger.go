package aimtp_trigger

import (
	"context"

	"aimtp/internal/aimtp_trigger/biz"
	"aimtp/pkg/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Trigger struct {
	consumer *kafka.Consumer
	biz      biz.TriggerBiz
}

func New(consumer *kafka.Consumer, biz biz.TriggerBiz) *Trigger {
	return &Trigger{
		consumer: consumer,
		biz:      biz,
	}
}

func (t *Trigger) Run(ctx context.Context) error {
	return t.consumer.StartConsuming(ctx, func(ctx context.Context, record *kgo.Record) error {
		//log.Infow("Trigger received record", "topic", record.Topic, "partition", record.Partition, "offset", record.Offset, "key_size", len(record.Key), "value_size", len(record.Value))
		return t.biz.HandleCreateDAG(ctx, record.Value)
	})
}
