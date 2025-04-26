#!/bin/bash

# This script helps set up GitHub secrets for GCP deployment
# You'll need the GitHub CLI (gh) installed

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "GitHub CLI (gh) is not installed. Please install it first."
    echo "Installation instructions: https://github.com/cli/cli#installation"
    exit 1
 fi

# Check if logged in to GitHub
if ! gh auth status &> /dev/null; then
    echo "You are not logged in to GitHub CLI. Please run 'gh auth login' first."
    exit 1
fi

# Get repository name
REPO=$(git remote get-url origin | sed 's/.*github.com[:\/]\(.*\)\.git/\1/')
if [ -z "$REPO" ]; then
    echo "Could not determine GitHub repository. Please enter it manually (format: username/repo):"
    read REPO
fi

# GCP Project ID
PROJECT_ID="amiable-alcove-456605-k2"

# Ask for GCP instance details
echo "Enter your GCP instance name (e.g., trademicro-api):"
read INSTANCE_NAME

echo "Enter your GCP zone (e.g., us-central1-a):"
read ZONE

# Path to service account key file
KEY_FILE="/Users/vikasavnish/Downloads/amiable-alcove-456605-k2-35120d51ad09.json"
if [ ! -f "$KEY_FILE" ]; then
    echo "Service account key file not found at $KEY_FILE"
    echo "Please enter the path to your service account key file:"
    read KEY_FILE
fi

# Set GitHub secrets
echo "Setting GitHub secrets..."

# GCP Project ID
echo "Setting GCP_PROJECT_ID..."
gh secret set GCP_PROJECT_ID --body "$PROJECT_ID" --repo "$REPO"

# GCP Instance Name
echo "Setting GCP_INSTANCE_NAME..."
gh secret set GCP_INSTANCE_NAME --body "$INSTANCE_NAME" --repo "$REPO"

# GCP Zone
echo "Setting GCP_ZONE..."
gh secret set GCP_ZONE --body "$ZONE" --repo "$REPO"

# GCP Service Account Key
echo "Setting GCP_SA_KEY..."
gh secret set GCP_SA_KEY --body "$(cat $KEY_FILE)" --repo "$REPO"

echo "All GCP secrets have been set up successfully!"
echo "You can now run the GCP deployment workflow from GitHub Actions."
