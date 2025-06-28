package main

import (
	"fmt"
	"log"
	"os"
	kafkaconsumer "service_notification/cmd/kafka_consumer"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
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
	brokers := []string{os.Getenv("KAFKA_BROKER")}

	if err := checkKafkaReady(brokers); err != nil {
		log.Fatalf("Kafka broker not ready: %v", err)
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
	go kafkaconsumer.ProductResponseConsumer(cb)
	fmt.Println("Service notification berjalan")

	select {}
}
