package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection
var rabbitChannel *amqp.Channel

// NotificationRequest - incoming request
type NotificationRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Type       string `json:"type" binding:"required,oneof=email push"`
	Title      string `json:"title" binding:"required"`
	Message    string `json:"message" binding:"required"`
	TemplateID string `json:"template_id"`
}

// NotificationPayload - what goes into the queue
type NotificationPayload struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	TemplateID     string `json:"template_id"`
}

// Response - standard response format
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message"`
}

func init() {
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

	// Declare exchange and queues
	err = rabbitChannel.ExchangeDeclare(
		"notifications.direct",
		"direct",
		true, // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Email queue
	_, err = rabbitChannel.QueueDeclare("email.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare email queue: %v", err)
	}
	rabbitChannel.QueueBind("email.queue", "email", "notifications.direct", false, nil)

	// Push queue
	_, err = rabbitChannel.QueueDeclare("push.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare push queue: %v", err)
	}
	rabbitChannel.QueueBind("push.queue", "push", "notifications.direct", false, nil)

	log.Println("RabbitMQ connected and queues declared")
}

func main() {
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	router := gin.Default()

	// Health check
	router.GET("/health", healthHandler)

	// Notification endpoints
	router.POST("/notifications", createNotification)
	router.GET("/notifications/:id", getNotificationStatus)

	log.Printf("API Gateway starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
	})
}

func createNotification(c *gin.Context) {
	var req NotificationRequest

	// Validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
			Message: "Invalid request",
		})
		return
	}

	// Generate notification ID
	notificationID := uuid.New().String()

	// Create payload
	payload := NotificationPayload{
		NotificationID: notificationID,
		UserID:         req.UserID,
		Type:           req.Type,
		Title:          req.Title,
		Message:        req.Message,
		TemplateID:     req.TemplateID,
	}

	// Marshal to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
			Message: "Failed to process request",
		})
		return
	}

	// Publish to appropriate queue
	routingKey := req.Type // "email" or "push"
	err = rabbitChannel.Publish(
		"notifications.direct",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
			Message: "Failed to queue notification",
		})
		return
	}

	// Success response
	c.JSON(http.StatusAccepted, Response{
		Success: true,
		Data: gin.H{
			"notification_id": notificationID,
			"status":          "queued",
			"type":            req.Type,
		},
		Message: "Notification queued successfully",
	})

	log.Printf("Notification %s queued: %s to %s\n", notificationID, req.Type, req.UserID)
}

func getNotificationStatus(c *gin.Context) {
	notificationID := c.Param("id")

	// TODO: Fetch from database/cache in full implementation
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"notification_id": notificationID,
			"status":          "queued",
		},
		Message: "Status retrieved",
	})
}
