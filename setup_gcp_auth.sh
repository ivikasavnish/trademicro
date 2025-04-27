#!/bin/bash

# Script to help set up GCP authentication for GitHub Actions

# Colors for output
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

echo -e "${YELLOW}This script will help you set up GCP authentication for GitHub Actions${NC}"

# Check if the service account key file exists
SA_KEY_FILE="/Users/vikasavnish/Downloads/amiable-alcove-456605-k2-35120d51ad09.json"
if [ ! -f "$SA_KEY_FILE" ]; then
    echo -e "${RED}Service account key file not found at $SA_KEY_FILE${NC}"
    echo -e "${YELLOW}Please download your service account key from Google Cloud Console${NC}"
    exit 1
fi

# Read the service account key file
SA_KEY=$(cat "$SA_KEY_FILE")

# Display instructions for setting up GitHub Secrets
echo -e "\n${GREEN}Follow these steps to set up GitHub Secrets:${NC}"
echo -e "${YELLOW}1. Go to your GitHub repository${NC}"
echo -e "${YELLOW}2. Click on 'Settings' > 'Secrets and variables' > 'Actions'${NC}"
echo -e "${YELLOW}3. Click on 'New repository secret'${NC}"
echo -e "${YELLOW}4. Add the following secrets:${NC}\n"

echo -e "${GREEN}Secret Name:${NC} GCP_PROJECT_ID"
echo -e "${GREEN}Secret Value:${NC} amiable-alcove-456605-k2\n"

echo -e "${GREEN}Secret Name:${NC} GCP_SA_KEY"
echo -e "${GREEN}Secret Value:${NC} [The entire content of your service account key JSON file]\n"

echo -e "${YELLOW}IMPORTANT: Make sure to copy the ENTIRE JSON content of your service account key file${NC}"
echo -e "${YELLOW}The JSON should include all fields like 'type', 'project_id', 'private_key', etc.${NC}\n"

echo -e "${GREEN}Would you like to print the service account key to copy? (y/n)${NC}"
read -r answer

if [[ "$answer" =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Here is your service account key:${NC}\n"
    echo "$SA_KEY"
    echo -e "\n${YELLOW}Copy the above JSON content exactly as shown${NC}"
fi

echo -e "\n${GREEN}After setting up these secrets, your GitHub Actions workflow should work correctly.${NC}"
