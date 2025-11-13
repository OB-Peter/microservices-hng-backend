-- Notification Status & Audit Tables
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
    delivered_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_user ON notification_status(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_status ON notification_status(status);
CREATE INDEX IF NOT EXISTS idx_notification_created ON notification_status(created_at);

CREATE TABLE IF NOT EXISTS notification_events (
    id SERIAL PRIMARY KEY,
    notification_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (notification_id) REFERENCES notification_status(notification_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_event_notification ON notification_events(notification_id);
CREATE INDEX IF NOT EXISTS idx_event_created ON notification_events(created_at);

CREATE TABLE IF NOT EXISTS dead_letter_messages (
    id SERIAL PRIMARY KEY,
    notification_id VARCHAR(255) NOT NULL,
    queue_name VARCHAR(100) NOT NULL,
    message_body TEXT NOT NULL,
    retry_count INT,
    error_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dlq_notification ON dead_letter_messages(notification_id);
CREATE INDEX IF NOT EXISTS idx_dlq_created ON dead_letter_messages(created_at);
