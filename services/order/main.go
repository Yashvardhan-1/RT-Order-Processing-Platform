package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	orderProto "build/proto/order"
	kafkaclient "kafkaclient"
)

type CreateOrderRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
}

func main() {
	brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if brokers == "" {
		log.Fatal("KAFKA_BOOTSTRAP_SERVERS not set")
	}

	producer, err := kafkaclient.NewProducer(kafkaclient.ProducerConfig{
		Brokers: brokers,
		Idempotent: true,
		TransactionalID: "order-service",
	})
	
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		orderID := uuid.NewString()

		event := &orderProto.OrderCreated{
			OrderId:   orderID,
			UserId:    req.UserID,
			Amount:    req.Amount,
			CreatedAt: time.Now().Unix(),
		}

		err = kafkaclient.Publish(
			context.Background(),
			producer,
			kafkaclient.PublishConfig{
				Topic: "orders.created",
				Key: orderID,
				Value: event,
			},
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(orderID))
	})

	log.Println("Order Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
