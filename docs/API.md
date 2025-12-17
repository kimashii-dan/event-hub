# Event Hub API Documentation

> Complete API reference for Event Hub REST API v1

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Authentication Endpoints](#authentication-endpoints)
  - [Event Endpoints](#event-endpoints)
  - [Registration Endpoints](#registration-endpoints)
  - [User Endpoints](#user-endpoints)
- [Data Models](#data-models)
- [Error Handling](#error-handling)

## Overview

### Base URL
```
http://localhost:8000
```

### Content Type
All requests and responses use `application/json` content type.

### Date Format
All datetime fields use ISO 8601 format: `2025-12-17T10:30:00Z`

### Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Missing or invalid authentication |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 500 | Internal Server Error |

## Authentication

Event Hub uses **JWT (JSON Web Token)** for authentication.

### Getting a Token

1. Register a new user or login with existing credentials
2. Include the token in the `Authorization` header for protected endpoints:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Expiration

Tokens expire after **24 hours** by default (configurable via `JWT_EXPIRATION_TIME` environment variable).

## Rate Limiting

Currently, there are no rate limits implemented. This will be added in future versions.

---

## Endpoints

## Authentication Endpoints

### Register User

Create a new user account.

**Endpoint:** `POST /auth/register`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

**Request Parameters:**

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| email | string | Yes | Valid email format, unique | User's email address |
| password | string | Yes | Min 8 characters | User's password (will be hashed) |
| name | string | Yes | Min 2 characters | User's full name |

**Success Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T10:30:00Z"
}
```

**Error Responses:**

*400 Bad Request - Validation Error:*
```json
{
  "error": "validation_failed",
  "message": "Email is required"
}
```

*409 Conflict - Email Already Exists:*
```json
{
  "error": "email_already_exists",
  "message": "User with this email already exists"
}
```

---

### Login

Authenticate and receive a JWT token.

**Endpoint:** `POST /auth/login`

**Authentication:** Not required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiZXhwIjoxNzM0NTEwNjAwfQ.xyz",
  "expires_at": "2025-12-18T10:30:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "created_at": "2025-12-17T10:30:00Z",
    "updated_at": "2025-12-17T10:30:00Z"
  }
}
```

**Error Responses:**

*401 Unauthorized - Invalid Credentials:*
```json
{
  "error": "invalid_credentials",
  "message": "Invalid email or password"
}
```

---

## Event Endpoints

### Get All Events

Retrieve a paginated list of events with optional filtering.

**Endpoint:** `GET /events`

**Authentication:** Not required

**Query Parameters:**

| Parameter | Type | Default | Description | Example |
|-----------|------|---------|-------------|---------|
| page | integer | 1 | Page number | `?page=2` |
| page_size | integer | 10 | Items per page (max: 100) | `?page_size=20` |
| status | string | - | Filter by status | `?status=published` |
| start_date_from | datetime | - | Events starting after this date | `?start_date_from=2025-01-01T00:00:00Z` |
| start_date_to | datetime | - | Events starting before this date | `?start_date_to=2025-12-31T23:59:59Z` |
| min_capacity | integer | - | Minimum capacity | `?min_capacity=50` |
| max_capacity | integer | - | Maximum capacity | `?max_capacity=500` |

**Event Status Values:**
- `draft` - Event is being created (not visible to public)
- `published` - Event is live and accepting registrations
- `cancelled` - Event has been cancelled

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Tech Conference 2025",
      "description": "Annual technology conference featuring latest innovations",
      "start_datetime": "2025-06-15T09:00:00Z",
      "end_datetime": "2025-06-15T18:00:00Z",
      "location": "Convention Center, New York",
      "capacity": 500,
      "status": "published",
      "created_at": "2025-12-17T10:30:00Z",
      "updated_at": "2025-12-17T10:30:00Z"
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

**Example Requests:**

```bash
# Get all published events
curl "http://localhost:8000/events?status=published"

# Get events in date range with pagination
curl "http://localhost:8000/events?start_date_from=2025-01-01&start_date_to=2025-12-31&page=1&page_size=20"

# Get events with capacity between 100-500
curl "http://localhost:8000/events?min_capacity=100&max_capacity=500"
```

---

### Get Event by ID

Retrieve a single event by its ID.

**Endpoint:** `GET /events/:id`

**Authentication:** Not required

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Event ID |

**Success Response (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "organizer": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "organizer@example.com",
    "name": "Jane Smith",
    "role": "user"
  },
  "title": "Tech Conference 2025",
  "description": "Annual technology conference",
  "start_datetime": "2025-06-15T09:00:00Z",
  "end_datetime": "2025-06-15T18:00:00Z",
  "location": "Convention Center, New York",
  "capacity": 500,
  "status": "published",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T10:30:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "event_not_found",
  "message": "Event with the specified ID does not exist"
}
```

---

### Create Event

Create a new event. Events are created in `draft` status by default.

**Endpoint:** `POST /events`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "title": "Tech Conference 2025",
  "description": "Annual technology conference featuring latest innovations in AI, Cloud, and Web3",
  "start_datetime": "2025-06-15T09:00:00Z",
  "end_datetime": "2025-06-15T18:00:00Z",
  "location": "Convention Center, New York",
  "capacity": 500
}
```

**Request Parameters:**

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| title | string | Yes | Min 3 characters | Event title |
| description | string | No | - | Event description |
| start_datetime | datetime | Yes | Future date | Event start date and time |
| end_datetime | datetime | Yes | After start_datetime | Event end date and time |
| location | string | Yes | - | Event location/venue |
| capacity | integer | Yes | Min 1 | Maximum number of attendees |

**Success Response (201 Created):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Tech Conference 2025",
  "description": "Annual technology conference featuring latest innovations in AI, Cloud, and Web3",
  "start_datetime": "2025-06-15T09:00:00Z",
  "end_datetime": "2025-06-15T18:00:00Z",
  "location": "Convention Center, New York",
  "capacity": 500,
  "status": "draft",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T10:30:00Z"
}
```

**Error Responses:**

*400 Bad Request - Validation Error:*
```json
{
  "error": "validation_failed",
  "message": "End time must be after start time"
}
```

*401 Unauthorized:*
```json
{
  "error": "unauthorized",
  "message": "Authentication required"
}
```

---

### Update Event

Update an existing event. Only the event organizer can update their events.

**Endpoint:** `PUT /events/:id`

**Authentication:** Required (JWT token, organizer only)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Event ID |

**Request Body:**
All fields are optional. Only include fields you want to update.

```json
{
  "title": "Updated Event Title",
  "description": "Updated description",
  "start_datetime": "2025-06-15T10:00:00Z",
  "end_datetime": "2025-06-15T19:00:00Z",
  "location": "New Location",
  "capacity": 600
}
```

**Success Response (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Updated Event Title",
  "description": "Updated description",
  "start_datetime": "2025-06-15T10:00:00Z",
  "end_datetime": "2025-06-15T19:00:00Z",
  "location": "New Location",
  "capacity": 600,
  "status": "draft",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T11:00:00Z"
}
```

**Error Responses:**

*403 Forbidden - Not Event Organizer:*
```json
{
  "error": "forbidden",
  "message": "Only the event organizer can update this event"
}
```

*404 Not Found:*
```json
{
  "error": "event_not_found",
  "message": "Event with the specified ID does not exist"
}
```

---

### Delete Event

Delete an event. Only the event organizer can delete their events.

**Endpoint:** `DELETE /events/:id`

**Authentication:** Required (JWT token, organizer only)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Event ID |

**Success Response (200 OK):**
```json
{
  "message": "Event deleted successfully"
}
```

**Error Responses:**

*403 Forbidden:*
```json
{
  "error": "forbidden",
  "message": "Only the event organizer can delete this event"
}
```

*404 Not Found:*
```json
{
  "error": "event_not_found",
  "message": "Event with the specified ID does not exist"
}
```

---

### Publish Event

Publish a draft event to make it visible and available for registration.

**Endpoint:** `POST /events/:id/publish`

**Authentication:** Required (JWT token, organizer only)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Event ID |

**Success Response (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Tech Conference 2025",
  "status": "published",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T11:30:00Z"
}
```

**Error Responses:**

*400 Bad Request - Already Published:*
```json
{
  "error": "invalid_status",
  "message": "Event is already published"
}
```

*403 Forbidden:*
```json
{
  "error": "forbidden",
  "message": "Only the event organizer can publish this event"
}
```

---

### Cancel Event

Cancel a published event.

**Endpoint:** `POST /events/:id/cancel`

**Authentication:** Required (JWT token, organizer only)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Event ID |

**Success Response (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Tech Conference 2025",
  "status": "cancelled",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Error Responses:**

*400 Bad Request - Already Cancelled:*
```json
{
  "error": "invalid_status",
  "message": "Event is already cancelled"
}
```

---

## Registration Endpoints

### Register for Event

Register the authenticated user for an event.

**Endpoint:** `POST /registrations`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "event_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Success Response (201 Created):**
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_id": "660e8400-e29b-41d4-a716-446655440001",
  "status": "confirmed",
  "registered_at": "2025-12-17T10:30:00Z"
}
```

**Error Responses:**

*400 Bad Request - Event Full:*
```json
{
  "error": "event_full",
  "message": "Event has reached maximum capacity"
}
```

*400 Bad Request - Already Registered:*
```json
{
  "error": "already_registered",
  "message": "User is already registered for this event"
}
```

*404 Not Found:*
```json
{
  "error": "event_not_found",
  "message": "Event with the specified ID does not exist"
}
```

---

### Get My Registrations

Get all registrations for the authenticated user.

**Endpoint:** `GET /registrations/my`

**Authentication:** Required (JWT token)

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "event_id": "660e8400-e29b-41d4-a716-446655440001",
      "event": {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "title": "Tech Conference 2025",
        "start_datetime": "2025-06-15T09:00:00Z",
        "location": "Convention Center, New York",
        "status": "published"
      },
      "status": "confirmed",
      "registered_at": "2025-12-17T10:30:00Z"
    }
  ]
}
```

---

### Cancel Registration

Cancel a registration for an event.

**Endpoint:** `DELETE /registrations/:id`

**Authentication:** Required (JWT token, user must own the registration)

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Registration ID |

**Success Response (200 OK):**
```json
{
  "message": "Registration cancelled successfully"
}
```

**Error Responses:**

*403 Forbidden:*
```json
{
  "error": "forbidden",
  "message": "You can only cancel your own registrations"
}
```

*404 Not Found:*
```json
{
  "error": "registration_not_found",
  "message": "Registration with the specified ID does not exist"
}
```

---

## User Endpoints

### Get Current User Profile

Get the profile of the authenticated user.

**Endpoint:** `GET /users/me`

**Authentication:** Required (JWT token)

**Success Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "created_at": "2025-12-17T10:30:00Z",
  "updated_at": "2025-12-17T10:30:00Z"
}
```

