# TradeMicro Go API

A modular Go API service for trading operations.

## Project Structure

```
trademicro/
├── cmd/
│   └── server/
│       └── main.go         # Entry point (minimal code)
├── internal/
│   ├── api/                # API initialization and routing
│   ├── config/             # Application configuration
│   ├── db/                 # Database setup and connections
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   ├── services/           # Business logic
│   ├── tasks/              # Background jobs and scheduled tasks
│   └── websocket/          # WebSocket handling
└── pkg/                    # Shared packages
```

## Migration to Modular Structure

TradeMicro is being migrated from a monolithic structure to a modular one. The modular version is in `cmd/server/main.go`, while the monolithic version is kept in `main.go` for backward compatibility.

## Getting Started

### Running the Server

```bash
make build
./trademicro
```

### Running Migrations

```bash
make migrate-up
```

### Importing Instrument Data

Update and import CSV data:

```bash
make update-csv-import
```

### Running Tests

```bash
go test ./...
```

## Environment Variables

TradeMicro uses the following environment variables:

- `PORT`: The HTTP port to listen on
- `POSTGRES_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `SECRET_KEY`: JWT secret key for authentication
- `SERVER_ROLE`: `micro` for frontend API, any other value for worker

## Modular Design

The application is designed following clean architecture principles:

- **Models** - Data structures
- **Services** - Business logic
- **Handlers** - HTTP request handling
- **Middleware** - HTTP middleware for auth, logging, etc.
- **Config** - Application configuration
- **Tasks** - Background jobs and scheduled tasks
- **WebSocket** - Real-time communication

## Deployment

Deploy using the included scripts:

```bash
./deploy_to_gcp.sh
```

## License

Copyright (c) 2025 TradeMicro - All Rights Reserved
