# Developer Guide - Event Hub

> Comprehensive guide for developers working on Event Hub

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Adding New Features](#adding-new-features)
- [Writing Tests](#writing-tests)
- [Debugging](#debugging)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

- **Go 1.21+** installed
- **Docker & Docker Compose** for database
- **PostgreSQL Client** (optional, for direct database access)
- **Postman** or **curl** for API testing
- **Git** for version control

### Initial Setup

```bash
# 1. Clone repository
git clone https://github.com/kimashii-dan/event-hub.git
cd event-hub/backend

# 2. Install Go dependencies
go mod download

# 3. Setup environment
cp docker/.env.example docker/.env
# Edit docker/.env with your configuration

# 4. Start database and redis with Docker
cd docker
docker compose up -d db redis

# 5. Run application
cd ..
go run cmd/app/main.go
```

### Project Dependencies

Key Go modules used:
```go
github.com/gin-gonic/gin          // HTTP web framework
gorm.io/gorm                      // ORM for database
gorm.io/driver/postgres           // PostgreSQL driver
github.com/golang-jwt/jwt/v5      // JWT authentication
github.com/joho/godotenv          // Environment variables
golang.org/x/crypto/bcrypt        // Password hashing
github.com/stretchr/testify       // Testing assertions
github.com/redis/go-redis/v9      // Redis client
```

## Development Workflow

### Branch Strategy

```
main (protected)
  ├── feature/initials-feature-name
  ├── bugfix/initials-bug-description
  ├── hotfix/initials-critical-fix
  └── docs/initials-documentation
```

### Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/jd-add-event-tags
   ```

2. **Make changes following guidelines**
   - Follow Clean Architecture principles
   - Write tests for new code
   - Update documentation

3. **Test your changes**
   ```bash
   # Run tests
   go test ./...
   
   # Check code formatting
   go fmt ./...
   
   # Run linter (if configured)
   golangci-lint run
   ```

4. **Commit with conventional commits**
   ```bash
   git add .
   git commit -m "feat: add tags support for events"
   ```

5. **Push and create PR**
   ```bash
   git push origin feature/jd-add-event-tags
   ```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting (no functional changes)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples:**
```bash
feat(events): add filtering by tags
fix(auth): resolve token expiration issue
docs(api): update event endpoints documentation
test(registration): add integration tests for capacity limits
refactor(service): simplify event validation logic
```

## Code Style Guidelines

### General Principles

1. **Follow Go idioms and conventions**
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://go.dev/doc/effective_go)
   - Use meaningful, clear variable names

2. **Clean Architecture layers**
   - Keep layers independent
   - Use dependency injection
   - Domain layer has no external dependencies

3. **Error handling**
   - Always handle errors explicitly
   - Wrap errors with context: `fmt.Errorf("failed to create event: %w", err)`
   - Return errors, don't panic (except in `main()`)

4. **Comments and documentation**
   - Write godoc comments for all exported types and functions
   - Explain "why", not just "what"
   - Keep comments up-to-date

### Naming Conventions

```go
// Types: PascalCase
type EventService struct {}
type CreateEventRequest struct {}

// Functions/Methods: camelCase (exported) or camelCase (unexported)
func NewEventService() {}
func (s *EventService) CreateEvent() {}
func getUserIDFromContext() {} // unexported

// Variables: camelCase
var eventService *EventService
var userID string

// Constants: PascalCase or UPPER_SNAKE_CASE
const MaxPageSize = 100
const DEFAULT_PAGE_SIZE = 10

// Interfaces: "-er" suffix when possible
type EventRepository interface {
    Create(event *Event) error
    GetByID(id string) (*Event, error)
}
```

### File Organization

```go
package handler

// 1. Imports (grouped: stdlib, external, internal)
import (
    "fmt"
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "github.com/Fixsbreaker/event-hub/backend/internal/domain"
    "github.com/Fixsbreaker/event-hub/backend/internal/service"
)

// 2. Type definitions
type EventHandler struct {
    eventService *service.EventService
}

// 3. Constructor
func NewEventHandler(...) *EventHandler { }

// 4. Public methods
func (h *EventHandler) CreateEvent() { }
func (h *EventHandler) UpdateEvent() { }

// 5. Helper functions (unexported)
func getUserIDFromContext() { }
func validateEventInput() { }
```

### Function Documentation

```go
// CreateEvent handles POST /events (protected)
// Creates a new event with the authenticated user as organizer.
//
// Request Body: domain.CreateEventRequest
//   - title: Event title (min 3 chars)
//   - description: Event description (optional)
//   - start_datetime: Event start (ISO 8601)
//   - end_datetime: Event end (ISO 8601)
//   - location: Venue location
//   - capacity: Max attendees (min 1)
//
// Success Response: 201 Created
//   Returns the created event with generated ID and timestamps
//
// Error Responses:
//   - 400 Bad Request: Invalid input or validation failed
//   - 401 Unauthorized: Missing or invalid JWT token
//
// Example:
//   POST /events
//   Authorization: Bearer <token>
//   Body: {"title": "Tech Conf 2025", "capacity": 500, ...}
func (h *EventHandler) CreateEvent(c *gin.Context) {
    // Implementation
}
```

## Adding New Features

### Example: Adding Event Tags Feature

#### 1. Domain Layer (`internal/domain/`)

```go
// event.go - Add tags field to Event struct
type Event struct {
    // ... existing fields
    Tags []string `gorm:"type:text[]" json:"tags"` // Event tags
}

// Add tags to DTOs
type CreateEventRequest struct {
    // ... existing fields
    Tags []string `json:"tags"` // Optional tags
}
```

#### 2. Database Migration

```sql
-- migrations/000004_add_event_tags.up.sql
ALTER TABLE events ADD COLUMN tags TEXT[];

-- migrations/000004_add_event_tags.down.sql
ALTER TABLE events DROP COLUMN tags;
```

#### 3. Repository Layer (`internal/repository/`)

```go
// event_repository.go - Add tag filtering
func (r *EventRepository) GetByTags(tags []string) ([]Event, error) {
    var events []Event
    err := r.db.Where("tags @> ?", pq.Array(tags)).Find(&events).Error
    return events, err
}
```

#### 4. Service Layer (`internal/service/`)

```go
// event_service.go - Add business logic for tags
func (s *EventService) GetEventsByTags(tags []string) ([]Event, error) {
    if len(tags) == 0 {
        return nil, fmt.Errorf("at least one tag required")
    }
    return s.eventRepo.GetByTags(tags)
}
```

#### 5. Handler Layer (`internal/handler/`)

```go
// event_handler.go - Add HTTP endpoint
func (h *EventHandler) GetEventsByTags(c *gin.Context) {
    tags := c.QueryArray("tags")
    
    events, err := h.eventService.GetEventsByTags(tags)
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    
    response.Success(c, 200, events)
}

// Register route in NewEventHandler
public.GET("/tags", h.GetEventsByTags)
```

#### 6. Write Tests

```go
// tests/unit/event_service_test.go
func TestEventService_GetEventsByTags(t *testing.T) {
    // Setup
    mockRepo := new(MockEventRepository)
    service := NewEventService(mockRepo)
    
    expectedEvents := []Event{
        {ID: "1", Title: "Tech Event", Tags: []string{"tech", "ai"}},
    }
    mockRepo.On("GetByTags", []string{"tech"}).Return(expectedEvents, nil)
    
    // Execute
    events, err := service.GetEventsByTags([]string{"tech"})
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, events, 1)
    assert.Equal(t, "Tech Event", events[0].Title)
}
```

#### 7. Update Documentation

- Update `docs/API.md` with new endpoint
- Update `README.md` if needed
- Add inline code comments

## Writing Tests

### Unit Tests

Test individual functions with mocked dependencies:

```go
// tests/unit/auth_service_test.go
package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
    args := m.Called(user)
    return args.Error(0)
}

// Test function
func TestAuthService_Register_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    authService := service.NewAuthService(mockRepo, "secret", "24h")
    
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    req := &domain.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "Test User",
    }
    
    // Act
    user, err := authService.Register(req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
    mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
    // Test error case
    mockRepo := new(MockUserRepository)
    authService := service.NewAuthService(mockRepo, "secret", "24h")
    
    mockRepo.On("Create", mock.Anything).Return(fmt.Errorf("email already exists"))
    
    req := &domain.CreateUserRequest{
        Email:    "existing@example.com",
        Password: "password123",
        Name:     "Test User",
    }
    
    user, err := authService.Register(req)
    
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "email already exists")
}
```

### Integration Tests

Test complete workflows with real database:

```go
// tests/integration/events_crud_integration_test.go
package integration

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestEventsCRUD_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(db)
    
    // Setup router with real dependencies
    router := setupTestRouter(db)
    
    // Create user and get token
    token := createTestUserAndLogin(t, router)
    
    // Test: Create event
    t.Run("CreateEvent", func(t *testing.T) {
        body := `{
            "title": "Test Event",
            "description": "Test Description",
            "start_datetime": "2025-06-15T09:00:00Z",
            "end_datetime": "2025-06-15T18:00:00Z",
            "location": "Test Location",
            "capacity": 100
        }`
        
        req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.Equal(t, "Test Event", response["title"])
        assert.Equal(t, "draft", response["status"])
    })
    
    // More test cases...
}
```

### Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage in browser
open coverage.html
```

