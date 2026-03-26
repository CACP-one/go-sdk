package cacp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				BaseURL: "http://localhost:4001",
				APIKey:  "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				APIKey: "test-key",
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &Config{
				BaseURL: "://invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientBaseURL(t *testing.T) {
	client, err := NewClient(&Config{
		BaseURL: "http://localhost:4001/",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Base URL should have trailing slash removed
	if client.BaseURL() != "http://localhost:4001" {
		t.Errorf("BaseURL() = %v, want %v", client.BaseURL(), "http://localhost:4001")
	}
}

func TestClientWebSocketURL(t *testing.T) {
	tests := []struct {
		baseURL  string
		wsURL    string
	}{
		{"http://localhost:4001", "ws://localhost:4001/ws/v1"},
		{"https://api.cacp.io", "wss://api.cacp.io/ws/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.baseURL, func(t *testing.T) {
			client, err := NewClient(&Config{BaseURL: tt.baseURL})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			if client.WebSocketURL() != tt.wsURL {
				t.Errorf("WebSocketURL() = %v, want %v", client.WebSocketURL(), tt.wsURL)
			}
		})
	}
}

func TestClientRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json")
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("Expected X-API-Key header to be test-key")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "test-id"})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var result map[string]string
	err = client.get(context.Background(), "/test", &result)
	if err != nil {
		t.Errorf("get() error = %v", err)
	}
	if result["id"] != "test-id" {
		t.Errorf("get() result = %v, want %v", result["id"], "test-id")
	}
}

func TestClientErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    string
	}{
		{
			name:       "authentication error",
			statusCode: http.StatusUnauthorized,
			response:   map[string]string{"error": "Invalid API key"},
			wantErr:    "Invalid API key",
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			response:   map[string]string{"error": "Agent not found"},
			wantErr:    "Agent not found",
		},
		{
			name:       "validation error",
			statusCode: http.StatusBadRequest,
			response:   map[string]interface{}{"error": "Invalid input", "details": map[string]string{"field": "name"}},
			wantErr:    "Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			var result map[string]string
			err = client.get(context.Background(), "/test", &result)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("Error = %v, want %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var result map[string]string
	err = client.get(context.Background(), "/test", &result)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

func TestClientClose(t *testing.T) {
	client, err := NewClient(&Config{
		BaseURL: "http://localhost:4001",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
