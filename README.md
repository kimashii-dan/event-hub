# Event Hub

> A modern, RESTful API for event management built with Go, enabling organizers to create, manage, and publish events while users can discover and register for events.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [API Documentation](#-api-documentation)
- [Development](#-development)
- [Testing](#-testing)
- [Contributing](#-contributing)
- [Documentation](#-documentation)

## ✨ Features

- 🔐 **Secure Authentication**: JWT-based authentication with role-based access control
- 📅 **Event Management**: Create, update, publish, and cancel events with comprehensive metadata
- 👥 **User Registration**: Users can register for events with capacity management
- 🔍 **Advanced Filtering**: Search and filter events by date, capacity, and status
- 📄 **Pagination**: Efficient data retrieval with configurable page sizes
- 🐳 **Docker Support**: Fully containerized with Docker Compose for easy deployment
- 🧪 **Comprehensive Testing**: Unit and integration tests for reliability
- 📊 **Health Checks**: Built-in endpoint for monitoring application status

## 🏗 Architecture

Event Hub follows **Clean Architecture** principles with clear separation of concerns:

```
backend/
├── cmd/app/           # Application entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── database/      # Database connection & migrations
│   ├── domain/        # Domain models & business rules
│   ├── handler/       # HTTP handlers (controllers)
│   ├── middleware/    # HTTP middleware (auth, logging)
│   ├── repository/    # Data access layer
│   └── service/       # Business logic layer
├── pkg/               # Reusable packages
│   ├── jwt/           # JWT token utilities
│   └── response/      # Standardized API responses
├── migrations/        # Database schema migrations
└── tests/             # Integration & unit tests
```

### Tech Stack

- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens
- **Containerization**: Docker & Docker Compose
- **Testing**: Go testing framework with testify

## 🚀 Quick Start

### Prerequisites

- **Docker** & **Docker Compose** (v20.10+)
- **Git**
- **Go** 1.21+ (for local development)

### Setup & Run

```bash
# 1. Clone the repository
git clone https://github.com/kimashii-dan/event-hub.git
cd event-hub/backend

# 2. Configure environment
cp docker/.env.example docker/.env
# Edit docker/.env with your configuration

# 3. Start the application
cd docker
docker compose up --build
```

🎉 **Application is running at**: `http://localhost:8000`

### Environment Configuration

Create `docker/.env` file with the following variables:

```bash
# Database Configuration
DB_HOST=db
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_PORT=5432
DB_NAME=event_hub

# Server Configuration
SERVER_PORT=8000

# JWT Configuration
JWT_SECRET=your_jwt_secret_key_min_32_chars
JWT_EXPIRATION_TIME=24h
```

> ⚠️ **Security Note**: Always use strong, unique values for `DB_PASSWORD` and `JWT_SECRET` in production!

## 📚 API Documentation

### Base URL
```
http://localhost:8000
```

### Authentication

Most endpoints require JWT authentication. Include the token in the `Authorization` header:
```
Authorization: Bearer <your_jwt_token>
```

### Core Endpoints

#### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register new user | No |
| POST | `/auth/login` | Login and get JWT token | No |

#### Events

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/events` | Get all events (with filtering & pagination) | No |
| GET | `/events/:id` | Get event by ID | No |
| POST | `/events` | Create new event | Yes |
| PUT | `/events/:id` | Update event | Yes (organizer only) |
| DELETE | `/events/:id` | Delete event | Yes (organizer only) |
| POST | `/events/:id/publish` | Publish event | Yes (organizer only) |
| POST | `/events/:id/cancel` | Cancel event | Yes (organizer only) |

#### Registrations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/registrations` | Register for event | Yes |
| GET | `/registrations/my` | Get user's registrations | Yes |
| DELETE | `/registrations/:id` | Cancel registration | Yes |

#### Users

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/users/me` | Get current user profile | Yes |
| GET | `/users/me/events` | Get user's created events | Yes |

### API Examples

<details>
<summary><b>Register New User</b></summary>

```bash
curl -X POST http://localhost:8000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123",
    "name": "John Doe"
  }'
```

**Response (201 Created)**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "created_at": "2025-12-17T10:30:00Z"
}
```
</details>

<details>
<summary><b>Login</b></summary>

```bash
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepass123"
  }'
```

**Response (200 OK)**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-18T10:30:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user"
  }
}
```
</details>

<details>
<summary><b>Create Event</b></summary>

```bash
curl -X POST http://localhost:8000/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Tech Conference 2025",
    "description": "Annual technology conference",
    "start_datetime": "2025-06-15T09:00:00Z",
    "end_datetime": "2025-06-15T18:00:00Z",
    "location": "Convention Center, NY",
    "capacity": 500
  }'
```

**Response (201 Created)**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Tech Conference 2025",
  "description": "Annual technology conference",
  "start_datetime": "2025-06-15T09:00:00Z",
  "end_datetime": "2025-06-15T18:00:00Z",
  "location": "Convention Center, NY",
  "capacity": 500,
  "status": "draft",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T10:30:00Z"
}
```
</details>

<details>
<summary><b>Get Events with Filtering</b></summary>

```bash
# Get published events with pagination
curl "http://localhost:8000/events?page=1&page_size=10&status=published"

# Filter by date range
curl "http://localhost:8000/events?start_date_from=2025-01-01&start_date_to=2025-12-31"

# Filter by capacity
curl "http://localhost:8000/events?min_capacity=100&max_capacity=500"
```

**Response (200 OK)**
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "Tech Conference 2025",
      "status": "published",
      ...
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 42,
    "total_pages": 5
  }
}
```
</details>

<details>
<summary><b>Register for Event</b></summary>

```bash
curl -X POST http://localhost:8000/registrations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "event_id": "660e8400-e29b-41d4-a716-446655440001"
  }'
