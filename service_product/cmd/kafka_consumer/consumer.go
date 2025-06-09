package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"service_product/internal/usecase"

	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func ProductRequestConsumer(usecase usecase.ProductUsecase) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "products-request",
		GroupID: "product-service",
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("error reading message:", err)
				continue
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				fmt.Println("error unmarshaling:", err)
				continue
			}

			storeID := uint(payload["store_id"].(float64))
			corrID := payload["correlation_id"].(string)

			if err := usecase.SendProductsResponse(storeID, corrID); err != nil {
				fmt.Printf("why error write : %v", err)
			}

		}
	}()
}

func ValidationStoreConsumer(redis *redis.Client) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "validation-store-response",
		GroupID: "product-service",
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Printf("error read message : %v", err)
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				fmt.Println("error unmarshaling:", err)
				continue
			}

			corrID := payload["correlation_id"].(string)
			isValid := payload["is_valid"].(bool)

			key := fmt.Sprintf("response:%s", corrID)
			redis.Set(context.Background(), key, isValid, 10*time.Second)

		}
	}()
}

func ValidationProductConsumer(usecase usecase.ProductUsecase) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9002"},
		Topic:   "product-validation-request",
		GroupID: "product-service",
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("error reading message:", err)
				continue
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				fmt.Println("error unmarshaling:", err)
				continue
			}

			corrID := payload["correlation_id"].(string)
			productId := payload["product_id"].(uint)

			if err := usecase.SendValidationCartResponse(productId, corrID); err != nil {
				fmt.Printf("why error write : %s", err)
			}
		}
	}()
}
