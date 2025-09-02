#!/bin/bash

echo "📊 Setting up Grafana Dashboards..."

# Import dashboards via Grafana API
GRAFANA_URL="http://localhost:3001"
GRAFANA_USER="admin"
GRAFANA_PASS="admin"

# Wait for Grafana to be ready
until curl -s "$GRAFANA_URL/api/health" > /dev/null; do
  echo "Waiting for Grafana..."
  sleep 2
done

echo "✅ Grafana is ready!"

# Add Prometheus data source
curl -X POST \
  -H "Content-Type: application/json" \
  -u "$GRAFANA_USER:$GRAFANA_PASS" \
  -d '{
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://prometheus:9090",
    "access": "proxy",
    "isDefault": true
  }' \
  "$GRAFANA_URL/api/datasources"

echo "✅ Prometheus data source added"

# Import popular dashboards
DASHBOARDS=(
  "1860"  # Node Exporter
  "3662"  # Prometheus Stats
  "6417"  # Kubernetes Cluster
  "11074" # Node Exporter for Prometheus
  "8588"  # Cadvisor exporter
)

for dashboard_id in "${DASHBOARDS[@]}"; do
  echo "Importing dashboard $dashboard_id..."
  curl -X POST \
    -H "Content-Type: application/json" \
    -u "$GRAFANA_USER:$GRAFANA_PASS" \
    -d "{
      \"dashboard\": {
        \"id\": null,
        \"uid\": null
      },
      \"overwrite\": true,
      \"inputs\": [{
        \"name\": \"DS_PROMETHEUS\",
        \"type\": \"datasource\",
        \"pluginId\": \"prometheus\",
        \"value\": \"Prometheus\"
      }]
    }" \
    "$GRAFANA_URL/api/dashboards/import"
done

echo "✅ Dashboards imported successfully!"
echo ""
echo "Access Grafana at: $GRAFANA_URL"
echo "Login: admin/admin"
