package main

import (
	"log"
	"github.com/streadway/amqp"
)

// SetupQueuesWithDLQ configures queues with dead letter exchange
func SetupQueuesWithDLQ(ch *amqp.Channel) error {
	// Declare Dead Letter Exchange
	err := ch.ExchangeDeclare(
		"dlx.exchange",  // name
		"fanout",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	// Declare Dead Letter Queue
	_, err = ch.QueueDeclare(
		"dead.letter.queue", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}

	// Bind DLQ to DLX
	err = ch.QueueBind(
		"dead.letter.queue", // queue name
		"",                  // routing key
		"dlx.exchange",      // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Setup Email Queue with DLQ
	err = setupQueueWithDLQ(ch, "email.queue")
	if err != nil {
		return err
	}

	// Setup Push Queue with DLQ
	err = setupQueueWithDLQ(ch, "push.queue")
	if err != nil {
		return err
	}

	log.Println("✅ Dead Letter Queues configured successfully")
	return nil
}

func setupQueueWithDLQ(ch *amqp.Channel, queueName string) error {
	args := amqp.Table{
		"x-dead-letter-exchange": "dlx.exchange",
		// Optional: Message TTL (30 seconds before moving to DLQ if not processed)
		// "x-message-ttl": 30000,
		// Optional: Max retries
		"x-max-length": 10000,
	}

	_, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments with DLQ config
	)
	if err != nil {
		return err
	}

	// Bind to notifications exchange
	err = ch.QueueBind(
		queueName,
		queueName,
		"notifications",
		false,
		nil,
	)
	
	log.Printf("✅ Queue '%s' configured with DLQ support", queueName)
	return err
}

// Example usage in your main.go
func ExampleUsage() {
	// In your main() function, after connecting to RabbitMQ:
	// conn, _ := amqp.Dial(rabbitmqURL)
	// ch, _ := conn.Channel()
	
	// Setup queues with DLQ
	// err := SetupQueuesWithDLQ(ch)
	// if err != nil {
	//     log.Fatal("Failed to setup DLQ:", err)
	// }
}

// DLQ Consumer - Monitor and log failed messages
func ConsumeDLQ(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		"dead.letter.queue",
		"dlq-consumer",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to consume DLQ:", err)
	}

	log.Println("📮 Monitoring Dead Letter Queue...")

	for msg := range msgs {
		// Log failed message details
		log.Printf("❌ DLQ Message Received:")
		log.Printf("   Body: %s", string(msg.Body))
		log.Printf("   Routing Key: %s", msg.RoutingKey)
		log.Printf("   Headers: %v", msg.Headers)
		
		// Optional: Store in database for manual review
		// storeFailed Message(msg)
		
		// Acknowledge the DLQ message
		msg.Ack(false)
	}
}
