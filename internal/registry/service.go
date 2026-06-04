package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/jeevanjoseph03/mcp-gateway/internal/config"
)

// BackendServer represents a registered backend MCP server
type BackendServer struct {
	Config    *config.ServerConfig
	Healthy   bool
	LastCheck time.Time
	Tools     []string
}

// Registry manages all backend servers
type Registry struct {
	servers map[string]*BackendServer
	mu      sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		servers: make(map[string]*BackendServer),
	}
}

// AddServer adds a server to the registry
func (r *Registry) AddServer(cfg *config.ServerConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.servers[cfg.ID] = &BackendServer{
		Config:  cfg,
		Healthy: true,
		Tools:   cfg.Tools,
	}
}

// GetServer retrieves a server by ID
func (r *Registry) GetServer(id string) (*BackendServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[id]
	if !exists {
		return nil, fmt.Errorf("server %s not found", id)
	}
	return server, nil
}

// GetHealthyServer finds a healthy server that can handle a tool
func (r *Registry) GetHealthyServer(toolName string) (*BackendServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, server := range r.servers {
		if !server.Healthy {
			continue
		}
		if len(server.Tools) == 0 || r.serverHasTool(server, toolName) {
			return server, nil
		}
	}
	return nil, fmt.Errorf("no healthy server found for tool: %s", toolName)
}

// GetAllServers returns all registered servers
func (r *Registry) GetAllServers() []*BackendServer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := make([]*BackendServer, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}
	return servers
}

// UpdateHealth updates health status
func (r *Registry) UpdateHealth(serverID string, healthy bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if server, exists := r.servers[serverID]; exists {
		server.Healthy = healthy
		server.LastCheck = time.Now()
	}
}

// SetServerTools updates tool list
func (r *Registry) SetServerTools(serverID string, tools []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if server, exists := r.servers[serverID]; exists {
		server.Tools = tools
	}
}

func (r *Registry) serverHasTool(server *BackendServer, toolName string) bool {
	for _, tool := range server.Tools {
		if tool == toolName {
			return true
		}
	}
	return false
}
