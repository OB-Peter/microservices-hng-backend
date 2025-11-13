package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type NotificationMessage struct {
	UserID  string `json:"user_id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

func startConsumer() {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://admin:admin@rabbitmq:5672/"
	}

	// Connect with retry
	var conn *amqp.Connection
	var err error
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Declare queue
	q, err := ch.QueueDeclare(
		"push.queue",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Set QoS
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("Push Consumer started. Waiting for messages...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var notification NotificationMessage
			err := json.Unmarshal(d.Body, &notification)
			if err != nil {
				log.Printf("Error parsing message: %v", err)
				d.Nack(false, false) // Send to DLQ
				continue
			}

			log.Printf("Processing push notification for user: %s", notification.UserID)

			// Simulate sending push notification
			err = sendPushNotification(notification)
			if err != nil {
				log.Printf("Failed to send push: %v", err)
				
				// Retry logic
				retryCount := getRetryCount(d.Headers)
				if retryCount < 3 {
					d.Nack(false, true) // Requeue
					log.Printf("Requeuing message (retry %d/3)", retryCount+1)
				} else {
					d.Nack(false, false) // Send to DLQ
					log.Printf("Max retries reached, sending to DLQ")
				}
				continue
			}

			log.Printf("✓ Push sent successfully to user: %s", notification.UserID)
			d.Ack(false)
		}
	}()

	<-forever
}

func sendPushNotification(notification NotificationMessage) error {
	// Simulate FCM/OneSignal API call
	log.Printf("Sending push: %s - %s", notification.Title, notification.Message)
	
	// Add your actual FCM/OneSignal logic here
	// For now, just simulate success
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if count, ok := headers["x-retry-count"].(int32); ok {
		return int(count)
	}
	return 0
}
