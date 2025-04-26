#!/bin/bash

# TradeMicro API Deployer Script
# This script handles both local and server-side deployment tasks
# Includes SSH key setup for passwordless authentication

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
BLUE="\033[0;34m"
NC="\033[0m" # No Color

# Default configuration
SERVER_IP="62.164.218.111"
SERVER_USER="root"
SERVER_PASSWORD="Sonam@7512"
APP_DIR="/opt/trademicro"
APP_PORT=8000
REDIS_URL="redis://localhost:6379/0"
POSTGRES_DSN=""
SECRET_KEY=""
SETUP_SSH=false
SSH_KEY_PATH="$HOME/.ssh/trademicro_rsa"

# Function to display usage information
show_usage() {
    echo -e "\n${BLUE}TradeMicro API Deployer${NC}"
    echo -e "Usage: $0 [options]"
    echo -e "\nOptions:"
    echo -e "  -h, --help\t\tShow this help message"
    echo -e "  -s, --server IP\t\tServer IP address (default: $SERVER_IP)"
    echo -e "  -u, --user USERNAME\tServer username (default: $SERVER_USER)"
    echo -e "  -p, --port PORT\t\tApplication port (default: $APP_PORT)"
    echo -e "  -d, --dir DIRECTORY\tApplication directory on server (default: $APP_DIR)"
    echo -e "  --postgres-dsn DSN\tPostgreSQL connection string"
    echo -e "  --redis-url URL\t\tRedis URL (default: $REDIS_URL)"
    echo -e "  --setup-ssh\t\tGenerate and setup SSH keys for passwordless authentication"
    echo -e "  --ssh-key PATH\t\tPath to SSH private key (default: $SSH_KEY_PATH)"
    echo -e "  --skip-build\t\tSkip building the Go binary"
    echo -e "  --skip-copy\t\tSkip copying files to server"
    echo -e "  --skip-deploy\t\tSkip server-side deployment"
    echo -e "  --local-only\t\tBuild binary only, don't deploy to server"
    echo -e "\nExample:"
    echo -e "  $0 --postgres-dsn 'postgresql://user:pass@host:port/dbname'"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            ;;
        -s|--server)
            SERVER_IP="$2"
            shift 2
            ;;
        -u|--user)
            SERVER_USER="$2"
            shift 2
            ;;
        -p|--port)
            APP_PORT="$2"
            shift 2
            ;;
        -d|--dir)
            APP_DIR="$2"
            shift 2
            ;;
        --postgres-dsn)
            POSTGRES_DSN="$2"
            shift 2
            ;;
        --redis-url)
            REDIS_URL="$2"
            shift 2
            ;;
        --setup-ssh)
            SETUP_SSH=true
            shift
            ;;
        --ssh-key)
            SSH_KEY_PATH="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-copy)
            SKIP_COPY=true
            shift
            ;;
        --skip-deploy)
            SKIP_DEPLOY=true
            shift
            ;;
        --local-only)
            SKIP_COPY=true
            SKIP_DEPLOY=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            ;;
    esac
done

# Check for required PostgreSQL DSN
if [[ -z "$POSTGRES_DSN" ]]; then
    # Check if it's in .env file
    if [[ -f ".env" ]]; then
        POSTGRES_DSN=$(grep POSTGRES_URL .env | cut -d '=' -f2)
    fi
    
    # If still empty, prompt user
    if [[ -z "$POSTGRES_DSN" ]]; then
        echo -e "${YELLOW}PostgreSQL DSN not provided.${NC}"
        read -p "Enter PostgreSQL DSN (postgresql://user:pass@host:port/dbname): " POSTGRES_DSN
        
        if [[ -z "$POSTGRES_DSN" ]]; then
            echo -e "${RED}PostgreSQL DSN is required. Exiting.${NC}"
            exit 1
        fi
    fi
fi

# Generate a random secret key if not provided
if [[ -z "$SECRET_KEY" ]]; then
    SECRET_KEY=$(openssl rand -hex 32)
fi

