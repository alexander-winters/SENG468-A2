package kafka

import (
	"context"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// CreateKafkaProducer creates a new Kafka producer
func CreateKafkaProducer(brokerURL string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokerURL),
		Topic:    "notifications",
		Balancer: &kafka.LeastBytes{},
	}
}

func SendNotification(producer *kafka.Writer, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := producer.WriteMessages(ctx, kafka.Message{
		Value: []byte(message),
	})
	if err != nil {
		log.Printf("failed to send notification: %v", err)
	}
}
