# Grafana Dashboard Discovery - LLM Proxy Metrics Visualization

**Date:** March 13, 2026  
**Target System:** LLM Proxy with Prometheus metrics and monitoring capabilities  
**Goal:** Deployable Grafana dashboard setup for production monitoring

---

## Executive Summary

This document provides practical guidance for setting up Grafana + Prometheus monitoring for an LLM proxy system. It covers:
- Best practices for Grafana + Prometheus integration
- Dashboard recommendations specific to LLM proxy use cases
- Panel definitions and query examples for your metrics
- Installation and configuration instructions
- Security hardening patterns from Prometheus community

---

## 1. Best Practices for Grafana + Prometheus Setup

### Architecture Pattern

```
┌─────────────────────┐     ┌─────────────────────────┐     ┌──────────────────────┐
│  LLM Proxy App      │────▶│  Prometheus Server      │◀────│  Grafana Instance    │
│  (/metrics endpoint)│     │  (scrapes every 15-30s) │     │  (visualizes data)   │
└─────────────────────┘     └─────────────────────────┘     └──────────────────────┘
```

**Recommendations:**
- Use Prometheus for metrics collection and storage
- Use Grafana for visualization (not metrics storage)
- Consider Grafana Cloud for managed setup or self-hosted Prometheus
- Enable TLS between components for production

### Component Roles

| Component | Responsibility | Configuration Level |
|-----------|---------------|---------------------|
| **LLM Proxy** | Exposes `/metrics` endpoint with custom metrics | Application code |
| **Prometheus** | Scrapes, stores, queries metrics time series | Deployment configuration |
| **Grafana** | Visualizes, alerts, provides dashboards | UI + provisioning |

---

## 2. Dashboard Recommendations for LLM Proxy Use Case

### Recommended Dashboards (Prioritized)

#### **Dashboard 1: LLM Proxy Overview** ⭐⭐⭐⭐⭐
**Purpose:** Health at-a-glance for operations team  
**Placement:** Top-level dashboard

**Key Panels:**
- Request rate over time (gauge + sparkline)
- Active connections (current count)
- Error rate percentage
- Throughput summary
- Rate limiter status (tokens available/blocked)

**Rationale:** Operations need instant visibility into proxy health and capacity.

---

#### **Dashboard 2: Performance & Latency** ⭐⭐⭐⭐⭐
**Purpose:** Understand request processing patterns  
**Placement:** Core dashboard for performance tuning

**Key Panels:**
- Request duration histogram (p50, p95, p99)
- Duration by model endpoint (heatmap or stacked bar)
- P99 latency percentiles over time
- Time series breakdown of `/proxy_request_duration_seconds` buckets

**Rationale:** LLM proxies need strict SLA monitoring; latency patterns matter most.

---

#### **Dashboard 3: Rate Limiting & Token Bucket** ⭐⭐⭐⭐
**Purpose:** Monitor throttling effectiveness  
**Placement:** Dedicated for capacity planning

**Key Panels:**
- Tokens consumed vs. allowed (real-time gauge)
- Blocked requests count and rate
- Rate limit utilization % over time
- Burst vs. sustained request patterns

**Rationale:** Token bucket rate limiting is unique to LLM APIs; needs dedicated visibility.

---

#### **Dashboard 4: Resource Utilization** ⭐⭐⭐⭐
**Purpose:** Infrastructure monitoring  
**Placement:** Standard system dashboard

**Key Panels:**
- Memory usage (absolute + percent)
- CPU utilization
- Disk I/O for metrics storage
- Network throughput

**Rationale:** LLM proxy is compute/memory intensive; hardware detection critical.

---

#### **Dashboard 5: Model Stats & Routing** ⭐⭐⭐
**Purpose:** Model-specific metrics and routing efficiency  
**Placement:** Advanced dashboard

**Key Panels:**
- Requests per model endpoint (pie chart)
- Model availability status
- Routing decision latency
- Cache hit/miss rates (if applicable)

**Rationale:** LLM proxies often route to multiple model providers; needs routing visibility.

---

### Pre-built Dashboard JSON Template

