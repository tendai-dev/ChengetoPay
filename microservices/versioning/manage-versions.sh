#!/bin/bash
# API Version Management Script

echo "Managing API versions..."

# Check version usage
echo "Checking version usage..."
curl -s "https://api.financialplatform.com/api/v1/health" | jq '.version'
curl -s "https://api.financialplatform.com/api/v2/health" | jq '.version'

# Generate deprecation notices
echo "Generating deprecation notices..."
cat > versioning/deprecation-notices.md << 'NOTICE_EOF'
# API Version Deprecation Notices

## Version v1
- **Status**: Current
- **Release Date**: 2024-01-01
- **Deprecation Date**: 2025-01-01
- **Sunset Date**: 2026-01-01

## Version v2
- **Status**: Beta
- **Release Date**: 2024-06-01
- **Deprecation Date**: 2026-06-01
- **Sunset Date**: 2027-06-01

## Migration Timeline
- **v1 to v2**: Available now
- **v2 to v3**: Planning phase
NOTICE_EOF

echo "Version management completed"
