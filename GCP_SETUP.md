# Google Cloud Platform (GCP) Setup for TradeMicro API

This guide will help you set up Google Cloud Platform for deploying the TradeMicro API.

## Prerequisites

- A Google Cloud Platform account
- Basic familiarity with GCP console
- `gcloud` CLI installed locally (optional, for testing)

## Step 1: Create a GCP Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click on the project dropdown at the top of the page
3. Click "New Project"
4. Enter a name for your project (e.g., "trademicro")
5. Click "Create"

## Step 2: Enable Required APIs

1. Go to the [API Library](https://console.cloud.google.com/apis/library)
2. Search for and enable the following APIs:
   - Compute Engine API
   - Cloud Build API
   - IAM API

## Step 3: Create a Compute Engine Instance

1. Go to [Compute Engine > VM Instances](https://console.cloud.google.com/compute/instances)
2. Click "Create Instance"
3. Configure your instance:
   - Name: `trademicro-api` (or your preferred name)
   - Region/Zone: Choose a region close to your users (e.g., `us-central1-a`)
   - Machine type: `e2-small` (2 vCPU, 2 GB memory) is sufficient for starting
   - Boot disk: Ubuntu 20.04 LTS
   - Allow HTTP and HTTPS traffic
4. Click "Create"

## Step 4: Create a Service Account for GitHub Actions

1. Go to [IAM & Admin > Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Click "Create Service Account"
3. Enter a name (e.g., "github-actions")
4. Click "Create and Continue"
5. Assign the following roles:
   - Compute Instance Admin (v1)
   - Service Account User
   - Storage Admin (if you plan to use GCS)
6. Click "Done"
7. Find your new service account in the list and click on it
8. Go to the "Keys" tab
9. Click "Add Key" > "Create new key"
10. Choose JSON format
11. Click "Create"
12. The key file will be downloaded to your computer

## Step 5: Add Secrets to GitHub

1. Go to your GitHub repository
2. Go to Settings > Secrets and variables > Actions
3. Add the following secrets:
   - `GCP_PROJECT_ID`: Your GCP project ID
   - `GCP_INSTANCE_NAME`: The name of your Compute Engine instance (e.g., `trademicro-api`)
   - `GCP_ZONE`: The zone of your instance (e.g., `us-central1-a`)
   - `GCP_SA_KEY`: The entire content of the JSON key file you downloaded
   - Ensure you already have these secrets set up:
     - `POSTGRES_URL`: Your PostgreSQL connection string
     - `REDIS_URL`: Your Redis connection string
     - `SECRET_KEY`: Your application secret key

## Step 6: Configure Firewall Rules

1. Go to [VPC Network > Firewall](https://console.cloud.google.com/networking/firewalls)
2. Click "Create Firewall Rule"
3. Configure the rule:
   - Name: `allow-trademicro-api`
   - Network: `default`
   - Direction of traffic: `Ingress`
   - Action on match: `Allow`
   - Targets: `All instances in the network`
   - Source filter: `IP ranges`
   - Source IP ranges: `0.0.0.0/0` (or restrict to specific IPs for better security)
   - Protocols and ports: `Specified protocols and ports` > `tcp:8000`
4. Click "Create"

## Step 7: Test the Deployment

1. Go to your GitHub repository
2. Go to the Actions tab
3. Select the "Deploy to GCP" workflow
4. Click "Run workflow" > "Run workflow"
5. Once the workflow completes, your API should be accessible at:
   `http://<YOUR_INSTANCE_IP>:8000`

## Troubleshooting

### SSH Connection Issues

If GitHub Actions cannot connect to your instance, check:
1. The service account has the correct permissions
2. The instance name and zone are correct
3. The instance is running

### Application Not Starting

If the application doesn't start, SSH into the instance and check:
1. The systemd service status: `sudo systemctl status trademicro`
2. Application logs: `sudo journalctl -u trademicro`
3. Environment file: `cat /opt/trademicro/.env`

### Database Connection Issues

Ensure your PostgreSQL database allows connections from your GCP instance's IP address.