## Debugging

### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugger
dlv debug cmd/app/main.go

# Set breakpoint
(dlv) break handler.CreateEvent
(dlv) continue

# Inspect variables
(dlv) print userID
(dlv) print req
```

### Logging

```go
// Add debug logging
import "log"

func (s *EventService) CreateEvent(userID string, req *CreateEventRequest) (*Event, error) {
    log.Printf("DEBUG: Creating event for user %s: %+v", userID, req)
    
    event := &Event{
        OrganizerID: userID,
        Title: req.Title,
        // ...
    }
    
    log.Printf("DEBUG: Event entity created: %+v", event)
    
    err := s.eventRepo.Create(event)
    if err != nil {
        log.Printf("ERROR: Failed to create event: %v", err)
        return nil, err
    }
    
    log.Printf("INFO: Event created successfully: %s", event.ID)
    return event, nil
}
```

### Database Debugging

```bash
# Access PostgreSQL in Docker
docker compose exec db psql -U postgres -d event_hub

# Check events table
event_hub=# SELECT id, title, status, organizer_id FROM events;

# Check registrations
event_hub=# SELECT * FROM registrations WHERE event_id = 'some-uuid';

# Check query performance
event_hub=# EXPLAIN ANALYZE SELECT * FROM events WHERE status = 'published';
```

## Common Tasks

### Add New Endpoint

1. Define DTO in `domain/`
2. Add method to service
3. Add handler method
4. Register route in handler constructor
5. Write tests
6. Update API documentation

### Add Database Index

```sql
-- migrations/000XXX_add_index.up.sql
CREATE INDEX idx_events_location ON events(location);

