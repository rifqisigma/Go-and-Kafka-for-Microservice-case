package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"service_user/cmd/database"
	"service_user/cmd/route"

	"service_user/internal/handler"
	"service_user/internal/repository"
	"service_user/internal/usecase"

	"github.com/segmentio/kafka-go"
)

func main() {

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("db : %v", err)
	}

	authRepo := repository.NewAuthRepo(db)

	writer := map[string]*kafka.Writer{
		"notification-request": kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{os.Getenv("KAFKA_BROKER")},
			Topic:    "notification-request",
			Balancer: &kafka.LeastBytes{},
		}),
	}
	authUC := usecase.NewAuthUsecase(authRepo, writer)
	authDelivery := handler.NewAuthHandler(authUC)

	r := route.SetupRoute(authDelivery)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("user service berjalan di port : %v", port)
	http.ListenAndServe(":"+port, r)
}
