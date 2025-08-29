#!/bin/bash
# Developer Portal Setup Script

echo "Setting up Developer Portal..."

# Install dependencies
echo "Installing dependencies..."
npm install -g @stoplight/elements
npm install -g @redocly/cli

# Generate portal content
echo "Generating portal content..."
mkdir -p portal/content/{docs,guides,examples}

# Copy API documentation
cp -r docs/api/* portal/content/docs/

echo "Developer Portal setup completed"
