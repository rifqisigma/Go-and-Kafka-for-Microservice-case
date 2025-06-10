package main

import (
	"log"
	"net/http"
	"os"
	"service_store/cmd/database"
	kafkaconsumer "service_store/cmd/kafka_consumer"
	"service_store/cmd/route"

	"service_store/internal/handler"
	"service_store/internal/repository"
	"service_store/internal/usecase"

	kafka "github.com/segmentio/kafka-go"
)

func main() {
	db, rdb, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("db or rdb : %v", err)
	}
	// Kafka

	kafkaWriter := map[string]*kafka.Writer{
		"products-request": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "products-request",
			Balancer: &kafka.LeastBytes{},
		}),
		"store-validation-response": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "store-validation-response",
			Balancer: &kafka.LeastBytes{},
		}),
	}

	go kafkaconsumer.ProductResponseConsumer(rdb)

	storeRepo := repository.NewStoreRepo(db, rdb)
	storeUC := usecase.NewStoreUsecase(storeRepo, kafkaWriter)
	storeHandler := handler.NewStoreHandler(storeUC)

	r := route.SetupRoute(storeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("service store berjalan pada port : %v", port)
	http.ListenAndServe(":"+port, r)

}
