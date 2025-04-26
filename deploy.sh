#!/bin/bash

# TradeMicro API Deployment Script
# Run this script on your server as root

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration - Edit these variables
APP_DIR="/var/www/trademicro"
APP_PORT=8000

# Database configuration - Replace with your remote PostgreSQL details
DB_HOST="your-db-host.com"
DB_PORT="5432"
DB_NAME="trademicro"
DB_USER="your-db-username"
DB_PASSWORD="your-db-password"

# Redis configuration
REDIS_URL="redis://localhost:6379/0"

# Generate a random secret key
SECRET_KEY=$(openssl rand -hex 32)

echo -e "${GREEN}Starting TradeMicro API deployment...${NC}"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}This script must be run as root${NC}"
    exit 1
fi

# Update system packages
echo -e "${YELLOW}Updating system packages...${NC}"
apt update && apt upgrade -y

# Install required packages
echo -e "${YELLOW}Installing required packages...${NC}"
apt install -y python3 python3-pip python3-dev build-essential libssl-dev libffi-dev python3-setuptools python3-venv redis-server nginx

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

# Copy application files
echo -e "${YELLOW}Copying application files...${NC}"
cp -r $(dirname "$0")/* $APP_DIR/

# Create Python virtual environment
echo -e "${YELLOW}Setting up Python virtual environment...${NC}"
cd $APP_DIR
python3 -m venv venv
source venv/bin/activate

# Install Python dependencies
echo -e "${YELLOW}Installing Python dependencies...${NC}"
pip install fastapi uvicorn sqlalchemy asyncpg aioredis python-jose passlib python-dotenv

# Create environment file
echo -e "${YELLOW}Creating environment file...${NC}"
cat > $APP_DIR/.env << EOF
POSTGRES_URL=postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME
REDIS_URL=$REDIS_URL
SECRET_KEY=$SECRET_KEY
EOF

# Update CORS settings
echo -e "${YELLOW}Updating CORS settings...${NC}"
SERVER_IP=$(hostname -I | awk '{print $1}')
sed -i "s|allow_origins=\[\"http://localhost:5173\", \"http://127.0.0.1:5173\"\]|allow_origins=[\"http://$SERVER_IP\", \"http://localhost:5173\", \"http://127.0.0.1:5173\"]|g" $APP_DIR/backend_api.py

# Create systemd service
echo -e "${YELLOW}Creating systemd service...${NC}"
cat > /etc/systemd/system/trademicro.service << EOF
[Unit]
Description=TradeMicro FastAPI application
After=network.target

[Service]
User=root
Group=root
WorkingDirectory=$APP_DIR
Environment="PATH=$APP_DIR/venv/bin"
EnvironmentFile=$APP_DIR/.env
ExecStart=$APP_DIR/venv/bin/uvicorn backend_api:app --host 0.0.0.0 --port $APP_PORT

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo -e "${YELLOW}Enabling and starting service...${NC}"
systemctl daemon-reload
systemctl enable trademicro
systemctl start trademicro

# Configure Nginx
echo -e "${YELLOW}Configuring Nginx...${NC}"
cat > /etc/nginx/sites-available/trademicro << EOF
server {
    listen 80;
    server_name $SERVER_IP;

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
    echo -e "${GREEN}Your API is accessible at: http://$SERVER_IP${NC}"
    echo -e "${GREEN}API documentation is available at: http://$SERVER_IP/docs${NC}"
    echo -e "\n${YELLOW}Database connection:${NC}"
    echo -e "Host: $DB_HOST"
    echo -e "Database: $DB_NAME"
    echo -e "User: $DB_USER"
    
    echo -e "\n${YELLOW}To check service status:${NC}"
    echo -e "systemctl status trademicro"
    
    echo -e "\n${YELLOW}To view logs:${NC}"
    echo -e "journalctl -u trademicro"
else
    echo -e "${RED}Deployment failed. Service is not running.${NC}"
    echo -e "${YELLOW}Check logs with: journalctl -u trademicro${NC}"
fi
