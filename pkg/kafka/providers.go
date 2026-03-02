package kafka

import (
	"fmt"

	"aimtp/pkg/options"
)

type TopicConfig struct {
	Topic string
}

func ProvideClient(opts *options.KafkaOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("kafka options is nil")
	}
	if len(opts.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers is empty")
	}
	return NewClient(opts)
}

func ProvideProducer(client *Client) (*Producer, error) {
	if client == nil {
		return nil, fmt.Errorf("kafka client is nil")
	}
	return NewProducer(client), nil
}

func ProvideTopicConfig(opts *options.KafkaOptions) (*TopicConfig, error) {
	if opts == nil {
		return nil, fmt.Errorf("kafka options is nil")
	}
	if opts.Topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}
	return &TopicConfig{Topic: opts.Topic}, nil
}
