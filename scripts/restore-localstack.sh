#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Set AWS credentials
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566

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

# Wait for LocalStack to be ready
echo -e "${GREEN}Waiting for LocalStack to be ready...${NC}"
check_service_health "s3" || exit 1

# Restore state from S3
echo -e "${GREEN}Restoring LocalStack state...${NC}"
aws --endpoint-url=$AWS_ENDPOINT_URL s3 cp s3://localstack-state/state.json /var/lib/localstack/state.json

# Restart LocalStack to apply the state
echo -e "${GREEN}Restarting LocalStack to apply state...${NC}"
docker-compose restart localstack

echo -e "${GREEN}State restoration complete!${NC}" 