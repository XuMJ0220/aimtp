package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer is a wrapper around Kafka client for consuming messages.
type Consumer struct {
	client *Client
}

// NewConsumer creates a new Consumer using the given Kafka client.
func NewConsumer(client *Client) *Consumer {
	return &Consumer{
		client: client,
	}
}

// HandlerFunc is the type for the message handler function.
type HandlerFunc func(ctx context.Context, record *kgo.Record) error

// StartConsuming starts consuming messages from the subscribed topics.
// It will block until the context is cancelled or an error occurs.
// This is a simple wrapper for synchronous consuming.
func (c *Consumer) StartConsuming(ctx context.Context, handler HandlerFunc) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.client.GetClient().PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				// All errors are retriable, so we just log them and continue.
				// In a real application, we might want to handle specific errors differently.
				for _, err := range errs {
					return fmt.Errorf("topic %s partition %d had error: %v", err.Topic, err.Partition, err.Err)
				}
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				if err := handler(ctx, record); err != nil {
					// Handle handler error (e.g. log, nack, etc.)
					// For now, we just return the error to stop consuming if handler fails.
					return err
				}
			}
		}
	}
}

// Close closes the underlying client.
func (c *Consumer) Close() {
	c.client.Close()
}
