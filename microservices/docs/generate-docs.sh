#!/bin/bash
# Documentation Generation Script

echo "Generating API documentation..."

# Generate OpenAPI documentation
echo "Generating OpenAPI documentation..."
swagger-codegen generate -i docs/api/openapi.yaml -l html2 -o docs/generated/html

# Generate Postman collection
echo "Generating Postman collection..."
swagger-codegen generate -i docs/api/openapi.yaml -l postman -o docs/generated/postman

echo "Documentation generation completed"
