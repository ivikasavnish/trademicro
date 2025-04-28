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
DOMAIN="trade.servloci.in"

echo -e "${YELLOW}TradeMicro API Deployment with Restart${NC}"
echo -e "${YELLOW}=====================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Build the application for Linux (cross-compilation)
echo -e "${YELLOW}Building TradeMicro API for Linux...${NC}"
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o trademicro -a -ldflags '-extldflags "-static"' .

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Create a deployment package
echo -e "${YELLOW}Creating deployment package...${NC}"
tar -czf deploy.tar.gz trademicro api

# Copy the deployment package to the instance
echo -e "${YELLOW}Copying deployment package to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE deploy.tar.gz $GCE_INSTANCE:/tmp/ --quiet

# Create deployment script
echo -e "${YELLOW}Creating deployment script...${NC}"
cat > deploy.sh << 'EOF'
#!/bin/bash

# Stop the service
systemctl stop trademicro

# Extract the deployment package
tar -xzf /tmp/deploy.tar.gz -C /tmp/

# Create API directory if it doesn't exist
mkdir -p /opt/trademicro/api

# Copy the binary and API files
cp /tmp/trademicro /opt/trademicro/
cp -r /tmp/api/* /opt/trademicro/api/

# Set permissions
chmod +x /opt/trademicro/trademicro

# Start the service
systemctl start trademicro

# Wait for the service to start
sleep 5

# Check the service status
systemctl status trademicro

# Test the health endpoint
echo "\nTesting health endpoint:\n"
curl -s http://localhost:8000/api/health
EOF

chmod +x deploy.sh

# Copy the deployment script to the instance
echo -e "${YELLOW}Copying deployment script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE deploy.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the deployment script on the instance
echo -e "${YELLOW}Deploying TradeMicro API...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/deploy.sh" --quiet

# Clean up local files
rm -f trademicro deploy.tar.gz deploy.sh

echo -e "\n${GREEN}Deployment completed!${NC}"
echo -e "${YELLOW}Testing API health endpoint...${NC}"
curl -s https://$DOMAIN/api/health

echo -e "\n\n${GREEN}Your TradeMicro API is now fully configured and accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/api/${NC}"
echo -e "\n${YELLOW}API Endpoints:${NC}"
echo -e "${GREEN}Health Check: https://$DOMAIN/api/health${NC}"
echo -e "${GREEN}Tasks: https://$DOMAIN/api/tasks${NC}"
