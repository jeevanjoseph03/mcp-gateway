#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo "========================================="
echo "MCP Gateway Test Suite"
echo "========================================="
echo ""

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Gateway Health Check${NC}"
curl -s ${BASE_URL}/health | jq '.'
echo ""
echo ""

# Test 2: Initialize
echo -e "${YELLOW}Test 2: Initialize Handshake${NC}"
INIT_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }')
echo "$INIT_RESPONSE" | jq '.'

# Check if initialize succeeded
if echo "$INIT_RESPONSE" | jq -e '.result' > /dev/null; then
    echo -e "${GREEN}✓ Initialize succeeded${NC}"
else
    echo -e "${RED}✗ Initialize failed${NC}"
fi
echo ""

# Test 3: List Tools (should merge from both servers)
echo -e "${YELLOW}Test 3: List Tools (Federation Test)${NC}"
TOOLS_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }')
echo "$TOOLS_RESPONSE" | jq '.'

# Check if tools are merged (should have at least tools from both servers)
TOOL_COUNT=$(echo "$TOOLS_RESPONSE" | jq '.result.tools | length')
if [ "$TOOL_COUNT" -ge 2 ]; then
    echo -e "${GREEN}✓ Tools merged successfully (found $TOOL_COUNT tools)${NC}"
else
    echo -e "${RED}✗ Tool merging failed (only $TOOL_COUNT tools found)${NC}"
fi
echo ""

# Test 4: Call echo tool
echo -e "${YELLOW}Test 4: Call 'echo' Tool${NC}"
ECHO_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "echo",
      "arguments": {
        "message": "Hello from test client!"
      }
    }
  }')
echo "$ECHO_RESPONSE" | jq '.'

# Check if echo succeeded
if echo "$ECHO_RESPONSE" | jq -e '.result.content[0].text' > /dev/null; then
    echo -e "${GREEN}✓ Echo tool succeeded${NC}"
else
    echo -e "${RED}✗ Echo tool failed${NC}"
fi
echo ""

# Test 5: Call add_numbers tool
echo -e "${YELLOW}Test 5: Call 'add_numbers' Tool${NC}"
ADD_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "add_numbers",
      "arguments": {
        "a": 15.5,
        "b": 27.3
      }
    }
  }')
echo "$ADD_RESPONSE" | jq '.'

if echo "$ADD_RESPONSE" | jq -e '.result.content[0].text' > /dev/null; then
    echo -e "${GREEN}✓ add_numbers tool succeeded${NC}"
else
    echo -e "${RED}✗ add_numbers tool failed${NC}"
fi
echo ""

# Test 6: Test Routing - Call same tool multiple times (should be load-balanced)
echo -e "${YELLOW}Test 6: Load Balancing Test${NC}"
for i in {1..4}; do
    RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
      -H "Content-Type: application/json" \
      -d '{
        "jsonrpc": "2.0",
        "id": '$((10+i))',
        "method": "tools/call",
        "params": {
          "name": "echo",
          "arguments": {
            "message": "Request #'$i'"
          }
        }
      }')
    
    # Extract which server handled the request
    TEXT=$(echo "$RESPONSE" | jq -r '.result.content[0].text')
    echo "Request $i -> $TEXT"
done
echo ""

# Test 7: Error Handling - Call non-existent tool
echo -e "${YELLOW}Test 7: Error Handling (Non-existent Tool)${NC}"
ERROR_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 20,
    "method": "tools/call",
    "params": {
      "name": "non_existent_tool",
      "arguments": {}
    }
  }')
echo "$ERROR_RESPONSE" | jq '.'

if echo "$ERROR_RESPONSE" | jq -e '.error' > /dev/null; then
    echo -e "${GREEN}✓ Proper error returned for non-existent tool${NC}"
else
    echo -e "${RED}✗ Missing error for non-existent tool${NC}"
fi
echo ""

# Test 8: Health Check Monitoring
echo -e "${YELLOW}Test 8: Health Check Monitoring${NC}"
echo "Killing test-server-1 (port 8081)..."
kill $(lsof -t -i:8081) 2>/dev/null || echo "Server not running"
sleep 12  # Wait for health checker to detect (10s interval + 2s buffer)

echo "Attempting tool call after server failure..."
FAILOVER_RESPONSE=$(curl -s -X POST ${BASE_URL}/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 30,
    "method": "tools/call",
    "params": {
      "name": "echo",
      "arguments": {
        "message": "Testing failover"
      }
    }
  }')
echo "$FAILOVER_RESPONSE" | jq '.'

if echo "$FAILOVER_RESPONSE" | jq -e '.result' > /dev/null; then
    TEXT=$(echo "$FAILOVER_RESPONSE" | jq -r '.result.content[0].text')
    if [[ "$TEXT" == *"test-server-2"* ]]; then
        echo -e "${GREEN}✓ Failover working! Request routed to healthy server${NC}"
    else
        echo -e "${YELLOW}⚠ Request handled, but not sure which server${NC}"
    fi
else
    echo -e "${RED}✗ Failover failed - no healthy server found${NC}"
fi
echo ""

echo "========================================="
echo "Test Suite Complete"
echo "========================================="