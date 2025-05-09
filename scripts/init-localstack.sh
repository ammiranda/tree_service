#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if a service is ready
check_service_health() {
    local service=$1
    local max_attempts=30
    local attempt=1

    echo -e "${GREEN}Checking $service health...${NC}"
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:4566/health | grep -q "\"$service\": \"running\""; then
            echo -e "${GREEN}$service is ready!${NC}"
            return 0
        fi
        echo "Waiting for $service to be ready... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    echo -e "${RED}$service failed to start within the timeout period${NC}"
    return 1
}

# Set AWS credentials
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566

# Wait for LocalStack to be ready
echo -e "${GREEN}Waiting for LocalStack to be ready...${NC}"
check_service_health "elasticache" || exit 1

# Function to create RDS instance
create_rds() {
    echo -e "${GREEN}Creating RDS instance...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL rds create-db-instance \
        --db-instance-identifier tree-api-db \
        --db-instance-class db.t3.micro \
        --engine postgres \
        --master-username postgres \
        --master-user-password postgres \
        --allocated-storage 20
}

# Function to create ElastiCache cluster
create_elasticache() {
    echo -e "${GREEN}Creating ElastiCache Redis cluster...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL elasticache create-cache-cluster \
        --cache-cluster-id tree-api-redis \
        --engine redis \
        --cache-node-type cache.t3.micro \
        --num-cache-nodes 1 \
        --port 6379
}

# Function to create Secrets Manager secret
create_secret() {
    echo -e "${GREEN}Creating Secrets Manager secret...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL secretsmanager create-secret \
        --name tree-api/rds/credentials \
        --secret-string '{
            "host": "postgres",
            "port": 5432,
            "username": "postgres",
            "password": "postgres",
            "dbname": "tree_db",
            "sslmode": "disable"
        }'
}

# Function to create Lambda function
create_lambda() {
    echo -e "${GREEN}Creating Lambda function...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL lambda create-function \
        --function-name tree-api \
        --runtime go1.x \
        --handler main \
        --zip-file fileb://build/lambda.zip \
        --role arn:aws:iam::000000000000:role/lambda-role
}

# Function to create API Gateway
create_apigateway() {
    echo -e "${GREEN}Creating API Gateway...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL apigateway create-rest-api \
        --name tree-api
}

# Run RDS, ElastiCache, and Secrets Manager creation in parallel
echo -e "${GREEN}Starting parallel resource creation...${NC}"
create_rds & rds_pid=$!
create_elasticache & elasticache_pid=$!
create_secret & secret_pid=$!

# Wait for parallel operations to complete
wait $rds_pid || echo -e "${RED}RDS creation failed${NC}"
wait $elasticache_pid || echo -e "${RED}ElastiCache creation failed${NC}"
wait $secret_pid || echo -e "${RED}Secrets Manager creation failed${NC}"

# Create Lambda and API Gateway sequentially (they depend on other services)
create_lambda
create_apigateway

echo -e "${GREEN}LocalStack initialization complete!${NC}"

# Save the state
echo -e "${GREEN}Saving LocalStack state...${NC}"
aws --endpoint-url=$AWS_ENDPOINT_URL s3 mb s3://localstack-state
aws --endpoint-url=$AWS_ENDPOINT_URL s3 cp /var/lib/localstack/state.json s3://localstack-state/state.json 