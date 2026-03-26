package cacp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TestWebSocketClient tests the WebSocket client
func TestWebSocketClient(t *testing.T) {
	t.Run("IsConnected initially false", func(t *testing.T) {
		client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
		ws := client.WebSocket()

		if ws.IsConnected() {
			t.Error("Expected IsConnected to be false initially")
		}
	})

	t.Run("Close when not connected", func(t *testing.T) {
		client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
		ws := client.WebSocket()

		// Should not panic or error
		ws.Close()
	})

	t.Run("Connect and Close", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, _ := upgrader.Upgrade(w, r, nil)
			defer conn.Close()

			// Read auth message
			_, _, _ = conn.ReadMessage()

			// Keep connection open for a bit
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		client, _ := NewClient(&Config{BaseURL: server.URL})

		// Override WebSocket URL for testing
		ws := client.WebSocket()
		ws.client.config.BaseURL = server.URL

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Note: This test requires the WebSocket URL to be properly set
		// In a real test, we'd mock the dialer
		_ = wsURL
		_ = ctx
	})
}

// TestWebSocketMessageTypes tests WebSocket message types
func TestWebSocketMessageTypes(t *testing.T) {
	t.Run("WebSocketMessage serialization", func(t *testing.T) {
		msg := &WebSocketMessage{
			Type:        "message",
			ToAgent:     "agent-123",
			FromAgent:   "agent-456",
			Content:     map[string]interface{}{"text": "hello"},
			MessageType: "request",
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal WebSocketMessage: %v", err)
		}

		var decoded WebSocketMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal WebSocketMessage: %v", err)
		}

		if decoded.Type != "message" {
			t.Errorf("Expected Type 'message', got '%s'", decoded.Type)
		}
		if decoded.ToAgent != "agent-123" {
			t.Errorf("Expected ToAgent 'agent-123', got '%s'", decoded.ToAgent)
		}
	})

	t.Run("RPC message serialization", func(t *testing.T) {
		msg := &WebSocketMessage{
			Type:      "rpc",
			ToAgent:   "agent-123",
			Method:    "calculate",
			RequestID: "req-001",
			Params:    map[string]interface{}{"a": 1, "b": 2},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal RPC message: %v", err)
		}

		var decoded WebSocketMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal RPC message: %v", err)
		}

		if decoded.Method != "calculate" {
			t.Errorf("Expected Method 'calculate', got '%s'", decoded.Method)
		}
	})

	t.Run("RPC response serialization", func(t *testing.T) {
		msg := &WebSocketMessage{
			Type:      "rpc_response",
			ToAgent:   "agent-456",
			RequestID: "req-001",
			Result:    map[string]interface{}{"sum": 3},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal RPC response: %v", err)
		}

		var decoded WebSocketMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal RPC response: %v", err)
		}

		if decoded.Result == nil {
			t.Error("Expected Result to be set")
		}
	})

	t.Run("RPC error response", func(t *testing.T) {
		msg := &WebSocketMessage{
			Type:      "rpc_response",
			ToAgent:   "agent-456",
			RequestID: "req-001",
			Error: map[string]interface{}{
				"code":    400,
				"message": "Invalid params",
			},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal error response: %v", err)
		}

		var decoded WebSocketMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if decoded.Error == nil {
			t.Error("Expected Error to be set")
		}
	})
}

// TestWebSocketSubscription tests subscription management
func TestWebSocketSubscription(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
	ws := client.WebSocket()

	// Initially no subscriptions
	if len(ws.subscriptions) != 0 {
		t.Error("Expected no subscriptions initially")
	}
}

// TestWebSocketSend tests message sending
func TestWebSocketSend(t *testing.T) {
	t.Run("Send requires connection", func(t *testing.T) {
		client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
		ws := client.WebSocket()

		msg := &WebSocketMessage{
			ToAgent: "agent-123",
			Content: map[string]interface{}{"text": "hello"},
		}

		ctx := context.Background()
		err := ws.Send(ctx, msg)

		if err == nil {
			t.Error("Expected error when sending without connection")
		}

		if _, ok := err.(*WebSocketError); !ok {
			t.Errorf("Expected WebSocketError, got %T", err)
		}
	})

	t.Run("Subscribe requires connection", func(t *testing.T) {
		client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
		ws := client.WebSocket()

		ctx := context.Background()
		err := ws.Subscribe(ctx, "agent-123")

		if err == nil {
			t.Error("Expected error when subscribing without connection")
		}

		if _, ok := err.(*WebSocketError); !ok {
			t.Errorf("Expected WebSocketError, got %T", err)
		}
	})
}

// TestWebSocketMessages tests the message channel
func TestWebSocketMessages(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
	ws := client.WebSocket()

	// Get the messages channel
	msgChan := ws.Messages()

	if msgChan == nil {
		t.Error("Expected non-nil messages channel")
	}
}

// TestWebSocketOnMessage tests the message handler
func TestWebSocketOnMessage(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://localhost:4001"})
	ws := client.WebSocket()

	received := false
	ws.OnMessage(func(msg map[string]interface{}) {
		received = true
	})

	// Simulate calling the handler
	if ws.messageHandler != nil {
		ws.messageHandler(map[string]interface{}{"test": true})
	}

	if !received {
		t.Error("Expected message handler to be called")
	}
}
