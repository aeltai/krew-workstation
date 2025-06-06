#!/bin/bash

echo "Starting troubleshooting..."

# Check if Docker is running
echo "1. Checking Docker status..."
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
else
    echo "✅ Docker is running"
fi

# Stop any existing containers
echo -e "\n2. Stopping existing containers..."
docker-compose down -v

# Start the services
echo -e "\n3. Starting services with Docker Compose..."
docker-compose up -d --build

# Wait for Rancher to be ready
echo -e "\n4. Waiting for Rancher to be ready..."
while ! curl -sk https://localhost:9443/ping >/dev/null 2>&1; do
    echo "   Waiting for Rancher to start..."
    sleep 5
done
echo "✅ Rancher is running"

# Check Rancher container logs
echo -e "\n5. Checking Rancher container logs..."
docker-compose logs rancher | tail -n 20

# Check backend container logs
echo -e "\n6. Checking backend container logs..."
docker-compose logs krew-manager-backend | tail -n 20

# Test Rancher API connection
echo -e "\n7. Testing Rancher API connection..."
curl -sk -H "Authorization: Bearer token-d5lc9:s6hg42hlqh8hqvnx7cb6vfbd8bpr664wlfxg59xzbtvvfqgcfbfbxr" \
    https://localhost:9443/v3/clusters \
    | grep -q "resourceType" && echo "✅ Rancher API connection successful" || echo "❌ Rancher API connection failed"

# Check if krew binary directory exists
echo -e "\n8. Checking krew binary directory..."
docker-compose exec krew-manager-backend ls -la /root/.krew/bin || echo "❌ Krew binary directory not found"

# Test backend API
echo -e "\n9. Testing backend API..."
curl -sk -H "Authorization: Bearer token-d5lc9:s6hg42hlqh8hqvnx7cb6vfbd8bpr664wlfxg59xzbtvvfqgcfbfbxr" \
    http://localhost:9000/clusters/local/plugins \
    | grep -q "plugins" && echo "✅ Backend API connection successful" || echo "❌ Backend API connection failed"

# Print container status
echo -e "\n10. Container status:"
docker-compose ps

echo -e "\nTroubleshooting complete. Check the results above for any issues."
echo "To view real-time logs, run: docker-compose logs -f" 