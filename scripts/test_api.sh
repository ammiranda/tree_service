#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
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
        echo -e "${GREEN}✓ Status code $actual_status matches expected $expected_status${NC}"
    else
        echo -e "${RED}✗ Status code $actual_status does not match expected $expected_status${NC}"
    fi
}

echo -e "${BLUE}Starting API Tests...${NC}"

# Test 1: Get initial tree (should be empty or have default root)
echo -e "\n${BLUE}1. Getting initial tree...${NC}"
response=$(curl -s -w "\n%{http_code}" $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Get Initial Tree" "$body" 200 $status_code

# Test 2: Create a root node
echo -e "\n${BLUE}2. Creating root node...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"label": "Test Root"}' \
    $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Create Root Node" "$body" 201 $status_code

# Extract root ID for subsequent tests
root_id=$(echo $body | jq -r '.id')

# Test 3: Create a child node
echo -e "\n${BLUE}3. Creating child node...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Child 1\", \"parentId\": $root_id}" \
    $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Create Child Node" "$body" 201 $status_code

# Test 4: Create another child node
echo -e "\n${BLUE}4. Creating another child node...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"Child 2\", \"parentId\": $root_id}" \
    $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Create Second Child Node" "$body" 201 $status_code

# Test 5: Get final tree structure
echo -e "\n${BLUE}5. Getting final tree structure...${NC}"
response=$(curl -s -w "\n%{http_code}" $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Get Final Tree" "$body" 200 $status_code

# Test 6: Try to create a node with invalid parent ID
echo -e "\n${BLUE}6. Testing invalid parent ID...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"label": "Invalid Child", "parentId": 99999}' \
    $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Create Node with Invalid Parent" "$body" 404 $status_code

# Test 7: Try to create a node with empty label
echo -e "\n${BLUE}7. Testing empty label...${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "{\"label\": \"\", \"parentId\": $root_id}" \
    $BASE_URL/tree)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed \$d)
print_result "Create Node with Empty Label" "$body" 400 $status_code

echo -e "\n${GREEN}API Tests Completed!${NC}" 