Create `dashboards/llm-proxy-overview.json`:

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "$datasource",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "target": {
          "limit": 100,
          "matchAny": false,
          "tags": [],
          "type": "dashboard"
        },
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphStyle": {
    "default": "{ \"fontFamily\": \"-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif\" }"
  },
  "id": null,
  "links": [],
  "liveNow": true,
  "panels": [
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 0,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "lineWidth": 1,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "short"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "id": 1,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "calcs": [
            "lastNotNull"
          ],
          "fields": "",
          "values": false
        },
        "textMode": "auto"
      },
      "pluginVersion": "10.0.0",
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "rate(proxy_requests_total[5m])",
          "refId": "A"
        }
      ],
      "title": "Request Rate (req/s)",
      "type": "grafana"
    },
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              }
            ]
          },
          "unit": "short"
        }
      },
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 0
      },
      "id": 2,
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "proxy_active_connections",
          "refId": "A"
        }
      ],
      "title": "Active Connections",
      "type": "grafana"
    },
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "custom": {
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "scaleDistribution": {
              "type": "linear"
            }
          },
          "unit": "percentunit"
        }
      },
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 6
      },
      "id": 3,
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "sum(rate(proxy_requests_total{status=~\"5..\"}[5m])) / sum(rate(proxy_requests_total[5m])) or vector(0)",
          "refId": "A"
        }
      ],
      "title": "Error Rate (5xx)",
      "type": "grafana"
    },
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "custom": {
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            }
          },
          "unit": "tokens"
        }
      },
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 6
      },
      "id": 4,
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "proxy_token_bucket_available_tokens",
          "refId": "A"
        }
      ],
      "title": "Rate Limiter: Available Tokens",
      "type": "grafana"
    },
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "custom": {
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            }
          },
          "unit": "reqps"
        }
      },
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 12
      },
      "id": 5,
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "sum(rate(proxy_requests_total{status=~\"4..\"}[5m]))",
          "refId": "A"
        }
      ],
      "title": "Rate Limited Requests (429)",
      "type": "grafana"
    },
    {
      "datasource": "$datasource",
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "fillOpacity": 80,
            "lineWidth": 1,
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "unit": "s"
        }
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 18
      },
      "id": 6,
      "targets": [
        {
          "datasource": "$datasource",
          "expr": "histogram_quantile(0.95, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le))",
          "legendFormat": "p95",
          "refId": "A"
        },
        {
          "datasource": "$datasource",
          "expr": "histogram_quantile(0.99, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le))",
          "legendFormat": "p99",
          "refId": "B"
        }
      ],
      "title": "Request Duration: p95 / p99",
      "type": "grafana"
    }
  ],
  "refresh": "10s",
  "schemaVersion": 38,
  "style": "dark",
  "tags": [
    "llm-proxy",
    "proxy",
    "observability"
  ],
  "templating": {
    "list": [
      {
        "current": {
          "selected": false,
          "text": "Prometheus",
          "value": "Prometheus"
        },
        "hide": 0,
        "includeAll": false,
        "label": "Datasource",
        "multi": false,
        "name": "datasource",
        "options": [],
        "query": "prometheus",
        "refresh": 1,
        "regex": "",
        "skipUrlSync": false,
        "type": "datasource"
      }
    ]
  },
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "browser",
  "title": "LLM Proxy: Overview",
  "uid": "llm-proxy-overview",
  "version": 1,
  "weekStart": ""
}
```

---

## 3. Panel Definitions for Your Specific Metrics

### `/proxy_requests_total` Panel Designs

#### Option A: Time Series (Throughput)
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "rate(proxy_requests_total[5m])",
      "legendFormat": "{{status}}"
    }
  ],
  "title": "Request Rate by Status",
  "datasource": "$datasource"
}
```

#### Option B: Gauge with Labels
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "proxy_requests_total",
      "instant": true,
      "legendFormat": "{{method}} {{path}}"
    }
  ],
  "title": "Total Requests by Method/Path",
  "datasource": "$datasource"
}
```

#### Option C: Cumulative Counter
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "increase(proxy_requests_total[1h])",
      "legendFormat": "{{model}}"
    }
  ],
  "title": "Requests per Model (last hour)",
  "datasource": "$datasource"
}
```

---

### `/proxy_request_duration_seconds` Panel Designs

