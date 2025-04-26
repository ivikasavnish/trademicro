#!/bin/bash

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration
SERVER_IP="62.164.218.111"
SERVER_USER="root"
SERVER_PASSWORD="Sonam@7512"

# Your PostgreSQL DSN
POSTGRES_DSN="your-postgres-dsn-here"

echo -e "${GREEN}Building and deploying TradeMicro Go API...${NC}"

# Step 1: Build the Go binary
echo -e "${YELLOW}Building Go binary...${NC}"
go mod download
go build -o trademicro main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed. Please fix the errors and try again.${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Step 2: Update the deployment script with your PostgreSQL DSN
echo -e "${YELLOW}Updating deployment script with your PostgreSQL DSN...${NC}"
sed -i "" "s|POSTGRES_DSN=\"your-postgres-dsn-here\"|POSTGRES_DSN=\"$POSTGRES_DSN\"|g" deploy_go.sh

# Step 3: Copy files to the server
echo -e "${YELLOW}Copying files to the server...${NC}"
scp -o StrictHostKeyChecking=no trademicro deploy_go.sh $SERVER_USER@$SERVER_IP:/tmp/

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to copy files to the server. Please check your server credentials and connectivity.${NC}"
    exit 1
fi

# Step 4: Execute the deployment script on the server
echo -e "${YELLOW}Executing deployment script on the server...${NC}"
ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "cd /tmp && chmod +x deploy_go.sh && ./deploy_go.sh"

if [ $? -ne 0 ]; then
    echo -e "${RED}Deployment failed. Please check the server logs for more information.${NC}"
    exit 1
fi

echo -e "${GREEN}Deployment completed successfully!${NC}"
echo -e "${GREEN}Your API is now accessible at: http://$SERVER_IP${NC}"
