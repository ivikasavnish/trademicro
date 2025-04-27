#!/bin/bash

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

echo -e "${YELLOW}This script will help you set up SSH access to your GCP VM${NC}"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo -e "${RED}You are not authenticated with gcloud. Please run:${NC}"
    echo -e "${YELLOW}gcloud auth activate-service-account --key-file=/Users/vikasavnish/trademicro/gcp-service-account.json${NC}"
    exit 1
fi

# Variables
PROJECT_ID="amiable-alcove-456605-k2"
ZONE="asia-south1-c"
INSTANCE_NAME="instance-20250422-132526"

# Check if SSH key exists, if not create it
SSH_KEY_FILE="$HOME/.ssh/gcp_vm_key"
if [ ! -f "$SSH_KEY_FILE" ]; then
    echo -e "${YELLOW}Creating new SSH key for GCP VM access...${NC}"
    ssh-keygen -t rsa -b 4096 -f "$SSH_KEY_FILE" -N "" -C "vikasavnish@gcp-vm"
    echo -e "${GREEN}SSH key created at $SSH_KEY_FILE${NC}"
else
    echo -e "${YELLOW}Using existing SSH key at $SSH_KEY_FILE${NC}"
fi

# Get the public key content
PUBLIC_KEY=$(cat "${SSH_KEY_FILE}.pub")

# Format the key with username
USERNAME="vikasavnish"
FORMATTED_KEY="$USERNAME:$PUBLIC_KEY"

# Add the SSH key to the instance metadata
echo -e "${YELLOW}Adding SSH key to GCP VM instance metadata...${NC}"
gcloud compute instances add-metadata "$INSTANCE_NAME" \
    --project="$PROJECT_ID" \
    --zone="$ZONE" \
    --metadata="ssh-keys=$FORMATTED_KEY"

echo -e "\n${GREEN}SSH key has been added to the VM instance.${NC}"
echo -e "${YELLOW}You can now connect to your VM using:${NC}"
echo -e "${GREEN}ssh -i $SSH_KEY_FILE $USERNAME@35.244.30.157${NC}\n"

# Check if the VM has the necessary firewall rules
echo -e "${YELLOW}Checking firewall rules...${NC}"
if ! gcloud compute firewall-rules list --project="$PROJECT_ID" --filter="name=allow-ssh" --format="value(name)" | grep -q "allow-ssh"; then
    echo -e "${YELLOW}Creating firewall rule to allow SSH access...${NC}"
    gcloud compute firewall-rules create allow-ssh \
        --project="$PROJECT_ID" \
        --direction=INGRESS \
        --priority=1000 \
        --network=default \
        --action=ALLOW \
        --rules=tcp:22 \
        --source-ranges=0.0.0.0/0
    echo -e "${GREEN}Firewall rule created to allow SSH access${NC}"
else
    echo -e "${GREEN}SSH firewall rule already exists${NC}"
fi

echo -e "\n${GREEN}Setup complete! You should now be able to SSH into your VM.${NC}"
echo -e "${YELLOW}If you still have issues, please check the following:${NC}"
echo -e "1. Make sure the VM is running"
echo -e "2. Make sure the VM has an external IP address"
echo -e "3. Check that the firewall rules allow SSH access"
echo -e "4. Try connecting with verbose output: ssh -v -i $SSH_KEY_FILE $USERNAME@35.244.30.157\n"
