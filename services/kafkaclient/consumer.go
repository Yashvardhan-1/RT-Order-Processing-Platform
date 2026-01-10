package kafkaclient

import (
	"context"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ConsumerConfig struct {
	Brokers string
	GroupID string
	AutoOffsetReset string
	EnableAutoCommit bool
}

func NewConsumer(config ConsumerConfig) (*kafka.Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  config.Brokers,
		"group.id":           config.GroupID,
		"auto.offset.reset":  config.AutoOffsetReset,
		"enable.auto.commit": config.EnableAutoCommit,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Poll(ctx context.Context, c *kafka.Consumer) (*kafka.Message, error) {
	ev := c.Poll(1000)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		return e, nil
	case kafka.Error:
		log.Printf("Kafka error: %v\n", e)
	}

	return nil, nil
}
