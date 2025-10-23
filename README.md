# Gin Boilerplate with DDD Architecture

A production-ready REST API boilerplate using Gin Framework with Domain-Driven Design (DDD) architecture, authentication system, multi-role authorization, S3-compatible file storage, and document management.

## 🚀 Features

- **Domain-Driven Design (DDD)**: Clean architecture with separated concerns
- **Authentication**: Email/password and Google OAuth 2.0
- **Authorization**: Role-based access control (User & Admin roles)
- **JWT Tokens**: Access and refresh token implementation
- **Database**: PostgreSQL with GORM ORM and auto-migration
- **File Storage**: S3-compatible storage (AWS S3, MinIO, DigitalOcean Spaces, etc.)
- **Document Management**: Complete CRUD operations with file upload/download
- **Security**: Password hashing with bcrypt, CORS, request validation, file type/size restrictions
- **Logging**: Structured logging with logrus
- **Middleware**: Authentication, role-based authorization, CORS, logging
- **Configuration**: Environment-based configuration with godotenv
- **API Documentation**: Complete Swagger/OpenAPI documentation
- **Docker Ready**: Dockerfile included
- **Testing**: Test structure with unit and integration tests

## 📁 Project Structure

```
gin-boilerplate/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Business Logic Layer
│   │   ├── entity/                 # Domain entities
│   │   │   ├── user.go
│   │   │   ├── token.go
│   │   │   └── document.go
│   │   ├── repository/             # Repository interfaces
│   │   │   ├── user_repository.go
│   │   │   ├── token_repository.go
│   │   │   └── document_repository.go
│   │   ├── errors.go               # Domain errors
│   │   └── service/                # Domain services
│   │       ├── password_service.go
│   │       └── token_service.go
│   ├── application/                # Application Layer
│   │   ├── dto/                    # Data Transfer Objects
│   │   │   ├── auth_dto.go
│   │   │   └── document_dto.go
│   │   └── usecase/                # Use cases
│   │       ├── register_usecase.go
│   │       ├── login_usecase.go
│   │       ├── google_auth_usecase.go
│   │       ├── refresh_token_usecase.go
│   │       ├── logout_usecase.go
│   │       ├── user_usecase.go
│   │       └── document_usecase.go
│   ├── infrastructure/             # Infrastructure Layer
│   │   ├── config/                 # Configuration
│   │   │   ├── config.go
│   │   │   └── google_oauth.go
│   │   ├── persistence/
│   │   │   └── postgres/
│   │   │       ├── database.go
│   │   │       ├── user_repository.go
│   │   │       ├── token_repository.go
│   │   │       └── document_repository.go
│   │   └── storage/
│   │       └── s3_client.go       # S3-compatible storage client
│   └── interfaces/                 # Presentation Layer
│       └── http/
│           ├── handler/            # HTTP handlers
│           │   ├── auth_handler.go
│           │   ├── user_handler.go
│           │   └── document_handler.go
│           ├── middleware/         # Middlewares
│           │   ├── auth_middleware.go
│           │   ├── role_middleware.go
│           │   ├── cors_middleware.go
│           │   └── logger_middleware.go
│           └── router/             # Route definitions
│               └── router.go
├── interfaces/dto/                 # Request/Response DTOs
├── pkg/                            # Public utilities
├── docs/                           # API documentation (generated)
├── .env.example                    # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🛠️ Tech Stack

- **Go 1.21+**
- **Gin Framework** - HTTP web framework
- **GORM** - PostgreSQL ORM with auto-migration
- **AWS SDK v2** - S3-compatible storage integration
- **Redis** - Caching and rate limiting
- **JWT** - JSON Web Tokens for authentication
- **Google OAuth 2.0** - Third-party authentication
- **Swagger/OpenAPI** - API documentation generation
- **bcrypt** - Password hashing
- **logrus** - Structured logging
- **gin-contrib/cors** - CORS middleware
- **godotenv** - Environment variable management

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Google OAuth credentials (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd gin-boilerplate
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up database**
   ```bash
   # Create database
   createdb gin_boilerplate

   # The application will auto-migrate tables on startup
   ```

5. **Run the application**
   ```bash
   make run
   ```

The API will be available at `http://localhost:8080`

