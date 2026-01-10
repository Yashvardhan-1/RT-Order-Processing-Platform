package kafkaclient

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ProducerConfig struct {
	Brokers string
	Idempotent bool
	TransactionalID string
}

func NewProducer(config ProducerConfig) (*kafka.Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Brokers,

		// Reliability
		"acks":                     "all",
		"enable.idempotence":       config.Idempotent,
		"retries":                  5,
		"linger.ms":                5,
		"max.in.flight.requests.per.connection": 5,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return p, nil
}
