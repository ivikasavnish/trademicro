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
API_PORT=8000

echo -e "${YELLOW}TradeMicro API Nginx Setup Script${NC}"
echo -e "${YELLOW}=================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create Nginx configuration script
echo -e "${YELLOW}Creating Nginx setup script...${NC}"
cat > nginx_setup.sh << 'EOF'
#!/bin/bash

# Install Nginx if not already installed
if ! command -v nginx &> /dev/null; then
    echo "Installing Nginx..."
    apt-get update
    apt-get install -y nginx
fi

# Create Nginx configuration for TradeMicro API
cat > /etc/nginx/sites-available/trademicro << 'NGINX_CONF'
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
NGINX_CONF

# Enable the site
ln -sf /etc/nginx/sites-available/trademicro /etc/nginx/sites-enabled/

# Remove default site if it exists
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
nginx -t

# Reload Nginx to apply changes
systemctl restart nginx

# Enable Nginx to start on boot
systemctl enable nginx

# Check status
systemctl status nginx --no-pager

# Open firewall for HTTP and HTTPS
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full'
fi

echo "Nginx has been configured as a reverse proxy for TradeMicro API!"
EOF

chmod +x nginx_setup.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying Nginx setup script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE nginx_setup.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Setting up Nginx on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/nginx_setup.sh" --quiet

# Clean up local files
rm -f nginx_setup.sh

# Get the external IP of the instance
EXTERNAL_IP=$(gcloud compute instances describe $GCE_INSTANCE --zone $GCE_INSTANCE_ZONE --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo -e "\n${GREEN}Nginx setup completed!${NC}"
echo -e "${YELLOW}Your TradeMicro API is now accessible at:${NC}"
echo -e "${GREEN}http://$EXTERNAL_IP/${NC}"

# Check if firewall rule for HTTP exists, if not create it
if ! gcloud compute firewall-rules list --filter="name=allow-http" --format="get(name)" | grep -q "allow-http"; then
    echo -e "${YELLOW}Creating firewall rule to allow HTTP traffic...${NC}"
    gcloud compute firewall-rules create allow-http \
        --project=$GCP_PROJECT_ID \
        --direction=INGRESS \
        --priority=1000 \
        --network=default \
        --action=ALLOW \
        --rules=tcp:80 \
        --source-ranges=0.0.0.0/0
    echo -e "${GREEN}Firewall rule created!${NC}"
else
    echo -e "${GREEN}HTTP firewall rule already exists.${NC}"
fi
