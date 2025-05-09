# GitHub Actions Workflows

This directory contains GitHub Actions workflows for testing and deploying the Tree API application.

## Workflows

### Test Workflow (`test.yml`)

Runs on:
- Push to main branch
- Pull requests to main branch

Actions:
1. Sets up Go environment
2. Installs dependencies
3. Runs tests
4. Runs linter

### Deploy Workflow (`deploy.yml`)

Runs on:
- Push to main branch

Actions:
1. Sets up Go environment
2. Installs dependencies
3. Builds Lambda function
4. Configures AWS credentials
5. Sets up Terraform
6. Initializes Terraform
7. Plans infrastructure changes
8. Applies infrastructure changes
9. Updates Lambda function code

## Required Secrets

The following secrets need to be configured in your GitHub repository:

- `AWS_ACCESS_KEY_ID`: AWS access key for deployment
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for deployment
- `DB_PASSWORD`: Database password for RDS instance

## Environment Protection

The deployment workflow uses the `production` environment, which can be configured in GitHub to:
- Require approval before deployment
- Restrict which branches can be deployed
- Add environment-specific secrets

## Usage

1. Configure the required secrets in your GitHub repository:
   - Go to Settings > Secrets and variables > Actions
   - Add the required secrets

2. Configure the production environment:
   - Go to Settings > Environments
   - Create a new environment named "production"
   - Configure protection rules as needed

3. Push to main branch to trigger deployment:
   ```bash
   git push origin main
   ```

## Notes

- The deployment workflow uses `-auto-approve` for Terraform apply. Remove this flag if you want manual approval.
- The Lambda function is built for Linux x86_64 architecture.
- The workflow uses Terraform version 1.5.7. Update the version in the workflow if needed. 