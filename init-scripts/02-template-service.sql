-- Template Service Database
CREATE TABLE IF NOT EXISTS templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    subject TEXT,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default templates
INSERT INTO templates (name, type, subject, body) VALUES 
('welcome_email', 'email', 'Welcome to Our Platform!', 'Hello {{name}}, welcome aboard!'),
('password_reset', 'email', 'Reset Your Password', 'Click here to reset: {{link}}'),
('order_confirmation', 'push', 'Order Confirmed', 'Your order #{{order_id}} has been confirmed!')
ON CONFLICT (name) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_templates_name ON templates(name);
CREATE INDEX IF NOT EXISTS idx_templates_type ON templates(type);
CREATE INDEX IF NOT EXISTS idx_templates_created ON templates(created_at);
