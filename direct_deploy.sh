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

echo -e "${YELLOW}TradeMicro API Direct Deployment Script${NC}"
echo -e "${YELLOW}=====================================${NC}\n"

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
export CGO_ENABLED=0  # Disable CGO for static linking

# Build the binary with explicit flags
go build -v -a -ldflags '-extldflags "-static"' -o trademicro
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful for x86-64 architecture!${NC}\n"

# Verify the binary format
file trademicro

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

# Create a temporary directory for deployment files
TMP_DIR=$(mktemp -d)
echo -e "${YELLOW}Created temporary directory: $TMP_DIR${NC}"

# Copy files to temporary directory
cp trademicro $TMP_DIR/
cp .env.deploy $TMP_DIR/

# Create setup script for deployment
cat > $TMP_DIR/setup.sh << 'EOF'
#!/bin/bash

# TradeMicro API Server Setup Script

# Configuration
APP_DIR="/opt/trademicro"

# Create application directory
mkdir -p $APP_DIR
mkdir -p $APP_DIR/logs
mkdir -p $APP_DIR/.ssh

# Copy the binary
cp /tmp/deployment/trademicro $APP_DIR/
chmod +x $APP_DIR/trademicro

# Copy the environment file
cp /tmp/deployment/.env.deploy $APP_DIR/.env

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

# Generate SSH key for worker communication if it doesn't exist
if [ ! -f "$APP_DIR/.ssh/worker_key" ]; then
    ssh-keygen -t rsa -b 4096 -f $APP_DIR/.ssh/worker_key -N "" -C "micro-to-worker"
    echo "Worker SSH public key (add this to worker instance authorized_keys):"
    cat $APP_DIR/.ssh/worker_key.pub
fi

# Check service status
systemctl status trademicro --no-pager

echo "TradeMicro API has been set up successfully!"
EOF

chmod +x $TMP_DIR/setup.sh

# Create a tarball of all deployment files
echo -e "${YELLOW}Creating deployment package...${NC}"
tar -czf deployment.tar.gz -C $TMP_DIR .

# Copy the tarball to the instance
echo -e "${YELLOW}Copying deployment package to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE deployment.tar.gz $GCE_INSTANCE:/tmp/ --quiet

# Extract and run setup script on the instance
echo -e "${YELLOW}Extracting and running setup script on instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="mkdir -p /tmp/deployment && tar -xzf /tmp/deployment.tar.gz -C /tmp/deployment && chmod +x /tmp/deployment/setup.sh && sudo /tmp/deployment/setup.sh" --quiet

# Clean up local files
rm -rf $TMP_DIR
rm -f deployment.tar.gz
rm -f trademicro

echo -e "\n${GREEN}Deployment process finished!${NC}"
echo -e "${YELLOW}You can check the status of your application with:${NC}"
echo -e "${GREEN}gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command=\"sudo systemctl status trademicro\"${NC}"
