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
API_PORT=8000

echo -e "${YELLOW}TradeMicro API Domain Mapping Script${NC}"
echo -e "${YELLOW}==================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create API mapping script
echo -e "${YELLOW}Creating API mapping script...${NC}"
cat > api_mapping.sh << 'EOF'
#!/bin/bash

# Configuration
DOMAIN="DOMAIN_PLACEHOLDER"
API_PORT=8000

# Update Nginx configuration to map API endpoints
cat > /etc/nginx/sites-available/trademicro << NGINX_CONF
server {
    listen 80;
    server_name ${DOMAIN};
    
    # Redirect all HTTP traffic to HTTPS
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    server_name ${DOMAIN};
    
    # SSL configuration (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
    
    # API root endpoint
    location / {
        proxy_pass http://localhost:${API_PORT};
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
    
    # API endpoints
    location /api/ {
        proxy_pass http://localhost:${API_PORT}/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
    
    # Health check endpoint
    location /api/health {
        proxy_pass http://localhost:${API_PORT}/api/health;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        # Add CORS headers for health check
        add_header 'Access-Control-Allow-Origin' '*';
        add_header 'Access-Control-Allow-Methods' 'GET, OPTIONS';
        add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range';
    }
    
    # Task management endpoints
    location /api/tasks {
        proxy_pass http://localhost:${API_PORT}/api/tasks;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        # Add CORS headers for task API
        add_header 'Access-Control-Allow-Origin' '*';
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS';
        add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization';
    }
}
NGINX_CONF

# Test Nginx configuration
nginx -t

# Reload Nginx to apply changes
systemctl restart nginx

echo "API endpoints have been mapped to ${DOMAIN}!"
echo "API is now accessible at https://${DOMAIN}/api/"
echo "Health check endpoint: https://${DOMAIN}/api/health"
echo "Tasks endpoint: https://${DOMAIN}/api/tasks"
EOF

# Replace placeholder with actual domain
sed -i "" "s/DOMAIN_PLACEHOLDER/$DOMAIN/g" api_mapping.sh
chmod +x api_mapping.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying API mapping script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE api_mapping.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Mapping API endpoints on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/api_mapping.sh" --quiet

# Clean up local files
rm -f api_mapping.sh

echo -e "\n${GREEN}API mapping completed!${NC}"
echo -e "${YELLOW}Your TradeMicro API is now accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/api/${NC}"
echo -e "\n${YELLOW}API Endpoints:${NC}"
echo -e "${GREEN}Health Check: https://$DOMAIN/api/health${NC}"
echo -e "${GREEN}Tasks: https://$DOMAIN/api/tasks${NC}"

# Test the API health endpoint
echo -e "\n${YELLOW}Testing API health endpoint...${NC}"
curl -s https://$DOMAIN/api/health || echo -e "${RED}Could not connect to API health endpoint${NC}"
