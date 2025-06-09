package main

import (
	"fmt"
	"log"
	"os"
	kafkaconsumer "service_notification/cmd/kafka_consumer"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func checkKafkaReady(brokers []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	brokers := []string{os.Getenv("KAFKA_BROKER")}

	// Cek koneksi Kafka broker dulu
	if err := checkKafkaReady(brokers); err != nil {
		log.Fatalf("Kafka broker not ready: %v", err)
	}

	go kafkaconsumer.ProductResponseConsumer()
	fmt.Println("Service notification berjalan")

	select {}
}
