package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8081, "Port to listen on")
	name := flag.String("name", "mock-server", "Server name")
	flag.Parse()

	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		method, _ := req["method"].(string)
		id := req["id"]

		log.Printf("Received: %s", method)

		switch method {
		case "ping":
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result":  map[string]string{"status": "ok"},
			}
			json.NewEncoder(w).Encode(response)

		case "initialize":
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"protocolVersion": "2025-03-26",
					"capabilities":    map[string]bool{"tools": true},
					"serverInfo":      map[string]string{"name": *name, "version": "1.0"},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "tools/list":
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name":        "echo",
							"description": "Echoes your message",
							"inputSchema": map[string]interface{}{
								"type":       "object",
								"properties": map[string]interface{}{"message": map[string]string{"type": "string"}},
								"required":   []string{"message"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "tools/call":
			var params map[string]interface{}
			if p, ok := req["params"].(map[string]interface{}); ok {
				params = p
			}
			toolName, _ := params["name"].(string)
			args, _ := params["arguments"].(map[string]interface{})
			msg, _ := args["message"].(string)

			resultText := fmt.Sprintf("Echo from %s: %s", *name, msg)
			if toolName != "echo" {
				resultText = "Unknown tool"
			}

			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": resultText},
					},
					"isError": false,
				},
			}
			json.NewEncoder(w).Encode(response)

		default:
			response := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"error":   map[string]interface{}{"code": -32601, "message": "Method not found"},
			}
			json.NewEncoder(w).Encode(response)
		}
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Mock server '%s' starting on %s", *name, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
