# Golang Prometheus Grafana Demo

A simple Go application that exposes Prometheus metrics and visualizes them using Grafana, all running in Docker containers.

## Project Structure

```
.
├── main.go              # Go application with Prometheus metrics
├── go.mod               # Go module dependencies
├── Dockerfile           # Docker image for Go app
├── docker-compose.yml   # Docker Compose configuration
├── prometheus.yml       # Prometheus configuration
└── README.md           # This file
```

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 1.29 or higher)

## Quick Start

### Start all services

```bash
docker-compose up -d
```

This will start three services:
- **golang-app**: Go application on port 8080
- **prometheus**: Prometheus server on port 9090
- **grafana**: Grafana dashboard on port 3000

## Accessing the Services

### Go Application
- URL: http://localhost:8080
- Metrics endpoint: http://localhost:8080/metrics

Available endpoints:
- `POST /api/transactions` - Transaction endpoint
- `GET /metrics` - Prometheus metrics

### Prometheus
- URL: http://localhost:9090
- Status > Targets: http://localhost:9090/targets

### Grafana
- URL: http://localhost:3000
- Default credentials:
  - Username: `admin`
  - Password: `admin`

## Setting up Grafana Dashboard

### Step 1: Add Prometheus Data Source

1. Open Grafana at http://localhost:3000
2. Log in with username `admin` and password `admin`
3. Go to **Configuration** (gear icon) > **Data Sources**
4. Click **Add data source**
5. Select **Prometheus**
6. Configure:
   - Name: `Prometheus`
   - URL: `http://prometheus:9090`
7. Click **Save & Test**
