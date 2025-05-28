#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Testing Krew Manager Endpoints..."
echo "================================"

# Test 1: Check if Rancher is accessible
echo -e "\n1. Testing Rancher accessibility..."
curl -k https://localhost:9443 -w "\nStatus Code: %{http_code}\n"

# Test 2: Check if backend service is running
echo -e "\n2. Testing backend service..."
curl -k http://localhost:9000/health -w "\nStatus Code: %{http_code}\n"

# Test 3: Test cluster listing with authentication
echo -e "\n3. Testing cluster listing..."
curl -k -H "Authorization: Bearer token-kg852:kj7z26jf7fzzdfbb8sx9jv75p27wml6nh4fvfxrbtz62w44489n94v" \
     -H "x-rancher-url: https://localhost:9443" \
     http://localhost:9000/v3/clusters -w "\nStatus Code: %{http_code}\n"

# Test 4: Test plugins listing for local cluster
echo -e "\n4. Testing plugins listing..."
curl -k -H "Authorization: Bearer token-kg852:kj7z26jf7fzzdfbb8sx9jv75p27wml6nh4fvfxrbtz62w44489n94v" \
     http://localhost:9000/clusters/local/plugins -w "\nStatus Code: %{http_code}\n"

echo -e "\nTests completed!" 