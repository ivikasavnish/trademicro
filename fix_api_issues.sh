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

echo -e "${YELLOW}TradeMicro API Fix Script${NC}"
echo -e "${YELLOW}=======================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create API fix script
echo -e "${YELLOW}Creating API fix script...${NC}"
cat > api_fix.sh << 'EOF'
#!/bin/bash

# Install Redis if not already installed
if ! command -v redis-server &> /dev/null; then
    echo "Installing Redis..."
    apt-get update
    apt-get install -y redis-server
    
    # Configure Redis to listen on all interfaces
    sed -i 's/bind 127.0.0.1 ::1/bind 0.0.0.0/g' /etc/redis/redis.conf
    
    # Restart Redis to apply changes
    systemctl restart redis-server
    systemctl enable redis-server
    
    echo "Redis installed and configured"
fi

# Update environment file to use the correct Redis URL
sed -i 's|redis://localhost:6379/0|redis://127.0.0.1:6379/0|g' /opt/trademicro/.env

# Create health endpoint handler
cat > /tmp/health_handler.go << 'GOCODE'
package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Message   string    `json:"message"`
}

// HealthHandler handles the health check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Message:   "TradeMicro API is running",
	}
	
	json.NewEncoder(w).Encode(response)
}
GOCODE

# Create directory if it doesn't exist
mkdir -p /opt/trademicro/api

# Copy the health handler to the API directory
cp /tmp/health_handler.go /opt/trademicro/api/health_handler.go

# Create main.go update script to add health endpoint
cat > /tmp/update_main.go << 'GOSCRIPT'
#!/bin/bash

# Check if the file exists
if [ ! -f "/opt/trademicro/main.go" ]; then
    echo "main.go not found"
    exit 1
fi

# Add health endpoint route
if ! grep -q "r.HandleFunc(\"/api/health\"" /opt/trademicro/main.go; then
    # Find the line where routes are defined
    LINE_NUM=$(grep -n "r := mux.NewRouter()" /opt/trademicro/main.go | cut -d ':' -f 1)
    if [ -z "$LINE_NUM" ]; then
        echo "Could not find router initialization"
        exit 1
    fi
    
    # Add the health endpoint after the router initialization
    HEALTH_ROUTE="\tr.HandleFunc(\"/api/health\", api.HealthHandler).Methods(\"GET\")"
    sed -i "$((LINE_NUM+2))i\$HEALTH_ROUTE" /opt/trademicro/main.go
    
    echo "Added health endpoint route"
fi
GOSCRIPT

chmod +x /tmp/update_main.go
/tmp/update_main.go

# Restart TradeMicro service
systemctl restart trademicro

# Wait for the service to start
sleep 5

# Test health endpoint
curl -s http://localhost:8000/api/health

echo "\nAPI fixes have been applied!"
EOF

chmod +x api_fix.sh

# Copy the script to the instance
echo -e "${YELLOW}Copying API fix script to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE api_fix.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Applying API fixes on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/api_fix.sh" --quiet

# Clean up local files
rm -f api_fix.sh

echo -e "\n${GREEN}API fixes completed!${NC}"
echo -e "${YELLOW}Testing API health endpoint...${NC}"
curl -s https://$DOMAIN/api/health

echo -e "\n\n${GREEN}Your TradeMicro API is now fully configured and accessible at:${NC}"
echo -e "${GREEN}https://$DOMAIN/api/${NC}"
echo -e "\n${YELLOW}API Endpoints:${NC}"
echo -e "${GREEN}Health Check: https://$DOMAIN/api/health${NC}"
echo -e "${GREEN}Tasks: https://$DOMAIN/api/tasks${NC}"
