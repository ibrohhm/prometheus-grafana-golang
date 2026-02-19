# The Dangers of High-Cardinality Labels in Prometheus

We're all familiar with the warnings: *"Don't use user_id as a Prometheus label,"* or *"Don't use transaction codes as labels — they can crash Prometheus."*

But do we really understand **why** these are so dangerous? Before that, we need to know how Prometheus works.

## How Prometheus Works

Prometheus is an open-source systems monitoring and alerting tool that collects and stores its metrics as time-series data. It periodically scrapes metrics from your services based on the configured interval.

This is an example of the config:

```yaml
scrape_configs:
  - job_name: 'golang-app'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 5s
```

This config will tell Prometheus to:

- **Target**: send an HTTP request GET to `http://localhost:8080/metrics`
- **Periodically**: for every 5 seconds
- **Label**: with `job=golang-app`

![How Prometheus Works](how-prometheus-works.png)

## Metric Types

Prometheus has three metric types.

### Gauge

Gauges represent current measurements and reflect the current state of a system, such as CPU usage and memory usage.

```go
import "github.com/prometheus/client_golang/prometheus/promauto"

// register metrics
var activeUsers = promauto.NewGauge(
    prometheus.GaugeOpts{
        Name: "active_users",
        Help: "Number of active users",
    },
)

// record the metric
activeUsers.Set(float64(50 + rand.Intn(50))) // Simulate active users changing
```

will have example metrics:

```
active_users 93
```

### Counter

Counters measure discrete events that continuously increase over time. Common examples are the number of HTTP requests received, CPU seconds spent, and bytes sent.

```go
import "github.com/prometheus/client_golang/prometheus/promauto"

// register metrics
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"path", "method", "status"},
)

// record the metric
httpRequestsTotal.WithLabelValues("/api/data", "GET", "200").Inc()
```

For example, suppose we have two endpoints: `GET: /api/data`, `GET: /api/users`. Each of which can return either a `200` or `500` status code. The counter metrics might look like this:

```
http_requests_total{method="GET", path="/api/data",  status="200"} 17
http_requests_total{method="GET", path="/api/data",  status="500"} 0
http_requests_total{method="GET", path="/api/users", status="200"} 10
http_requests_total{method="GET", path="/api/users", status="500"} 2
```

### Histogram

A Histogram tracks the distribution of observed values. For a base metric name `<basename>`, it exposes multiple related time series:

- **`<basename>_bucket{le="..."}`** — Cumulative counters representing the number of observations that fall within each bucket boundary
- **`<basename>_sum`** — The total sum of all observed values
- **`<basename>_count`** — The count of events that have been observed

```go
import "github.com/prometheus/client_golang/prometheus/promauto"

// register metrics
var httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request latency in seconds",
        Buckets: prometheus.DefBuckets, // []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
    },
    []string{"path", "method", "status"},
)

// record the metrics
httpRequestDuration.WithLabelValues("/api/data", "GET", "200").Observe(duration)
```

Example output:

```
http_request_duration_seconds_bucket{le="0.5"}
http_request_duration_seconds_bucket{le="1"}
http_request_duration_seconds_bucket{le="+Inf"}
http_request_duration_seconds_sum
http_request_duration_seconds_count
```

> Each additional label applied to a histogram multiplies **all buckets**, as well as the `_sum` and `_count` series.

## Time Series Database (TSDB)

Prometheus collects and stores metrics as time series. Each time series is uniquely identified by a metric name and a set of labels, while each sample within the series contains a timestamp and a value. Each unique combination of labels (method, path, and status) represents a separate time series whose value increases as more requests are processed, with total `method x path x status`.

In the Counter metrics example, suppose we have two endpoints: `GET: /api/data`, `GET: /api/users`, and each of which can return either a `200` or `500` status code. This results in the following metrics:

