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

echo -e "${YELLOW}TradeMicro API SSL Setup Script${NC}"
echo -e "${YELLOW}=============================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create Let's Encrypt SSL setup script
echo -e "${YELLOW}Creating Let's Encrypt setup script...${NC}"
cat > letsencrypt_setup.sh << EOF
#!/bin/bash

# Install Certbot and Nginx plugin
apt-get update
apt-get install -y certbot python3-certbot-nginx

# Update Nginx configuration to include the domain name
cat > /etc/nginx/sites-available/trademicro << 'NGINX_CONF'
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
NGINX_CONF

# Replace placeholder with actual domain
sed -i "s/\$DOMAIN/$DOMAIN/g" /etc/nginx/sites-available/trademicro

# Enable the site
ln -sf /etc/nginx/sites-available/trademicro /etc/nginx/sites-enabled/

# Remove default site if it exists
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
nginx -t

# Reload Nginx to apply changes
systemctl restart nginx

# Get SSL certificate from Let's Encrypt
certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@servloci.in

# Check Certbot timer for automatic renewal
systemctl status certbot.timer --no-pager

echo "Let's Encrypt SSL certificate has been set up for $DOMAIN!"
EOF

# Replace placeholder with actual domain
sed -i "" "s/\$DOMAIN/$DOMAIN/g" letsencrypt_setup.sh
chmod +x letsencrypt_setup.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying Let's Encrypt setup script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE letsencrypt_setup.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Setting up Let's Encrypt SSL on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/letsencrypt_setup.sh" --quiet

# Clean up local files
rm -f letsencrypt_setup.sh

echo -e "\n${GREEN}SSL setup completed!${NC}"
echo -e "${YELLOW}Your TradeMicro API is now accessible securely at:${NC}"
echo -e "${GREEN}https://$DOMAIN/${NC}"

# Check if firewall rule for HTTPS exists, if not create it
if ! gcloud compute firewall-rules list --filter="name=allow-https" --format="get(name)" | grep -q "allow-https"; then
    echo -e "${YELLOW}Creating firewall rule to allow HTTPS traffic...${NC}"
    gcloud compute firewall-rules create allow-https \
        --project=$GCP_PROJECT_ID \
        --direction=INGRESS \
        --priority=1000 \
        --network=default \
        --action=ALLOW \
        --rules=tcp:443 \
        --source-ranges=0.0.0.0/0
    echo -e "${GREEN}HTTPS firewall rule created!${NC}"
else
    echo -e "${GREEN}HTTPS firewall rule already exists.${NC}"
fi