# Display configuration
echo -e "\n${BLUE}TradeMicro API Deployment Configuration${NC}"
echo -e "Server IP:\t\t${YELLOW}$SERVER_IP${NC}"
echo -e "Server User:\t\t${YELLOW}$SERVER_USER${NC}"
echo -e "App Directory:\t${YELLOW}$APP_DIR${NC}"
echo -e "App Port:\t\t${YELLOW}$APP_PORT${NC}"
echo -e "PostgreSQL DSN:\t${YELLOW}$(echo $POSTGRES_DSN | sed 's/:.*/:*****@/')${NC}"
echo -e "Redis URL:\t\t${YELLOW}$REDIS_URL${NC}"

# Confirm deployment
read -p "Continue with deployment? (y/n): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo -e "${RED}Deployment cancelled.${NC}"
    exit 0
fi

# Function to setup SSH keys for passwordless authentication
setup_ssh_keys() {
    echo -e "\n${YELLOW}Setting up SSH keys for passwordless authentication...${NC}"
    
    # Check if ssh-keygen is installed
    if ! command -v ssh-keygen &> /dev/null; then
        echo -e "${RED}ssh-keygen is not installed. Please install OpenSSH and try again.${NC}"
        return 1
    fi
    
    # Check if sshpass is installed
    if ! command -v sshpass &> /dev/null; then
        echo -e "${YELLOW}Installing sshpass for password authentication...${NC}"
        brew install hudochenkov/sshpass/sshpass || { 
            echo -e "${RED}Failed to install sshpass. Please install it manually.${NC}"
            return 1
        }
    fi
    
    # Generate SSH key if it doesn't exist
    if [[ ! -f "$SSH_KEY_PATH" ]]; then
        echo -e "${YELLOW}Generating new SSH key at $SSH_KEY_PATH...${NC}"
        ssh-keygen -t rsa -b 4096 -f "$SSH_KEY_PATH" -N "" -C "trademicro-deployer"
        
        if [[ $? -ne 0 ]]; then
            echo -e "${RED}Failed to generate SSH key.${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}Using existing SSH key at $SSH_KEY_PATH...${NC}"
    fi
    
    # Copy SSH key to server
    echo -e "${YELLOW}Copying SSH public key to server...${NC}"
    
    # Create .ssh directory on server if it doesn't exist
    sshpass -p "$SERVER_PASSWORD" ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "mkdir -p ~/.ssh && chmod 700 ~/.ssh"
    
    # Copy public key to server
    sshpass -p "$SERVER_PASSWORD" ssh-copy-id -i "${SSH_KEY_PATH}.pub" -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP"
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Failed to copy SSH key to server.${NC}"
        return 1
    fi
    
    # Test SSH connection
    echo -e "${YELLOW}Testing SSH connection...${NC}"
    ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "echo 'SSH connection successful!'"
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}SSH connection test failed.${NC}"
        return 1
    fi
    
    echo -e "${GREEN}SSH keys setup successfully!${NC}"
    echo -e "${YELLOW}You can now use the --ssh-key option to authenticate without a password.${NC}"
    
    # Save SSH credentials to secrets file
    echo -e "${YELLOW}Saving SSH credentials to .github/secrets.env...${NC}"
    mkdir -p .github
    cat > .github/secrets.env << EOF
SERVER_HOST=$SERVER_IP
SERVER_USER=$SERVER_USER
SERVER_PASSWORD=$SERVER_PASSWORD
SSH_PRIVATE_KEY=$(cat "$SSH_KEY_PATH" | base64 -w 0)
POSTGRES_DSN=$POSTGRES_DSN
REDIS_URL=$REDIS_URL
SECRET_KEY=$SECRET_KEY
EOF
    
    echo -e "${GREEN}SSH credentials saved to .github/secrets.env${NC}"
    echo -e "${YELLOW}You can use this file to set up GitHub Actions secrets.${NC}"
    
    return 0
}

# Step 1: Setup SSH keys if requested
if [[ "$SETUP_SSH" == true ]]; then
    setup_ssh_keys
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}SSH key setup failed. Continuing with password authentication.${NC}"
    fi