```
http_requests_total{method="GET", path="/api/data",  status="200"} 17
http_requests_total{method="GET", path="/api/data",  status="500"} 0
http_requests_total{method="GET", path="/api/users", status="200"} 10
http_requests_total{method="GET", path="/api/users", status="500"} 2
```

Because each time series represents a *unique combination of labels*, these four label combinations produce *four distinct time series*. In the time-series database (TSDB), each of these time series is stored independently:

```
// time series 1
2026-02-19 09:00:00 | {__name__="http_requests_total", method="GET", path="/api/data", status="200"} | 15
2026-02-19 09:00:05 | {__name__="http_requests_total", method="GET", path="/api/data", status="200"} | 16
2026-02-19 09:00:10 | {__name__="http_requests_total", method="GET", path="/api/data", status="200"} | 17

// time series 2
2026-02-19 09:00:00 | {__name__="http_requests_total", method="GET", path="/api/data", status="500"} | 0
2026-02-19 09:00:05 | {__name__="http_requests_total", method="GET", path="/api/data", status="500"} | 0
2026-02-19 09:00:10 | {__name__="http_requests_total", method="GET", path="/api/data", status="500"} | 0

// time series 3
2026-02-19 09:00:00 | {__name__="http_requests_total", method="GET", path="/api/users", status="200"} | 8
2026-02-19 09:00:05 | {__name__="http_requests_total", method="GET", path="/api/users", status="200"} | 9
2026-02-19 09:00:10 | {__name__="http_requests_total", method="GET", path="/api/users", status="200"} | 10

// time series 4
2026-02-19 09:00:00 | {__name__="http_requests_total", method="GET", path="/api/users", status="500"} | 1
2026-02-19 09:00:05 | {__name__="http_requests_total", method="GET", path="/api/users", status="500"} | 2
2026-02-19 09:00:10 | {__name__="http_requests_total", method="GET", path="/api/users", status="500"} | 2
```

Now, take a look at the Histogram metrics example. In there, we have:

- name: `http_request_duration_seconds`
- label: `method`, `path`, `status`
- bucket: `[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}`

For each generated histogram metric we will have:

- **`http_request_duration_seconds_sum`** → method x path x status = 1 x 2 x 2 = **4 time series**
- **`http_request_duration_seconds_count`** → method x path x status = 1 x 2 x 2 = **4 time series**
- **`http_request_duration_seconds_bucket`** → method x path x status x bucket = 1 x 2 x 2 x 12 = **48 time series**

With just this simple histogram metric, we have **56 time series**.

Example bucket output for a single label combination:

```
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.005"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.01"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.025"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.05"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.1"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.25"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="0.5"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="1"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="2.5"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="5"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="10"}
http_request_duration_seconds_bucket{method="GET", path="/api/data", status="200", le="+Inf"}
```

## The Dangers

Let's go back to the warning: *"Don't use user_id as a Prometheus label,"* or *"Don't use transaction codes as labels — they can crash Prometheus."*

Now, imagine you want to record transaction activity using metric labels such as:

- `status`: `pending`, `paid`, `success`, `failed`
- `payment_type`: `wallet`, `cash`
- `transaction_type`: `trx-a`, `trx-b`, `trx-c`
- `transaction_code`: a unique identifier for each transaction

Here, `status` has 4 possible values, `payment_type` has 2 possible values, and `transaction_type` has 3 possible values. However, `transaction_code` is unique for every transaction and grows continuously with request volume. As a result, the number of possible values for `transaction_code` is **unbounded** and **increases over time.**

```
Status (4) × Payment Type (2) × Transaction Type (3) × Transaction Code (∞) = ∞ time series
```

This single unbounded label is enough to turn an otherwise manageable metric into a high-cardinality time-series explosion.

High cardinality labels cause:

- **Huge memory usage** — each unique label set creates a new time series
- **Rapid disk growth**
- **Slow queries**
- **Prometheus crashes**

> A wise man says, *"Never use a label whose value grows with users, requests, or time."*

**Good labels describe what something *is*, not *who* or which exact instance.**
