# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important Requirements

**Keep CLAUDE.md updated**: When implementation changes or Claude Code instructions differ from CLAUDE.md, always update CLAUDE.md to reflect the latest state.

## Project Overview

**umi.mikan** is a full-stack diary application with Go backend (gRPC) and SvelteKit frontend. The backend uses PostgreSQL with JWT authentication and CSRF protection, while the frontend is built with SvelteKit, TypeScript, and Tailwind CSS with comprehensive security headers. The system includes automated AI summary generation via Redis Pub/Sub, scheduled background processing, distributed locking, and comprehensive monitoring with Prometheus and Grafana.

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
dc up -d  # Starts all services (backend, frontend, postgres, postgres_test, redis, scheduler, subscriber, prometheus, grafana, loki, promtail, cadvisor)
```

**Service URLs:**

- Backend gRPC: http://localhost:2001
- Frontend: http://localhost:2000
- PostgreSQL: localhost:2002
- PostgreSQL Test: localhost:2003
- Redis: localhost:2004
- Subscriber Metrics: http://localhost:2005/metrics
- Scheduler Metrics: http://localhost:2006/metrics
- Prometheus: http://localhost:2007
- Grafana: http://localhost:2008 (admin/admin)
- cAdvisor: http://localhost:2009
- Loki: http://localhost:2010
- Grafana Alloy: http://localhost:2011

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

### Monitoring Services

```bash
# Prometheus (metrics collection)
docker compose logs prometheus       # View Prometheus logs
docker compose exec prometheus sh    # Access Prometheus container

# Grafana (monitoring dashboard)
docker compose logs grafana          # View Grafana logs
docker compose exec grafana sh       # Access Grafana container

# Loki (log aggregation)
docker compose logs loki             # View Loki logs
docker compose exec loki sh          # Access Loki container

# Grafana Alloy (log collection)
docker compose logs alloy            # View Alloy logs
docker compose exec alloy sh         # Access Alloy container

# cAdvisor (container metrics)
docker compose logs cadvisor         # View cAdvisor logs

# Access monitoring endpoints
curl http://localhost:2005/metrics   # Subscriber metrics
curl http://localhost:2006/metrics   # Scheduler metrics
curl http://localhost:2009/metrics   # cAdvisor metrics
curl http://localhost:2010/ready     # Loki health check
curl http://localhost:2011/metrics   # Grafana Alloy metrics
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
grpc_cli ls localhost:2001                                           # List services
grpc_cli ls localhost:2001 diary.DiaryService -l                     # Service details
grpc_cli type localhost:2001 diary.CreateDiaryEntryRequest           # Show message type
grpc_cli call localhost:2001 DiaryService.CreateDiaryEntry 'title: "test",content:"test"'  # Test call
grpc_cli call localhost:2001 DiaryService.SearchDiaryEntries 'userID:"id" keyword:"%日記%"'  # Search entries
```

## Architecture

### Backend Structure

- **Clean Architecture**: Domain models, services, and infrastructure layers
- **Dependency Injection**: uber-go/dig container with centralized DI management (`backend/container/container.go`)
- **gRPC Services**: AuthService and DiaryService
- **JWT Authentication**: 15-minute access tokens, 30-day refresh tokens with secure HTTP-only cookies
- **CSRF Protection**: Token-based CSRF protection with timing-safe validation
- **Database**: PostgreSQL with xo-generated models (separate test DB)
- **Hot Reload**: Air tool for automatic backend reloading
- **Async Processing**: Scheduler and Subscriber services with Redis Pub/Sub
- **Distributed Locking**: Redis-based locks with Lua scripts for task coordination
- **Monitoring**: Comprehensive monitoring stack with Prometheus, Grafana, Loki, Grafana Alloy, and cAdvisor

### Frontend Structure

- **Atomic Design**: Components organized as atoms/molecules/organisms
- **Route Protection**: Separated (authenticated) and (guest) route groups
- **State Management**: Svelte stores for user state and UI state
- **Type Safety**: Full TypeScript with generated gRPC types
- **Internationalization**: svelte-i18n with Japanese and English support
- **Progressive Web App**: @vite-pwa/sveltekit with offline support and app installation
- **Security Headers**: Content Security Policy (CSP), X-Frame-Options, X-Content-Type-Options, and Referrer Policy
- **CSRF Protection**: Client-side CSRF token handling for form submissions

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
                          ↓
                    Distributed Locks
                          ↓
                    Prometheus Metrics
```

- **Scheduler**: `backend/cmd/scheduler` - Periodic task execution
  - Identifies users with auto-summary enabled
  - Queues summary generation tasks (excluding today/current month)
  - Uses generic `ScheduledJob` interface for extensibility
  - Exposes metrics on port 2006 for monitoring