fi

# Step 2: Build the Go binary
if [[ "$SKIP_BUILD" != true ]]; then
    echo -e "\n${YELLOW}Step 2: Building Go binary...${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Go is not installed. Please install Go and try again.${NC}"
        exit 1
    fi
    
    # Download dependencies and build
    go mod download
    go build -o trademicro main.go
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Build failed. Please fix the errors and try again.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Build successful!${NC}"
else
    echo -e "\n${YELLOW}Skipping build step...${NC}"
    
    # Check if binary exists
    if [[ ! -f "trademicro" ]]; then
        echo -e "${RED}Binary 'trademicro' not found. Run without --skip-build option.${NC}"
        exit 1
    fi
fi

# Create server deployment script
echo -e "\n${YELLOW}Creating server deployment script...${NC}"
cat > server_deploy.sh << EOF
#!/bin/bash

# TradeMicro API Server Deployment Script
# Generated by deployer.sh

# Configuration
APP_DIR="$APP_DIR"
APP_PORT=$APP_PORT
POSTGRES_DSN="$POSTGRES_DSN"
REDIS_URL="$REDIS_URL"
SECRET_KEY="$SECRET_KEY"

# Colors for output
GREEN="\\033[0;32m"
YELLOW="\\033[1;33m"
RED="\\033[0;31m"
NC="\\033[0m" # No Color

echo -e "${GREEN}Starting TradeMicro API server deployment...${NC}"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}This script must be run as root${NC}"
    exit 1
fi

# Update system packages
echo -e "${YELLOW}Updating system packages...${NC}"
apt update && apt upgrade -y

# Install Redis if not already installed
echo -e "${YELLOW}Installing Redis...${NC}"
apt install -y redis-server

# Create application directory
echo -e "${YELLOW}Creating application directory...${NC}"
mkdir -p $APP_DIR

# Check if the directory already has files
if [ "$(ls -A $APP_DIR)" ]; then
    echo -e "${YELLOW}Directory $APP_DIR is not empty. Backing up existing files...${NC}"
    BACKUP_DIR="${APP_DIR}_backup_$(date +%Y%m%d%H%M%S)"
    mv $APP_DIR $BACKUP_DIR
    mkdir -p $APP_DIR
fi

# Copy the Go binary to the server
echo -e "${YELLOW}Copying application binary...${NC}"
cp ./trademicro $APP_DIR/

# Create environment file
echo -e "${YELLOW}Creating environment file...${NC}"
cat > $APP_DIR/.env << ENVEOF
POSTGRES_URL=$POSTGRES_DSN
REDIS_URL=$REDIS_URL
SECRET_KEY=$SECRET_KEY
PORT=$APP_PORT
ENVEOF

# Create systemd service
echo -e "${YELLOW}Creating systemd service...${NC}"
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
echo -e "${YELLOW}Enabling and starting service...${NC}"
systemctl daemon-reload
systemctl enable trademicro
systemctl start trademicro

# Configure Nginx
echo -e "${YELLOW}Installing and configuring Nginx...${NC}"
apt install -y nginx

cat > /etc/nginx/sites-available/trademicro << NGINXEOF
server {
    listen 80;
    server_name \\\$hostname;

    location / {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \\\$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \\\$host;
        proxy_set_header X-Real-IP \\\$remote_addr;
    }
}
NGINXEOF

# Enable Nginx site
ln -sf /etc/nginx/sites-available/trademicro /etc/nginx/sites-enabled/
nginx -t && systemctl restart nginx

# Check if service is running
if systemctl is-active --quiet trademicro; then
    echo -e "${GREEN}TradeMicro API has been successfully deployed!${NC}"
    SERVER_IP=$(hostname -I | awk '{print $1}')
    echo -e "${GREEN}Your API is accessible at: http://$SERVER_IP${NC}"
    echo -e "${GREEN}API documentation is available at: http://$SERVER_IP/docs${NC}"
    
    echo -e "\\n${YELLOW}To check service status:${NC}"
    echo -e "systemctl status trademicro"
    
    echo -e "\\n${YELLOW}To view logs:${NC}"
    echo -e "journalctl -u trademicro"
