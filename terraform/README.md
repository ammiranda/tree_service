# Tree API Infrastructure

This directory contains Terraform configurations to provision the AWS infrastructure for the Tree API application.

## Prerequisites

- Terraform v1.0.0 or later
- AWS CLI configured with appropriate credentials
- AWS account with necessary permissions

## Infrastructure Components

The configuration provisions the following AWS resources:

- VPC with public and private subnets
- RDS PostgreSQL instance
- Lambda function
- API Gateway
- Security groups
- IAM roles and policies
- Secrets Manager secret for RDS credentials

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Create a `terraform.tfvars` file with your specific values:
   ```hcl
   aws_region = "us-west-2"
   project_name = "tree-api"
   db_password = "your-secure-password"
   ```

3. Plan the deployment:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

5. After successful deployment, the outputs will show important information like:
   - API Gateway endpoint URL
   - RDS instance endpoint
   - Lambda function name
   - VPC and subnet IDs
   - Security group IDs

## Security Considerations

- The RDS password is stored in AWS Secrets Manager
- The Lambda function runs in a VPC with private subnets
- Security groups restrict access to RDS and Lambda
- IAM roles follow the principle of least privilege

## Cleanup

To destroy all provisioned resources:
```bash
terraform destroy
```

## Variables

| Name | Description | Type | Default |
|------|-------------|------|---------|
| aws_region | AWS region to deploy resources | string | "us-west-2" |
| project_name | Name of the project | string | "tree-api" |
| vpc_cidr | CIDR block for VPC | string | "10.0.0.0/16" |
| availability_zones | List of availability zones | list(string) | ["us-west-2a", "us-west-2b"] |
| private_subnet_cidrs | CIDR blocks for private subnets | list(string) | ["10.0.1.0/24", "10.0.2.0/24"] |
| public_subnet_cidrs | CIDR blocks for public subnets | list(string) | ["10.0.101.0/24", "10.0.102.0/24"] |
| rds_instance_class | RDS instance class | string | "db.t3.micro" |
| db_name | Name of the database | string | "tree_db" |
| db_username | Username for RDS instance | string | "postgres" |
| db_password | Password for RDS instance | string | - |
| tags | Tags to apply to all resources | map(string) | { Environment = "production", Project = "tree-api" } |

## Outputs

| Name | Description |
|------|-------------|
| api_endpoint | API Gateway endpoint URL |
| rds_endpoint | RDS instance endpoint |
| lambda_function_name | Name of the Lambda function |
| vpc_id | ID of the VPC |
| private_subnet_ids | IDs of the private subnets |
| public_subnet_ids | IDs of the public subnets |
| rds_security_group_id | Security group ID for RDS |
| lambda_security_group_id | Security group ID for Lambda | 