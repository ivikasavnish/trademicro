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

echo -e "${YELLOW}TradeMicro API Router Order Fix${NC}"
echo -e "${YELLOW}=============================${NC}\n"

# Create a modified version of the main.go file with the catch-all handler moved to the end
echo -e "${YELLOW}Creating router fix script...${NC}"
cat > router_fix.sh << 'EOF'
#!/bin/bash

# Stop the TradeMicro service
systemctl stop trademicro

# Create a backup of the original main.go file
cp /opt/trademicro/main.go /opt/trademicro/main.go.bak

# Extract the catch-all handler and remove it from its current position
grep -A 10 "r.PathPrefix(\"/\").HandlerFunc" /opt/trademicro/main.go > /tmp/catchall_handler.txt
sed -i '/r.PathPrefix("\\/").HandlerFunc/,/})/d' /opt/trademicro/main.go

# Add the catch-all handler at the end of the router setup, just before starting the server
sed -i '/\/\/ Start the WebSocket broadcast handler/i \	// Catch-all handler for serving the SPA\n\tr.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n\t\t// For API requests, let the router handle them\n\t\tif strings.HasPrefix(r.URL.Path, "/api/") {\n\t\t\thttp.NotFound(w, r)\n\t\t\treturn\n\t\t}\n\n\t\t// For all other requests, serve the index.html file\n\t\thttp.ServeFile(w, r, "web/index.html")\n\t})\n' /opt/trademicro/main.go

# Rebuild and restart the TradeMicro service
cd /opt/trademicro
go build -o trademicro
chmod +x trademicro
systemctl start trademicro

# Wait for the service to start
sleep 5

# Test the login endpoint
echo "Testing login endpoint directly on server..."
curl -s -X POST -H "Content-Type: application/json" -d '{"username":"vikasavnish","password":"Servloci@54321"}' http://localhost:8000/api/login
echo

# Check the service status
systemctl status trademicro
EOF

chmod +x router_fix.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying router fix script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE router_fix.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Applying router fix on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/router_fix.sh" --quiet

# Clean up local files
rm -f router_fix.sh

echo -e "\n${GREEN}Router order fix completed!${NC}"
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
