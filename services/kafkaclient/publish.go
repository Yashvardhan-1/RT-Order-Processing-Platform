package kafkaclient

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

type PublishConfig struct {
	Topic string
	Key string
	Value proto.Message
}

func Publish(
	ctx context.Context,
	producer *kafka.Producer,
	config PublishConfig,
) error {

	bytes, err := proto.Marshal(config.Value)
	if err != nil {
		return err
	}

	return producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &config.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(config.Key),
		Value: bytes,
	}, nil)
}
