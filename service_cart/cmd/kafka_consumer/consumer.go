package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"service_cart/internal/usecase"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func ValidationResponseConsumer(redis *redis.Client, usecase usecase.CartUsecase) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "product-validation-response",
		GroupID: "cart-service",
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
			deleted := payload["deleted"].(bool)
			stock := payload["stock"].(int)
			productId := payload["product_id"].(uint)

			if deleted {
				usecase.UpdateIsDeleteProduct(productId)
			}

			data := map[string]interface{}{
				"deleted": deleted,
				"stock":   stock,
			}
			jsonData, _ := json.Marshal(data)

			key := fmt.Sprintf("response:%s", corrID)
			redis.Set(context.Background(), key, jsonData, 10*time.Second)

		}
	}()
}
