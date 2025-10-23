# Project Prompt: Gin DDD Boilerplate with Auth & Multi-Role

## Project Overview
Buatkan REST API boilerplate menggunakan Gin Framework dengan DDD architecture. Project ini mencakup authentication system (email/password & Google OAuth), authorization dengan multi-role (User & Admin), dan menggunakan GORM + PostgreSQL sebagai database.

## Core Requirements

### 1. Authentication System
- **Email/Password Authentication**
  - Register dengan email, password, dan name
  - Login dengan email & password
  - Password hashing menggunakan bcrypt
  - JWT token generation (access token & refresh token)
  - Refresh token endpoint

- **Google OAuth Integration**
  - Login/Register via Google OAuth 2.0
  - Auto-create user jika belum exist
  - Merge dengan existing account jika email sama
  - Support Google profile data (name, email, avatar)

### 2. Authorization & Multi-Role
- **Roles:**
  - `USER` (default role untuk registrasi)
  - `ADMIN` (role khusus dengan elevated permissions)

- **Role-based Access Control (RBAC):**
  - Middleware untuk check role
  - Endpoint yang hanya bisa diakses ADMIN
  - Endpoint yang bisa diakses semua authenticated users

### 3. User Management
- **User Entity Fields:**
  - ID (UUID)
  - Email (unique, required)
  - Password (nullable - untuk OAuth users)
  - Name (required)
  - Role (enum: USER, ADMIN)
  - Provider (enum: LOCAL, GOOGLE)
  - ProviderID (nullable - untuk OAuth)
  - Avatar (nullable - URL)
  - EmailVerified (boolean)
  - CreatedAt, UpdatedAt

- **User Endpoints:**
  - GET /api/v1/users/me - Get current user profile
  - PUT /api/v1/users/me - Update current user profile
  - GET /api/v1/users - List all users (ADMIN only)
  - GET /api/v1/users/:id - Get user by ID (ADMIN only)
  - DELETE /api/v1/users/:id - Delete user (ADMIN only)

### 4. Auth Endpoints
```
POST   /api/v1/auth/register          - Register dengan email/password
POST   /api/v1/auth/login             - Login dengan email/password
POST   /api/v1/auth/refresh           - Refresh access token
POST   /api/v1/auth/logout            - Logout (invalidate token)
GET    /api/v1/auth/google            - Redirect ke Google OAuth
GET    /api/v1/auth/google/callback   - Google OAuth callback
```

## Technical Stack
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL
- **Auth:** JWT (golang-jwt/jwt)
- **OAuth:** Google OAuth 2.0 (golang.org/x/oauth2)
- **Password:** bcrypt
- **Validation:** go-playground/validator
- **Config:** godotenv / viper
- **Migration:** GORM Auto Migrate

## DDD Structure Implementation

### Domain Layer
1. **Entities:**
   - User entity dengan validation methods
   - Token entity (untuk refresh token tracking)

2. **Repository Interfaces:**
   - UserRepository (CRUD + FindByEmail, FindByProviderID)
   - TokenRepository (Store & Validate refresh tokens)

3. **Domain Services (optional):**
   - PasswordService (hash, verify)
   - TokenService (generate, verify JWT)

### Application Layer
1. **Use Cases:**
   - RegisterUseCase
   - LoginUseCase
   - GoogleAuthUseCase
   - RefreshTokenUseCase
   - GetUserProfileUseCase
   - UpdateUserProfileUseCase
   - ListUsersUseCase (admin)
   - DeleteUserUseCase (admin)

2. **DTOs:**
   - RegisterRequest, LoginRequest
   - GoogleCallbackRequest
   - UpdateProfileRequest
   - UserResponse, AuthResponse (with tokens)

### Infrastructure Layer
1. **Persistence:**
   - PostgreSQL UserRepository implementation
   - PostgreSQL TokenRepository implementation
   - Database connection setup
   - GORM models (if different from entities)

2. **Config:**
   - Database config (DSN, pool settings)
   - JWT config (secret, expiry)
   - Google OAuth config (client ID, secret, redirect URL)
   - Server config (port, environment)

3. **External Services:**
   - Google OAuth client wrapper

### Interface Layer
1. **Handlers:**
   - AuthHandler (register, login, google callback, refresh, logout)
   - UserHandler (profile, list, delete)

2. **Middlewares:**
   - AuthMiddleware (verify JWT)
   - RoleMiddleware (check user role)
   - CORSMiddleware
   - LoggerMiddleware

3. **Router:**
   - Public routes (register, login, google auth)
   - Protected routes (require auth)
   - Admin routes (require admin role)

## Configuration (.env example)
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gin_boilerplate
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-super-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Server
SERVER_PORT=8080
SERVER_ENV=development
```

## Security Requirements
- Password minimum 8 karakter
- JWT tokens dengan proper expiry
- Refresh token stored di database (dapat di-revoke)
- CORS configuration
- Rate limiting (optional, tapi recommended)
- Input validation di semua endpoints
- Secure password hashing (bcrypt cost 10+)

## Error Handling
- Custom error types di domain layer
- Error wrapping dengan context
- Proper HTTP status codes
- Consistent error response format:
  ```json
  {
    "error": {
      "code": "INVALID_CREDENTIALS",
      "message": "Email or password is incorrect"
    }
  }
  ```

## Success Response Format
```json
{
  "data": {
    "user": {...},
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

## Database Migration
- Auto migration menggunakan GORM pada startup
- Seeder untuk admin user pertama (optional)

## Testing Considerations
- Unit tests untuk use cases (dengan mock repositories)
- Integration tests untuk repositories
- API tests untuk handlers (optional)

## Additional Features (Nice to Have)
- Email verification flow
- Password reset flow
- Logout from all devices
- User activity logging
- Profile picture upload
- Account deletion with soft delete

## Deliverables
1. Complete project structure sesuai DDD pattern
2. All endpoints working dengan proper authentication & authorization
3. Database migrations
4. README.md dengan setup instructions
5. .env.example file
6. Postman collection / API documentation (optional)

## Notes
- Gunakan UUID untuk user ID (bukan auto-increment)
- Implement proper logging di setiap layer
- Handle concurrent requests dengan proper transaction
- Validate semua input dari user
- Return appropriate HTTP status codes
- Jangan expose sensitive data di error messages
- Follow Go best practices & naming conventions

---

**Start dari domain layer (entities & repository interfaces), lalu ke application layer (use cases), infrastructure (implementations), dan terakhir interface layer (handlers & routes).**
