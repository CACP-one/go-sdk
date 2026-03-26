package cacp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessagingAPISend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("Expected /v1/messages path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Message{
			ID:          "msg_123",
			ToAgent:     "agent_456",
			Content:     map[string]interface{}{"text": "Hello"},
			MessageType: MessageTypeMessage,
			Status:      MessageStatusPending,
			Priority:    "normal",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	message, err := client.Messaging().Send(context.Background(), &MessageSend{
		ToAgent: "agent_456",
		Content: map[string]interface{}{"text": "Hello"},
	})
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}
	if message.ID != "msg_123" {
		t.Errorf("Send() ID = %v, want %v", message.ID, "msg_123")
	}
}

func TestMessagingAPIGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages/msg_123" {
			t.Errorf("Expected /v1/messages/msg_123 path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:          "msg_123",
			ToAgent:     "agent_456",
			FromAgent:   "agent_789",
			Content:     map[string]interface{}{"text": "Hello"},
			MessageType: MessageTypeMessage,
			Status:      MessageStatusDelivered,
			Priority:    "normal",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	message, err := client.Messaging().Get(context.Background(), "msg_123")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if message.ID != "msg_123" {
		t.Errorf("Get() ID = %v, want %v", message.ID, "msg_123")
	}
	if message.Status != MessageStatusDelivered {
		t.Errorf("Get() Status = %v, want %v", message.Status, MessageStatusDelivered)
	}
}

func TestMessagingAPIGetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:     "msg_123",
			Status: MessageStatusCompleted,
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	status, err := client.Messaging().GetStatus(context.Background(), "msg_123")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}
	if status != MessageStatusCompleted {
		t.Errorf("GetStatus() = %v, want %v", status, MessageStatusCompleted)
	}
}

func TestMessagingAPIRPCCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages/rpc" {
			t.Errorf("Expected /v1/messages/rpc path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RPCResponse{
			ID:            "rpc_123",
			FromAgent:     "agent_456",
			Result:        map[string]interface{}{"sum": float64(30)},
			ExecutionTime: 0.05,
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	response, err := client.Messaging().RPCCall(context.Background(), &RPCRequest{
		ToAgent: "agent_456",
		Method:  "add",
		Params:  map[string]interface{}{"a": 10, "b": 20},
	})
	if err != nil {
		t.Errorf("RPCCall() error = %v", err)
	}
	if response.ID != "rpc_123" {
		t.Errorf("RPCCall() ID = %v, want %v", response.ID, "rpc_123")
	}
}

func TestMessagingAPIRPCCallError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RPCResponse{
			ID:        "rpc_123",
			FromAgent: "agent_456",
			Error: map[string]interface{}{
				"code":    float64(500),
				"message": "Internal error",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.Messaging().RPCCall(context.Background(), &RPCRequest{
		ToAgent: "agent_456",
		Method:  "add",
	})
	if err == nil {
		t.Errorf("Expected RPCError, got nil")
	}
}

func TestMessagingAPIBroadcast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages/broadcast" {
			t.Errorf("Expected /v1/messages/broadcast path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"id":           "msg_1",
					"to_agent":     "agent_1",
					"content":      map[string]interface{}{"event": "test"},
					"message_type": "broadcast",
					"status":       "pending",
					"priority":     "normal",
				},
				map[string]interface{}{
					"id":           "msg_2",
					"to_agent":     "agent_2",
					"content":      map[string]interface{}{"event": "test"},
					"message_type": "broadcast",
					"status":       "pending",
					"priority":     "normal",
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	messages, err := client.Messaging().Broadcast(context.Background(), &BroadcastOptions{
		Content: map[string]interface{}{"event": "test"},
	})
	if err != nil {
		t.Errorf("Broadcast() error = %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("Broadcast() length = %v, want %v", len(messages), 2)
	}
}

func TestMessagingAPICancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages/msg_123" {
			t.Errorf("Expected /v1/messages/msg_123 path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Messaging().Cancel(context.Background(), "msg_123")
	if err != nil {
		t.Errorf("Cancel() error = %v", err)
	}
}

func TestMessagingAPIRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages/msg_123/retry" {
			t.Errorf("Expected /v1/messages/msg_123/retry path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:          "msg_new",
			ToAgent:     "agent_456",
			Content:     map[string]interface{}{"text": "Hello"},
			MessageType: MessageTypeMessage,
			Status:      MessageStatusPending,
			Priority:    "normal",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	message, err := client.Messaging().Retry(context.Background(), "msg_123")
	if err != nil {
		t.Errorf("Retry() error = %v", err)
	}
	if message.ID != "msg_new" {
		t.Errorf("Retry() ID = %v, want %v", message.ID, "msg_new")
	}
}
