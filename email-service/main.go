package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection
var rabbitChannel *amqp.Channel

type NotificationPayload struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	TemplateID     string `json:"template_id"`
}

func init() {
	_ = godotenv.Load()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5673/"
	}

	var err error
	rabbitConn, err = amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}

	// Declare email queue
	_, err = rabbitChannel.QueueDeclare("email.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare email queue: %v", err)
	}

	log.Println("Email Service connected to RabbitMQ")
}

func main() {
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "email-service",
		})
	})

	// Start consuming from queue in a goroutine
	go consumeEmailMessages()

	log.Printf("Email Service starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func consumeEmailMessages() {
	msgs, err := rabbitChannel.Consume(
		"email.queue",
		"",   // consumer tag
		true, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("Email Service: Waiting for messages...")

	for msg := range msgs {
		var payload NotificationPayload
		err := json.Unmarshal(msg.Body, &payload)
		if err != nil {
			log.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		// Send email (mock for MVP)
		sendEmail(payload)
	}
}

func sendEmail(payload NotificationPayload) {
	log.Printf("📧 SENDING EMAIL")
	log.Printf("   Notification ID: %s\n", payload.NotificationID)
	log.Printf("   User ID: %s\n", payload.UserID)
	log.Printf("   Title: %s\n", payload.Title)
	log.Printf("   Message: %s\n", payload.Message)
	log.Printf("   Template: %s\n", payload.TemplateID)

	// TODO: In production, integrate with:
	// - SendGrid API
	// - Gmail SMTP
	// - Mailgun
	// For now, we just log it

	log.Printf("   ✅ Email sent successfully!\n")
}
