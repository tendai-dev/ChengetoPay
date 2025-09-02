#!/bin/bash

echo "🔐 Generating SSL Certificates..."

# Create SSL directory
mkdir -p ssl

# Generate private key
openssl genrsa -out ssl/server.key 2048

# Generate certificate signing request
openssl req -new -key ssl/server.key -out ssl/server.csr -subj "/C=US/ST=State/L=City/O=ChengeToPay/CN=api.chengetopay.local"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 -in ssl/server.csr -signkey ssl/server.key -out ssl/server.crt

# Generate Diffie-Hellman parameters for extra security
openssl dhparam -out ssl/dhparam.pem 2048

echo "✅ SSL certificates generated in ssl/ directory"
echo ""
echo "Files created:"
echo "  - ssl/server.key (private key)"
echo "  - ssl/server.crt (certificate)"
echo "  - ssl/server.csr (certificate signing request)"
echo "  - ssl/dhparam.pem (DH parameters)"
