# API Documentation

## Base URL
- Development: `http://localhost:8080`
- WebSocket: `ws://localhost:8080/ws`

## Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <access_token>
```

---

## Authentication Endpoints

### POST /api/auth/register
Register a new user.

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": "uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

---

### POST /api/auth/login
Login with credentials.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "securepassword123"
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": "uuid",
    "username": "johndoe",
    "email": "john@example.com"
  },
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

---

### POST /api/auth/refresh
Refresh access token using refresh token.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc..."
}
```

---

### POST /api/auth/logout
Logout and invalidate refresh token.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

## Conversation Endpoints

### GET /api/conversations
List all conversations for the authenticated user.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "conversations": [
    {
      "id": "uuid",
      "participants": ["user_id_1", "user_id_2"],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /api/conversations
Create or get existing conversation with another user.

**Headers:** `Authorization: Bearer <access_token>`

**Request Body:**
```json
{
  "user_id": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "conversation": {
    "id": "uuid",
    "participants": ["current_user_id", "other_user_id"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### GET /api/conversations/:id
Get conversation details.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "conversation": {
    "id": "uuid",
    "participants": ["user_id_1", "user_id_2"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### GET /api/conversations/:id/messages
Get message history for a conversation with cursor-based pagination.

**Headers:** `Authorization: Bearer <access_token>`

**Query Parameters:**
- `cursor` (optional): Timestamp for pagination
- `limit` (optional, default: 50): Number of messages to return

**Response:** `200 OK`
```json
{
  "messages": [
    {
      "id": "uuid",
      "sender_id": "uuid",
      "conversation_id": "uuid",
      "content": "Hello!",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z"
}
```

---

### POST /api/conversations/:id/messages
Send a message in a conversation.

**Headers:** `Authorization: Bearer <access_token>`

**Request Body:**
```json
{
  "content": "Hello, how are you?"
}
```

**Response:** `201 Created`
```json
{
  "message": {
    "id": "uuid",
    "sender_id": "uuid",
    "conversation_id": "uuid",
    "content": "Hello, how are you?",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Group Endpoints

### GET /api/groups
List all groups for the authenticated user.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "groups": [
    {
      "id": "uuid",
      "name": "Team Chat",
      "created_by": "uuid",
      "members": ["user_id_1", "user_id_2", "user_id_3"],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /api/groups
Create a new group.

**Headers:** `Authorization: Bearer <access_token>`

**Request Body:**
```json
{
  "name": "Team Chat"
}
```

**Response:** `201 Created`
```json
{
  "group": {
    "id": "uuid",
    "name": "Team Chat",
    "created_by": "uuid",
    "members": ["creator_user_id"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### GET /api/groups/:id
Get group details.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "group": {
    "id": "uuid",
    "name": "Team Chat",
    "created_by": "uuid",
    "members": ["user_id_1", "user_id_2"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### POST /api/groups/:id/members
Add a member to the group.

**Headers:** `Authorization: Bearer <access_token>`

**Request Body:**
```json
{
  "user_id": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "message": "Member added successfully"
}
```

---

### DELETE /api/groups/:id/members/:userId
Remove a member from the group.

**Headers:** `Authorization: Bearer <access_token>`

**Response:** `200 OK`
```json
{
  "message": "Member removed successfully"
}
```

---

### GET /api/groups/:id/messages
Get message history for a group with cursor-based pagination.

**Headers:** `Authorization: Bearer <access_token>`

**Query Parameters:**
- `cursor` (optional): Timestamp for pagination
- `limit` (optional, default: 50): Number of messages to return

**Response:** `200 OK`
```json
{
  "messages": [
    {
      "id": "uuid",
      "sender_id": "uuid",
      "group_id": "uuid",
      "content": "Hello team!",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z"
}
```

---

### POST /api/groups/:id/messages
Send a message in a group.

**Headers:** `Authorization: Bearer <access_token>`

**Request Body:**
```json
{
  "content": "Hello team!"
}
```

**Response:** `201 Created`
```json
{
  "message": {
    "id": "uuid",
    "sender_id": "uuid",
    "group_id": "uuid",
    "content": "Hello team!",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## WebSocket Protocol

### Connection

Connect to WebSocket with JWT token:
```
ws://localhost:8080/ws?token=<access_token>
```

### Message Types

**Incoming Messages:**

1. **New Message**
```json
{
  "type": "message",
  "data": {
    "id": "uuid",
    "sender_id": "uuid",
    "conversation_id": "uuid",
    "group_id": null,
    "content": "Hello!",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

2. **Error**
```json
{
  "type": "error",
  "message": "Error description"
}
```

**Outgoing Messages:**

1. **Send Message**
```json
{
  "type": "message",
  "conversation_id": "uuid",
  "group_id": null,
  "content": "Hello!"
}
```

---

## Error Responses

All endpoints may return the following error responses:

**400 Bad Request**
```json
{
  "error": "Invalid request body"
}
```

**401 Unauthorized**
```json
{
  "error": "Invalid or expired token"
}
```

**403 Forbidden**
```json
{
  "error": "You don't have permission to access this resource"
}
```

**404 Not Found**
```json
{
  "error": "Resource not found"
}
```

**429 Too Many Requests**
```json
{
  "error": "Rate limit exceeded"
}
```

**500 Internal Server Error**
```json
{
  "error": "Internal server error"
}
```
