# Distributed Notification System

A microservices-based notification system built with Go, RabbitMQ, and PostgreSQL.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Git

### Installation
```bash
# Clone the repository
git clone <your-repo-url>
cd microservices

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

## Services

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8001 | Entry point for requests |
| User Service | 8002 | User data management |
| Template Service | 8003 | Template management |
| Email Service | 8004 | Email notifications |
| Push Service | 8005 | Push notifications |
| RabbitMQ | 5673 | Message queue |
| RabbitMQ UI | 15673 | Management interface |

## Testing

### Send Email Notification
```bash
curl -X POST http://localhost:8001/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "type": "email",
    "title": "Welcome",
    "message": "Hello from our service!",
    "template_id": "welcome_001"
  }'
```

### Send Push Notification
```bash
curl -X POST http://localhost:8001/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "type": "push",
    "title": "New Message",
    "message": "You have a new notification!",
    "template_id": "message_001"
  }'
```

### Check Service Health
```bash
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health
curl http://localhost:8004/health
curl http://localhost:8005/health
```

## Architecture

See `SYSTEM_DESIGN.md` for detailed architecture documentation.

## CI/CD

GitHub Actions pipeline automatically:
- Builds all services
- Runs tests
- Builds Docker images
- Tests with docker-compose
- Deploys on main branch push

## Stopping Services
```bash
docker-compose down
```

## Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api-gateway
docker-compose logs -f email-service
docker-compose logs -f push-service
```

## Project Structure
```
.
├── api-gateway/          # API Gateway service
├── user-service/         # User service
├── template-service/     # Template service
├── email-service/        # Email service
├── push-service/         # Push service
├── docker-compose.yml    # Docker Compose configuration
├── .github/
│   └── workflows/
│       └── deploy.yml    # CI/CD pipeline
├── README.md             # This file
└── SYSTEM_DESIGN.md      # Architecture documentation
```

## Implementation Status

✅ **Completed:**
- All 5 microservices
- Docker containerization
- RabbitMQ message queues
- Basic API endpoints
- Health checks
- Docker Compose orchestration
- CI/CD pipeline
- System architecture

📋 **To Be Added:**
- Database integration (User & Template services)
- SMTP integration (Email service)
- FCM integration (Push service)
- Template variable substitution
- Retry logic with dead letter queue
- Circuit breaker pattern
- Distributed tracing

## Team

- [Team Member 1]
- [Team Member 2]
- [Team Member 3]
- [Team Member 4]
