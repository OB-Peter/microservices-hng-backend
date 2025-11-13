#!/bin/bash

# ============================================
# USER SERVICE .env
# ============================================
cat > user-service/.env << 'EOF'
# Database Configuration (Supabase PostgreSQL)
DB_HOST=aws-1-eu-west-1.pooler.supabase.com
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres.xjvvuzcyhqjxswoxuhvs
DB_PASSWORD=SsJiBL62uVUSXRDN
DB_SSL=true

# Service Configuration
PORT=8082
SERVICE_NAME=user-service
EOF

echo "✓ Created user-service/.env"

# ============================================
# EMAIL SERVICE .env
# ============================================
cat > email-service/.env << 'EOF'
# Database Configuration (Supabase PostgreSQL)
DB_HOST=aws-1-eu-west-1.pooler.supabase.com
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres.xjvvuzcyhqjxswoxuhvs
DB_PASSWORD=SsJiBL62uVUSXRDN
DB_SSL=true

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourapp.com

# Service Configuration
PORT=8084
SERVICE_NAME=email-service
EOF

echo "✓ Created email-service/.env"

# ============================================
# TEMPLATE SERVICE .env
# ============================================
cat > template-service/.env << 'EOF'
# Database Configuration (Supabase PostgreSQL)
DB_HOST=aws-1-eu-west-1.pooler.supabase.com
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres.xjvvuzcyhqjxswoxuhvs
DB_PASSWORD=SsJiBL62uVUSXRDN
DB_SSL=true

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672

# Service Configuration
PORT=8085
SERVICE_NAME=template-service
EOF

echo "✓ Created template-service/.env"

# ============================================
# API GATEWAY .env (if you have one)
# ============================================
cat > api-gateway/.env << 'EOF'
# JWT Configuration
JWT_SECRET=9LfWD84EAnrWOv1AVzqvqbnXFcDmKJmIDDkvVpOKVfo=
JWT_EXPIRY=7d

# Redis Configuration
REDIS_URL=redis://localhost:6379

# Service URLs
USER_SERVICE_URL=http://localhost:8082
EMAIL_SERVICE_URL=http://localhost:8084
TEMPLATE_SERVICE_URL=http://localhost:8085

# Application Settings
NODE_ENV=development
PORT=3007
EOF

echo "✓ Created api-gateway/.env"

# ============================================
# PUSH NOTIFICATION SERVICE .env (if you have one)
# ============================================
cat > push-service/.env << 'EOF'
# Database Configuration (Supabase PostgreSQL)
DB_HOST=aws-1-eu-west-1.pooler.supabase.com
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres.xjvvuzcyhqjxswoxuhvs
DB_PASSWORD=SsJiBL62uVUSXRDN
DB_SSL=true

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672

# Firebase Cloud Messaging
FCM_SERVER_KEY=t4c0XRQO-NWDeVukJAsp7fNvtmrGC1ctfwm1Cp_ENwc

# Service Configuration
PORT=8086
SERVICE_NAME=push-service
EOF

echo "✓ Created push-service/.env"

echo ""
echo "================================================"
echo "All .env files created successfully!"
echo "================================================"
echo ""
echo "⚠️  IMPORTANT: Update the following before running:"
echo "  - SMTP credentials in email-service/.env"
echo "  - FCM_SERVER_KEY in push-service/.env (if using)"
echo ""
echo "📝 Service Ports:"
echo "  - User Service: 8082"
echo "  - Email Service: 8084"
echo "  - Template Service: 8085"
echo "  - Push Service: 8086"
echo "  - API Gateway: 3007"
echo ""
