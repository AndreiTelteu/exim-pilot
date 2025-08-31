# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Exim-Pilot is a comprehensive web-based management interface for Exim mail servers. It's a full-stack application with a Go backend and React frontend that provides real-time mail queue monitoring, log analysis, and deliverability reporting.

## Common Development Commands

### Building and Running
- `./make install` - Install Go dependencies and frontend packages
- `./make dev` - Start development server with hot reload using Air
- `./make build` - Build production binary with embedded frontend assets
- `./make build-dev` - Build development binary (no embedded frontend)
- `./make build-frontend` - Build frontend only
- `./make start` - Start production binary with config
- `./make test` - Run Go tests
- `./make clean` - Clean build artifacts

### Frontend Development
- `cd web && npm run dev` - Start frontend development server
- `cd web && npm run build` - Build frontend for production
- `cd web && npm run lint` - Lint frontend code

### Production Deployment
- `./make build-all` - Build all binaries (main app + config tool)
- `./make verify` - Verify embedded assets in production binary

## Architecture Overview

### Backend (Go)
- **cmd/exim-pilot/main.go**: Main application entry point with service initialization
- **internal/api/**: REST API server with comprehensive endpoints for queue, logs, reports
- **internal/config/**: YAML-based configuration with environment variable overrides
- **internal/database/**: SQLite database layer with migrations, models, and repositories
- **internal/logprocessor/**: Real-time Exim log parsing and processing service
- **internal/queue/**: Exim queue management operations
- **internal/websocket/**: WebSocket service for real-time updates
- **internal/auth/**: Authentication and user management
- **internal/static/**: Static file serving with embedded/dev mode switching

### Frontend (React/TypeScript)
- **web/src/App.tsx**: Main application with routing and authentication
- **web/src/components/**: Modular React components (Dashboard, Queue, Reports, etc.)
- **web/src/context/**: React contexts for auth and app state
- **web/src/services/**: API client and WebSocket service
- Built with Vite, TailwindCSS, React Router, and ECharts for visualization

### Key Services Integration
- **Log Processing**: Real-time monitoring of Exim logs with WebSocket broadcasting
- **Queue Management**: Direct interaction with Exim spool directory and command-line tools
- **Database**: SQLite with automatic migrations and connection pooling
- **WebSocket**: Real-time updates for dashboard metrics and log entries

### Configuration
- Primary config: `config/test-config.yaml` (development)
- Production default: `/opt/exim-pilot/config/config.yaml`
- Environment overrides: `EXIM_PILOT_*` variables
- Air config: `.air.toml` for hot reload in development

### Build System
- **Embedded Assets**: Production builds embed frontend in Go binary using build tags
- **Development Mode**: Serves frontend from filesystem in dev builds
- **Multi-binary**: Builds both main app and configuration tool
- **Hot Reload**: Air watches Go files, Vite handles frontend hot reload

## Development Workflow

### Starting Development
1. Run `./make install` to install dependencies
2. Use `./make dev` for full-stack development with hot reload
3. Or run backend (`./make build-dev && ./tmp/main.exe -config config/test-config.yaml`) and frontend (`cd web && npm run dev`) separately

### Testing and Quality
- Always run `./make test` after making changes
- Frontend has linting configured: `cd web && npm run lint`
- Production builds automatically optimize and minify assets

### Database Management
- Automatic migrations run on startup
- Manual migration commands: `--migrate-up` / `--migrate-down` flags
- SQLite database with connection pooling and transaction support

### Configuration Management
- Development uses `config/test-config.yaml`
- Environment variables override config values
- Configuration validation occurs on startup
- Default admin user created if none exists (username: admin, password configurable)

## Key Implementation Patterns

### API Structure
- RESTful endpoints under `/api/v1/`
- Authentication middleware for protected routes
- WebSocket endpoint at `/ws` for real-time communication
- Comprehensive error handling and response formatting

### Frontend Architecture
- Context-based state management (AuthContext, AppContext)
- Component-based architecture with reusable UI elements
- TypeScript for type safety
- Responsive design with TailwindCSS

### Real-time Features
- WebSocket service broadcasts log entries and dashboard updates
- Frontend automatically reconnects on authentication
- Periodic dashboard metric updates every 30 seconds

### Security Considerations
- Session-based authentication with configurable timeouts
- CORS configuration for API endpoints
- Security headers middleware
- Content redaction for sensitive log information
- Audit logging for all user actions