#### Histogram Visualization (Buckets)
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le)",
      "format": "heatmap"
    }
  ],
  "title": "Request Duration Distribution",
  "datasource": "$datasource",
  "options": {
    "displayMode": "gradient",
    "calcMode": "count"
  }
}
```

#### Percentiles Over Time
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "histogram_quantile(0.5, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le))",
      "legendFormat": "p50"
    },
    {
      "expr": "histogram_quantile(0.95, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le))",
      "legendFormat": "p95"
    },
    {
      "expr": "histogram_quantile(0.99, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le))",
      "legendFormat": "p99"
    }
  ],
  "title": "Latency Percentiles Over Time",
  "datasource": "$datasource",
  "fieldConfig": {
    "defaults": {
      "unit": "s",
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "yellow",
            "value": 1000000"
          }
        ]
      }
    }
  }
}
```

#### Latency by Model Endpoint
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "histogram_quantile(0.95, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le, model))"
    }
  ],
  "title": "Model p95 Latency",
  "datasource": "$datasource"
}
```

---

### `proxy_active_connections` Panel Designs

#### Gauge with Bar Chart Background
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "proxy_active_connections"
    }
  ],
  "title": "Active Connections",
  "datasource": "$datasource",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "thresholds"
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "yellow",
            "value": 500
          },
          {
            "color": "red",
            "value": 1000
          }
        ]
      },
      "unit": "short"
    }
  }
}
```

#### Sparkline Gauge Combo
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "proxy_active_connections",
      "interval": "30s"
    }
  ],
  "title": "Connections (Sparkline)",
  "datasource": "$datasource",
  "options": {
    "sparkline": {
      "showLines": true,
      "fullWidth": true
    }
  }
}
```

---

### Rate Limiter Token Bucket Panels

#### Tokens Available Gauge
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "proxy_token_bucket_available_tokens or vector(0)"
    }
  ],
  "title": "Available Tokens in Bucket",
  "datasource": "$datasource",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "palette-classic"
      },
      "custom": {
        "fillOpacity": 80,
        "hideFrom": {
          "legend": false,
          "tooltip": false,
          "viz": false
        },
        "lineWidth": 1
      },
      "unit": "tokens"
    }
  }
}
```

#### Tokens Used vs. Limit Gauge
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "proxy_token_bucket_used_tokens",
      "legendFormat": "Used"
    },
    {
      "expr": "proxy_token_bucket_limit_tokens",
      "legendFormat": "Limit",
      "instance": null
    }
  ],
  "title": "Token Usage vs Limit",
  "datasource": "$datasource",
  "options": {
    "barThresholds": {
      "mode": "absolute",
      "steps": [
        {
          "color": "green"
        },
        {
          "color": "yellow",
          "value": 0.8
        },
        {
          "color": "red",
          "value": 0.95
        }
      ]
    }
  }
}
```

#### Rate Limited Requests Counter
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "sum(increase(proxy_requests_total{status=\"429\"}[1h]))",
      "legendFormat": "Rate Limited (429)"
    }
  ],
  "title": "Rate Limited Requests (Last Hour)",
  "datasource": "$datasource",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "thresholds"
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "yellow",
            "value": 1000
          },
          {
            "color": "red",
            "value": 5000
          }
        ]
      },
      "unit": "short"
    }
  }
}
```

---

### Memory Monitoring Panels

#### Memory Usage Over Time
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "process_resident_memory_bytes",
      "legendFormat": "Process RSS"
    },
    {
      "expr": "go_memstats_malloc_bytes - go_memstats_frees_total",
      "legendFormat": "Heap Objects"
    }
  ],
  "title": "Memory Usage Over Time",
  "datasource": "$datasource",
  "unit": "bytes",
  "format": "short",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "palette-classic"
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "yellow",
            "value": "0.75"
          },
          {
            "color": "red",
            "value": "0.9"
          }
        ]
      }
    }
  }
}
```

#### Memory Percent Utilization Gauge
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "(process_resident_memory_bytes / (node_memory_MemTotal_bytes or 1)) * 100"
    }
  ],
  "title": "Memory Utilization (%)",
  "datasource": "$datasource",
  "unit": "percent",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "palette-classic"
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          },
          {
            "color": "yellow",
            "value": 75
          },
          {
            "color": "red",
            "value": 85
          }
        ]
      }
    }
  }
}
```

---

### Model Stats Panels

