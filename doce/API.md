# API Reference

This document provides detailed API specifications for the AuthService.

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

- `code`: integer gRPC code (e.g., `3` = INVALID\_ARGUMENT, `5` = NOT\_FOUND, `6` = ALREADY\_EXISTS, `7` = PERMISSION\_DENIED)
- `message`: descriptive error message

---

## AuthService.Register

**Request**

```proto
RegisterRequest {
  string email    = 1; // user email (required)
  string password = 2; // plaintext password (required)
}
```

**Response**

```proto
AuthResponse {
  string token = 1; // JWT token
}
```

**Errors**

- `INVALID_ARGUMENT` (3): missing or invalid email/password format
- `ALREADY_EXISTS` (6): email already registered

---

## AuthService.Login

**Request**

```proto
LoginRequest {
  string email    = 1;
  string password = 2;
}
```

**Response**

```proto
AuthResponse { string token = 1; }
```

**Errors**

- `INVALID_ARGUMENT` (3): missing credentials
- `UNAUTHENTICATED` (16) / `UNKNOWN` (2): invalid credentials
- `RESOURCE_EXHAUSTED` (8): too many failed attempts

---

## AuthService.Logout

**Request**

```proto
LogoutRequest { string token = 1; }
```

**Response**

```proto
Empty {}
```

**Errors**

- `INVALID_ARGUMENT` (3): missing token
- `UNAUTHENTICATED` (16): invalid or expired token

---

## AuthService.ListUsers

**Request**

```proto
ListUsersRequest {
  string filter_name  = 1; // optional regex
  string filter_email = 2; // optional regex
  int32  page         = 3; // 1-based
  int32  size         = 4; // page size
}
```

**Response**

```proto
ListUsersResponse {
  repeated User users     = 1;
  int32        total_count = 2;
}
```

**Errors**

- `UNAUTHENTICATED` (16): missing/invalid auth
- `PERMISSION_DENIED` (7): insufficient rights

---

## AuthService.GetProfile

**Request**

```proto
GetProfileRequest { string id = 1; }
```

**Response**

```proto
User { string id = 1; string email = 2; string createdAt = 3; }
```

**Errors**

- `UNAUTHENTICATED` (16): missing/invalid auth
- `PERMISSION_DENIED` (7): requesting other user's data
- `NOT_FOUND` (5): user not found or deleted

---

## AuthService.UpdateProfile

**Request**

```proto
UpdateProfileRequest { string id = 1; string email = 2; }
```

**Response**

```proto
User
```

**Errors**

- same as GetProfile
- `ALREADY_EXISTS` (6): email conflict

---

## AuthService.DeleteProfile

**Request**

```proto
DeleteProfileRequest { string id = 1; }
```

**Response**

```proto
Empty {}
```

**Errors**

- same as GetProfile

---

## AuthService.RequestPasswordReset

**Request**

```proto
PasswordResetRequest { string email = 1; }
```

**Response**

```proto
Empty {}
```

**Notes**

- Always returns `Empty` even if email not found (to avoid user enumeration).

---

## AuthService.ResetPassword

**Request**

```proto
ResetPasswordRequest { string token = 1; string newPassword = 2; }
```

**Response**

```proto
Empty {}
```

**Errors**

- `INVALID_ARGUMENT` (3): missing token or password
- `NOT_FOUND` (5): reset token not found or expired
- `INVALID_ARGUMENT` (3): password fails strength validation

---

