#!/bin/bash

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration
GCP_PROJECT_ID="amiable-alcove-456605-k2"
GCE_INSTANCE="instance-20250422-132526"
GCE_INSTANCE_ZONE="asia-south1-c"
APP_DIR="/opt/trademicro"
APP_PORT=8000

echo -e "${YELLOW}TradeMicro API Deployment Script${NC}"
echo -e "${YELLOW}=============================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Build the Go binary
echo -e "\n${YELLOW}Building TradeMicro API...${NC}"
go build -v -o trademicro
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful!${NC}\n"

# Check if this is a first-time deployment
echo -e "${YELLOW}Checking if this is a first-time deployment...${NC}"
FIRST_DEPLOY=$(gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="test -f $APP_DIR/trademicro || echo 'first'" || echo "first")

# Create environment variables file
echo -e "${YELLOW}Creating environment file...${NC}"
cat > .env.deploy << EOF
POSTGRES_URL=YOUR_POSTGRES_CONNECTION_STRING
REDIS_URL=YOUR_REDIS_URL
SECRET_KEY=YOUR_SECRET_KEY
PORT=$APP_PORT
SERVER_ROLE=micro
WORKER_HOST=34.47.230.168
WORKER_USER=root
WORKER_SSH_KEY=$APP_DIR/.ssh/worker_key
TASK_LOG_DIR=$APP_DIR/logs
EOF

echo -e "${YELLOW}Please edit the .env.deploy file to add your database credentials${NC}"
echo -e "${YELLOW}Press Enter when ready to continue...${NC}"
read

# Create setup script for first-time deployment
if [[ $FIRST_DEPLOY == *"first"* ]]; then
    echo -e "${YELLOW}Creating first-time setup script...${NC}"
    cat > server_setup.sh << 'EOF'
#!/bin/bash

# TradeMicro API Server Setup Script

# Configuration
APP_DIR="/opt/trademicro"
APP_PORT=8000

# Create application directory
mkdir -p $APP_DIR
mkdir -p $APP_DIR/logs
mkdir -p $APP_DIR/.ssh

# Copy the binary
cp /tmp/trademicro $APP_DIR/
chmod +x $APP_DIR/trademicro

# Copy the environment file
cp /tmp/.env.deploy $APP_DIR/.env

# Create systemd service
cat > /etc/systemd/system/trademicro.service << SERVICEEOF
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
SERVICEEOF

# Enable and start service
systemctl daemon-reload
systemctl enable trademicro
systemctl start trademicro

# Generate SSH key for worker communication
ssh-keygen -t rsa -b 4096 -f $APP_DIR/.ssh/worker_key -N "" -C "micro-to-worker"
echo "Worker SSH public key (add this to worker instance authorized_keys):"
cat $APP_DIR/.ssh/worker_key.pub

echo "TradeMicro API has been set up successfully!"
EOF
    chmod +x server_setup.sh

    echo -e "${YELLOW}Performing first-time deployment...${NC}"
    
    # Copy files to the instance
    echo -e "${YELLOW}Copying files to instance...${NC}"
    gcloud compute scp --zone $GCE_INSTANCE_ZONE trademicro $GCE_INSTANCE:/tmp/ --quiet
    gcloud compute scp --zone $GCE_INSTANCE_ZONE server_setup.sh $GCE_INSTANCE:/tmp/ --quiet
    gcloud compute scp --zone $GCE_INSTANCE_ZONE .env.deploy $GCE_INSTANCE:/tmp/ --quiet
    
    # Execute setup script
    echo -e "${YELLOW}Executing setup script...${NC}"
    gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo chmod +x /tmp/server_setup.sh && sudo /tmp/server_setup.sh" --quiet
    
    echo -e "\n${GREEN}First-time deployment completed!${NC}"
    echo -e "${YELLOW}Important: Copy the worker SSH public key displayed above and add it to the worker instance's authorized_keys file.${NC}"
    echo -e "${YELLOW}You can connect to the worker instance with: ssh root@34.47.230.168${NC}"
    echo -e "${YELLOW}Then add the key to: /root/.ssh/authorized_keys${NC}"
    
    # Clean up local files
    rm -f server_setup.sh
else
    echo -e "${YELLOW}Performing update deployment...${NC}"
    
    # Copy binary to the instance
    echo -e "${YELLOW}Copying binary to instance...${NC}"
    gcloud compute scp --zone $GCE_INSTANCE_ZONE trademicro $GCE_INSTANCE:/tmp/ --quiet
    
    # Update binary and restart service
    echo -e "${YELLOW}Updating binary and restarting service...${NC}"
    gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo systemctl stop trademicro && sudo cp /tmp/trademicro $APP_DIR/ && sudo chmod +x $APP_DIR/trademicro && sudo systemctl start trademicro" --quiet
    
    echo -e "\n${GREEN}Update deployment completed!${NC}"
fi

# Clean up local files
rm -f .env.deploy
rm -f trademicro

echo -e "\n${GREEN}Deployment process finished!${NC}"
echo -e "${YELLOW}You can check the status of your application with:${NC}"
echo -e "${GREEN}gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command=\"sudo systemctl status trademicro\"${NC}"
