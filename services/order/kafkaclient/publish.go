package kafkaclient

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

func Publish(
	ctx context.Context,
	producer *kafka.Producer,
	topic string,
	key string,
	value proto.Message,
) error {

	bytes, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: bytes,
	}, nil)
}