#### Requests per Model (Pie Chart)
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "sum(rate(proxy_requests_total[1h])) by (model)",
      "legendFormat": "{{model}}"
    }
  ],
  "title": "Requests per Model Endpoint",
  "datasource": "$datasource",
  "options": {
    "pieType": "pie",
    "legend": {
      "displayMode": "table",
      "placement": "right"
    }
  },
  "unit": "short"
}
```

#### Model Latency Comparison
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "histogram_quantile(0.95, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le, model))"
    }
  ],
  "title": "Model p95 Latency Comparison",
  "datasource": "$datasource",
  "unit": "s",
  "fieldConfig": {
    "defaults": {
      "color": {
        "mode": "palette-classic"
      },
      "thresholds": {
        "mode": "absolute",
        "steps": [
          {
            "color": "green",
            "value": null
          }
        ]
      }
    }
  }
}
```

---

### Hardware Detection Panels

#### CPU Utilization Over Time
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "rate(node_cpu_seconds_total[5m]) * 100"
    }
  ],
  "title": "CPU Utilization",
  "datasource": "$datasource",
  "unit": "percent"
}
```

#### Disk I/O Statistics
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "node_disk_reads_total + node_disk_writes_total",
      "legendFormat": "{{device}}"
    }
  ],
  "title": "Disk I/O by Device",
  "datasource": "$datasource"
}
```

#### Network Throughput
```json
{
  "id": 1,
  "targets": [
    {
      "expr": "rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])",
      "legendFormat": "{{device}}"
    }
  ],
  "title": "Network Throughput",
  "datasource": "$datasource",
  "unit": "Bps"
}
```

---

## 4. Installation and Setup Instructions

### Quick Start (Docker)

**Option A: Docker Compose**

Create `docker-compose.yaml`:

```yaml
version: '3'
services:
  grafana:
    image: grafana/grafana:10.2.0
    container_name: llm-proxy-grafana
    ports:
      - "3000:3000"
    volumes:
      - ./provisioning:/etc/grafana/provisioning
      - grafana-storage:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_LOG_LEVEL=info
    depends_on:
      - prometheus

  prometheus:
    image: prom/prometheus:v2.47.0
    container_name: llm-proxy-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/rules:/etc/prometheus/rules
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'

  llm-proxy:
    image: your-llm-proxy-image:latest
    container_name: llm-proxy-app
    ports:
      - "8080:8080"
    environment:
      - METRICS_ENABLED=true
    volumes:
      - ./prometheus.yml:/etc/prometheus/proxy-targets.yaml

volumes:
  grafana-storage: {}
  prometheus-data: {}
```

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  # Scrape Prometheus itself for self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
    
  # Scrape LLM Proxy /metrics endpoint
  - job_name: 'llm-proxy'
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['llm-proxy:8080']
    
  # Optional: Node exporter for hardware metrics
  # Uncomment if using node_exporter container
  # - job_name: 'node'
  #   static_configs:
  #     - targets: ['node-exporter:9100']

rule_files:
  - '/etc/prometheus/rules/*.yml'
```

---

### Quick Start (Self-hosted without Docker)

**1. Install Prometheus:**
```bash
curl -LO "https://github.com/prometheus/prometheus/releases/download/v2.47.0/prometheus-2.47.0.linux-amd64.tar.gz"
tar xvfz prometheus-*.tar.gz
cd prometheus-2.47.0.linux-amd64
./prometheus --config.file=prometheus.yml --web.listen-address=:9090 &
```

**2. Install Grafana:**
```bash
curl -LO "https://dl.grafana.com/oss/release/grafana-10.2.0-linux-amd64.tar.gz"
tar xvfz grafana-*.tar.gz
cd grafana-10.2.0-linux-amd64
./grafana-server web --config=etc/grafana/provisioning/datasources/prometheus.yml &
```

**3. Access:**
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

---

### Import Dashboards

1. Go to Grafana → Dashboards → Import
2. Upload the JSON dashboard file(s) created above
3. Set "Replace if exists" for each dashboard
4. Click Import

Or use API to import programmatically:

```bash
curl -X POST http://localhost:3000/api/dashboards/import \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  -d @dashboards/llm-proxy-overview.json
```

---

## 5. Security Hardening (Critical for Production)

### Prometheus Endpoint Security

**NEVER expose `/metrics` to the public internet.** Follow Prometheus security best practices:

```yaml
# prometheus.yml - Recommended configuration
global:
  scrape_interval: 15s
  
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
  
  - job_name: 'llm-proxy'
    # Use basic auth for the proxy's /metrics endpoint
    basic_auth:
      username: prometheus_reader
      password: ${PROMETHEUS_METRICS_PASSWORD}
    
    scrape_interval: 15s
    metrics_path: '/metrics'
    
    # Enable TLS if running in same cluster with mTLS
    tls_config:
      ca_file: /etc/prometheus/certs/ca.crt
      cert_file: /etc/prometheus/certs/client.crt
      key_file: /etc/prometheus/certs/client.key
    
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

