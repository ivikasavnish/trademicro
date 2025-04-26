# TradeMicro API

A Go-based backend API for the TradeMicro trading management application.

## Features

- RESTful API for trade management
- WebSocket support for real-time updates
- JWT authentication
- PostgreSQL database integration
- Redis for WebSocket message broadcasting

## Prerequisites

- Go 1.16 or higher
- PostgreSQL database
- Redis (optional, for WebSocket functionality)

## Local Development

### Setup

1. Clone the repository
2. Install dependencies:
   ```
   go mod download
   ```
3. Create a `.env` file with the following variables:
   ```
   POSTGRES_URL=postgresql://username:password@localhost:5432/trademicro
   REDIS_URL=redis://localhost:6379/0
   SECRET_KEY=your_secret_key_here
   PORT=8000
   ```
4. Run the application:
   ```
   go run main.go
   ```

## Building for Production

To build the application for production:

```bash
go build -o trademicro main.go
```

This will create a single binary file that can be deployed to your server.

## Deployment

### Simple Deployment

1. Build the application on your local machine:
   ```bash
   go build -o trademicro main.go
   ```

2. Copy the binary to your server:
   ```bash
   scp trademicro root@your-server-ip:/tmp/
   ```

3. SSH into your server and run the deployment script:
   ```bash
   ssh root@your-server-ip
   cd /tmp
   chmod +x deploy_go.sh
   ./deploy_go.sh
   ```

### Using the Deployment Script

The `deploy_go.sh` script automates the deployment process. Before running it:

1. Edit the script to set your database credentials and other configuration
2. Make the script executable: `chmod +x deploy_go.sh`
3. Run the script on your server

## API Endpoints

### Authentication

- `POST /token` - Login and get JWT token

### Trades

- `GET /api/trades` - Get all trades
- `POST /api/trades` - Create a new trade
- `GET /api/trades/{id}` - Get a specific trade
- `PUT /api/trades/{id}` - Update a trade

### Symbols

- `GET /api/symbols` - Get all symbols
- `POST /api/symbols` - Create a new symbol

### Broker Tokens

- `GET /api/broker-tokens` - Get all broker tokens
- `POST /api/broker-tokens` - Create a new broker token

### Users

- `GET /api/users` - Get all users
- `POST /api/users` - Create a new user

### WebSocket

- `GET /ws` - WebSocket endpoint for real-time updates

## Benefits of Go for Deployment

- Single binary deployment (no dependencies to install)
- Cross-platform compatibility
- Excellent performance
- Low memory footprint
- Built-in concurrency support
