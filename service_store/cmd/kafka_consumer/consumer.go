package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"service_store/internal/usecase"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

// consumer get response
func ProductResponseConsumer(redisClient *redis.Client, breaker *gobreaker.CircuitBreaker) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "products-response",
		GroupID: "store-service",
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
			data, _ := json.Marshal(payload["data"])

			key := fmt.Sprintf("response:%s", corrID)
			_, errBreaker := breaker.Execute(func() (interface{}, error) {
				return redisClient.Set(context.Background(), key, data, 10*time.Second).Result()
			})
			if errBreaker != nil {
				fmt.Println("Redis SET failed or breaker open:", err)
				continue
			}
		}
	}()
}

// consumer send response
func ValidationRequestConsumer(usecase usecase.StoreUsecase, breaker *gobreaker.CircuitBreaker) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "store-validation-request",
		GroupID: "store-service",
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
			userId, _ := payload["user_id"].(uint)
			storeId, _ := payload["store_id"].(uint)

			_, errBreaker := breaker.Execute(func() (interface{}, error) {
				if err := usecase.SendValidationResponse(userId, storeId, corrID); err != nil {
					return nil, err
				}
				return nil, nil
			})
			if errBreaker != nil {
				fmt.Println("send validation failed or breaker open:", err)
				continue
			}

		}
	}()
}