- **Redis Pub/Sub**: Message queue with `diary_events` channel
  - JSON message format with type-based routing
  - Message types: `daily_summary`, `monthly_summary`
  - Uses rueidis client for high performance

- **Subscriber**: `backend/cmd/subscriber` - Async message processor
  - Consumes messages from Redis queue
  - Generates summaries via LLM APIs
  - Saves results to database with conflict resolution
  - Exposes metrics on port 2005 for monitoring

- **Distributed Locking**: `backend/infrastructure/lock` - Redis-based coordination
  - Prevents duplicate task execution across multiple instances
  - Uses Lua scripts for atomic lock operations
  - Separate locks for daily and monthly summary generation

- **Comprehensive Monitoring Stack**: Prometheus + Grafana + Loki + Alloy + cAdvisor
  - **Prometheus**: Collects metrics from scheduler and subscriber services
  - **Grafana**: Custom dashboards for pub/sub monitoring and container resource monitoring
  - **Loki**: Log aggregation system for centralized log management
  - **Grafana Alloy**: Modern log collection agent that ships logs to Loki (replacement for Promtail)
  - **cAdvisor**: Container resource usage and performance metrics
  - Tracks job execution rates, duration, success rates, container resources, and logs

## Security & Authentication Flow

### Authentication
1. **Registration/Login**: Password-based via AuthService
2. **Token Storage**: JWT tokens in secure HTTP-only cookies with proper SameSite settings
3. **Authorization**: Bearer tokens in gRPC metadata headers
4. **Middleware**: Automatic token validation for protected endpoints
5. **User Context**: Injected user info available in all services

### Security Features
1. **CSRF Protection**: Token-based protection with timing-safe validation
2. **Content Security Policy**: Restrictive CSP headers with environment-specific rules
3. **Security Headers**: X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy
4. **Cookie Security**: Secure, HttpOnly, SameSite=Strict cookies with environment-aware settings
5. **Timing Attack Prevention**: Constant-time string comparison for token validation
6. **Registration Key Protection**: Optional registration key (REGISTER_KEY) to restrict new user signups

## Development Workflow

1. **Code Changes**: Backend uses Air for hot reload, frontend uses Vite
2. **Database Changes**: Update schema files, run `make db-init`, then `make xo`
3. **Proto Changes**: Update .proto files, run `make grpc`
4. **Frontend**: Uses pnpm for package management, Biome for formatting
5. **Backend**: Uses Go modules, standard Go formatting
6. **DI Container**: Add new dependencies to `backend/container/container.go` provider functions

## Key Files

- `compose.yml`: Development environment configuration
- `Makefile`: All development commands
- `proto/`: gRPC service definitions
- `schema/`: Database migration files
- `backend/cmd/server/main.go`: Backend entry point (uses DI container)
- `backend/cmd/scheduler/main.go`: Scheduler service entry point (uses DI container)
- `backend/cmd/subscriber/main.go`: Subscriber service entry point (uses DI container)
- `backend/container/container.go`: Central dependency injection configuration
- `frontend/src/routes/+layout.server.ts`: Authentication logic
- `frontend/src/hooks.server.ts`: Security headers and CSP configuration
- `frontend/src/lib/server/csrf.ts`: CSRF protection utilities
- `frontend/src/lib/utils/cookie-utils.ts`: Secure cookie configuration utilities
- `frontend/src/locales/`: Internationalization files (ja.json, en.json)
- `frontend/vite.config.ts`: PWA configuration with @vite-pwa/sveltekit
- `frontend/src/lib/components/PWA*`: PWA install/update components
- `frontend/static/icons/`: PWA app icons (72px-512px)
- `adr/`: Architecture Decision Records
  - `0004-pubsub.md`: Redis Pub/Sub implementation details
  - `0005-scheduler.md`: Scheduler system architecture
- `monitoring/`: Monitoring configuration
  - `prometheus.yml`: Metrics collection configuration
  - `loki/loki-config.yml`: Loki log aggregation configuration
  - `alloy/alloy-config.alloy`: Grafana Alloy log collection configuration
  - `grafana/`: Dashboard and data source provisioning
    - `dashboards/umi-mikan-pubsub.json`: Pub/Sub monitoring dashboard
    - `dashboards/container-monitoring.json`: Container resource monitoring dashboard
    - `dashboards/container-logs.json`: Container logs monitoring dashboard
    - `provisioning/datasources/`: Prometheus and Loki data source configurations
- `backend/infrastructure/lock/`: Distributed locking system
  - `distributed_lock.go`: Redis-based lock implementation
- `backend/container/`: Dependency injection container
  - `container.go`: Central DI container with uber-go/dig
  - `container_test.go`: Comprehensive container tests

## Development Guidelines

### Port Usage

