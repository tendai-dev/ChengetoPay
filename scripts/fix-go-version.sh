#!/bin/bash

# Fix Go version in all go.mod files from 1.24 to 1.21
echo "🔧 Fixing Go version in all services..."

# Find all go.mod files and update Go version
find /Users/mukurusystemsadministrator/Desktop/Project_X/microservices -name "go.mod" -type f | while read -r file; do
    echo "Updating: $file"
    sed -i '' 's/go 1\.24/go 1.21/g' "$file"
done

echo "✅ Go version updated to 1.21 in all services"