## 📖 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/register` | Register new user | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/refresh` | Refresh access token | No |
| POST | `/api/v1/auth/logout` | Logout (current device) | Yes |
| POST | `/api/v1/auth/logout-all` | Logout (all devices) | Yes |
| GET | `/api/v1/auth/google` | Initiate Google OAuth | No |
| GET | `/api/v1/auth/google/callback` | Google OAuth callback | No |

### User Endpoints

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| GET | `/api/v1/users/me` | Get current user profile | Yes | User/Admin |
| PUT | `/api/v1/users/me` | Update current user profile | Yes | User/Admin |
| GET | `/api/v1/users` | List all users (paginated) | Yes | Admin |
| GET | `/api/v1/users/:id` | Get user by ID | Yes | Admin |
| DELETE | `/api/v1/users/:id` | Delete user | Yes | Admin |
| POST | `/api/v1/users/:id/promote` | Promote user to admin | Yes | Admin |
| POST | `/api/v1/users/:id/demote` | Demote admin to user | Yes | Admin |

### Avatar Endpoints

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/users/avatar` | Upload avatar image | Yes | User/Admin |
| DELETE | `/api/v1/users/avatar` | Remove current avatar | Yes | User/Admin |
| GET | `/api/v1/users/avatar/:id` | Serve user avatar image | No | Public |

### Document Endpoints

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/documents/upload` | Upload document with file | Yes | User/Admin |
| GET | `/api/v1/documents` | List user documents (paginated) | Yes | User/Admin |
| GET | `/api/v1/documents/:id` | Get document by ID | Yes | User/Admin |
| PUT | `/api/v1/documents/:id` | Update document metadata | Yes | User/Admin |
| DELETE | `/api/v1/documents/:id` | Delete document and file | Yes | User/Admin |
| GET | `/api/v1/documents/:id/download` | Get presigned download URL | Yes | User/Admin |

### API Examples

#### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Get Current User
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access-token>"
```

#### Upload Document
```bash
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -H "Authorization: Bearer <access-token>" \
  -F "title=My Document" \
  -F "description=A sample document" \
  -F "file=@/path/to/document.pdf"
```

#### Get Presigned Download URL
```bash
curl -X GET http://localhost:8080/api/v1/documents/:id/download \
  -H "Authorization: Bearer <access-token>"
```

#### Upload Avatar
```bash
curl -X POST http://localhost:8080/api/v1/users/avatar \
  -H "Authorization: Bearer <access-token>" \
  -F "avatar=@/path/to/avatar.jpg"
```

#### Remove Avatar
```bash
curl -X DELETE http://localhost:8080/api/v1/users/avatar \
  -H "Authorization: Bearer <access-token>"
```

#### Access User Avatar
```bash
# For any user avatar
curl -X GET http://localhost:8080/api/v1/users/avatar/:user_id

# This will redirect to:
# - Google avatar URL (if from OAuth)
# - S3 presigned URL (if uploaded)
```

#### Access API Documentation
Visit `http://localhost:8080/swagger/index.html` for interactive API documentation.

## ⚙️ Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure the following:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gin_boilerplate
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# S3-Compatible Storage Configuration
S3_ENDPOINT=https://s3.amazonaws.com  # For AWS S3. For MinIO: http://localhost:9000
S3_ACCESS_KEY_ID=your-s3-access-key
S3_SECRET_ACCESS_KEY=your-s3-secret-key
S3_REGION=us-east-1
S3_BUCKET=your-bucket-name
S3_USE_SSL=true

# Server Configuration
SERVER_PORT=8080
SERVER_ENV=development
```

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google+ API
4. Create OAuth 2.0 client ID
5. Add authorized redirect URI: `http://localhost:8080/api/v1/auth/google/callback`
6. Copy Client ID and Client Secret to your `.env` file

### S3-Compatible Storage Setup

#### AWS S3
1. Create an AWS account and S3 bucket
2. Create IAM user with programmatic access
3. Attach policy with S3 permissions (GetObject, PutObject, DeleteObject)
4. Copy credentials to your `.env` file

#### MinIO (Local Development)
1. Install MinIO: `docker run -p 9000:9000 -p 9001:9001 minio/minio server /data`
2. Create bucket through MinIO console
3. Configure with endpoint: `http://localhost:9000`

#### DigitalOcean Spaces
1. Create a Space in your DigitalOcean account
2. Generate API keys
3. Configure with your Space endpoint and credentials

