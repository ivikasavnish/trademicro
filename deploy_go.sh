#!/bin/bash

# TradeMicro Go API Deployment Script
# Run this script on your server

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration - Edit these variables
APP_DIR="/opt/trademicro"
APP_PORT=8000

# Database configuration - Replace with your PostgreSQL DSN
POSTGRES_DSN="your-postgres-dsn-here"

# Redis configuration
REDIS_URL="redis://localhost:6379/0"

# Generate a random secret key
SECRET_KEY=$(openssl rand -hex 32)

echo -e "${GREEN}Starting TradeMicro Go API deployment...${NC}"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}This script must be run as root${NC}"
    exit 1
fi

# Update system packages
echo -e "${YELLOW}Updating system packages...${NC}"
apt update && apt upgrade -y

# Install Redis if not already installed
echo -e "${YELLOW}Installing Redis...${NC}"
apt install -y redis-server

# Create application directory
echo -e "${YELLOW}Creating application directory...${NC}"
mkdir -p $APP_DIR

# Check if the directory already has files
if [ "$(ls -A $APP_DIR)" ]; then
    echo -e "${YELLOW}Directory $APP_DIR is not empty. Backing up existing files...${NC}"
    BACKUP_DIR="${APP_DIR}_backup_$(date +%Y%m%d%H%M%S)"
    mv $APP_DIR $BACKUP_DIR
    mkdir -p $APP_DIR
fi

# Copy the Go binary to the server
echo -e "${YELLOW}Copying application binary...${NC}"
cp ./trademicro $APP_DIR/

# Create environment file
echo -e "${YELLOW}Creating environment file...${NC}"
cat > $APP_DIR/.env << EOF
POSTGRES_URL=$POSTGRES_DSN
REDIS_URL=$REDIS_URL
SECRET_KEY=$SECRET_KEY
PORT=$APP_PORT
EOF

# Create systemd service
echo -e "${YELLOW}Creating systemd service...${NC}"
cat > /etc/systemd/system/trademicro.service << EOF
[Unit]
Description=TradeMicro Go API
After=network.target

[Service]
User=root
Group=root
WorkingDirectory=$APP_DIR
EnvironmentFile=$APP_DIR/.env
ExecStart=$APP_DIR/trademicro
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo -e "${YELLOW}Enabling and starting service...${NC}"
systemctl daemon-reload
systemctl enable trademicro
systemctl start trademicro

# Configure Nginx
echo -e "${YELLOW}Installing and configuring Nginx...${NC}"
apt install -y nginx

cat > /etc/nginx/sites-available/trademicro << EOF
server {
    listen 80;
    server_name \$hostname;

    location / {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF

# Enable Nginx site
ln -sf /etc/nginx/sites-available/trademicro /etc/nginx/sites-enabled/
nginx -t && systemctl restart nginx

# Check if service is running
if systemctl is-active --quiet trademicro; then
    echo -e "${GREEN}TradeMicro API has been successfully deployed!${NC}"
    SERVER_IP=$(hostname -I | awk '{print $1}')
    echo -e "${GREEN}Your API is accessible at: http://$SERVER_IP${NC}"
    echo -e "\n${YELLOW}Database connection:${NC}"
    echo -e "Using PostgreSQL DSN from configuration"
    
    echo -e "\n${YELLOW}To check service status:${NC}"
    echo -e "systemctl status trademicro"
    
    echo -e "\n${YELLOW}To view logs:${NC}"
    echo -e "journalctl -u trademicro"
else
    echo -e "${RED}Deployment failed. Service is not running.${NC}"
    echo -e "${YELLOW}Check logs with: journalctl -u trademicro${NC}"
fi