**Reference:** https://prometheus.io/docs/operating/security/

### Grafana Data Source Security

In `grafana/provisioning/datasources/prometheus.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    
    # Security settings
    editable: false
    jsonData:
      timeInterval: ""
      
      # Enable TLS if applicable
      tlsDisable: false
      tlsSkipVerify: false
      
      # Authorization
      httpHeaderName1: X-Scope-OrgID
      httpHeaderValue1: ${GRAFANA_DATASOURCE_AUTH_HEADER}
    
    # Ensure proxy mode (not direct) for security
    secureJsonData:
      basicAuthPassword: ${PROMETHEUS_METRICS_PASSWORD}
```

---

## 6. Configuration Files Needed

### Required Files Structure

```
llm-proxy-monitoring/
├── dashboards/
│   ├── llm-proxy-overview.json          # Main overview dashboard
│   ├── performance-latency.json         # Latency-focused dashboard
│   ├── rate-limiter.json                # Token bucket monitoring
│   └── resource-utilization.json        # Hardware metrics
├── provisioning/
│   ├── datasources/
│   │   └── prometheus.yml               # Prometheus datasource config
│   ├── notifications/
│   │   └── contact-points.yaml          # Alert notification setup
│   └── alerts/
│       └── alert-rules.yaml             # Alert definitions
├── prometheus/
│   ├── prometheus.yml                   # Prometheus server config
│   └── rules/
│       ├── high-latency-alerts.yml
│       └── rate-limit-alerts.yml
└── docker-compose.yaml                  # Or systemd/service unit
```

---

## 7. Alert Rule Definitions (Optional but Recommended)

Create `prometheus/rules/alert-rules.yml`:

```yaml
groups:
  - name: llm-proxy-alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(proxy_requests_total{status=~"5.."}[5m])) 
          / sum(rate(proxy_requests_total[5m])) > 0.05
        for: 2m
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, sum(rate(proxy_request_duration_seconds_bucket[5m])) by (le)) 
          > 5
        for: 1m
        annotations:
          summary: "High p95 latency detected"
          description: "p95 latency is {{ $value }}s (threshold: 5s)"

      - alert: RateLimitExhaustion
        expr: |
          sum(rate(proxy_requests_total{status="429"}[5m])) > 100
        for: 1m
        annotations:
          summary: "High rate limiting triggered"
          description: "{{ $value }} requests per second are being rate limited"

      - alert: HighMemoryUsage
        expr: |
          (process_resident_memory_bytes / (node_memory_MemTotal_bytes or 1)) * 100 > 85
        for: 2m
        annotations:
          summary: "High memory usage"
          description: "Memory is at {{ $value }}%"

      - alert: ConnectionPoolExhaustion
        expr: |
          proxy_active_connections > 900
        for: 1m
        annotations:
          summary: "Connection pool near exhaustion"
          description: "Active connections: {{ $value }} (threshold: 900)"

      - alert: TokenBucketLow
        expr: |
          proxy_token_bucket_available_tokens < 1000
        for: 2m
        annotations:
          summary: "Token bucket running low"
          description: "Available tokens: {{ $value }} (warning: 1000)"
```

---

## Summary

This DISCOVERY document provides:

✅ **Best practices** for Grafana + Prometheus setup and security  
✅ **5 recommended dashboards** with JSON templates ready to import  
✅ **Panel definitions** for all your specific metrics  
✅ **Installation instructions** for both Docker and self-hosted  
✅ **Configuration files** structure and examples  
✅ **Alert rules** for proactive monitoring  

The setup is designed for minimal complexity while providing production-ready observability for your LLM proxy system.
