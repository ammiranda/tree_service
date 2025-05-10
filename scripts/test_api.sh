#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL for the API
BASE_URL="http://localhost:8080/api"

# Function to print test results
print_result() {
    local test_name=$1
    local response=$2
    local expected_status=$3
    local actual_status=$4

    echo -e "\n${BLUE}Test: ${test_name}${NC}"
    echo "Response: $response"
    
    if [ $actual_status -eq $expected_status ]; then
        echo -e "${GREEN}✓ PASS: Status code $actual_status matches expected $expected_status${NC}"
    else
        echo -e "${RED}✗ FAIL: Status code $actual_status does not match expected $expected_status${NC}"
    fi
}

# Function to make API calls and print responses with status check
make_request() {
    local method=$1
    local endpoint=$2
    local payload=$3
    local expected_status=$4
    local test_name=$5

    echo -e "\n${BLUE}Making $method request to $endpoint${NC}"
    if [ -n "$payload" ]; then
        echo "With payload: $payload"
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json" -d "$payload")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed \$d)
    
    if [ $status_code -eq $expected_status ]; then
        echo -e "${GREEN}✓ PASS: $test_name${NC}"
        echo "Response: $body"
    else
        echo -e "${RED}✗ FAIL: $test_name${NC}"
        echo "Expected status: $expected_status"
        echo "Actual status: $status_code"
        echo "Response: $body"
    fi
    echo "----------------------------------------"
}

echo -e "${BLUE}Starting API Tests...${NC}"

# Create a root node
echo -e "\n${YELLOW}Creating tree structure...${NC}"
root_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d '{"label": "Root Node"}')
root_id=$(echo $root_response | jq -r '.id')
echo -e "${GREEN}✓ Root node created with ID: $root_id${NC}"

# Create first level children
child1_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Child 1\", \"parentId\": $root_id}")
child1_id=$(echo $child1_response | jq -r '.id')

child2_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Child 2\", \"parentId\": $root_id}")
child2_id=$(echo $child2_response | jq -r '.id')
echo -e "${GREEN}✓ First level children created with IDs: $child1_id, $child2_id${NC}"

# Create second level children (grandchildren)
grandchild1_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Grandchild 1.1\", \"parentId\": $child1_id}")
grandchild1_id=$(echo $grandchild1_response | jq -r '.id')

grandchild2_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Grandchild 1.2\", \"parentId\": $child1_id}")
grandchild2_id=$(echo $grandchild2_response | jq -r '.id')

grandchild3_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Grandchild 2.1\", \"parentId\": $child2_id}")
grandchild3_id=$(echo $grandchild3_response | jq -r '.id')
echo -e "${GREEN}✓ Second level children created with IDs: $grandchild1_id, $grandchild2_id, $grandchild3_id${NC}"

# Create third level children (great-grandchildren)
great_grandchild1_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Great-Grandchild 1.1.1\", \"parentId\": $grandchild1_id}")
great_grandchild1_id=$(echo $great_grandchild1_response | jq -r '.id')

great_grandchild2_response=$(curl -s -X POST "$BASE_URL/tree" \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Great-Grandchild 1.1.2\", \"parentId\": $grandchild1_id}")
great_grandchild2_id=$(echo $great_grandchild2_response | jq -r '.id')
echo -e "${GREEN}✓ Third level children created with IDs: $great_grandchild1_id, $great_grandchild2_id${NC}"

echo -e "\n${YELLOW}Running API Tests...${NC}"

# Test getting the tree with different pagination parameters
make_request "GET" "/tree" "" 200 "Get tree (default pagination)"
make_request "GET" "/tree?page=2" "" 200 "Get tree (page 2)"
make_request "GET" "/tree?pageSize=5" "" 200 "Get tree (page size 5)"
make_request "GET" "/tree?page=2&pageSize=5" "" 200 "Get tree (page 2, size 5)"

# Test updating a node
make_request "PUT" "/node/$child1_id" "{\"label\": \"Updated Child 1\"}" 200 "Update node label"

# Test getting the tree after update
make_request "GET" "/tree" "" 200 "Get tree after update"

# Test creating a node with invalid input
make_request "POST" "/tree" "{\"label\": \"\"}" 400 "Create node with empty label"
make_request "POST" "/tree" "{\"label\": \"Valid Label\", \"parentId\": 999999}" 404 "Create node with invalid parent"

# Test getting the tree one final time
make_request "GET" "/tree" "" 200 "Get final tree state"

echo -e "\n${GREEN}API Tests Completed!${NC}" 