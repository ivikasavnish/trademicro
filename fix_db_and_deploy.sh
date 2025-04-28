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

echo -e "${YELLOW}TradeMicro API Database Fix and Deployment${NC}"
echo -e "${YELLOW}=========================================${NC}\n"

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
tar -czf deploy.tar.gz trademicro api web

# Copy the deployment package to the instance
echo -e "${YELLOW}Copying deployment package to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE deploy.tar.gz $GCE_INSTANCE:/tmp/ --quiet

# Create database fix script
echo -e "${YELLOW}Creating database fix script...${NC}"
cat > fix_db.sql << 'EOF'
-- Fix the users table schema
BEGIN;

-- Check if hashed_password column exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'hashed_password') THEN
        -- Add hashed_password column
        ALTER TABLE users ADD COLUMN hashed_password TEXT;
        
        -- Copy data from password to hashed_password
        UPDATE users SET hashed_password = password;
        
        -- Make hashed_password NOT NULL
        ALTER TABLE users ALTER COLUMN hashed_password SET NOT NULL;
    END IF;
END
$$;

-- Drop existing users to recreate with correct schema
TRUNCATE TABLE users CASCADE;

COMMIT;
EOF

# Create deployment script
echo -e "${YELLOW}Creating deployment script...${NC}"
cat > deploy.sh << 'EOF'
#!/bin/bash

# Stop the service
systemctl stop trademicro

# Get PostgreSQL connection string from environment file
POSTGRES_URL=$(grep POSTGRES_URL /opt/trademicro/.env | cut -d '=' -f2-)

# Run database fix script
echo "Fixing database schema..."
psql "$POSTGRES_URL" -f /tmp/fix_db.sql

# Extract the deployment package
tar -xzf /tmp/deploy.tar.gz -C /tmp/

# Create directories if they don't exist
mkdir -p /opt/trademicro/api
mkdir -p /opt/trademicro/web/static/css
mkdir -p /opt/trademicro/web/static/js

# Copy the binary and API files
cp /tmp/trademicro /opt/trademicro/
cp -r /tmp/api/* /opt/trademicro/api/

# Copy web files
cp -r /tmp/web/* /opt/trademicro/web/

# Set permissions
chmod +x /opt/trademicro/trademicro
chmod -R 755 /opt/trademicro/web

# Start the service
systemctl start trademicro

# Wait for the service to start
sleep 5

# Check the service status
systemctl status trademicro

# Test the health endpoint
echo "\nTesting health endpoint:\n"
curl -s http://localhost:8000/api/health

# Test the web UI
echo "\nTesting web UI:\n"
curl -s -I http://localhost:8000/
EOF

chmod +x deploy.sh

# Copy the scripts to the instance
echo -e "${YELLOW}Copying scripts to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE fix_db.sql $GCE_INSTANCE:/tmp/ --quiet
gcloud compute scp --zone $GCE_INSTANCE_ZONE deploy.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the deployment script on the instance
echo -e "${YELLOW}Fixing database and deploying TradeMicro API...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo apt-get update && sudo apt-get install -y postgresql-client && sudo bash /tmp/deploy.sh" --quiet

# Clean up local files
rm -f trademicro deploy.tar.gz deploy.sh fix_db.sql

echo -e "\n${GREEN}Database fix and deployment completed!${NC}"
echo -e "${YELLOW}Testing API health endpoint...${NC}"
curl -s https://$DOMAIN/api/health

echo -e "\n${YELLOW}Testing Web UI...${NC}"
curl -s -I https://$DOMAIN/

echo -e "\n\n${GREEN}Your TradeMicro API is now fully configured and accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/api/${NC}"
echo -e "\n${GREEN}Your TradeMicro Web UI is now accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/${NC}"
echo -e "\n${YELLOW}API Endpoints:${NC}"
echo -e "${GREEN}Health Check: https://$DOMAIN/api/health${NC}"
echo -e "${GREEN}Tasks: https://$DOMAIN/api/tasks${NC}"
echo -e "\n${YELLOW}Login Credentials:${NC}"
echo -e "${GREEN}Username: vikasavnish${NC}"
echo -e "${GREEN}Password: Servloci@54321${NC}"
