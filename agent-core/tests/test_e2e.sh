#!/bin/bash
# End-to-end test script for agent-core generate API
# Usage: ./test_e2e.sh [port]

PORT=${1:-8081}
BASE_URL="http://localhost:${PORT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Agent Core E2E Test"
echo "Target: ${BASE_URL}"
echo "=========================================="

# Check if server is running
echo -e "\n${YELLOW}1. Health Check${NC}"
HEALTH_RESPONSE=$(curl -s "${BASE_URL}/health")
if [[ "$HEALTH_RESPONSE" == *'"status":"ok"'* ]]; then
    echo -e "${GREEN}✓ Server is healthy${NC}"
else
    echo -e "${RED}✗ Server is not responding${NC}"
    echo "Response: $HEALTH_RESPONSE"
    exit 1
fi

# Test root endpoint
echo -e "\n${YELLOW}2. Root Endpoint${NC}"
ROOT_RESPONSE=$(curl -s "${BASE_URL}/")
if [[ "$ROOT_RESPONSE" == *'"service":"agent-core"'* ]]; then
    echo -e "${GREEN}✓ Root endpoint working${NC}"
else
    echo -e "${RED}✗ Root endpoint failed${NC}"
    echo "Response: $ROOT_RESPONSE"
fi

# Test echo endpoint (from Phase 2)
echo -e "\n${YELLOW}3. Echo Endpoint (SSE)${NC}"
ECHO_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/echo" \
    -H "Content-Type: application/json" \
    -d '{"message": "test", "count": 2, "delay": 0.1}' \
    --max-time 10)
if [[ "$ECHO_RESPONSE" == *'event: message'* ]]; then
    echo -e "${GREEN}✓ Echo endpoint working${NC}"
else
    echo -e "${RED}✗ Echo endpoint failed${NC}"
    echo "Response: $ECHO_RESPONSE"
fi

# Test generate endpoint with empty images (should fail with 400)
echo -e "\n${YELLOW}4. Generate Endpoint - Empty Images (should return 400)${NC}"
EMPTY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/v1/generate" \
    -H "Content-Type: application/json" \
    -d '{"session_id": "test-1", "images": [], "options": {}}')
if [[ "$EMPTY_RESPONSE" == "400" ]]; then
    echo -e "${GREEN}✓ Correctly rejected empty images with 400${NC}"
else
    echo -e "${RED}✗ Expected 400, got $EMPTY_RESPONSE${NC}"
fi

# Test generate endpoint with invalid request (should fail with 422)
echo -e "\n${YELLOW}5. Generate Endpoint - Missing session_id (should return 422)${NC}"
INVALID_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/v1/generate" \
    -H "Content-Type: application/json" \
    -d '{"images": [{"id": "1", "base64": "data:image/png;base64,test", "order": 0}]}')
if [[ "$INVALID_RESPONSE" == "422" ]]; then
    echo -e "${GREEN}✓ Correctly rejected invalid request with 422${NC}"
else
    echo -e "${RED}✗ Expected 422, got $INVALID_RESPONSE${NC}"
fi

# Test generate/sync endpoint with empty images
echo -e "\n${YELLOW}6. Generate Sync Endpoint - Empty Images (should return 400)${NC}"
SYNC_EMPTY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/v1/generate/sync" \
    -H "Content-Type: application/json" \
    -d '{"session_id": "test-2", "images": [], "options": {}}')
if [[ "$SYNC_EMPTY_RESPONSE" == "400" ]]; then
    echo -e "${GREEN}✓ Correctly rejected empty images with 400${NC}"
else
    echo -e "${RED}✗ Expected 400, got $SYNC_EMPTY_RESPONSE${NC}"
fi

echo -e "\n${YELLOW}=========================================="
echo "E2E Tests Complete"
echo "==========================================${NC}"

# Note about full generation test
echo -e "\n${YELLOW}Note:${NC} Full generation tests require LLM API keys configured."
echo "To test full generation flow:"
echo "  1. Set LLM_PROVIDER and corresponding API key in environment"
echo "  2. Use the test_generate_full.py script with actual images"
