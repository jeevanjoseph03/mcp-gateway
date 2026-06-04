package types

import (
	"fmt"
)

// JSONRPCRequest represents a standard JSON-RPC 2.0 request
// JSON-RPC is the protocol MCP uses for communication - it's like a formal way
// for the client and server to talk to each other with request/response pairs
type JSONRPCRequest struct {
	// JSONRPC must be "2.0" - this identifies the protocol version
	JSONRPC string `json:"jsonrpc"`

	// ID can be a string, number, or null - used to match responses to requests
	// Think of it like a tracking number for your API call
	ID interface{} `json:"id"`

	// Method is the action to perform (e.g., "tools/list", "tools/call")
	Method string `json:"method"`

	// Params are the arguments for the method - type varies by method
	Params interface{} `json:"params,omitempty"`
}

// JSONRPCResponse is the successful response to a request
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result"`
}

// JSONRPCError is returned when something goes wrong
type JSONRPCError struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   ErrorObject `json:"error"`
}

// ErrorObject contains the details of what went wrong
// Following JSON-RPC 2.0 error codes standard
type ErrorObject struct {
	Code    int         `json:"code"`           // -32768 to 32000 are standard codes
	Message string      `json:"message"`        // Human-readable description
	Data    interface{} `json:"data,omitempty"` // Optional additional info
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700 // Invalid JSON was received
	InvalidRequest = -32600 // The JSON sent is not a valid Request object
	MethodNotFound = -32601 // The method does not exist / is not available
	InvalidParams  = -32602 // Invalid method parameter(s)
	InternalError  = -32603 // Internal JSON-RPC error
)

// InitializeRequest is sent by client to start an MCP session
// This is the first message in any MCP conversation
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"` // e.g., "2025-03-26"
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities tells the server what the client can do
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`    // Can the client provide file roots?
	Sampling *SamplingCapability `json:"sampling,omitempty"` // Can the client do LLM sampling?
}

// RootsCapability - client can provide root directories the server can access
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"` // Notify when roots change?
}

// SamplingCapability - client can generate text completions (LLM calls)
type SamplingCapability struct{}

// ClientInfo identifies the client application
type ClientInfo struct {
	Name    string `json:"name"`    // e.g., "My MCP Client"
	Version string `json:"version"` // e.g., "1.0.0"
}

// InitializeResponse is the server's reply to initialize
type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities tells the client what this server can do
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`     // Can this server provide tools?
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`   // Can this server provide prompts?
	Resources *ResourcesCapability `json:"resources,omitempty"` // Can this server provide resources?
}

// ToolsCapability - server provides executable tools/functions
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"` // Notify when tools change?
}

// PromptsCapability - server provides reusable prompt templates
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability - server provides read-only resources (files, configs, etc.)
type ResourcesCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
	Subscribe   bool `json:"subscribe,omitempty"` // Can client subscribe to changes?
}

// ServerInfo identifies the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents a function the LLM can call
// Think of this as an API endpoint the LLM can invoke
type Tool struct {
	Name        string                 `json:"name"`        // Unique identifier for this tool
	Description string                 `json:"description"` // What does this tool do? (LLM reads this)
	InputSchema map[string]interface{} `json:"inputSchema"` // JSON Schema defining parameters
}

// ToolListRequest asks the server to list all available tools
type ToolListRequest struct {
	Cursor string `json:"cursor,omitempty"` // For pagination
}

// ToolListResponse contains the list of tools
type ToolListResponse struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"` // For pagination
}

// ToolCallRequest asks to execute a specific tool
type ToolCallRequest struct {
	Name      string                 `json:"name"`      // Which tool to call
	Arguments map[string]interface{} `json:"arguments"` // Parameters for the tool
}

// ToolCallResponse contains the result of executing a tool
type ToolCallResponse struct {
	Content []Content `json:"content"` // The actual result data
	IsError bool      `json:"isError"` // True if the tool execution failed
}

// Content represents different types of tool output (text, images, etc.)
type Content struct {
	Type string `json:"type"` // "text", "image", "resource", etc.

	// For text content
	Text string `json:"text,omitempty"`

	// For image content
	Data     string `json:"data,omitempty"` // base64 encoded
	MimeType string `json:"mimeType,omitempty"`

	// For resource content
	Resource *Resource `json:"resource,omitempty"`
}

// Resource represents a read-only data source
type Resource struct {
	URI         string `json:"uri"`  // Unique identifier (like file:// or config://)
	Name        string `json:"name"` // Human-readable name
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        int64  `json:"size,omitempty"` // Size in bytes if known
}

// Helper function to create an error response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) JSONRPCError {
	return JSONRPCError{
		JSONRPC: "2.0",
		ID:      id,
		Error: ErrorObject{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// Helper function to create a success response
func NewSuccessResponse(id interface{}, result interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// Validate checks if a request has all required fields
func (r *JSONRPCRequest) Validate() error {
	if r.JSONRPC != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %s, expected 2.0", r.JSONRPC)
	}
	if r.ID == nil {
		return fmt.Errorf("missing request id")
	}
	if r.Method == "" {
		return fmt.Errorf("missing method")
	}
	return nil
}
