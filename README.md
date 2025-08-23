# Exim Control Panel (Exim-Pilot)

A comprehensive web-based management interface for Exim mail servers running on Ubuntu/Debian systems.

## Project Structure

```
exim-pilot/
├── cmd/
│   └── exim-pilot/          # Main application entry point
├── internal/                # Internal Go packages (will be populated)
├── web/                     # Frontend React/TypeScript project (will be populated)
├── .air.toml               # Air configuration for hot reload
├── make                    # Build automation script
├── go.mod                  # Go module dependencies
└── README.md               # This file
```

## Development

### Prerequisites
- Go 1.23.4+
- Air (for hot reload): `go install github.com/cosmtrek/air@latest`

### Getting Started

1. Install dependencies:
   ```bash
   ./make install
   ```

2. Run in development mode with hot reload:
   ```bash
   ./make dev
   ```

3. Build for production:
   ```bash
   ./make build
   ```

### Available Build Commands

- `./make install` - Install Go dependencies
- `./make dev` - Start development server with hot reload
- `./make build-dev` - Build development binary
- `./make build` - Build production binary with embedded frontend
- `./make test` - Run tests
- `./make clean` - Clean build artifacts

## Features (To be implemented)

- Real-time mail queue monitoring and management
- Comprehensive log analysis and search
- Deliverability reporting and analytics
- Message tracing and delivery history
- Web-based interface with responsive design
- Security and audit logging

## Requirements

See `.kiro/specs/exim-control-panel/requirements.md` for detailed requirements.

## Design

See `.kiro/specs/exim-control-panel/design.md` for architectural design.

## Implementation Plan

See `.kiro/specs/exim-control-panel/tasks.md` for the implementation task list.