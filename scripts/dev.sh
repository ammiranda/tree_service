#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for a service to be ready
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    
    echo -e "${GREEN}Waiting for $service to be ready...${NC}"
    while ! nc -z $host $port >/dev/null 2>&1; do
        sleep 1
    done
    echo -e "${GREEN}$service is ready!${NC}"
}

# Function to initialize LocalStack
init_localstack() {
    echo -e "${GREEN}Initializing LocalStack...${NC}"
    
    # Wait for LocalStack to be ready
    while ! curl -s http://localhost:4566/health | grep -q '"elasticache": "running"'; do
        sleep 1
    done
    
    # Set AWS credentials
    export AWS_ACCESS_KEY_ID=test
    export AWS_SECRET_ACCESS_KEY=test
    export AWS_DEFAULT_REGION=us-west-2
    export AWS_ENDPOINT_URL=http://localhost:4566
    
    # Create RDS instance
    echo -e "${GREEN}Creating RDS instance...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL rds create-db-instance \
        --db-instance-identifier tree-api-db \
        --db-instance-class db.t3.micro \
        --engine postgres \
        --master-username postgres \
        --master-user-password postgres \
        --allocated-storage 20
    
    # Create ElastiCache Redis cluster
    echo -e "${GREEN}Creating ElastiCache Redis cluster...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL elasticache create-cache-cluster \
        --cache-cluster-id tree-api-redis \
        --engine redis \
        --cache-node-type cache.t3.micro \
        --num-cache-nodes 1 \
        --port 6379
    
    # Create Secrets Manager secret for RDS
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
    
    # Create Lambda function
    echo -e "${GREEN}Creating Lambda function...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL lambda create-function \
        --function-name tree-api \
        --runtime go1.x \
        --handler main \
        --zip-file fileb://build/lambda.zip \
        --role arn:aws:iam::000000000000:role/lambda-role
    
    # Create API Gateway
    echo -e "${GREEN}Creating API Gateway...${NC}"
    aws --endpoint-url=$AWS_ENDPOINT_URL apigateway create-rest-api \
        --name tree-api
    
    echo -e "${GREEN}LocalStack initialization complete!${NC}"
}

# Function to start services
start_services() {
    echo -e "${GREEN}Starting services...${NC}"
    
    # Check if Docker is installed
    if ! command_exists docker; then
        echo "Error: Docker is not installed"
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command_exists docker-compose; then
        echo "Error: Docker Compose is not installed"
        exit 1
    fi
    
    # Start services
    docker-compose up -d
    
    # Wait for services to be ready
    wait_for_service localhost 5432 "PostgreSQL"
    wait_for_service localhost 6379 "Redis"
    wait_for_service localhost 4566 "LocalStack"
    
    # Initialize LocalStack
    init_localstack
    
    echo -e "${GREEN}All services are up and running!${NC}"
}

# Function to stop services
stop_services() {
    echo -e "${GREEN}Stopping services...${NC}"
    docker-compose down
    echo -e "${GREEN}Services stopped!${NC}"
}

# Main script
case "$1" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        start_services
        ;;
    *)
        echo "Usage: $0 {start|stop|restart}"
        exit 1
        ;;
esac

exit 0 