# Distributed Notification System - Architecture

## System Overview

This is a microservices-based notification system that sends emails and push notifications asynchronously using RabbitMQ message queues.

## Architecture Diagram

[See the diagram in the artifact above - can be viewed in GitHub markdown]

## Service Descriptions

### 1. API Gateway (Port 8001)
- **Entry point** for all notification requests
- **Validates** incoming requests (user_id, type, message, etc.)
- **Routes** messages to appropriate RabbitMQ queues
- **Tracks** notification status
- **Response Format**: Standard JSON with snake_case

### 2. User Service (Port 8002)
- **Manages** user contact information (email, push tokens)
- **Stores** notification preferences
- **Provides** REST API for user data lookup
- **Database**: PostgreSQL on port 5434

### 3. Template Service (Port 8003)
- **Stores** notification templates
- **Handles** variable substitution (e.g., {{user_name}})
- **Supports** multiple languages
- **Keeps** version history
- **Database**: PostgreSQL on port 5433

### 4. Email Service (Port 8004)
- **Consumes** messages from `email.queue`
- **Retrieves** user email via User Service
- **Fetches** template via Template Service
- **Sends** emails via SMTP (Gmail, SendGrid, etc.)
- **Logs** delivery status

### 5. Push Service (Port 8005)
- **Consumes** messages from `push.queue`
- **Retrieves** device tokens via User Service
- **Fetches** template via Template Service
- **Sends** push notifications via FCM or OneSignal
- **Logs** delivery status

## Infrastructure

### Message Queue (RabbitMQ - Port 5673)
```
Exchange: notifications.direct
├── email.queue  → Email Service
├── push.queue   → Push Service
└── failed.queue → Dead Letter Queue (future)
```

### Databases
- **PostgreSQL User DB** (Port 5434): User data, preferences
- **PostgreSQL Template DB** (Port 5433): Templates, versions
- **Redis Cache** (Port 6380): Rate limiting, preference caching

## Data Flow

1. **Request**: Client sends notification request to API Gateway
2. **Validation**: API Gateway validates and authenticates
3. **Routing**: Routes to `email.queue` or `push.queue` based on type
4. **Consumption**: Email/Push services consume from respective queues
5. **Enrichment**: Services fetch user data and templates from their APIs
6. **Sending**: Services send via SMTP or FCM
7. **Status**: Services log status back to system

## Deployment

### Docker Compose
All services run in Docker containers:
```bash
docker-compose up -d
```

Services:
- api-gateway (8001)
- user-service (8002)
- template-service (8003)
- email-service (8004)
- push-service (8005)
- RabbitMQ (5673, 15673 management)
- PostgreSQL User DB (5434)
- PostgreSQL Template DB (5433)
- Redis (6380)

### CI/CD Pipeline
GitHub Actions workflow that:
1. Builds all services
2. Runs tests
3. Builds Docker images
4. Tests with docker-compose
5. Deploys on main branch

## Scaling Strategy

- **Horizontal Scaling**: All services can run multiple instances
- **Load Balancing**: Use nginx/HAProxy in front of API Gateway
- **Queue Scaling**: RabbitMQ can handle multiple consumers per queue
- **Database Scaling**: PostgreSQL replication for high availability
- **Caching**: Redis for frequently accessed data

## Performance Targets

- **Throughput**: 1,000+ notifications/minute
- **API Response Time**: < 100ms
- **Delivery Success Rate**: 99.5%
- **All services**: Horizontally scalable

## Future Enhancements

- [ ] Dead letter queue for failed messages
- [ ] Retry logic with exponential backoff
- [ ] Circuit breaker pattern
- [ ] Distributed tracing (correlation IDs)
- [ ] Metrics & monitoring (Prometheus, Grafana)
- [ ] Service mesh (Istio)
- [ ] API rate limiting per user
- [ ] Notification preferences enforcement
