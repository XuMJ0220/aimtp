package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer is a wrapper around Kafka client for producing messages.
type Producer struct {
	client *Client
}

// NewProducer creates a new Producer using the given Kafka client.
func NewProducer(client *Client) *Producer {
	return &Producer{
		client: client,
	}
}

// SendMessage sends a message synchronously.
func (p *Producer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	if err := p.client.GetClient().ProduceSync(ctx, record).FirstErr(); err != nil {
		return err
	}
	return nil
}

// SendMessageAsync sends a message asynchronously.
// The callback is optional.
func (p *Producer) SendMessageAsync(ctx context.Context, topic string, key, value []byte, cb func(*kgo.Record, error)) {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	p.client.GetClient().Produce(ctx, record, cb)
}

// Close closes the underlying client.
// Note: If the client is shared with a consumer, closing here will close it for the consumer too.
func (p *Producer) Close() {
	p.client.Close()
}
