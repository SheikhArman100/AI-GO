# Go Authentication Service

A modern authentication service built with Go, featuring both traditional email/password authentication and OAuth2 social login capabilities. This service provides secure user management, profile handling, and token-based authentication using JWT, with a focus on security and scalability.

## Features

### Authentication
- Email/Password Registration and Login
- JWT Token-based Authentication
- Refresh Token Management
- Email Verification System
- Secure Password Hashing
- Session Management

### Social Authentication
- Google OAuth2 Integration
- Social Profile Management
- Automatic Account Linking
- Profile Picture Support

### User Management
- Profile Management
- User Details Storage
- Role-based Authorization
- Account Verification

### Security Features
- CORS Protection
- Input Validation
- Secure Cookie Management
- Rate Limiting
- SQL Injection Prevention (GORM)

## Tech Stack

- **Go** - Core programming language
- **Gin** - Web framework
- **GORM** - ORM library
- **MySQL** - Database
- **JWT** - Token-based authentication
- **OAuth2** - Social authentication
- **Docker** - Containerization

## Prerequisites

- Go 1.16 or higher
- MySQL
- Docker (optional)
- Google OAuth2 credentials (for social login)

## Environment Setup

Create a `.env` file in the root directory:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=your_password
DB_DATABASE=auth_db

# JWT
JWT_EMAIL_VERIFY_SECRET=your_email_verify_secret
JWT_EMAIL_VERIFY_EXPIRES_IN=3000
JWT_ACCESS_TOKEN_SECRET=your_access_token_secret
JWT_ACCESS_TOKEN_EXPIRES_IN=180
JWT_REFRESH_TOKEN_SECRET=your_refresh_token_secret
JWT_REFRESH_TOKEN_EXPIRES_IN=600

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password

# Google OAuth
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
```

## Quick Start

### Using Make Commands

Build and run with tests:
```bash
make all
```

Start the development server:
```bash
make run
```

Create database container:
```bash
make docker-run
```

Run tests:
```bash
make test
```

Live reload during development:
```bash
make watch
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/signup` - Register new user
- `POST /api/v1/auth/signin` - User login
- `PUT /api/v1/auth/verify-email/:token` - Verify email
- `GET /api/v1/auth/update-token` - Refresh access token
- `POST /api/v1/auth/signout` - User logout

### Social Authentication
- `GET /api/v1/auth/google/signin` - Initiate Google OAuth
- `GET /api/v1/auth/google/callback` - Google OAuth callback

### User Management
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile

## Project Structure

```
.
├── cmd/
│   └── api/                # Application entry point with graceful shutdown
├── internal/
│   ├── database/          # Database configuration, MySQL connection, migrations
│   ├── handler/           # HTTP request handlers for auth and user operations
│   │   ├── auth.go        # Authentication handlers (signup, signin, verify)
│   │   └── user.go        # User profile and management handlers
│   ├── helper/            # Utility functions
│   │   ├── random.go      # Secure random string generation for OAuth state
│   │   └── sendEmail.go   # Email service 
│   ├── middleware/        # HTTP middleware components
│   │   ├── auth.go        # JWT authentication middleware
│   │   └── validator.go   # Request validation middleware
│   ├── model/             # Database models and relationships
│   │   ├── user.go        # User model with profile relations
│   │   ├── userDetail.go  # Extended user profile information
│   │   ├── refreshToken.go# JWT refresh token management
│   │   ├── socialProfile.go# OAuth provider profile data
│   │   └── image.go       # User profile image handling
│   ├── oauth/             # OAuth integration
│   │   └── google.go      # Google OAuth2 configuration and user info
│   ├── response/          # Standardized API responses
│   │   ├── apiError.go    # Error response handling
│   │   └── sendResponse.go# Success response formatting
│   ├── server/            # Server configuration
│   │   ├── routes.go      # API route definitions and grouping
│   │   └── server.go      # HTTP server setup and configuration
│   └── validation/        # Request validation rules
│       └── auth.go        # Authentication request validation schema
├── .env                   # Environment variables configuration
└── Makefile              # Build and development commands
```

### Package Descriptions

- **handler**: Implements the business logic for all API endpoints
  - Authentication flows (traditional and social)
  - User profile management
  - Token refresh and validation

- **database**: Manages database connections and operations
  - MySQL connection setup
  - Database service interface
  - Connection pooling and configuration

- **model**: Defines data structures and relationships
  - User and profile models
  - Social authentication integration
  - Token management
  - GORM model definitions and relationships

- **middleware**: HTTP request processing
  - JWT authentication verification
  - Request validation
  - Database connection injection
  - CORS configuration

- **response**: Standardized API response handling
  - Consistent error formatting
  - Success response structure
  - HTTP status code management

- **validation**: Input validation rules
  - Request payload validation
  - Data format verification
  - Required field checking

- **server**: Core server functionality
  - Route configuration
  - Middleware integration
  - Server lifecycle management

- **oauth**: Social authentication
  - Google OAuth2 integration
  - User profile fetching
  - Token exchange and validation

- **helper**: Utility functions
  - Secure random string generation
  - Email service integration
  - Common helper functions

## Key Features Implementation

### Authentication Flow
1. User registers with email/password or social login
2. Email verification for traditional signup
3. JWT access token and refresh token issued on login
4. Automatic token refresh using refresh token

### Social Authentication
1. User initiates Google OAuth flow
2. OAuth state validation for security
3. User profile creation/update with social data
4. Automatic account linking if email exists

### Security Measures
- Password hashing using secure algorithms
- JWT token expiration and refresh mechanism
- CORS protection for API endpoints
- Input validation for all requests
- Secure cookie handling



## License

This project is licensed under the MIT License - see the LICENSE file for details.
