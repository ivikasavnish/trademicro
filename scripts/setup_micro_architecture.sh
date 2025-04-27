#!/bin/bash

# TradeMicro Micro-Worker Architecture Setup Script
# This script configures the micro instance as a frontend API server
# and sets up SSH communication with the worker instance for long-running tasks

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# Configuration
MICRO_INSTANCE="instance-20250422-132526"
WORKER_INSTANCE="instance-20250416-112838"
SSH_KEY_DIR="/opt/trademicro/.ssh"
SSH_KEY_NAME="worker_key"
ZONE="asia-south1-c"
PROJECT_ID="amiable-alcove-456605-k2"

echo -e "${YELLOW}Setting up TradeMicro Micro-Worker Architecture...${NC}"

# Determine if this is the micro or worker instance
HOSTNAME=$(hostname)
if [[ "$HOSTNAME" == "$MICRO_INSTANCE" ]]; then
    IS_MICRO=true
    echo -e "${GREEN}Detected MICRO instance: Setting up as frontend API server${NC}"
else
    IS_MICRO=false
    echo -e "${GREEN}Detected WORKER instance: Setting up as worker for long-running tasks${NC}"
fi

# Create SSH directory if it doesn't exist
mkdir -p $SSH_KEY_DIR
chmod 700 $SSH_KEY_DIR

if [ "$IS_MICRO" = true ]; then
    # We're on the micro instance - set up as frontend API server
    
    # Generate SSH key pair if it doesn't exist
    if [ ! -f "$SSH_KEY_DIR/$SSH_KEY_NAME" ]; then
        echo -e "${YELLOW}Generating SSH key pair for worker communication...${NC}"
        ssh-keygen -t rsa -b 4096 -f "$SSH_KEY_DIR/$SSH_KEY_NAME" -N "" -C "trademicro-micro-to-worker"
        chmod 600 "$SSH_KEY_DIR/$SSH_KEY_NAME"
        chmod 644 "$SSH_KEY_DIR/$SSH_KEY_NAME.pub"
    fi
    
    # Get the public key
    PUBLIC_KEY=$(cat "$SSH_KEY_DIR/$SSH_KEY_NAME.pub")
    
    # Add the public key to the worker instance's authorized_keys
    echo -e "${YELLOW}Adding SSH key to worker instance...${NC}"
    gcloud compute ssh --zone "$ZONE" "$WORKER_INSTANCE" --project "$PROJECT_ID" --command "mkdir -p ~/.ssh && echo '$PUBLIC_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
    
    # Test the connection
    echo -e "${YELLOW}Testing SSH connection to worker instance...${NC}"
    if ssh -i "$SSH_KEY_DIR/$SSH_KEY_NAME" -o StrictHostKeyChecking=no -o ConnectTimeout=5 root@$WORKER_INSTANCE "echo 'Connection successful'"; then
        echo -e "${GREEN}SSH connection to worker established successfully!${NC}"
    else
        echo -e "${RED}Failed to establish SSH connection to worker. Please check the configuration.${NC}"
        exit 1
    fi
    
    # Create environment file for the micro instance
    echo -e "${YELLOW}Creating environment configuration...${NC}"
    cat > /opt/trademicro/.env << EOF
# TradeMicro Micro Instance Configuration
SERVER_ROLE=micro
WORKER_HOST=$WORKER_INSTANCE
WORKER_USER=root
WORKER_SSH_KEY=$SSH_KEY_DIR/$SSH_KEY_NAME
TASK_LOG_DIR=/opt/trademicro/logs/tasks
EOF
    
    # Create task log directory
    mkdir -p /opt/trademicro/logs/tasks
    chmod 755 /opt/trademicro/logs/tasks
    
    echo -e "${GREEN}Micro instance setup completed successfully!${NC}"
    echo -e "${YELLOW}The micro instance is now configured as a frontend API server.${NC}"
    echo -e "${YELLOW}It will handle API requests and delegate long-running tasks to the worker instance.${NC}"
    
else
    # We're on the worker instance - set up as worker for long-running tasks
    
    # Create environment file for the worker instance
    echo -e "${YELLOW}Creating environment configuration...${NC}"
    cat > /opt/trademicro/.env << EOF
# TradeMicro Worker Instance Configuration
SERVER_ROLE=worker
EOF
    
    # Create scripts directory for worker tasks
    mkdir -p /opt/trademicro/scripts
    chmod 755 /opt/trademicro/scripts
    
    # Create a simple worker status script
    cat > /opt/trademicro/scripts/worker_status.sh << EOF
#!/bin/bash
echo "TradeMicro Worker Status"
echo "========================"
echo "Hostname: \$(hostname)"
echo "CPU Usage: \$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - \$1}')%"
echo "Memory Usage: \$(free -m | awk 'NR==2{printf "%.2f%%", \$3*100/\$2}')"
echo "Disk Usage: \$(df -h / | awk 'NR==2{print \$5}')"
echo "Uptime: \$(uptime -p)"
EOF
    chmod +x /opt/trademicro/scripts/worker_status.sh
    
    echo -e "${GREEN}Worker instance setup completed successfully!${NC}"
    echo -e "${YELLOW}The worker instance is now configured to handle long-running tasks.${NC}"
    echo -e "${YELLOW}It will receive task requests from the micro instance via SSH.${NC}"
fi

echo -e "${GREEN}TradeMicro Micro-Worker Architecture setup completed!${NC}"
