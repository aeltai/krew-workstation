#!/bin/bash

# Get Rancher container ID
RANCHER_CONTAINER=$(docker ps | grep "rancher/rancher" | awk '{print $1}')
if [ -z "$RANCHER_CONTAINER" ]; then
    echo "Error: Rancher container not found"
    exit 1
fi

# Build the Docker image
docker build -t krew-manager-backend:latest .

# Tag and copy image to Rancher container
docker save krew-manager-backend:latest | docker exec -i $RANCHER_CONTAINER ctr images import -

# Create secret for Rancher credentials
docker exec $RANCHER_CONTAINER kubectl create secret generic rancher-creds \
  --from-literal=url="https://localhost:80" \
  --from-literal=token="$RANCHER_TOKEN" \
  --dry-run=client -o yaml | docker exec -i $RANCHER_CONTAINER kubectl apply -f -

# Apply the deployment
cat k8s/deployment.yaml | docker exec -i $RANCHER_CONTAINER kubectl apply -f -

# Wait for deployment
docker exec $RANCHER_CONTAINER kubectl rollout status deployment/krew-manager-backend

echo "Backend deployment complete. Service available inside Rancher container at krew-manager-backend:3000" 