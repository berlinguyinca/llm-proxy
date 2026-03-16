# LLM Proxy Monitoring Setup

This directory contains monitoring configuration for Grafana dashboards and Prometheus metrics collection.

## Quick Start

### Option 1: Standalone Monitoring Stack

```bash
# Start the monitoring stack (requires proxy service already running)
docker-compose -f grafana-docker-compose.yml up -d

# Access Grafana
open http://localhost:3000
# Admin credentials: admin / admin (or set GRAFANA_ADMIN_PASSWORD in .env)

# Access Prometheus
open http://localhost:9090
```

### Option 2: Full Stack with Proxy

If you haven't started the proxy service yet:

```bash
# Start everything together
docker-compose up -d

# Or run just monitoring alongside your existing setup
docker-compose -f grafana-docker-compose.yml up -d
```

## Configuration Files

### Prometheus (`prometheus.yml`)

Scrapes metrics from the LLM proxy at `http://proxy:9999/metrics`.

**Key settings:**
- `scrape_interval`: 15s
- `scrape_timeout`: 10s
- Targets: Prometheus self-monitoring + LLM Proxy

### Grafana Dashboards

Five pre-configured dashboards available for import:

| Dashboard | UID | Description |
|-----------|-----|-------------|
| **Overview** | `llm-proxy-overview` | Main health at-a-glance with request rate, connections, error rates, latency percentiles, token bucket status |
| **Performance & Latency** | `llm-proxy-latency` | Detailed latency analysis with p50/p95/p99 percentiles, model comparison, duration histograms |
| **Rate Limiting** | `llm-proxy-ratelimit` | Token bucket monitoring, available tokens, refill rate, blocked requests |
| **Resource Utilization** | `llm-proxy-resources` | CPU, memory, disk I/O, network throughput |
| **Model Stats** | `llm-proxy-models` | Requests per model, routing efficiency, model latency comparison |

### Alert Rules (`prometheus/rules/alert-rules.yml`)

Six alert rules for proactive monitoring:

1. **HighErrorRate** - Error rate > 5% for 2 minutes
2. **HighLatency** - p95 latency > 5 seconds for 1 minute
3. **RateLimitExhaustion** - >100 req/s being rate limited for 1 minute
4. **HighMemoryUsage** - Memory usage > 85% for 2 minutes
5. **ConnectionPoolExhaustion** - Active connections > 900 for 1 minute
6. **TokenBucketLow** - Available tokens < 1000 for 2 minutes

## Prometheus Metrics Exposed

The LLM proxy exposes these metrics at `/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `proxy_requests_total` | Counter | Total requests by status and model |
| `proxy_request_duration_seconds` | Histogram | Request latency distribution by model |
| `proxy_active_connections` | Gauge | Current active connection count |
| `proxy_token_bucket_available_tokens` | Gauge | Available tokens in rate limiter |
| `process_resident_memory_bytes` | Gauge | Process memory usage |

## Security Considerations

1. **Never expose `/metrics` to the public internet** - Use internal networking only
2. Set strong passwords for Grafana admin and Prometheus metrics
3. Configure TLS if running in production clusters
4. Review Prometheus security best practices: https://prometheus.io/docs/operating/security/

## Deployment Checklist

- [ ] Start proxy service with `METRICS_ENABLED=true` (default)
- [ ] Verify metrics endpoint is accessible: `curl http://localhost:9999/metrics | head -20`
- [ ] Start monitoring stack with `docker-compose -f grafana-docker-compose.yml up -d`
- [ ] Access Grafana at http://localhost:3000 (admin/admin)
- [ ] Import dashboards via Grafana UI or upload JSON files
- [ ] Review alert rules and adjust thresholds if needed
- [ ] Configure notification channels (Slack, email, PagerDuty) for alerts

## Troubleshooting

### Metrics not being scraped?

```bash
# Check Prometheus status
curl http://localhost:9090/-/ready

# Check targets in Prometheus UI
open http://localhost:9090/targets
```

### Dashboard import fails?

Ensure Grafana is fully initialized:
```bash
docker-compose -f grafana-docker-compose.yml logs grafana | grep "Database not ready"
```

### No data showing in dashboards?

Verify scrape configuration:
```bash
curl http://localhost:9090/api/v1/query \
  -d 'query=proxy_requests_total' | jq '.data.result[0].value'
```

## Environment Variables

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
nano .env  # Edit with your values
```

Key variables:
- `METRICS_ENABLED=true` - Enable Prometheus metrics endpoint
- `MONITORING_PORT=9999` - Port for metrics endpoint
- `GRAFANA_ADMIN_PASSWORD` - Grafana admin password

## Next Steps

1. Import dashboards in Grafana UI (Dashboards → Import)
2. Configure notification channels in Grafana (Configuration → Alerting → Contact Points)
3. Set up alert rules in Prometheus UI or use provided `alert-rules.yml`
4. Monitor metrics and adjust thresholds as needed