else
    echo -e "${RED}Deployment failed. Service is not running.${NC}"
    echo -e "${YELLOW}Check logs with: journalctl -u trademicro${NC}"
fi
EOF

chmod +x server_deploy.sh

# Step 3: Copy files to the server
if [[ "$SKIP_COPY" != true ]]; then
    echo -e "\n${YELLOW}Step 3: Copying files to the server...${NC}"
    
    # Check if sshpass is installed for password-based authentication
    if [[ -n "$SERVER_PASSWORD" ]]; then
        if ! command -v sshpass &> /dev/null; then
            echo -e "${YELLOW}Installing sshpass for password authentication...${NC}"
            brew install hudochenkov/sshpass/sshpass || { 
                echo -e "${RED}Failed to install sshpass. Please install it manually or use SSH keys.${NC}"
                exit 1
            }
        fi
        
        # Copy files using sshpass
        sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no trademicro server_deploy.sh "$SERVER_USER@$SERVER_IP:/tmp/"
    else
        # Copy files using regular scp (requires SSH key authentication)
        if [[ -f "$SSH_KEY_PATH" ]]; then
            scp -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no trademicro server_deploy.sh "$SERVER_USER@$SERVER_IP:/tmp/"
        elif [[ -n "$SERVER_PASSWORD" ]]; then
            sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no trademicro server_deploy.sh "$SERVER_USER@$SERVER_IP:/tmp/"
        else
            # Copy files using regular scp (requires SSH key authentication)
            scp -o StrictHostKeyChecking=no trademicro server_deploy.sh "$SERVER_USER@$SERVER_IP:/tmp/"
        fi
    fi
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Failed to copy files to the server. Please check your server credentials and connectivity.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Files copied successfully!${NC}"
else
    echo -e "\n${YELLOW}Skipping copy step...${NC}"
fi

# Step 4: Execute the deployment script on the server
if [[ "$SKIP_DEPLOY" != true ]]; then
    echo -e "\n${YELLOW}Step 4: Executing deployment script on the server...${NC}"
    
    # Execute script using SSH key if available, otherwise use password
    if [[ -f "$SSH_KEY_PATH" ]]; then
        ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "cd /tmp && chmod +x server_deploy.sh && ./server_deploy.sh"
    elif [[ -n "$SERVER_PASSWORD" ]]; then
        sshpass -p "$SERVER_PASSWORD" ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "cd /tmp && chmod +x server_deploy.sh && ./server_deploy.sh"
    else
        # Execute script using regular ssh (requires SSH key authentication)
        ssh -o StrictHostKeyChecking=no "$SERVER_USER@$SERVER_IP" "cd /tmp && chmod +x server_deploy.sh && ./server_deploy.sh"
    fi
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Deployment failed. Please check the server logs for more information.${NC}"
        exit 1
    fi
    
    echo -e "\n${GREEN}Deployment completed successfully!${NC}"
    echo -e "${GREEN}Your API is now accessible at: http://$SERVER_IP${NC}"
    echo -e "${GREEN}API documentation is available at: http://$SERVER_IP/docs${NC}"
else
    echo -e "\n${YELLOW}Skipping server deployment step...${NC}"
    
    if [[ "$SKIP_COPY" == true ]]; then
        echo -e "${GREEN}Local build completed successfully!${NC}"
        echo -e "${YELLOW}To deploy manually, copy the 'trademicro' binary and 'server_deploy.sh' to your server and run the script.${NC}"
    else
        echo -e "${GREEN}Files copied to server successfully!${NC}"
        echo -e "${YELLOW}To complete deployment, run the following commands on your server:${NC}"
        echo -e "  cd /tmp"
        echo -e "  chmod +x server_deploy.sh"
        echo -e "  ./server_deploy.sh"
    fi
fi

echo -e "\n${BLUE}TradeMicro API deployment process completed!${NC}"