#### File Upload Restrictions
- **Maximum file size**: 10MB (documents), 2MB (avatars)
- **Allowed file types**:
  - Documents: Images (JPEG, PNG, GIF), PDF, Text, Word documents
  - Avatars: Images only (JPEG, PNG, GIF, WebP)
- **User isolation**: Users can only access their own files
- **Automatic cleanup**: Files are deleted from storage when documents/avatars are deleted

## 🧪 Development

### Make Commands

```bash
make help          # Show all available commands
make run           # Run the application
make dev           # Run with hot reload (requires air)
make test          # Run tests
make test-coverage # Run tests with coverage
make build         # Build the application
make clean         # Clean build artifacts
make lint          # Run linter
make fmt           # Format code
make tidy          # Clean up dependencies
make docs          # Generate Swagger docs
make swagger       # Alias for docs command
make redis-up      # Start Redis container (for development)
make docker-build  # Build Docker image
make docker-run    # Run Docker container
make setup         # Quick setup for development
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/application/usecase -v
```

### Hot Reload

For development with hot reload:

1. Install air:
   ```bash
   go install github.com/cosmtrek/air@latest
   ```

2. Run with hot reload:
   ```bash
   make dev
   ```

## 🐳 Docker

### Build Docker Image
```bash
make docker-build
```

### Run with Docker
```bash
# Ensure .env file exists
make docker-run
```

### Docker Compose
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
    depends_on:
      - postgres
    env_file:
      - .env

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gin_boilerplate
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 🔒 Security Features

- **Password Hashing**: Uses bcrypt with configurable cost
- **JWT Security**: Short-lived access tokens (15m) and refresh tokens (7d)
- **Input Validation**: Request validation using struct tags
- **CORS**: Configurable CORS middleware
- **Role-Based Access Control**: Middleware for role verification
- **File Security**: File type validation, size limits, and user isolation
- **Storage Security**: Presigned URLs with expiration for secure file access
- **Rate Limiting**: IP-based and user-based rate limiting with Redis
- **Caching**: Redis integration for performance optimization
- **SQL Injection Prevention**: GORM ORM provides protection
- **HTTPS Ready**: Production deployment should use HTTPS

## 📝 Architecture Patterns

### Domain-Driven Design (DDD)

The project follows DDD principles with clear separation of concerns:

1. **Domain Layer**: Contains business logic, entities, and repository interfaces
2. **Application Layer**: Contains use cases and orchestration logic
3. **Infrastructure Layer**: Contains external dependencies like database and external services
4. **Interface Layer**: Contains HTTP handlers and middleware

### Dependency Injection

Dependencies are injected in `main.go` following the dependency inversion principle:
- Domain layer doesn't depend on external libraries
- Infrastructure implements domain interfaces
- Application orchestrates between layers

## 🚀 Deployment

### Production Build
```bash
# Build for production
make prod-build

# Run production binary
make prod-run
```

### Environment Setup

1. **Database**: Set up PostgreSQL database
2. **Environment**: Set production environment variables
3. **SSL**: Configure SSL certificates
4. **Reverse Proxy**: Use Nginx or similar for load balancing
5. **Monitoring**: Set up logging and monitoring

### Docker Production
```bash
# Build and run with Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Common Issues

1. **Database Connection Error**
   - Check PostgreSQL is running
   - Verify database credentials in `.env`
   - Ensure database exists

2. **Google OAuth Error**
   - Verify Client ID and Secret
   - Check redirect URI configuration
   - Ensure Google+ API is enabled

3. **JWT Token Error**
   - Check JWT_SECRET is set
   - Verify token hasn't expired
   - Ensure proper Authorization header format

4. **Import Errors**
   - Run `go mod tidy`
   - Check Go version compatibility
   - Verify module path

### Getting Help

- Check the [Issues](../../issues) page
- Create a new issue with detailed description
- Check existing documentation and examples

## 📚 Additional Resources

- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [AWS SDK v2 Documentation](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2)
- [Swagger/OpenAPI Documentation](https://swagger.io/specification/)
- [JWT Documentation](https://jwt.io/)
- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [S3-Compatible Storage](https://docs.aws.amazon.com/s3/)
- [Domain-Driven Design](https://en.wikipedia.org/wiki/Domain-driven_design)

---

Made with ❤️ using Go and Gin Framework