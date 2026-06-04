package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jeevanjoseph03/mcp-gateway/internal/types"
	"go.uber.org/zap"
)

// HealthChecker monitors server health
type HealthChecker struct {
	registry *Registry
	logger   *zap.Logger
	stopCh   chan struct{}
	client   *http.Client
}

// NewHealthChecker creates a health checker
func NewHealthChecker(registry *Registry, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		logger:   logger,
		stopCh:   make(chan struct{}),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start begins health checking
func (hc *HealthChecker) Start() {
	go hc.run()
	hc.logger.Info("health checker started")
}

// Stop halts health checking
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.logger.Info("health checker stopped")
}

func (hc *HealthChecker) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllServers()
		case <-hc.stopCh:
			return
		}
	}
}

func (hc *HealthChecker) checkAllServers() {
	servers := hc.registry.GetAllServers()

	for _, server := range servers {
		if time.Since(server.LastCheck) >= server.Config.HealthCheck.Interval {
			go hc.checkServer(server)
		}
	}
}

func (hc *HealthChecker) checkServer(server *BackendServer) {
	ctx, cancel := context.WithTimeout(context.Background(), server.Config.HealthCheck.Timeout)
	defer cancel()

	healthy := hc.pingServer(ctx, server)
	hc.registry.UpdateHealth(server.Config.ID, healthy)

	if healthy {
		hc.logger.Debug("health check passed", zap.String("server", server.Config.ID))
	} else {
		hc.logger.Warn("health check failed", zap.String("server", server.Config.ID))
	}
}

func (hc *HealthChecker) pingServer(ctx context.Context, server *BackendServer) bool {
	pingRequest := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "health-check",
		Method:  "ping",
		Params:  nil,
	}

	requestBody, err := json.Marshal(pingRequest)
	if err != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "POST", server.Config.URL, bytes.NewReader(requestBody))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := hc.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
