package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total requests",
		},
		[]string{"method", "status"},
	)

	// ADD THIS NEW METRIC:
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
		},
		[]string{"server"},
	)
)

func main() {
	http.HandleFunc("/mcp", handleMCP)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	fmt.Println("Gateway on http://localhost:8080")
	// Initialize circuit breaker as closed (0)
	circuitBreakerState.WithLabelValues("test-server-1").Set(0)
	fmt.Println("Metrics on http://localhost:8080/metrics")
	http.ListenAndServe(":8080", nil)
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req map[string]interface{}
	json.Unmarshal(body, &req)
	method, _ := req["method"].(string)
	id := req["id"]

	// Forward to mock server
	resp, err := http.Post("http://localhost:8081/mcp", "application/json", bytes.NewReader(body))

	status := "success"
	var response []byte

	if err != nil {
		// Mock server down - return simple response
		status = "mock_fallback"
		response, _ = json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  map[string]string{"message": "Fallback response"},
		})
	} else {
		defer resp.Body.Close()
		response, _ = io.ReadAll(resp.Body)
	}

	requestsTotal.WithLabelValues(method, status).Inc()
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
