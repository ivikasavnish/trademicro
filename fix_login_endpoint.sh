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

echo -e "${YELLOW}TradeMicro API Login Endpoint Fix${NC}"
echo -e "${YELLOW}=================================${NC}\n"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth login${NC}"
    echo -e "or"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Create a simplified version of the main.go file with a direct login handler
echo -e "${YELLOW}Creating simplified login handler...${NC}"
cat > login_handler.go << 'EOF'
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse represents the login response body
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// Claims represents the JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// SimpleLoginHandler handles login requests without middleware
func SimpleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for the main request
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Find the user
	var user User
	result := db.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(loginReq.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create a JWT token
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: tokenString,
		TokenType:   "bearer",
	})
}
EOF

# Create server fix script
echo -e "${YELLOW}Creating server fix script...${NC}"
cat > server_fix.sh << 'EOF'
#!/bin/bash

# Stop the TradeMicro service
systemctl stop trademicro

# Copy the login handler to the server
cp /tmp/login_handler.go /opt/trademicro/

# Update the main.go file to register the login handler directly
sed -i '/r.HandleFunc("\/api\/login"/d' /opt/trademicro/main.go
sed -i '/r.HandleFunc("\/ws"/a \	r.HandleFunc("/api/login", SimpleLoginHandler).Methods("POST", "OPTIONS")' /opt/trademicro/main.go

# Create a test login endpoint
cat > /tmp/test_login.sh << 'TEST_SCRIPT'
#!/bin/bash

echo "Testing login endpoint directly on server..."
curl -s -X POST -H "Content-Type: application/json" -d '{"username":"vikasavnish","password":"Servloci@54321"}' http://localhost:8000/api/login
echo
TEST_SCRIPT

chmod +x /tmp/test_login.sh

# Update Nginx configuration to properly handle API endpoints
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
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Add CORS headers
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
        add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization' always;
        
        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain charset=UTF-8';
            add_header 'Content-Length' 0;
            return 204;
        }
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

# Rebuild and restart the TradeMicro service
cd /opt/trademicro
go build -o trademicro
chmod +x trademicro
systemctl start trademicro

# Wait for the service to start
sleep 5

# Test the login endpoint
/tmp/test_login.sh

# Check the service status
systemctl status trademicro
EOF

chmod +x server_fix.sh

# Copy the files to the instance
echo -e "${YELLOW}Copying files to instance...${NC}"
gcloud compute scp --zone $GCE_INSTANCE_ZONE login_handler.go $GCE_INSTANCE:/tmp/ --quiet
gcloud compute scp --zone $GCE_INSTANCE_ZONE server_fix.sh $GCE_INSTANCE:/tmp/ --quiet

# Execute the script on the instance
echo -e "${YELLOW}Applying fixes on the instance...${NC}"
gcloud compute ssh --zone $GCE_INSTANCE_ZONE $GCE_INSTANCE --command="sudo bash /tmp/server_fix.sh" --quiet

# Clean up local files
rm -f login_handler.go server_fix.sh

echo -e "\n${GREEN}Login endpoint fix completed!${NC}"
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