---

### Get My Events

Get all events created by the authenticated user.

**Endpoint:** `GET /users/me/events`

**Authentication:** Required (JWT token)

**Success Response (200 OK):**
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "organizer_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Tech Conference 2025",
      "description": "Annual technology conference",
      "start_datetime": "2025-06-15T09:00:00Z",
      "end_datetime": "2025-06-15T18:00:00Z",
      "location": "Convention Center, New York",
      "capacity": 500,
      "status": "published",
      "created_at": "2025-12-17T10:30:00Z",
      "updated_at": "2025-12-17T10:30:00Z"
    }
  ]
}
```

---

## Data Models

### User

```go
{
  "id": "UUID",              // Unique identifier
  "email": "string",         // Unique email address
  "name": "string",          // Full name
  "role": "string",          // User role: "user", "organizer", "admin"
  "created_at": "datetime",  // Account creation timestamp
  "updated_at": "datetime"   // Last update timestamp
}
```

### Event

```go
{
  "id": "UUID",                    // Unique identifier
  "organizer_id": "UUID",          // ID of user who created the event
  "organizer": User,               // Organizer details (optional, in some responses)
  "title": "string",               // Event title (min 3 chars)
  "description": "string",         // Event description (optional)
  "start_datetime": "datetime",    // Event start date and time
  "end_datetime": "datetime",      // Event end date and time
  "location": "string",            // Event venue/location
  "capacity": "integer",           // Maximum number of attendees (min 1)
  "status": "string",              // Event status: "draft", "published", "cancelled"
  "created_at": "datetime",        // Creation timestamp
  "updated_at": "datetime"         // Last update timestamp
}
```

### Registration

```go
{
  "id": "UUID",                  // Unique identifier
  "user_id": "UUID",             // ID of registered user
  "event_id": "UUID",            // ID of event
  "event": Event,                // Event details (optional, in some responses)
  "status": "string",            // Registration status: "confirmed", "cancelled"
  "registered_at": "datetime"    // Registration timestamp
}
```

### Pagination Response

```go
{
  "data": [],                    // Array of items
  "pagination": {
    "page": "integer",           // Current page number
    "page_size": "integer",      // Items per page
    "total": "integer",          // Total number of items
    "total_pages": "integer"     // Total number of pages
  }
}
```

---

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "specific_field",
    "reason": "additional_context"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `validation_failed` | 400 | Input validation error |
| `invalid_credentials` | 401 | Invalid email or password |
| `unauthorized` | 401 | Missing or invalid JWT token |
| `forbidden` | 403 | Insufficient permissions |
| `event_not_found` | 404 | Event doesn't exist |
| `user_not_found` | 404 | User doesn't exist |
| `registration_not_found` | 404 | Registration doesn't exist |
| `email_already_exists` | 409 | Email is already registered |
| `already_registered` | 409 | User already registered for event |
| `event_full` | 400 | Event reached capacity |
| `invalid_status` | 400 | Invalid status transition |
| `internal_error` | 500 | Server error |

---

## Best Practices

### Security
- Always use HTTPS in production
- Store JWT tokens securely (e.g., httpOnly cookies)
- Never expose sensitive information in error messages
- Implement rate limiting (planned for v2.0)

### Performance
- Use pagination for large datasets
- Cache frequently accessed data
- Use appropriate filters to reduce data transfer

### Integration
- Handle token expiration gracefully
- Implement retry logic for failed requests
- Log all API interactions for debugging

---

**Need help?** Open an issue on [GitHub](https://github.com/kimashii-dan/event-hub/issues)
