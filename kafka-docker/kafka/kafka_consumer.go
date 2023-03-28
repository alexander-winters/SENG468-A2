package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

func createKafkaConsumer(brokerURL string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerURL},
		Topic:     "notifications",
		Partition: 0,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})
}

func processNotifications(consumer *kafka.Reader) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := consumer.ReadMessage(ctx)
		cancel()
		if err != nil {
			log.Printf("failed to read notification: %v", err)
		} else {
			fmt.Printf("Received notification: %s\n", string(msg.Value))
			// Process the notification (e.g. store it in the database or send an email)
		}
	}
}
