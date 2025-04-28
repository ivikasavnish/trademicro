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
LOCAL_PORT=8888
REMOTE_IP="35.244.30.157"

echo -e "${YELLOW}TradeMicro API Fast Deployment Script${NC}"
echo -e "${YELLOW}====================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Build the Go binary for x86-64 architecture
echo -e "\n${YELLOW}Building TradeMicro API for x86-64 architecture...${NC}"
# Set environment variables for cross-compilation
export GOOS=linux
export GOARCH=amd64

# Build the binary
go build -v -o trademicro
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful for x86-64 architecture!${NC}\n"

# Check if this is a first-time deployment
echo -e "${YELLOW}Checking if this is a first-time deployment...${NC}"
FIRST_DEPLOY=$(gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="test -f $APP_DIR/trademicro || echo 'first'" || echo "first")

# Create environment variables file if it doesn't exist
if [ ! -f ".env.deploy" ]; then
    echo -e "${YELLOW}Creating environment file...${NC}"
    cat > .env.deploy << EOF
PORT=8000
SERVER_ROLE=micro
WORKER_HOST=34.47.230.168
WORKER_USER=root
WORKER_SSH_KEY=$APP_DIR/.ssh/worker_key
TASK_LOG_DIR=$APP_DIR/logs
REDIS_URL=redis://localhost:6379/0
SECRET_KEY=6b55274139509604d9b7f5866f82d4f4bfb7044ba7b8d0ae3bccbe92df298d6d9cedae05b30b027637f097111bb0c700c1bc5478d4fc6000bb5a42aa2f54af1c
POSTGRES_URL=postgresql://neondb_owner:npg_DCdpL46RGfVr@ep-solitary-feather-a1y0yrpq-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require
EOF
    echo -e "${GREEN}Environment file created!${NC}"
fi

# Create setup script for first-time deployment
if [[ $FIRST_DEPLOY == *"first"* ]]; then
    echo -e "${YELLOW}Creating first-time setup script...${NC}"
    cat > server_setup.sh << 'EOF'
#!/bin/bash

# TradeMicro API Server Setup Script

# Configuration
APP_DIR="/opt/trademicro"

# Create application directory
mkdir -p $APP_DIR
mkdir -p $APP_DIR/logs
mkdir -p $APP_DIR/.ssh

# Copy the binary
mv /tmp/trademicro $APP_DIR/
chmod +x $APP_DIR/trademicro

# Copy the environment file
mv /tmp/.env.deploy $APP_DIR/.env

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
    
    # Create download script for first-time deployment
    cat > download_setup.sh << 'EOF'
#!/bin/bash

# Download binary and setup script
cd /tmp
wget http://LOCAL_IP:LOCAL_PORT/trademicro -O trademicro
wget http://LOCAL_IP:LOCAL_PORT/server_setup.sh -O server_setup.sh
wget http://LOCAL_IP:LOCAL_PORT/.env.deploy -O .env.deploy

# Make setup script executable
chmod +x /tmp/server_setup.sh

# Run setup script
sudo /tmp/server_setup.sh
EOF
    
    # Replace placeholders with actual values
    sed -i "" "s/LOCAL_IP/$LOCAL_IP/g" download_setup.sh
    sed -i "" "s/LOCAL_PORT/$LOCAL_PORT/g" download_setup.sh
    chmod +x download_setup.sh
    
    echo -e "${YELLOW}Starting local web server for file transfer...${NC}"
    # Start a local web server in the background
    python3 -m http.server $LOCAL_PORT &
    WEB_SERVER_PID=$!
    
    # Wait for web server to start
    sleep 2
    
    # Get local IP address
    LOCAL_IP=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)
    echo -e "${GREEN}Local web server started at http://$LOCAL_IP:$LOCAL_PORT${NC}"
    
    # Update download script with actual IP
    sed -i "" "s/LOCAL_IP/$LOCAL_IP/g" download_setup.sh
    
    # Copy download script to the instance
    echo -e "${YELLOW}Copying download script to instance...${NC}"
    gcloud compute scp --zone $GCE_INSTANCE_ZONE download_setup.sh $GCE_INSTANCE:/tmp/ --quiet
    
    # Execute download script
    echo -e "${YELLOW}Executing download script on remote instance...${NC}"
    gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="chmod +x /tmp/download_setup.sh && /tmp/download_setup.sh" --quiet
    
    # Kill the web server
    kill $WEB_SERVER_PID
    
    echo -e "\n${GREEN}First-time deployment completed!${NC}"
    echo -e "${YELLOW}Important: Copy the worker SSH public key displayed above and add it to the worker instance's authorized_keys file.${NC}"
    echo -e "${YELLOW}You can connect to the worker instance with: ssh root@34.47.230.168${NC}"
    echo -e "${YELLOW}Then add the key to: /root/.ssh/authorized_keys${NC}"
    
    # Clean up local files
    rm -f server_setup.sh download_setup.sh
else
    echo -e "${YELLOW}Performing update deployment...${NC}"
    
    # Create download script for update
    cat > download_update.sh << 'EOF'
#!/bin/bash

# Download binary
cd /tmp
wget http://LOCAL_IP:LOCAL_PORT/trademicro -O trademicro

# Update binary and restart service
sudo systemctl stop trademicro
sudo cp /tmp/trademicro /opt/trademicro/
sudo chmod +x /opt/trademicro/trademicro
sudo systemctl start trademicro
echo "TradeMicro API has been updated successfully!"
EOF
    
    # Replace placeholders with actual values
    sed -i "" "s/LOCAL_IP/$LOCAL_IP/g" download_update.sh
    sed -i "" "s/LOCAL_PORT/$LOCAL_PORT/g" download_update.sh
    chmod +x download_update.sh
    
    echo -e "${YELLOW}Starting local web server for file transfer...${NC}"
    # Start a local web server in the background
    python3 -m http.server $LOCAL_PORT &
    WEB_SERVER_PID=$!
    
    # Wait for web server to start
    sleep 2
    
    # Get local IP address
    LOCAL_IP=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)
    echo -e "${GREEN}Local web server started at http://$LOCAL_IP:$LOCAL_PORT${NC}"
    
    # Update download script with actual IP
    sed -i "" "s/LOCAL_IP/$LOCAL_IP/g" download_update.sh
    
    # Copy download script to the instance
    echo -e "${YELLOW}Copying download script to instance...${NC}"
    gcloud compute scp --zone $GCE_INSTANCE_ZONE download_update.sh $GCE_INSTANCE:/tmp/ --quiet
    
    # Execute download script
    echo -e "${YELLOW}Executing download script on remote instance...${NC}"
    gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="chmod +x /tmp/download_update.sh && /tmp/download_update.sh" --quiet
    
    # Kill the web server
    kill $WEB_SERVER_PID
    
    echo -e "\n${GREEN}Update deployment completed!${NC}"
    
    # Clean up local files
    rm -f download_update.sh
fi

# Clean up local files
rm -f trademicro

echo -e "\n${GREEN}Deployment process finished!${NC}"
echo -e "${YELLOW}You can check the status of your application with:${NC}"
echo -e "${GREEN}gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command=\"sudo systemctl status trademicro\"${NC}"
