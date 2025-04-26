# GitHub Actions Deployment

This directory contains GitHub Actions workflows for automatically deploying the TradeMicro API.

## Setup Instructions

### 1. Configure GitHub Secrets

You need to add the following secrets to your GitHub repository:

- `SERVER_HOST`: Your server IP address (e.g., 62.164.218.111)
- `SERVER_USER`: Your server username (e.g., root)
- `SERVER_PASSWORD`: Your server password
- `POSTGRES_DSN`: Your PostgreSQL connection string
- `REDIS_URL`: Your Redis URL (default: redis://localhost:6379/0)
- `SECRET_KEY`: A secure random key for JWT token signing

To add these secrets:
1. Go to your GitHub repository
2. Click on "Settings" tab
3. Click on "Secrets and variables" > "Actions"
4. Click "New repository secret" and add each secret

### 2. Push to Main Branch

Once the secrets are configured, the workflow will automatically run whenever you push to the `main` or `master` branch.

### 3. Manual Deployment

You can also manually trigger the deployment:
1. Go to the "Actions" tab in your GitHub repository
2. Select the "Deploy TradeMicro API" workflow
3. Click "Run workflow"
4. Select the environment (production or staging)
5. Click "Run workflow"

## Workflow Details

The deployment workflow:

1. Builds the Go application
2. Creates a deployment script with your environment variables
3. Copies the binary and deployment script to your server
4. Executes the deployment script on your server

## Troubleshooting

If the deployment fails, check the GitHub Actions logs for details. Common issues include:

- Incorrect server credentials
- Database connection issues
- Permission problems on the server

You can also check the server logs with:
```bash
journalctl -u trademicro
```
