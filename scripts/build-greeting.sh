#!/bin/bash
set -e

echo "Building greeting extension..."

# Build the extension
cd pkg/greeting
yarn install

# Create chart package
cd ../../
helm package ./charts/greeting/0.1.0 -d ./dist-pkg

echo "Build complete!" 