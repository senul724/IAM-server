# IAM & Auth Server

A lightweight Identity & Access Management (IAM) and authentication server built with **Go**, **Gin**, **GORM**, **PostgreSQL**, **Upstash Redis**, and **Resend**.

## How It Works

The server implements a **triple-token JWT architecture** for secure, stateless authentication with session tracking:

1. **User registers or logs in** via email/password, OTP, or Google OAuth 2.0.
2. The server issues three tokens:
   - **Access Token** (30 min) — short-lived, used in `Authorization: Bearer <token>` headers to access protected endpoints. Contains `sub` (user ID) and `email`. Its JTI is stored in Redis for revocation checks.
   - **Refresh Token** (15 days) — long-lived, used to obtain new access tokens without re-authenticating. Contains `sub`. Its JTI is stored in Redis, linked to the access token JTI.
   - **Session Token** (15 days) — client-facing session info containing `sub`, `name`, and `email`. Set as an HTTP-only cookie for web clients.
3. **Protected routes** use the `ProtectRoute` middleware, which validates the access token's signature *and* cross-checks its JTI against Redis to support instant revocation.
4. **Token refresh** rotates both the access and refresh tokens atomically — the old refresh JTI is revoked and a new pair is issued along with a new DB session record.
5. **Logout** revokes all token pairs (from cookies, `Authorization` header, and `X-Refresh-Token` header) and deletes the associated DB session.

### Session Management

Each login creates a `Session` record in PostgreSQL that tracks:
- Device details, IP address, and location
- First and last login timestamps
- The associated refresh token JTI

Users can view all their active sessions and selectively remove them.

---

## Quick Start

### 1. Configure `.env`
Create a `.env` file in the root directory:

```env
DATABASE_URL=""

# Upstash Redis
REDIS_URL=""

# Resend Email Service
RESEND_API_KEY=""
RESEND_FROM_EMAIL=""

# JWT Secrets
ACCESS_SECRET="your-access-secret"
REFRESH_SECRET="your-refresh-secret"
SESSION_SECRET="your-session-secret"

# Google OAuth (optional)
GOOGLE_CLIENT_ID="your-google-client-id"
GOOGLE_CLIENT_SECRET="your-google-client-secret"
GOOGLE_REDIRECT_URL="http://localhost:3030/api/auth/google/callback"
```

### 2. Run
```bash
go run cmd/server/main.go
```
The server runs on `http://localhost:3030`. Database tables are automatically migrated on startup.

### 3. API Documentation (Swagger)

Full interactive API documentation is available via Swagger UI at:

```
http://localhost:3030/swagger/index.html
```

To regenerate the Swagger docs after making changes to handler annotations:
```bash
cd cmd/server
swag init
```

---

## API Routes

### Auth (`/api/auth`)

| Method | Endpoint              | Auth     | Description                          |
|--------|-----------------------|----------|--------------------------------------|
| POST   | `/register`           | —        | Register a new user with OTP         |
| POST   | `/login`              | —        | Login with email & password          |
| POST   | `/login-otp`          | —        | Login with OTP                       |
| POST   | `/check-email`        | —        | Check if email exists & has password |
| POST   | `/send-login-otp`     | —        | Send a login OTP to email            |
| POST   | `/forgot-password`    | —        | Send a password reset OTP            |
| POST   | `/reset-password`     | —        | Reset password with OTP              |
| POST   | `/refresh`            | —        | Refresh access & refresh tokens      |
| POST   | `/logout`             | —        | Logout & revoke tokens               |
| GET    | `/sessions`           | Bearer   | Get all active sessions              |
| POST   | `/session/remove`     | Bearer   | Remove a specific session            |
| GET    | `/google/login`       | —        | Initiate Google OAuth flow           |
| GET    | `/google/callback`    | —        | Google OAuth callback                |

### Notes (`/api/notes`) — *All routes require Bearer token*

| Method | Endpoint | Description              |
|--------|----------|--------------------------|
| GET    | `/`      | List all notes for user  |
| GET    | `/:id`   | Get a specific note      |
| POST   | `/`      | Create a new note        |
| PUT    | `/:id`   | Update a note            |
| DELETE | `/:id`   | Delete a note            |

---

## Project Structure

```
IAM-server/
├── cmd/server/
│   ├── main.go              # Entry point, CORS, Swagger & route setup
│   └── docs/                # Auto-generated Swagger docs
├── internal/
│   ├── connections/         # PostgreSQL (GORM) & Upstash Redis clients
│   ├── handlers/            # HTTP handlers (auth, notes, sessions, logout)
│   │   └── oauth/           # Google OAuth handlers
│   ├── middleware/          # Auth middleware (JWT + Redis JTI validation)
│   ├── models/              # GORM models (Customer, Note, Session)
│   ├── routes/              # Route group definitions
│   ├── services/            # Business logic (email, OTP, tokens, sessions, notes)
│   └── utils/tokens/        # JWT token creation & verification
├── .env                     # Environment variables (not committed)
├── go.mod
└── go.sum
```

---

## Core Components

### Models (`internal/models/`)
- **Customer** — User entity with `Name`, `Email` (unique), nullable `Password` (supports OAuth/passwordless), and optional `PhotoURL`.
- **Note** — Belongs to a Customer, with `Title`, `Content`, and `Pinned` status.
- **Session** — Tracks active login sessions per user with `RefreshID`, `DeviceDetails`, `IPAddress`, `Location`, and login timestamps.

### Services (`internal/services/`)
- **email.go** — Email/password checks, OTP-based login & registration flows.
- **otp.go** — OTP generation, Redis storage, and verification.
- **password.go** — Forgot/reset password flows with OTP.
- **session.go** — Session CRUD (create, update, delete, list by user).
- **token_store.go** — Redis-backed access/refresh JTI storage and revocation.
- **note.go** — CRUD operations for user notes.

### Middleware (`internal/middleware/`)
- **ProtectRoute** — Extracts `Bearer` token, verifies JWT signature, cross-checks JTI in Redis, and injects claims into the Gin context.