```

**Response (201 Created)**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_id": "660e8400-e29b-41d4-a716-446655440001",
  "status": "confirmed",
  "registered_at": "2025-12-17T10:30:00Z"
}
```
</details>

### Query Parameters for Event Filtering

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `page` | integer | Page number (default: 1) | `?page=2` |
| `page_size` | integer | Items per page (default: 10, max: 100) | `?page_size=20` |
| `status` | string | Filter by status: draft, published, cancelled | `?status=published` |
| `start_date_from` | datetime | Filter events starting from this date | `?start_date_from=2025-01-01` |
| `start_date_to` | datetime | Filter events starting before this date | `?start_date_to=2025-12-31` |
| `min_capacity` | integer | Minimum event capacity | `?min_capacity=50` |
| `max_capacity` | integer | Maximum event capacity | `?max_capacity=500` |

### Error Responses

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "specific_field",
    "reason": "validation_failed"
  }
}
```

Common HTTP status codes:
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

## 🛠 Development

### Local Development Setup

```bash
# 1. Clone repository
git clone https://github.com/kimashii-dan/event-hub.git
cd event-hub/backend

# 2. Install dependencies
go mod download

# 3. Setup local database (optional, or use Docker)
# Create PostgreSQL database named 'event_hub'

# 4. Configure environment
cp .env.example .env
# Edit .env with local database credentials

# 5. Run migrations
# Migrations run automatically on application start

