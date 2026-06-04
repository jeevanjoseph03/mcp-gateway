//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeevanjoseph03/mcp-gateway/internal/config"
	"github.com/jeevanjoseph03/mcp-gateway/internal/metrics"
	"github.com/jeevanjoseph03/mcp-gateway/internal/registry"
	"github.com/jeevanjoseph03/mcp-gateway/internal/router"
)

func main() {
	cfg, _ := config.Load("configs/config.yaml")
	
	// Create metrics
	metricsCollector := metrics.NewMetricsCollector()
	
	// Create registry
	reg := registry.NewRegistry()
	for _, serverCfg := range cfg.Servers {
		reg.AddServer(&serverCfg)
	}
	
	// Create router
	rtr := router.NewRouter(reg, metricsCollector)
	
	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		response, _ := rtr.Route(r.Context(), body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.Handle("/metrics", metricsCollector.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"healthy"}`))
	})
	
	server := &http.Server{Addr: ":8080", Handler: mux}
	
	go server.ListenAndServe()
	fmt.Println("Gateway running on :8080")
	
	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	server.Shutdown(context.Background())
}
