#!/bin/bash
# AWS CloudFront Setup Script

DISTRIBUTION_NAME="financial-platform-cdn"
ORIGIN_DOMAIN="api-gateway.financialplatform.com"

echo "Setting up CloudFront distribution..."

# Create CloudFront distribution
aws cloudfront create-distribution \
  --distribution-config "{
    \"CallerReference\": \"$(date +%s)\",
    \"Comment\": \"Financial Platform CDN\",
    \"DefaultRootObject\": \"index.html\",
    \"Origins\": {
      \"Quantity\": 1,
      \"Items\": [
        {
          \"Id\": \"api-gateway\",
          \"DomainName\": \"$ORIGIN_DOMAIN\",
          \"OriginPath\": \"\",
          \"CustomOriginConfig\": {
            \"HTTPPort\": 80,
            \"HTTPSPort\": 443,
            \"OriginProtocolPolicy\": \"https-only\"
          }
        }
      ]
    },
    \"DefaultCacheBehavior\": {
      \"TargetOriginId\": \"api-gateway\",
      \"ViewerProtocolPolicy\": \"redirect-to-https\",
      \"TrustedSigners\": {
        \"Enabled\": false,
        \"Quantity\": 0
      },
      \"ForwardedValues\": {
        \"QueryString\": true,
        \"Cookies\": {
          \"Forward\": \"all\"
        },
        \"Headers\": {
          \"Quantity\": 0
        },
        \"QueryStringCacheKeys\": {
          \"Quantity\": 0
        }
      },
      \"MinTTL\": 0,
      \"DefaultTTL\": 300,
      \"MaxTTL\": 3600
    },
    \"Enabled\": true,
    \"PriceClass\": \"PriceClass_100\"
  }"

echo "CloudFront distribution created successfully"
