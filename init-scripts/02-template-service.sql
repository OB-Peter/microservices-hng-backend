-- Template Service Database
CREATE TABLE IF NOT EXISTS templates (
    id SERIAL PRIMARY KEY,
    template_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_id, version)
);

CREATE INDEX IF NOT EXISTS idx_templates_id ON templates(template_id);
CREATE INDEX IF NOT EXISTS idx_templates_created ON templates(created_at);
