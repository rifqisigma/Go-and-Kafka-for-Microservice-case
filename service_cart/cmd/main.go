package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"service_cart/cmd/database"
	kafkaconsumer "service_cart/cmd/kafka_consumer"
	"service_cart/cmd/route"
	"service_cart/internal/handler"
	"service_cart/internal/repository"
	"service_cart/internal/usecase"

	"github.com/segmentio/kafka-go"
)

func main() {
	db, rdb, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("db err : %s", err)
	}

	cartRepo := repository.NewCartRepo(db, rdb)

	writes := map[string]*kafka.Writer{
		"product-validation-response": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "product-validation-response",
			Balancer: &kafka.LeastBytes{},
		}),
	}
	cartUC := usecase.NewCartUsecase(cartRepo, writes)
	cartHandler := handler.NewCartpHandler(cartUC)

	r := route.SetupRoute(cartHandler)

	go kafkaconsumer.ValidationResponseConsumer(rdb, cartUC)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}

	fmt.Printf("service cart berjalan pada port:%s", port)
	http.ListenAndServe(":"+port, r)

}