-- migrations/000XXX_add_index.down.sql
DROP INDEX idx_events_location;
```

### Update Environment Variable

1. Add to `docker/.env.example`
2. Add to `internal/config/config.go`
3. Update README.md
4. Use in code via `cfg.YourNewVar`

### Add New Dependency

```bash
# Add dependency
go get github.com/some/package

# Update go.mod and go.sum
go mod tidy

# Verify
go mod verify
```

## Troubleshooting

### Common Issues

**Issue: "Port already in use"**
```bash
# Find process using port 8000
lsof -i :8000

# Kill process
kill -9 <PID>

# Or change port in .env
SERVER_PORT=8080
```

**Issue: "Database connection refused"**
```bash
# Check if database is running
docker compose ps

# Restart database
docker compose restart db

# Check logs
docker compose logs db
```

**Issue: "Test database issues"**
```bash
# Drop and recreate test database
docker compose exec db psql -U postgres -c "DROP DATABASE IF EXISTS event_hub_test;"
docker compose exec db psql -U postgres -c "CREATE DATABASE event_hub_test;"
```

**Issue: "Import cycle detected"**
- Check dependency direction (inner layers shouldn't depend on outer)
- Refactor to use interfaces
- Consider extracting shared code to `pkg/`

**Issue: "GORM preload not working"**
```go
// Make sure to use Preload
db.Preload("Organizer").First(&event, "id = ?", id)

// Check struct tags
type Event struct {
    OrganizerID string `gorm:"type:uuid"`
    Organizer   *User  `gorm:"foreignKey:OrganizerID"` // Must match!
}
```

### Performance Optimization

**Slow queries:**
```go
// Enable SQL logging
db.Debug().Where(...).Find(&events)

// Add index
CREATE INDEX idx_events_start_datetime ON events(start_datetime);

// Use pagination
.Limit(pageSize).Offset((page - 1) * pageSize)

// Eager load relationships
.Preload("Organizer").Find(&events)
```

**Memory issues:**
```go
// Stream large result sets
rows, err := db.Model(&Event{}).Where(...).Rows()
defer rows.Close()

for rows.Next() {
    var event Event
    db.ScanRows(rows, &event)
    // Process event
}
```

## Resources

### Documentation
- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Guide](https://gorm.io/docs/)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)

### Tools
- [Postman Collection](../backend/postman_collection.json) - API testing
- [VS Code Go Extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
- [TablePlus](https://tableplus.com/) - Database GUI

### Internal Docs
- [Architecture Documentation](./ARCHITECTURE.md)
- [API Documentation](./API.md)
- [README](../README.md)

---

**Questions?** Open an issue or ask in team discussions!
