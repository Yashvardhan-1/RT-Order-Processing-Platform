package main

import (
	"context"
	"log"
	"os"
	"time"

	kafkaclient "kafkaclient"

	orderProto "build/proto/order"
	paymentProto "build/proto/payment"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func main() {
	brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

	consumer, err := kafkaclient.NewConsumer(kafkaclient.ConsumerConfig{
		Brokers: brokers,
		GroupID: "payment-service-group",
		AutoOffsetReset: "earliest",
		EnableAutoCommit: true,
	})

	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	producer, err := kafkaclient.NewProducer(kafkaclient.ProducerConfig{
		Brokers: brokers,
		Idempotent: true,
		TransactionalID: "payment-service",
	})
	
	if err != nil {
		log.Fatal(err)
	}

	err = consumer.Subscribe("orders.created", nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Payment Service started")

	for {
		msg, err := kafkaclient.Poll(context.Background(), consumer)
		if err != nil || msg == nil {
			continue
		}

		var order orderProto.OrderCreated
		if err := proto.Unmarshal(msg.Value, &order); err != nil {
			log.Println("Invalid message, skipping")
			consumer.CommitMessage(msg)
			continue
		}

		// --- Idempotency check would go here ---
		// (DB lookup: has order_id been processed?)

		log.Printf("Processing payment for order %s\n", order.OrderId)

		paymentEvent := &paymentProto.PaymentCompleted{
			OrderId:       order.OrderId,
			TransactionId: uuid.NewString(),
			Status:        paymentProto.PaymentStatus_PAYMENT_STATUS_SUCCESS,
			ProcessedAt:   time.Now().Unix(),
		}

		err = kafkaclient.Publish(
			context.Background(),
			producer,
			kafkaclient.PublishConfig{
				Topic: "payments.completed",
				Key: order.OrderId,
				Value: paymentEvent,
			},
		)

		if err != nil {
			log.Println("Failed to publish payment event")
			continue
		}

		// Commit only AFTER producing output
		consumer.CommitMessage(msg)
	}
}

func ptr(s string) *string {
	return &s
}