# 6. Run application
go run cmd/app/main.go
```

### Project Structure Explained

```
backend/
├── cmd/app/main.go              # Application entry point, dependency injection
├── internal/
│   ├── config/config.go         # Configuration loader (env variables)
│   ├── database/postgres.go     # Database connection & auto-migrations
│   ├── domain/                  # Business entities & DTOs
│   │   ├── user.go              # User model, validation, DTOs
│   │   ├── event.go             # Event model, validation, DTOs
│   │   └── registration.go      # Registration model, DTOs
│   ├── handler/                 # HTTP request handlers (controllers)
│   │   ├── auth_handler.go      # Authentication endpoints
│   │   ├── event_handler.go     # Event CRUD endpoints
│   │   ├── registration_handler.go  # Registration endpoints
│   │   └── user_handler.go      # User profile endpoints
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go              # JWT authentication middleware
│   │   └── logger.go            # Request logging middleware
│   ├── repository/              # Data access layer (database operations)
│   │   ├── user_repository.go
│   │   ├── event_repository.go
│   │   └── registration_repository.go
│   └── service/                 # Business logic layer
│       ├── auth_service.go      # Authentication & user management
│       ├── event_service.go     # Event management logic
│       └── registration_service.go  # Registration logic
├── pkg/                         # Reusable packages (can be extracted)
│   ├── jwt/jwt.go               # JWT token generation & validation
│   └── response/response.go     # Standardized API response helpers
├── migrations/                  # SQL migration files
│   ├── 000001_create_users_table.up.sql
│   ├── 000002_create_events_table.up.sql
│   └── 000003_create_registrations_table.up.sql
└── tests/
    ├── integration/             # Integration tests (with test DB)
    └── unit/                    # Unit tests (mocked dependencies)
```

### Code Style Guidelines

- Follow **Clean Architecture** principles
- Use **meaningful variable names** in English
- Write **godoc comments** for all exported functions and types
- Keep functions **focused and small** (single responsibility)
- Use **dependency injection** for testability
- Handle errors explicitly (no silent failures)

### Database Migrations

Migrations run automatically on application startup. To create new migrations:

```bash
# Create new migration files manually in migrations/ directory
# Follow naming convention: 000XXX_description.up.sql and 000XXX_description.down.sql
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run only unit tests
go test ./tests/unit/...

# Run only integration tests
go test ./tests/integration/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

- **Unit Tests**: Test individual functions with mocked dependencies
  - Located in `tests/unit/`
  - Mock database and external dependencies
  - Fast execution, no external services required

- **Integration Tests**: Test complete workflows with real database
  - Located in `tests/integration/`
  - Use test database (separate from production)
  - Test full request/response cycles

### Writing Tests

Example test structure:

```go
func TestEventService_CreateEvent(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer cleanupTestDB(db)
    
    eventRepo := repository.NewEventRepository(db)
    eventService := service.NewEventService(eventRepo)
    
    // Test case
    req := &domain.CreateEventRequest{
        Title:         "Test Event",
        Description:   "Test Description",
        StartDatetime: time.Now().Add(24 * time.Hour),
        EndDatetime:   time.Now().Add(48 * time.Hour),
        Location:      "Test Location",
        Capacity:      100,
    }
    
    // Execute
    event, err := eventService.CreateEvent("user-123", req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, event)
    assert.Equal(t, "Test Event", event.Title)
    assert.Equal(t, "draft", event.Status)
}
```

## 📦 Docker Commands

```bash
# Build and start services
docker compose up --build

# Start services in background
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f

# Rebuild specific service
docker compose up --build backend

# Access database
docker compose exec db psql -U postgres -d event_hub

# Run tests in container
docker compose exec backend go test ./...
```

- Features: `feature/jd-user-auth`
- Bug fixes: `bugfix/jd-fix-login`
- Hotfixes: `hotfix/jd-security-patch`

### Daily Workflow

```bash
# Start work
git checkout main && git pull origin main
git checkout -b feature/initials-feature-name

# Finish work
git add . && git commit -m "feat: description"
git push origin feature/initials-feature-name
# Create PR on GitHub
```

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

### Git Workflow

#### Branch Naming Conventions

- **Feature**: `feature/initials-short-description`
  - Example: `feature/jd-event-filtering`
- **Bug Fix**: `bugfix/initials-issue-description`
  - Example: `bugfix/jd-registration-validation`
- **Hotfix**: `hotfix/initials-critical-fix`
  - Example: `hotfix/jd-auth-security`
- **Documentation**: `docs/initials-what-docs`
  - Example: `docs/jd-api-endpoints`

