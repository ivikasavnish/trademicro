#!/bin/bash
# Migration management script

# Default values
MIGRATIONS_PATH="./migrations"
ENV_FILE=".env"
DATABASE_URL_ENV_VAR="POSTGRES_URL"

# Check if migrate command is available
if ! command -v migrate &> /dev/null; then
    echo "Error: 'migrate' command not found."
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Function to display help
show_help() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo
    echo "Options:"
    echo "  -e, --env ENV_FILE       Specify environment file (default: .env)"
    echo "  -u, --url DATABASE_URL   Specify database URL directly"
    echo "  -p, --path PATH          Migration files path (default: ./migrations)"
    echo "  -h, --help               Show this help message"
    echo
    echo "Commands:"
    echo "  up        Run all pending migrations"
    echo "  down      Rollback all migrations"
    echo "  version   Show current migration version"
    echo "  create NAME     Create new migration files"
    echo "  force VERSION   Force migration version (use with caution)"
    echo "  reset     Reset all migrations (drops and recreates everything)"
    echo "  status    Check migration status"
    echo
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -e|--env)
            ENV_FILE="$2"
            shift 2
            ;;
        -u|--url)
            DATABASE_URL="$2"
            shift 2
            ;;
        -p|--path)
            MIGRATIONS_PATH="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

# Check for command
if [[ $# -eq 0 ]]; then
    show_help
    exit 1
fi

COMMAND="$1"
shift

# Load environment variables if no direct URL provided
if [[ -z "$DATABASE_URL" ]]; then
    if [[ -f "$ENV_FILE" ]]; then
        echo "Loading environment from $ENV_FILE"
        export $(grep -v '^#' "$ENV_FILE" | xargs)
        DATABASE_URL="${!DATABASE_URL_ENV_VAR}"
    else
        echo "Warning: Environment file $ENV_FILE not found."
    fi
fi

# Check if we have a database URL
if [[ -z "$DATABASE_URL" ]]; then
    echo "Error: No database URL provided."
    echo "Either specify it with -u/--url, set in environment file, or set $DATABASE_URL_ENV_VAR environment variable."
    exit 1
fi

# Process commands
case "$COMMAND" in
    up)
        echo "Running migrations up..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" up
        ;;
    down)
        echo "Running migrations down..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" down
        ;;
    version)
        echo "Current migration version:"
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" version
        ;;
    create)
        if [[ $# -eq 0 ]]; then
            echo "Error: Migration name is required"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        name="$1"
        timestamp=$(date +%Y%m%d%H%M%S)
        up_file="${MIGRATIONS_PATH}/${timestamp}_${name}.up.sql"
        down_file="${MIGRATIONS_PATH}/${timestamp}_${name}.down.sql"
        
        mkdir -p "$MIGRATIONS_PATH"
        echo "-- Write your UP migration SQL here" > "$up_file"
        echo "-- Write your DOWN migration SQL here" > "$down_file"
        
        echo "Created migration files:"
        echo "  $up_file"
        echo "  $down_file"
        ;;
    force)
        if [[ $# -eq 0 ]]; then
            echo "Error: Version number is required"
            echo "Usage: $0 force <version>"
            exit 1
        fi
        echo "Forcing migration version to $1..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" force "$1"
        ;;
    reset)
        echo "WARNING: This will reset ALL migrations and data. Press Ctrl+C to cancel..."
        sleep 5
        
        echo "Forcing migration version to 0..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" force 0
        
        echo "Running migrations up..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" up
        ;;
    status)
        echo "Checking migration status..."
        migrate -database "$DATABASE_URL" -path "$MIGRATIONS_PATH" version
        ;;
    *)
        echo "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac

exit_code=$?
if [ $exit_code -eq 0 ]; then
    echo "Command completed successfully."
else
    echo "Command failed with exit code $exit_code."
fi

exit $exit_code