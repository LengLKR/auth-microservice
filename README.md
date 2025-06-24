[![Go](https://img.shields.io/badge/go-1.20+-00ADD8?logo=go)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()
# Auth Microservice

## Overview

Authentication & User Management microservice built with:

- **gRPC**
- **Golang**
- **MongoDB**
- **JWT**

---

## Prerequisites

- Go 1.20+
- Docker (for MongoDB)
- Protobuf compiler (`protoc`)
- `protoc-gen-go` & `protoc-gen-go-grpc` plugins
- (Optional) `grpcurl` or Postman GRPC client

---

## Setup

1. **Clone repository**

   ```bash
   git clone https://github.com/LengLKR/auth-microservice.git
   cd auth-microservice
   ```

2. **Environment variables**  
   Create a `.env` file in project root:

   ```dotenv
   MONGO_URI=mongodb://localhost:27017
   MONGO_DB=authdb
   JWT_SECRET=a-string-secret-at-least-256-bits-long
   ```

3. **Run MongoDB**

   ```bash
   docker run --name auth-mongo -d -p 27017:27017 mongo:latest
   ```

4. **Install Protobuf plugins**

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

5. **Generate Go code from proto**

   ```bash
   protoc \
     --proto_path=proto \
     --go_out=internal/transport/proto --go_opt=paths=source_relative \
     --go-grpc_out=internal/transport/proto --go-grpc_opt=paths=source_relative \
     proto/auth.proto
   ```

6. **Build & Run the server**

   ```bash
   go build -o auth-server ./cmd/auth-server
   ./auth-server
   ```

---

## Error Response Format

All error responses follow the gRPC standard error shape:

```json
{
  "code": <gRPC status code>,
  "message": "<error message>",
  "details": []
}
```

- `code`: integer gRPC code (e.g., `3` = INVALID_ARGUMENT, `5` = NOT_FOUND, `6` = ALREADY_EXISTS, `7` = PERMISSION_DENIED)
- `message`: descriptive error message

### Example: `Register` error (already exists)
```bash
grpcurl -plaintext -d '{"email":"alice@example.com","password":"P@ssw0rd!"}' localhost:50051 auth.AuthService/Register

Status code: 2 UNKNOWN
Time: 31 ms
details: write exception: write errors: [E11000 duplicate key error collection: authdb.users index: email_1 dup key: { email: "alice@example.com" }]
```

See full spec in [docs/API.md](doce/API.md).

---

## API Usage (examples with `grpcurl` / Postman GRPC)

### 1. Register

```bash
grpcurl -plaintext -d '{"email":"alice@example.com","password":"P@ssw0rd!"}' localhost:50051 auth.AuthService/Register
```

### 2. Login

```bash
grpcurl -plaintext -d '{"email":"alice@example.com","password":"P@ssw0rd!"}' localhost:50051 auth.AuthService/Login
```

📥 Response: `{ "token":"<JWT_TOKEN>" }`

### 3. List Users

```bash
grpcurl -plaintext \
  -H 'authorization: Bearer <JWT_TOKEN>' \
  -d '{"filterName":"","filterEmail":"","page":1,"size":10}' \
  localhost:50051 auth.AuthService/ListUsers
```

### 4. Get Profile

```bash
grpcurl -plaintext \
  -H 'authorization: Bearer <JWT_TOKEN>' \
  -d '{"id":"<USER_ID>"}' \
  localhost:50051 auth.AuthService/GetProfile
```

### 5. Update Profile

```bash
grpcurl -plaintext \
  -H 'authorization: Bearer <JWT_TOKEN>' \
  -d '{"id":"<USER_ID>","email":"new@example.com"}' \
  localhost:50051 auth.AuthService/UpdateProfile
```

### 6. Delete Profile

```bash
grpcurl -plaintext \
  -H 'authorization: Bearer <JWT_TOKEN>' \
  -d '{"id":"<USER_ID>"}' \
  localhost:50051 auth.AuthService/DeleteProfile
```

### 7. Request Password Reset

```bash
grpcurl -plaintext -d '{"email":"alice@example.com"}' localhost:50051 auth.AuthService/RequestPasswordReset
```

🔔 Token logged in server output.

### 8. Reset Password

```bash
grpcurl -plaintext \
  -d '{"token":"<RESET_TOKEN>","newPassword":"N3wP@ss!"}' \
  localhost:50051 auth.AuthService/ResetPassword
```

---

## API Reference

See detailed spec in [docs/API.md](docs/API.md).

---

## Architectural Overview

- **Service Layer** (`internal/service`): Business logic for authentication, profile management, and password reset.
- **Repository Layer** (`internal/repository`): MongoDB data access for users, tokens, and reset tokens. Implements interfaces for swapping DBs.
- **Transport Layer** (`internal/transport`): gRPC server implementation. Translates gRPC requests to service calls.
- **Proto Definitions** (`proto/auth.proto`): Defines RPCs and message contracts.
- **Config** (`config`): Loads environment variables and initializes MongoDB client.

**Design Decisions & Trade-offs**:

- **Soft Delete**: Mark users as deleted for data retention without hard removal.
- **In-Memory Rate Limiting**: Simple mutex+map for login attempts—suitable for single-instance; replace with Redis for multi-instance.
- **JWT Blacklist**: Stored in MongoDB with TTL for logout token invalidation.
- **Password Reset Flow**: Stateless tokens stored in DB with TTL index.

---


> **Next Steps**: Add unit & integration tests, improve error handling, and containerize service for production deployment.