#### Development Process

1. **Create feature branch**
   ```bash
   git checkout -b feature/your-initials-feature-name
   ```

2. **Make changes following code style guidelines**
   - Write tests for new features
   - Update documentation
   - Follow Go conventions and Clean Architecture

3. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add event filtering by date range"
   git commit -m "fix: resolve registration capacity overflow"
   git commit -m "docs: update API documentation"
   ```

   **Commit types:**
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation changes
   - `style`: Code style changes (formatting)
   - `refactor`: Code refactoring
   - `test`: Adding or updating tests
   - `chore`: Maintenance tasks

4. **Push and create Pull Request**
   ```bash
   git push origin feature/your-initials-feature-name
   ```
   - Create PR on GitHub
   - Provide clear description of changes
   - Link related issues

5. **Code Review & Merge**
   - Team lead reviews the PR
   - Address feedback and make changes if needed
   - Once approved, team lead merges to `main`

### Pull Request Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass (`go test ./...`)
- [ ] New features include tests
- [ ] Documentation updated (README, code comments)
- [ ] No breaking changes (or clearly documented)
- [ ] Commit messages follow conventional commits
- [ ] Branch is up to date with `main`

## Database

- **PostgreSQL 15** with golang/migrate
- **ORM**: GORM

## ❗ Troubleshooting

**Port conflicts**: Change ports in `.env` or `docker-compose.yaml`
**Database issues**: Check `.env` values and run `docker compose logs db`
**Build failures**: Run `docker compose down && docker compose up --build`

## 📝 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## � Documentation

Comprehensive documentation is available in the `/docs` directory:

| Document | Description |
|----------|-------------|
| **[Documentation Index](./docs/README.md)** | Complete documentation overview and navigation |
| **[API Reference](./docs/API.md)** | Detailed REST API documentation with examples |
| **[Architecture Guide](./docs/ARCHITECTURE.md)** | System design, patterns, and technical decisions |
| **[Developer Guide](./docs/DEVELOPER_GUIDE.md)** | Development workflow, coding standards, testing |
| **[Deployment Guide](./docs/DEPLOYMENT.md)** | Production deployment for various platforms |

### Quick Links

- 🚀 **New to the project?** Start with [Quick Start](#-quick-start)
- 🔌 **Integrating the API?** Check [API Documentation](./docs/API.md)
- 💻 **Contributing code?** Read [Developer Guide](./docs/DEVELOPER_GUIDE.md)
- 🚢 **Deploying to production?** See [Deployment Guide](./docs/DEPLOYMENT.md)
- 🏗 **Understanding the system?** Review [Architecture](./docs/ARCHITECTURE.md)

## �👥 Team

**Event Hub Development Team**
- Repository: [github.com/kimashii-dan/event-hub](https://github.com/kimashii-dan/event-hub)

## 🆘 Support & Contact

- **Issues**: [GitHub Issues](https://github.com/kimashii-dan/event-hub/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kimashii-dan/event-hub/discussions)
- **Documentation**: [docs/README.md](./docs/README.md)

## 🗺 Roadmap

### Current Version (v1.0)
- ✅ User authentication & authorization
- ✅ Event CRUD operations
- ✅ Event registration system
- ✅ Advanced filtering & pagination
- ✅ Docker deployment
- ✅ Comprehensive documentation

### Planned Features (v2.0)
- [ ] Email notifications for event updates
- [ ] Event categories and tags
- [ ] Advanced search with full-text search
- [ ] Event attendee management dashboard
- [ ] Export registrations to CSV
- [ ] Event analytics and reports
- [ ] Webhook support for integrations
- [ ] Rate limiting and API throttling

---

**Built with ❤️ using Go, PostgreSQL, and Docker**

**📖 [View Full Documentation](./docs/README.md)** | **🐛 [Report Issues](https://github.com/kimashii-dan/event-hub/issues)** | **💬 [Join Discussions](https://github.com/kimashii-dan/event-hub/discussions)**
