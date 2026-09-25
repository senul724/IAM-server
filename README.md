# IAM & Auth Demo Server

A lightweight Identity & Access Management (IAM) and authentication demo server built with Go, Gin, GORM, PostgreSQL, Upstash Redis, and Resend.

## Quick Start

### 1. Configure `.env`
Create a `.env` file in the root directory:

```env
DATABASE_URL=""

# Upstash Redis
UPSTASH_REDIS_URL=""

# Resend Email Service
RESEND_API_KEY=""
RESEND_FROM_EMAIL=""

# JWT Secrets
ACCESS_SECRET=""
REFRESH_SECRET=""
SESSION_SECRET=""
```

### 2. Run
```bash
go run cmd/server/main.go
```
The server runs on `http://localhost:3030`. Database tables are automatically migrated on startup.

---

## Core Components

### 1. Models (`internal/models/`)
- **Customer**: User entity with `Name`, `Email` (unique), nullable `Password` (supports OAuth/passwordless), and `Notes`.
- **Note**: Belongs to a Customer (`CustomerID` foreign key), with `Pinned` (bool) and `Content` (text).

### 2. Triple-Token Auth (`internal/utils/tokens/`)
- **Access Token** (30 min): Short-lived token for API authorization. Contains `sub` (User ID) and `email`.
- **Refresh Token** (15 days): Long-lived token used to issue new access tokens. Contains `sub`.
- **Session Token** (15 days): Client session token containing `sub`, `name`, and `email`.

### 3. Services (`internal/services/email.go`)
- `CheckEmail(email)`: Checks if a user exists and if they have a password set.
- `CheckPassword(email, password)`: Verifies plain password against the stored bcrypt hash.
- `GetLoginOTP(email)` (alias `SendLoginOTP`): Generates a 6-digit OTP, stores it in Redis (`login_{email}`), and emails it via Resend.
- `LoginWithOTP(c, email, otp)`: Verifies OTP, checks user existence, issues Refresh & Session tokens, and sets HTTP-only cookies.
- `RegisterWithOTP(c, name, email, otp)`: Verifies OTP, checks user is not already registered, creates the user, issues tokens, and sets HTTP-only cookies.
