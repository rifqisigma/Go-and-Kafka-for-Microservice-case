package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"service_product/cmd/database"
	kafkaconsumer "service_product/cmd/kafka_consumer"
	"service_product/cmd/route"
	"service_product/internal/handler"
	"service_product/internal/repository"
	"service_product/internal/usecase"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

func main() {

	db, rdb, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("db error : %v", err)
	}

	productRepo := repository.NewStoreRepo(db, rdb)

	writer := map[string]*kafka.Writer{
		"products-response": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "products-response",
			Balancer: &kafka.LeastBytes{},
		}),
		"validation-store-request": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "validation-store-request",
			Balancer: &kafka.LeastBytes{},
		}),
		"validation-product-response": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "validation-product-response",
			Balancer: &kafka.LeastBytes{},
		}),
	}

	productUC := usecase.NewProductUsecase(productRepo, writer)
	productHandler := handler.NewStoreHandler(productUC)

	r := route.SetupRoute(productHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "consumerBreaker",
		Interval:    30 * time.Second,
		MaxRequests: 5,
		Timeout:     5 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("[Circuit Breaker: %s] status berubah dari %s ➜ %s\n", name, from.String(), to.String())
		},
	})
	go kafkaconsumer.ProductRequestConsumer(productUC, cb)
	go kafkaconsumer.ValidationStoreConsumer(rdb, cb)
	go kafkaconsumer.ProductRequestConsumer(productUC, cb)

	fmt.Printf("service product berjalan pada port:%s", port)
	http.ListenAndServe(":"+port, r)

}
