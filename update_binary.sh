#!/bin/bash

# TradeMicro API Server Update Script
# This script only updates the binary and restarts the service

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration
APP_DIR="/opt/trademicro"
BINARY_PATH="$APP_DIR/trademicro"
SERVICE_NAME="trademicro"

echo -e "${GREEN}Starting TradeMicro API binary update...${NC}"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}This script must be run as root${NC}"
    exit 1
fi

# Check if service exists
if ! systemctl list-unit-files | grep -q "$SERVICE_NAME.service"; then
    echo -e "${RED}Error: $SERVICE_NAME service does not exist. Please run the full deployment script first.${NC}"
    exit 1
fi

# Check if application directory exists
if [ ! -d "$APP_DIR" ]; then
    echo -e "${RED}Error: Application directory $APP_DIR does not exist. Please run the full deployment script first.${NC}"
    exit 1
fi

# Stop the service
echo -e "${YELLOW}Stopping $SERVICE_NAME service...${NC}"
systemctl stop $SERVICE_NAME

# Backup the existing binary
if [ -f "$BINARY_PATH" ]; then
    echo -e "${YELLOW}Backing up existing binary...${NC}"
    mv "$BINARY_PATH" "${BINARY_PATH}.bak.$(date +%Y%m%d%H%M%S)"
fi

# Copy the new binary
echo -e "${YELLOW}Copying new binary...${NC}"
cp ./trademicro "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Start the service
echo -e "${YELLOW}Starting $SERVICE_NAME service...${NC}"
systemctl start $SERVICE_NAME

# Check if service is running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${GREEN}TradeMicro API has been successfully updated!${NC}"
    echo -e "${YELLOW}To check service status:${NC}"
    echo -e "systemctl status $SERVICE_NAME"
    
    echo -e "\n${YELLOW}To view logs:${NC}"
    echo -e "journalctl -u $SERVICE_NAME"
else
    echo -e "${RED}Update failed. Service is not running.${NC}"
    echo -e "${YELLOW}Check logs with: journalctl -u $SERVICE_NAME${NC}"
    exit 1
fi
