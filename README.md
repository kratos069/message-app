# GoMessenger

> Production-grade end-to-end encrypted messaging application built with Go

A scalable, secure real-time messaging platform demonstrating modern backend development practices and production-ready code quality.

## 🎯 Project Overview

GoMessenger is a full-featured chat application backend built to handle thousands of concurrent users while maintaining sub-second response times. This project showcases enterprise-level software engineering practices including clean architecture, comprehensive testing, and production-grade observability.

**Proficiency in:
- Backend system design and scalability
- Security-first development (E2E encryption, PASETO tokens)
- Production-ready code (testing, graceful shutdown)
- Modern Go ecosystem and best practices

## ✨ Key Features

### Core Functionality
- 💬 **Real-time Chat** - Direct messaging with instant delivery
- 👥 **User Management** - Registration, authentication, and profile management
- 📧 **Email Verification** - Async email verification system using Redis queues
- 📱 **Read Receipts** - Track message delivery and read status
- ⌨️ **Typing Indicators** - Real-time typing status updates
- 🔍 **Message Search** - Full-text search across conversations (on encrypted metadata)

### Technical Excellence
- 🚀 **High Performance** - Handles 300+ concurrent users on 4 cores with <2s p95 latency
- 🔒 **Security First** - PASETO tokens, bcrypt password hashing, SQL injection prevention
- 📊 **Production Ready** - Health checks, graceful shutdown, connection pooling
- 🧪 **Comprehensive Testing** - Unit tests, integration tests, load tests with k6
- 📈 **Scalable Architecture** - Horizontal scaling ready, database optimizations
- 🛡️ **Error Handling** - Proper error responses, idempotency for messages

## 🏗️ Architecture

### System Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS/REST
       ▼
┌─────────────────────────────────────┐
│         Gin HTTP Server             │
│  ┌──────────────────────────────┐  │
│  │   Middleware Stack           │  │
│  │  - Authentication (PASETO)   │  │
│  │  - Request Validation        │  │
│  │  - Error Handling            │  │
│  │  - Active Request Tracking   │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │   API Handlers               │  │
│  │  - Auth (Login/Register)     │  │
│  │  - Conversations             │  │
│  │  - Messages                  │  │
│  │  - Users                     │  │
│  │  - Admin                     │  │
│  └──────────────────────────────┘  │
└───────┬─────────────────┬───────────┘
        │                 │
        ▼                 ▼
┌──────────────┐  ┌──────────────────┐
│  PostgreSQL  │  │  Redis (Asynq)   │
│              │  │                  │
│  - Users     │  │  - Email Queue   │
│  - Messages  │  │  - Tasks         │
│  - Convos    │  │                  │
│  - Sessions  │  │                  │
└──────────────┘  └──────────────────┘
```

### Database Schema

```
Users ──────┐
            │
            ├──── Sessions (Auth)
            │
            ├──── ConversationParticipants
            │           │
            │           ▼
            │     Conversations
            │           │
            └──────┐    │
                   ▼    ▼
                Messages (E2E Encrypted)
```

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.25.4
- **Web Framework**: Gin (high-performance HTTP router)
- **Database**: PostgreSQL 17 (with pgx driver)
- **Query Builder**: SQLC (type-safe SQL)
- **Authentication**: PASETO v4 (secure tokens)
- **Password Hashing**: bcrypt

### Infrastructure
- **Caching/Queue**: Redis (with Asynq for async tasks)
- **Email Service**: SMTP integration for verification emails
- **Database Migrations**: golang-migrate
- **Connection Pooling**: pgxpool

### Development & Testing
- **Testing**: Go testing package + testify
- **Load Testing**: Grafana k6 (realistic production scenarios)
- **Logging**: zerolog (structured, fast logging)
- **Code Quality**: golangci-lint, go vet

### DevOps
- **CI/CD**: GitHub Actions
- **Process Management**: Graceful shutdown, health checks (K8s-ready)

## 🚀 Getting Started

### Prerequisites

```bash
# Required
Go 1.25+
PostgreSQL 17+
Redis 7+
```

### Technical Skills
- **Advanced Go**: Context management, goroutines, channels, error handling patterns
- **Database Design**: Schema design for scalability, transaction management, query optimization
- **Security**: E2E encryption implementation, secure token management, vulnerability prevention
- **System Design**: async processing, horizontal scaling patterns
- **Testing**: Unit testing, integration testing, load testing with realistic scenarios
- **DevOps**: Docker containerization, CI/CD pipelines, production deployment strategies

### Best Practices
- Clean architecture and separation of concerns
- Type-safe database operations with SQLC
- Comprehensive error handling and logging
- Production-ready code quality (testing, graceful degradation)
- API design following REST principles
- Security-first development mindset

Built with ❤️ using Go
