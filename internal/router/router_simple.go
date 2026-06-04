package router

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/jeevanjoseph03/mcp-gateway/internal/metrics"
	"github.com/jeevanjoseph03/mcp-gateway/internal/registry"
	"github.com/jeevanjoseph03/mcp-gateway/internal/types"
)

type Router struct {
	registry *registry.Registry
	client   *http.Client
	metrics  *metrics.MetricsCollector
}

func NewRouter(reg *registry.Registry, metrics *metrics.MetricsCollector) *Router {
	return &Router{
		registry: reg,
		client:   &http.Client{Timeout: 30 * time.Second},
		metrics:  metrics,
	}
}

func (r *Router) Route(ctx context.Context, rawRequest []byte) ([]byte, error) {
	var baseReq struct {
		Method string      `json:"method"`
		ID     interface{} `json:"id"`
	}
	json.Unmarshal(rawRequest, &baseReq)

	switch baseReq.Method {
	case "tools/call":
		return r.handleToolCall(rawRequest, baseReq.ID)
	default:
		return r.defaultResponse(baseReq.ID)
	}
}

func (r *Router) handleToolCall(rawRequest []byte, id interface{}) ([]byte, error) {
	start := time.Now()
	
	var req types.JSONRPCRequest
	json.Unmarshal(rawRequest, &req)
	
	var toolCall types.ToolCallRequest
	paramsBytes, _ := json.Marshal(req.Params)
	json.Unmarshal(paramsBytes, &toolCall)
	
	// Forward to backend
	request := types.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "call",
		Method:  "tools/call",
		Params:  toolCall,
	}
	body, _ := json.Marshal(request)
	
	resp, err := r.client.Post("http://localhost:8081/mcp", "application/json", bytes.NewReader(body))
	
	// Record metrics
	status := "success"
	if err != nil {
		status = "failure"
	}
	r.metrics.RequestsTotal.WithLabelValues("tools/call", "test-server-1", status).Inc()
	r.metrics.RequestDuration.WithLabelValues("tools/call", "test-server-1").Observe(float64(time.Since(start).Milliseconds()))
	
	if err != nil {
		return r.errorResponse(id, -32603, err.Error())
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	
	var rpcResponse types.JSONRPCResponse
	json.Unmarshal(respBody, &rpcResponse)
	
	return json.Marshal(types.NewSuccessResponse(id, rpcResponse.Result))
}

func (r *Router) defaultResponse(id interface{}) ([]byte, error) {
	return json.Marshal(types.NewSuccessResponse(id, map[string]string{"status": "ok"}))
}

func (r *Router) errorResponse(id interface{}, code int, message string) ([]byte, error) {
	return json.Marshal(types.NewErrorResponse(id, code, message, nil))
}
