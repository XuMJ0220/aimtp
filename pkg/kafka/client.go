package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"aimtp/pkg/options"
)

// Client is a wrapper around kgo.Client.
type Client struct {
	client *kgo.Client
	opts   *options.KafkaOptions
}

// NewClient creates a new Kafka client.
func NewClient(opts *options.KafkaOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("kafka options is nil")
	}

	c, err := opts.NewClient()
	if err != nil {
		return nil, err
	}

	return &Client{
		client: c,
		opts:   opts,
	}, nil
}

// Close closes the underlying kafka client.
func (c *Client) Close() {
	c.client.Close()
}

// GetClient returns the underlying kgo.Client.
func (c *Client) GetClient() *kgo.Client {
	return c.client
}

// Ping checks the connection to the kafka cluster.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}
