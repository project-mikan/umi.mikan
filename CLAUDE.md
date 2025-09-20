# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**umi.mikan** is a full-stack diary application with Go backend (gRPC) and SvelteKit frontend. The backend uses PostgreSQL with JWT authentication, while the frontend is built with SvelteKit, TypeScript, and Tailwind CSS. The system includes automated AI summary generation via Redis Pub/Sub and scheduled background processing.

## Development Setup

### Prerequisites

```bash
sudo pacman -S protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
npm install -g @grpc/proto-loader
```

### Starting Development Environment

```bash
dc up -d  # Starts all services (backend, frontend, postgres, redis, scheduler, subscriber)
```

**Service URLs:**

- Backend gRPC: http://localhost:8080
- Frontend: http://localhost:5173
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Common Development Commands

### Frontend Development

```bash
make pnpm-dev          # Start frontend dev server with hot reload
make f-format          # Format frontend code with Biome
make f-sh              # Access frontend container shell
make f-log
```

If you want to use pnpm commands, use `docker compose exec frontend pnpm`.

backend/infrastructure/grpc dir is automatically generated. DO NOT EDIT MANUALLY.

When installing the library, `docker compose exec frontend pnpm install` instead of writing it in package.json, and give the -D option if necessary.

When you change the frontend, make sure that

- `make f-lint`
- `make f-test`
- `make f-log`

are OK.

### Backend Development

```bash
make b-sh              # Access backend container shell
make tidy              # Run go mod tidy
make b-log             # View backend logs
```

If you want to use the go command, use `docker compose exec backend go`.

frontend/src/lib/grpc dir is automatically generated. DO NOT EDIT MANUALLY.

When you change the backend, make sure that

- `make b-lint`
- `make b-test`
- `make b-log`

are OK.

### Async Processing Services

```bash
# Scheduler service (periodic task execution)
docker compose logs scheduler        # View scheduler logs
docker compose exec scheduler sh     # Access scheduler container

# Subscriber service (async message processing)
docker compose logs subscriber       # View subscriber logs
docker compose exec subscriber sh    # Access subscriber container

# Redis (message queue)
docker compose logs redis            # View Redis logs
docker compose exec redis redis-cli  # Access Redis CLI
```

### Database Operations

```bash
make db                # Connect to PostgreSQL
make db-init           # Reset and reinitialize database
make p-log             # View postgres logs
```

### Code Generation

```bash
make grpc              # Generate both Go and TypeScript gRPC code
make grpc-go           # Generate Go gRPC code only
make grpc-ts           # Generate TypeScript gRPC code only
make xo                # Generate database models from schema
```

### gRPC Debugging

```bash
grpc_cli ls localhost:8080                                           # List services
grpc_cli ls localhost:8080 diary.DiaryService -l                     # Service details
grpc_cli type localhost:8080 diary.CreateDiaryEntryRequest           # Show message type
grpc_cli call localhost:8080 DiaryService.CreateDiaryEntry 'title: "test",content:"test"'  # Test call
```

## Architecture

### Backend Structure

- **Clean Architecture**: Domain models, services, and infrastructure layers
- **gRPC Services**: AuthService and DiaryService
- **JWT Authentication**: 15-minute access tokens, 30-day refresh tokens
- **Database**: PostgreSQL with xo-generated models
- **Hot Reload**: Air tool for automatic backend reloading
- **Async Processing**: Scheduler and Subscriber services with Redis Pub/Sub

### Frontend Structure

- **Atomic Design**: Components organized as atoms/molecules/organisms
- **Route Protection**: Separated (authenticated) and (guest) route groups
- **State Management**: Svelte stores for user state and UI state
- **Type Safety**: Full TypeScript with generated gRPC types
- **Internationalization**: svelte-i18n with Japanese and English support

### Database Schema

- **users**: UUID primary keys, email-based authentication
- **diaries**: One diary per user per date (unique constraint)
- **user_password_authes**: Separate password authentication table
- **user_llms**: LLM provider settings and auto-summary preferences
- **diary_summary_days**: AI-generated daily summaries
- **diary_summary_months**: AI-generated monthly summaries
- **Migrations**: Numbered SQL files in /schema directory

### Async Processing Architecture

```
Scheduler (5min interval) → Redis Pub/Sub → Subscriber → LLM APIs → Database
```

- **Scheduler**: `backend/cmd/scheduler` - Periodic task execution
  - Identifies users with auto-summary enabled
  - Queues summary generation tasks (excluding today/current month)
  - Uses generic `ScheduledJob` interface for extensibility

- **Redis Pub/Sub**: Message queue with `diary_events` channel
  - JSON message format with type-based routing
  - Message types: `daily_summary`, `monthly_summary`
  - Uses rueidis client for high performance

- **Subscriber**: `backend/cmd/subscriber` - Async message processor
  - Consumes messages from Redis queue
  - Generates summaries via LLM APIs
  - Saves results to database with conflict resolution

## Authentication Flow

1. **Registration/Login**: Password-based via AuthService
2. **Token Storage**: JWT tokens in HTTP-only cookies (frontend)
3. **Authorization**: Bearer tokens in gRPC metadata headers
4. **Middleware**: Automatic token validation for protected endpoints
5. **User Context**: Injected user info available in all services

## Development Workflow

1. **Code Changes**: Backend uses Air for hot reload, frontend uses Vite
2. **Database Changes**: Update schema files, run `make db-init`, then `make xo`
3. **Proto Changes**: Update .proto files, run `make grpc`
4. **Frontend**: Uses pnpm for package management, Biome for formatting
5. **Backend**: Uses Go modules, standard Go formatting

## Key Files

- `compose.yml`: Development environment configuration
- `Makefile`: All development commands
- `proto/`: gRPC service definitions
- `schema/`: Database migration files
- `backend/cmd/server/main.go`: Backend entry point
- `backend/cmd/scheduler/main.go`: Scheduler service entry point
- `backend/cmd/subscriber/main.go`: Subscriber service entry point
- `frontend/src/routes/+layout.server.ts`: Authentication logic
- `frontend/src/locales/`: Internationalization files (ja.json, en.json)
- `adr/`: Architecture Decision Records
  - `0004-pubsub.md`: Redis Pub/Sub implementation details
  - `0005-scheduler.md`: Scheduler system architecture

## Development Guidelines

### Internationalization (i18n)

- **Always use i18n for user-facing text**: Use `$_("key")` for all UI text
- **Translation files**: Update both `frontend/src/locales/ja.json` and `frontend/src/locales/en.json`
- **Import requirements**: Include `import { _ } from "svelte-i18n";` and `import "$lib/i18n";`
- **Key structure**: Use nested objects (e.g., `timeProgress.yearProgress`)

### Component Creation

- **New components must support i18n**: All user-facing text should be translatable
- **Follow atomic design**: Place components in appropriate atoms/molecules/organisms directories
- **Consistent imports**: Always include necessary i18n imports

## Production Notes

- Copy `compose-prod.example.yml` to `compose-prod.yml` for production
- gRPC reflection is enabled in development (TODO: disable in production)
- JWT_SECRET should be changed from "hogehoge" in production
- Frontend builds with `dcoker compose exec frontend pnpm build`, backend builds with `docker compopse exec backend go build`
