package kafkaService

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
	kafka "github.com/segmentio/kafka-go"
)

type KafkaService struct {
	Producer *kafka.Writer
	Consumer *kafka.Reader
}

func NewKafkaService(producer *kafka.Writer, consumer *kafka.Reader) *KafkaService {
	return &KafkaService{
		Producer: producer,
		Consumer: consumer,
	}
}

func CreateKafkaProducer(brokerURL string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokerURL),
		Topic:    "notifications",
		Balancer: &kafka.LeastBytes{},
	}
}

func (ks *KafkaService) SendUserNotification(notification models.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define the topic and serialize the notification
	topic := fmt.Sprintf("notifications-%s", notification.Recipient)
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		log.Printf("failed to serialize notification: %v", err)
		return err
	}

	err = ks.Producer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: notificationBytes,
	})
	if err != nil {
		log.Printf("failed to send notification: %v", err)
	}
	return err
}

func CreateKafkaConsumer(brokerURL, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerURL},
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})
}

func (ks *KafkaService) ConsumeUserNotifications(username string) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := ks.Consumer.ReadMessage(ctx)
		cancel()
		if err != nil {
			log.Printf("failed to read notification: %v", err)
		} else {
			var notification models.Notification
			err = json.Unmarshal(msg.Value, &notification)
			if err != nil {
				log.Printf("failed to deserialize notification: %v", err)
				continue
			}

			fmt.Printf("Received notification for user %s: %s\n", username, notification.Content)

			notificationsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("notifications")
			_, err = notificationsCollection.InsertOne(ctx, notification)
			if err != nil {
				log.Printf("failed to store notification in database: %v", err)
			}
		}
	}
}
