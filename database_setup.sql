// db_setup.sql - Run this to create status tracking tables

package main

const statusTableSchema = `
CREATE TABLE IF NOT EXISTS notification_status (
    id SERIAL PRIMARY KEY,
    notification_id VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS notification_events (
    id SERIAL PRIMARY KEY,
    notification_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (notification_id) REFERENCES notification_status(notification_id),
    INDEX idx_notification_id (notification_id),
    INDEX idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS dead_letter_messages (
    id SERIAL PRIMARY KEY,
    notification_id VARCHAR(255) NOT NULL,
    queue_name VARCHAR(100) NOT NULL,
    message_body TEXT NOT NULL,
    retry_count INT,
    error_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notification_id (notification_id),
    INDEX idx_created_at (created_at)
);
`

// Example usage in API Gateway to store initial status
func storeNotificationStatus(notificationID, userID, notificationType string) {
	query := `
	INSERT INTO notification_status (notification_id, user_id, notification_type, status)
	VALUES ($1, $2, $3, 'queued')
	`
	// db.Exec(query, notificationID, userID, notificationType)
}

// Update status when email/push service processes
func updateNotificationStatus(notificationID, status, error string, retryCount int) {
	query := `
	UPDATE notification_status
	SET status = $1, last_error = $2, retry_count = $3, updated_at = CURRENT_TIMESTAMP
	WHERE notification_id = $4
	`
	// db.Exec(query, status, error, retryCount, notificationID)
}

// Mark as delivered
func markNotificationDelivered(notificationID string) {
	query := `
	UPDATE notification_status
	SET status = 'delivered', delivered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
	WHERE notification_id = $1
	`
	// db.Exec(query, notificationID)
}

// Store failed notification to DLQ table
func storeDeadLetterMessage(notificationID, queueName string, messageBody []byte, retryCount int, reason string) {
	query := `
	INSERT INTO dead_letter_messages (notification_id, queue_name, message_body, retry_count, error_reason)
	VALUES ($1, $2, $3, $4, $5)
	`
	// db.Exec(query, notificationID, queueName, messageBody, retryCount, reason)
}

// Record event for audit trail
func recordNotificationEvent(notificationID, eventType, eventData string) {
	query := `
	INSERT INTO notification_events (notification_id, event_type, event_data)
	VALUES ($1, $2, $3)
	`
	// db.Exec(query, notificationID, eventType, eventData)
}
