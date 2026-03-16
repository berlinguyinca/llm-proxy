# Grafana Dashboards for LLM Proxy

Production-ready Grafana dashboards for monitoring your LLM Proxy with Prometheus.

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Use the provided docker-compose.yml to deploy the full stack
docker-compose -f grafana-docker-compose.yml up -d

# Access Grafana at http://localhost:3000
# Login: admin / admin123
# Dashboards will automatically load from provisioned config
```

### Option 2: Manual Import

If you have an existing Prometheus + Grafana setup:

```bash
# For each dashboard, visit:
# http://your-grafana-url/dashboard/new/import

# Or use grafana-cli to import directly:
grafana-cli dashboard import llm-proxy-overview.json --id 10001
grafana-cli dashboard import llm-proxy-performance.json --id 10002
grafana-cli dashboard import llm-proxy-rate-limit.json --id 10003
grafana-cli dashboard import llm-proxy-resources.json --id 10004
grafana-cli dashboard import llm-proxy-model-stats.json --id 10005
```

## Available Dashboards

### 1. LLM Proxy Overview (Overview)
- **Purpose**: Main health and performance dashboard
- **Panels**: Request rate, throughput, active connections, status code distribution, latency percentiles
- **Refresh**: Every 5 seconds

### 2. Performance & Latency
- **Purpose**: Deep dive into request latency and response times
- **Panels**: 
  - p95/p99 latency over time
  - Response time distribution
  - Model-specific latency comparison
- **Key Metrics**: Histogram quantiles for 50th, 95th, 99th percentiles

### 3. Rate Limiting & Token Bucket
- **Purpose**: Monitor rate limiter effectiveness
- **Panels**:
  - Rate limited requests count (429 responses)
  - Request rate by status code
  - Response code distribution
- **Key for**: Ensuring your rate limiting is working correctly

### 4. Resource Utilization
- **Purpose**: Track system resource usage
- **Panels**:
  - CPU utilization percentage
  - Memory usage (bytes)
  - Active connections count
- **Key for**: Capacity planning and overload detection

### 5. Model Stats & Routing
- **Purpose**: Monitor model-specific metrics and routing efficiency
- **Panels**:
  - Registered models count
  - Requests per model
  - Latency by model
  - Response distribution per model
- **Key for**: Understanding which models are being used most

## Configuration Files

### prometheus.yml
Configures Prometheus to scrape `/metrics` endpoint from your proxy.

```yaml
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: 'llm_proxy'
    metrics_path: '/metrics'
    static_configs:
      - targets: ['proxy:8080']
```

### prometheus-alerts.yml
Defines alert rules for common issues:
- High error rate (>10%)
- Excessive rate limiting (429 responses)
- High latency (p95 > 5 seconds)
- Memory pressure (>1.5GB)
- Connection overload (>800 connections)

### grafana-dashboards.yml
Automatically imports all dashboards into Grafana on startup.

## Deploying with Existing Stack

If you're already running your proxy and want to add monitoring:

```bash
# 1. Start Prometheus and Grafana separately
docker-compose -f grafana-docker-compose.yml up -d prometheus grafana

# 2. Verify metrics endpoint is accessible
curl http://localhost:9999/metrics

# 3. Import dashboards via Grafana UI or use grafana-cli as shown above
```

## Updating Dashboards

Edit any JSON file in the `dashboards/` directory and reload in Grafana UI,
or restart the containers if using docker-compose.

## Customizing

All panels are configured with standard Prometheus metrics. To customize:

1. Open dashboard in edit mode (gear icon → Edit)
2. Click "Dashboard settings" → "Update from JSON"
3. Paste modified dashboard JSON

Or use Grafana's built-in editing features directly in the UI.

## Support & Troubleshooting

- Check Prometheus is accessible at `http://localhost:9090`
- Verify metrics endpoint returns data: `curl http://localhost:9999/metrics`
- If dashboards don't load, check Grafana logs for provisioning errors
