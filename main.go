package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
		next(rw, r)
		status := strconv.Itoa(rw.status)
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method, status).Observe(time.Since(start).Seconds())
	}
}

func main() {
	// Register HTTP handlers
	http.HandleFunc("/", metricsMiddleware(handleHome))
	http.HandleFunc("/api/users", metricsMiddleware(handleUsers))
	http.HandleFunc("/api/data", metricsMiddleware(handleData))
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting server on :8080")
	log.Println("Metrics available at http://localhost:8080/metrics")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	activeUsers.Set(float64(50 + rand.Intn(50))) // Simulate active users changing
	w.Write([]byte("Welcome to Prometheus Demo App!"))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing time
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users": ["alice", "bob", "charlie"]}`))
}

func handleData(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing time
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"data": "sample data"}`))
}
