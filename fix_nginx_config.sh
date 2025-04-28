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

echo -e "${YELLOW}TradeMicro API Nginx Configuration Fix${NC}"
echo -e "${YELLOW}====================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create Nginx configuration fix script
echo -e "${YELLOW}Creating Nginx configuration fix script...${NC}"
cat > nginx_fix.sh << 'EOF'
#!/bin/bash

# Create updated Nginx configuration
cat > /etc/nginx/sites-available/default << 'NGINX_CONF'
server {
    listen 80;
    server_name trade.servloci.in;

    # Redirect HTTP to HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name trade.servloci.in;

    ssl_certificate /etc/letsencrypt/live/trade.servloci.in/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/trade.servloci.in/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    ssl_session_tickets off;

    # API endpoints
    location /api/ {
        proxy_pass http://localhost:8000/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Static files
    location /static/ {
        proxy_pass http://localhost:8000/static/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Root path - serve the SPA
    location / {
        proxy_pass http://localhost:8000/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
NGINX_CONF

# Test Nginx configuration
nginx -t

# Reload Nginx if configuration is valid
if [ $? -eq 0 ]; then
    systemctl reload nginx
    echo "Nginx configuration updated and reloaded successfully."
else
    echo "Error in Nginx configuration. Please check the syntax."
    exit 1
fi

# Test API endpoints
echo "Testing API endpoints..."
echo "Health endpoint: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/api/health)"
echo "Login endpoint: $(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"username":"test","password":"test"}' http://localhost:8000/api/login)"
EOF

chmod +x nginx_fix.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying Nginx fix script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE nginx_fix.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Applying Nginx configuration fix on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/nginx_fix.sh" --quiet

# Clean up local files
rm -f nginx_fix.sh

echo -e "\n${GREEN}Nginx configuration fix completed!${NC}"
echo -e "${YELLOW}Testing API login endpoint...${NC}"
curl -s -X POST -H "Content-Type: application/json" -d '{"username":"vikasavnish","password":"Servloci@54321"}' https://$DOMAIN/api/login

echo -e "\n\n${GREEN}Your TradeMicro API is now fully configured and accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/api/${NC}"
echo -e "\n${GREEN}Your TradeMicro Web UI is now accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/${NC}"
echo -e "\n${YELLOW}API Endpoints:${NC}"
echo -e "${GREEN}Health Check: https://$DOMAIN/api/health${NC}"
echo -e "${GREEN}Login: https://$DOMAIN/api/login${NC}"
echo -e "${GREEN}Tasks: https://$DOMAIN/api/tasks${NC}"
echo -e "\n${YELLOW}Login Credentials:${NC}"
echo -e "${GREEN}Username: vikasavnish${NC}"
echo -e "${GREEN}Password: Servloci@54321${NC}"