- **Always use 2000 series ports**: All services must use ports in the 2000-2099 range
- **Port allocation**: Follow the existing port scheme documented in Service URLs
- **Current port allocation**:
  - 2000: Frontend
  - 2001: Backend gRPC
  - 2002: PostgreSQL
  - 2003: PostgreSQL Test
  - 2004: Redis
  - 2005: Subscriber Metrics
  - 2006: Scheduler Metrics
  - 2007: Prometheus
  - 2008: Grafana
  - 2009: cAdvisor
  - 2010: Loki
  - 2011: Grafana Alloy
- **Custom services**: New services should use available ports in the 2000 range (e.g., 2012+)

### Internationalization (i18n)

- **Always use i18n for user-facing text**: Use `$_("key")` for all UI text
- **Translation files**: Update both `frontend/src/locales/ja.json` and `frontend/src/locales/en.json`
- **Import requirements**: Include `import { _ } from "svelte-i18n";` and `import "$lib/i18n";`
- **Key structure**: Use nested objects (e.g., `timeProgress.yearProgress`)

### Component Creation

- **New components must support i18n**: All user-facing text should be translatable
- **Follow atomic design**: Place components in appropriate atoms/molecules/organisms directories
- **Consistent imports**: Always include necessary i18n imports
- **PWA considerations**: Ensure components work offline when cached data is available

### Security Guidelines

- **Always use CSRF protection**: Include CSRF tokens in all state-changing forms
- **Secure cookie configuration**: Use `getSecureCookieOptions()` for all authentication cookies
- **CSP compliance**: Ensure all new features comply with the defined Content Security Policy
- **Input validation**: Always validate and sanitize user inputs on both client and server sides
- **Timing attack prevention**: Use timing-safe comparisons for sensitive operations

### Code Language Guidelines

- **Comments in Japanese**: All code comments must be written in Japanese
- **Test case names in Japanese**: All test case names and descriptions should be written in Japanese

### TypeScript Guidelines

- **No any or unknown types**: Always use specific, properly typed interfaces and types
- **Type safety**: Ensure all variables, function parameters, and return values have explicit types
- **Generated types**: Use the auto-generated gRPC types from `frontend/src/lib/grpc/`

## Configuration Options

### Scheduler Configuration

Environment variables for controlling scheduler behavior:

- `SCHEDULER_DAILY_INTERVAL`: Interval for daily summary job execution (default: `5m`)
- `SCHEDULER_MONTHLY_INTERVAL`: Interval for monthly summary job execution (default: `5m`)

Examples:
```bash
SCHEDULER_DAILY_INTERVAL=10m    # Run daily summaries every 10 minutes
SCHEDULER_MONTHLY_INTERVAL=1h   # Run monthly summaries every hour
```

### Subscriber Configuration

Environment variables for controlling async message processing:

- `SUBSCRIBER_MAX_CONCURRENT_JOBS`: Maximum number of concurrent message processing jobs (default: `10`)

Examples:
```bash
SUBSCRIBER_MAX_CONCURRENT_JOBS=5    # Limit to 5 concurrent jobs
SUBSCRIBER_MAX_CONCURRENT_JOBS=20   # Allow up to 20 concurrent jobs
```

### Registration Key Configuration

Environment variable for restricting new user registrations:

- `REGISTER_KEY`: Secret key required for new user registration (optional, no default)

**Usage:**
- **Not set**: Anyone can register without a key (default behavior)
- **Set**: Users must provide the correct registration key during signup

**Configuration:**
1. Set `REGISTER_KEY` value in the `backend` service in `compose.yml`
2. Backend validates the key during registration
3. Frontend always displays the registration key field (users can leave it empty if not required)

Examples:
```yaml
# compose.yml - backend service only
environment:
  REGISTER_KEY: "your-secret-registration-key"
```

**Security Notes:**
- Use a strong, unique key for production environments
- Keys are validated with timing-safe comparison to prevent timing attacks
- Invalid or missing keys return appropriate error codes:
  - `codes.InvalidArgument`: Registration key is required but not provided
  - `codes.PermissionDenied`: Provided registration key is incorrect

## Production Notes

- Copy `compose-prod.example.yml` to `compose-prod.yml` for production
- gRPC reflection is enabled in development (TODO: disable in production)
- JWT_SECRET should be changed from "hogehoge" in production
- Frontend builds with `docker compose exec frontend pnpm build`, backend builds with `docker compose exec backend go build`
- PWA manifest and service worker are automatically generated during build
- PWA icons are pre-generated in `frontend/static/icons/` directory

### Security in Production

- **HTTPS Required**: All cookies are configured with `secure: true` in production environments
- **CSP Headers**: Content Security Policy headers are automatically applied via `hooks.server.ts`
- **CSRF Protection**: CSRF tokens are mandatory for all state-changing operations
- **Cookie Security**: HTTP-only, secure, SameSite=Strict cookies for authentication tokens
- **CI/CD Optimization**: Build workflow optimized for performance (removed disk space cleanup step)
