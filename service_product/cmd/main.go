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

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("env : %v", err)
	}

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

	go kafkaconsumer.ProductRequestConsumer(productUC)
	go kafkaconsumer.ValidationStoreConsumer(rdb)
	go kafkaconsumer.ProductRequestConsumer(productUC)

	fmt.Printf("service product berjalan pada port:%s", port)
	http.ListenAndServe(":"+port, r)

}
