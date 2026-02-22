package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	httpRequestTransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_transaction_duration_seconds",
			Help:    "Transaction request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status", "payment_type"},
	)

	httpRequestTransactionWithCodeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_transaction_with_code_duration_seconds",
			Help:    "Transaction request duration in seconds with code",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status", "payment_type", "code"},
	)
)

func main() {
	router := httprouter.New()

	// Register HTTP handlers
	router.POST("/api/transactions", handleTransactionWithRouter)
	router.Handler("GET", "/metrics", promhttp.Handler())

	log.Println("Starting server on :8080")
	log.Println("Metrics available at http://localhost:8080/metrics")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func handleTransactionWithRouter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	start := time.Now()

	transactionID := int64(rand.Intn(1000000))
	code := "TRX-" + strconv.FormatInt(transactionID, 10)
	randomStatus := TransactionStatusList[rand.Intn(len(TransactionStatusList))]
	randomPaymentMethod := PaymentMethodList[rand.Intn(len(PaymentMethodList))]
	duration := time.Duration(rand.Intn(1000000))

	transaction := Transaction{
		ID:          transactionID,
		Code:        code,
		Status:      randomStatus,
		PaymentType: randomPaymentMethod,
	}

	// Simulate some processing time
	time.Sleep(duration * time.Microsecond)

	httpRequestTransactionDuration.WithLabelValues("/api/transactions", "POST", string(randomStatus), string(randomPaymentMethod)).Observe(time.Since(start).Seconds())
	httpRequestTransactionWithCodeDuration.WithLabelValues("/api/transactions", "POST", string(randomStatus), string(randomPaymentMethod), code).Observe(time.Since(start).Seconds())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"id": ` + strconv.FormatInt(transaction.ID, 10) + `, "code": "` + transaction.Code + `", "status": "` + string(transaction.Status) + `", "payment_type": "` + string(transaction.PaymentType) + `"}`))
}
