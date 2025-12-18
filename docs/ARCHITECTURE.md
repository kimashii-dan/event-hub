# Event Hub - Architecture Documentation

> Technical architecture and design decisions for Event Hub

## Table of Contents

- [Overview](#overview)
- [Architecture Pattern](#architecture-pattern)
- [Project Structure](#project-structure)
- [Layer Responsibilities](#layer-responsibilities)
- [Data Flow](#data-flow)
- [Database Schema](#database-schema)
- [Authentication & Authorization](#authentication--authorization)
- [Design Patterns](#design-patterns)
- [Technology Stack](#technology-stack)

## Overview

Event Hub is built using **Clean Architecture** principles, ensuring:
- **Separation of Concerns**: Each layer has clear responsibilities
- **Testability**: Easy to test with mocked dependencies
- **Maintainability**: Changes in one layer don't affect others
- **Scalability**: Easy to add new features without breaking existing code

## Architecture Pattern

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────┐
│           External Interfaces Layer              │
│  (HTTP Handlers, CLI, gRPC - Future)            │
│                   handler/                       │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│            Application Layer                     │
│  (Business Logic, Use Cases)                    │
│                  service/                        │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│            Domain Layer                          │
│  (Entities, Business Rules)                     │
│                  domain/                         │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│         Infrastructure Layer                     │
│  (Database, External Services)                  │
│           repository/, database/                 │
└─────────────────────────────────────────────────┘
```

### Dependency Rule

**Inner layers don't know about outer layers**:
- `domain/` → No dependencies
- `service/` → Depends only on `domain/` and `repository/` interfaces
- `repository/` → Depends on `domain/`
- `handler/` → Depends on `service/` and `domain/`

## Project Structure

```
backend/
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
│                                   # - Dependency injection
│                                   # - Server setup
│                                   # - Route configuration
│
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management
│   │                               # - Environment variables
│   │                               # - App settings
│   │
│   ├── database/
│   │   └── postgres.go             # Database connection
│   │                               # - Connection pooling
│   │                               # - Auto-migrations
│   │
│   ├── domain/                     # DOMAIN LAYER
│   │   ├── user.go                 # User entity & DTOs
│   │   ├── user_dto.go             # User update DTOs
│   │   ├── event.go                # Event entity & DTOs
│   │   ├── registration.go         # Registration entity & DTOs
│   │   ├── notific.go              # Notification entity
│   │   └── motific_dto.go          # Notification DTOs
│   │                               # - Entities (structs)
│   │                               # - Validation rules
│   │                               # - DTOs for requests/responses
│   │
│   ├── handler/                    # PRESENTATION LAYER
│   │   ├── auth_handler.go         # Authentication endpoints
│   │   ├── event_handler.go        # Event CRUD endpoints
│   │   ├── registration_handler.go # Registration endpoints
│   │   ├── user_handler.go         # User profile endpoints
│   │   └── notific_handler.go      # Notification endpoints
│   │                               # - HTTP request/response handling
│   │                               # - Input validation
│   │                               # - Routing setup
│   │
│   ├── middleware/                 # HTTP MIDDLEWARE
│   │   ├── auth.go                 # JWT authentication
│   │   └── logger.go               # Request logging
│   │                               # - Cross-cutting concerns
│   │                               # - Request/Response modification
│   │
│   ├── repository/                 # INFRASTRUCTURE LAYER
│   │   ├── user_repository.go      # User data access
│   │   ├── event_repository.go     # Event data access
│   │   ├── registration_repository.go  # Registration data access
│   │   └── notific_repo.go         # Notification data access
│   │                               # - Database operations (CRUD)
│   │                               # - Query building
│   │                               # - Data mapping
│   │
│   └── service/                    # APPLICATION LAYER
│       ├── auth_service.go         # Authentication logic
│       ├── event_service.go        # Event business logic
│       ├── registration_service.go # Registration logic
│       ├── user_service.go         # User management logic
│       └── notific_service.go      # Notification logic
│                                   # - Use cases
│                                   # - Business rules
│                                   # - Business rules
│                                   # - Orchestration
│   │
│   ├── worker/                     # BACKGROUND JOBS
│   │   └── notification_pool.go    # Worker pool for notifications
│   │
│   ├── cache/                      # CACHING LAYER
│   │   └── redis_cache.go          # Redis client wrapper
│   │
├── pkg/                            # SHARED UTILITIES
│   ├── jwt/
│   │   └── jwt.go                  # JWT token utilities
│   └── response/
│       └── response.go             # API response helpers
│
├── migrations/                     # DATABASE MIGRATIONS
│   ├── 000001_create_users_table.up.sql
│   ├── 000002_create_events_table.up.sql
│   └── 000003_create_registrations_table.up.sql
│
└── tests/
    ├── integration/                # Integration tests
    └── unit/                       # Unit tests
```

## Layer Responsibilities

### 1. Domain Layer (`domain/`)

**Purpose**: Core business entities and rules

**Responsibilities**:
- Define domain models (User, Event, Registration)
- Implement business validation logic
- Define DTOs for data transfer
- No external dependencies

**Example: Event Validation**
```go
func (e *Event) Validate() error {
    if e.Title == "" {
        return fmt.Errorf("title is required")
    }
    if e.Capacity < 1 {
        return fmt.Errorf("capacity must be at least 1")
    }
    if e.EndDatetime.Before(e.StartDatetime) {
        return fmt.Errorf("end time must be after start time")
    }
    return nil
}
```

### 2. Service Layer (`service/`)

**Purpose**: Business logic and use cases

**Responsibilities**:
- Implement business operations (create event, register user)
- Coordinate between repositories
- Enforce business rules
- Transaction management

**Example: Create Event Service**
```go
func (s *EventService) CreateEvent(userID string, req *domain.CreateEventRequest) (*domain.Event, error) {
    // 1. Create domain entity
    event := &domain.Event{
        OrganizerID: userID,
        Title: req.Title,
        // ... other fields
        Status: "draft",
    }
    
    // 2. Validate business rules
    if err := event.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 3. Persist to database
    if err := s.eventRepo.Create(event); err != nil {
        return nil, fmt.Errorf("failed to create event: %w", err)
    }
    
    return event, nil
}
```

### 3. Repository Layer (`repository/`)

**Purpose**: Data persistence abstraction

**Responsibilities**:
- CRUD operations
- Query construction
- Data mapping between database and domain models
- Handle database-specific errors

**Example: Event Repository**
```go
func (r *EventRepository) GetByID(id string) (*domain.Event, error) {
    var event domain.Event
    err := r.db.Preload("Organizer").First(&event, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("event not found")
        }
        return nil, err
    }
    return &event, nil
}
```

### 3.5. Cache Layer (`cache/`)

**Purpose**: High-speed data retrieval

**Responsibilities**:
- Key-value storage (Redis)
- Cache invalidation (TTL)
- Read-through caching strategies

### 3.6. Worker Layer (`worker/`)

**Purpose**: Asynchronous background processing

**Responsibilities**:
- Offload heavy tasks from HTTP handlers
- Manage concurrency with Worker Pool
- Graceful shutdown of background jobs

### 4. Handler Layer (`handler/`)

**Purpose**: HTTP interface

**Responsibilities**:
- Parse HTTP requests
- Validate input (basic format validation)
- Call appropriate service methods
- Format HTTP responses
- Handle HTTP-specific errors

**Example: Event Handler**
```go
func (h *EventHandler) CreateEvent(c *gin.Context) {
    // 1. Get user from context (set by auth middleware)
    userID, _ := getUserIDFromContext(c)
    
    // 2. Parse request body
    var req domain.CreateEventRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid_input", err.Error())
        return
    }
    
    // 3. Call service
    event, err := h.eventService.CreateEvent(userID, &req)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "creation_failed", err.Error())
        return
    }
    
    // 4. Return response
    response.Success(c, http.StatusCreated, event)
}
```

### 5. Middleware Layer (`middleware/`)

**Purpose**: Cross-cutting concerns

**Responsibilities**:
- Authentication & Authorization
- Request/Response logging
- Error handling
- Rate limiting (future)

**Example: Auth Middleware**
```go
func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from header
        token := extractToken(c)
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        
        // 2. Validate token
        claims, err := jwt.ValidateToken(token, jwtSecret)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid_token"})
            return
        }
        
        // 3. Set user context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

## Data Flow

### Example: Create Event Flow

```
1. HTTP Request
   POST /events
   Headers: Authorization: Bearer <token>
   Body: { "title": "...", "capacity": 100, ... }
   │
   ▼
2. Middleware (auth.go)
   - Validates JWT token
   - Extracts user_id from token
   - Sets user_id in Gin context
   │
   ▼
3. Handler (event_handler.go)
   - Parses request body
   - Validates input format
   - Calls EventService.CreateEvent()
   │
   ▼
4. Service (event_service.go)
   - Creates Event domain entity
   - Validates business rules (dates, capacity)
   - Calls EventRepository.Create()
   │
   ▼
5. Repository (event_repository.go)
   - Executes SQL INSERT
   - Handles database errors
   - Returns created entity
   │
   ▼
6. Response Path (reverse)
   Service → Handler → HTTP Response (201 Created)
```

### Example: Get Events with Filtering Flow

```
1. HTTP Request
   GET /events?status=published&page=1&page_size=10
   │
   ▼
2. Handler (event_handler.go)
   - Parses query parameters
   - Validates pagination params
   - Calls EventService.GetAllEvents()
   │
   ▼
3. Service (event_service.go)
   - Applies business logic filters
   - Calls EventRepository.GetAll()
   │
   ▼
4. Repository (event_repository.go)
   - Builds SQL query with filters
   - Applies pagination
   - Executes query
   - Returns events + pagination data
   │
   ▼
5. Response
   Handler → HTTP Response (200 OK)
   { "data": [...], "pagination": {...} }
```

### Example: Async Notification Flow

```
1. Trigger (e.g., Event Cancelled)
   Service detects state change
   │
   ▼
2. Push to Worker Pool
   notificationService.SendNotification(...)
   → Adds job to buffered channel
   → Service returns immediately (Non-blocking)
   │
   ▼
3. Worker Processing (Background)
   Worker goroutine picks up job
   → Simulates email sending (2s delay)
   → Updates DB status to "sent"
```

## Database Schema

### Entity-Relationship Diagram

```
┌─────────────────┐
│      users      │
├─────────────────┤
│ id (PK)         │──┐
│ email (UNIQUE)  │  │
│ password_hash   │  │
│ name            │  │
│ role            │  │
│ created_at      │  │
│ updated_at      │  │
│ deleted_at      │  │
└─────────────────┘  │
                     │
         ┌───────────┴───────────────────┐
         │                               │
         │                               │
    ┌────▼──────────┐            ┌──────▼─────────────┐
    │    events     │            │  notifications     │
    ├───────────────┤            ├────────────────────┤
    │ id (PK)       │──┐         │ id (PK)            │
    │ organizer_id  │  │         │ user_id (FK)       │
    │ (FK → users)  │  │         │ title              │
    │ title         │  │         │ message            │
    │ description   │  │         │ read               │
    │ start_datetime│  │         │ created_at         │
    │ end_datetime  │  │         └────────────────────┘
    │ location      │  │
    │ capacity      │  │    ┌────▼────────────────┐
    │ status        │  │    │   registrations     │
    │ created_at    │  │    ├─────────────────────┤
    │ updated_at    │  └────│ id (PK)             │
    │ deleted_at    │       │ user_id (FK)        │
    └───────────────┘       │ event_id (FK)       │
                            │ status              │
                            │ registered_at       │
                            │ created_at          │
                            │ updated_at          │
                            │ deleted_at          │
                            └─────────────────────┘
```

### Table Relationships

- **users** (1) → (N) **events**: One user can organize many events
- **users** (1) → (N) **registrations**: One user can register for many events
- **users** (1) → (N) **notifications**: One user can have many notifications
- **events** (1) → (N) **registrations**: One event can have many registrations

### Indexes

**Performance optimization**:
- `users.email` - UNIQUE index for fast email lookups
- `events.organizer_id` - Index for filtering events by organizer
- `events.start_datetime` - Index for date range filtering
- `events.status` - Index for status filtering
- `registrations.user_id` - Index for user's registrations
- `registrations.event_id` - Index for event's registrations
- `notifications.user_id` - Index for user's notifications
- `notifications.read` - Index for filtering unread notifications

### Soft Deletes

All entities use `deleted_at` field for soft deletion:
- Allows data recovery
- Maintains referential integrity
- Historical data preservation

## Authentication & Authorization

### JWT Token Structure

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "exp": 1734510600
}
```

### Authentication Flow

```
1. User Registration/Login
   ↓
2. Server generates JWT token
   - Includes user_id and expiration
   - Signed with JWT_SECRET
   ↓
3. Client stores token
   - Recommended: httpOnly cookie or secure storage
   ↓
4. Client includes token in requests
   - Header: Authorization: Bearer <token>
   ↓
5. Auth Middleware validates token
   - Verifies signature
   - Checks expiration
   - Extracts user_id
   ↓
6. Request proceeds with user context
```

### Authorization Rules

| Resource | Action | Rule |
|----------|--------|------|
| Event | Create | Any authenticated user |
| Event | Read | Public (no auth required) |
| Event | Update | Only event organizer |
| Event | Delete | Only event organizer |
| Event | Publish | Only event organizer |
| Registration | Create | Authenticated user, event not full |
| Registration | Delete | Registration owner only |

## Design Patterns

### 1. Repository Pattern
**Purpose**: Abstracts data access logic

**Benefits**:
- Easy to swap database implementations
- Testable with mock repositories
- Centralized data access logic

### 2. Dependency Injection
**Purpose**: Loose coupling between layers

**Implementation**:
```go
// In main.go
userRepo := repository.NewUserRepository(db)
authService := service.NewAuthService(userRepo, jwtSecret, jwtExpiration)
handler.NewAuthHandler(router, authService)
```

**Benefits**:
- Easy to test with mocks
- Clear dependencies
- Flexible configuration

### 3. DTO Pattern
**Purpose**: Separate internal models from API contracts

**Benefits**:
- API stability (internal changes don't affect API)
- Input validation
- Security (hide internal fields)

### 4. Middleware Pattern
**Purpose**: Cross-cutting concerns

**Benefits**:
- Reusable logic
- Clean separation
- Clean separation
- Easy to add/remove

### 5. Worker Pool Pattern
**Purpose**: Manage concurrency for background tasks

**Benefits**:
- Limits max concurrent goroutines (prevent OOM)
- Asynchronous processing (fast API response)
- Graceful handling of shutdowns

## Technology Stack

### Core Framework
- **Gin**: High-performance HTTP web framework
- **GORM**: ORM for database operations
- **GORM**: ORM for database operations
- **PostgreSQL**: Relational database
- **Redis**: In-memory data store for caching

### Libraries
- **jwt-go**: JWT token generation and validation
- **godotenv**: Environment variable loading
- **testify**: Testing assertions

### Tools
- **Docker**: Containerization
- **Docker Compose**: Multi-container orchestration
- **golang-migrate**: Database migrations (manual)

## Performance Considerations

### Database Optimization
- **Connection Pooling**: GORM handles connection pooling
- **Indexes**: Strategic indexes on frequently queried fields
- **Eager Loading**: Use `Preload` to avoid N+1 queries
- **Pagination**: Always paginate large datasets

### Caching Strategy (Implemented)
- Redis for frequently accessed data (e.g., Event Details)
- Cache invalidation on updates
- TTL-based expiration (10 minutes)

### Scaling Strategy
- **Horizontal Scaling**: Stateless application (JWT tokens)
- **Database Replication**: Read replicas for read-heavy workloads
- **Load Balancing**: Distribute traffic across instances

## Error Handling Strategy

### Error Types

1. **Validation Errors** (400)
   - Invalid input format
   - Business rule violations

2. **Authentication Errors** (401)
   - Missing token
   - Invalid token
   - Expired token

3. **Authorization Errors** (403)
   - Insufficient permissions

4. **Not Found Errors** (404)
   - Resource doesn't exist

5. **Conflict Errors** (409)
   - Duplicate resources
   - Business constraint violations

6. **Server Errors** (500)
   - Unexpected errors
   - Database failures

### Error Response Format

```json
{
  "error": "error_code",
  "message": "Human-readable message",
  "details": {
    "field": "specific_field",
    "reason": "additional_context"
  }
}
```

## Testing Strategy

### Unit Tests
- Test individual functions
- Mock all dependencies
- Fast execution
- High coverage target: 80%+

### Integration Tests
- Test complete workflows
- Use test database
- Real dependencies
- Cover critical paths

### Test Organization
```
tests/
├── unit/
│   ├── auth_service_test.go
│   ├── setup.go              # Test helpers
│   └── test_helpers.go       # Mock factories
└── integration/
    ├── events_crud_integration_test.go
    ├── registration_integration_test.go
    └── helpers_event_insert_test.go
```

## Future Improvements

### Planned Architecture Changes

1. **Event-Driven Architecture**
   - Event sourcing for audit trail
   - Message queues (RabbitMQ/Kafka)
   - Asynchronous processing

2. **Microservices**
   - Separate services for Auth, Events, Notifications
   - API Gateway
   - Service mesh

3. **Caching Layer**
   - Redis for session management
   - Query result caching
   - Distributed caching

4. **Observability**
   - Structured logging (zerolog)
   - Metrics (Prometheus)
   - Distributed tracing (Jaeger)

5. **API Versioning**
   - `/api/v1/`, `/api/v2/`
   - Backward compatibility

---


