package cacp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentsAPIRegister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents" {
			t.Errorf("Expected /v1/agents path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Agent{
			ID:           "agent_123",
			Name:         "test-agent",
			Description:  "Test agent",
			Capabilities: []string{"chat", "code"},
			Status:       AgentStatusOffline,
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

	agent, err := client.Agents().Register(context.Background(), &AgentRegistration{
		Name:         "test-agent",
		Description:  "Test agent",
		Capabilities: []string{"chat", "code"},
	})
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
	if agent.ID != "agent_123" {
		t.Errorf("Register() ID = %v, want %v", agent.ID, "agent_123")
	}
}

func TestAgentsAPIGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_123" {
			t.Errorf("Expected /v1/agents/agent_123 path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Agent{
			ID:     "agent_123",
			Name:   "test-agent",
			Status: AgentStatusOnline,
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

	agent, err := client.Agents().Get(context.Background(), "agent_123")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if agent.ID != "agent_123" {
		t.Errorf("Get() ID = %v, want %v", agent.ID, "agent_123")
	}
}

func TestAgentsAPIList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents" {
			t.Errorf("Expected /v1/agents path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AgentList{
			Agents: []*Agent{
				{ID: "agent_1", Name: "agent-1", Status: AgentStatusOnline},
				{ID: "agent_2", Name: "agent-2", Status: AgentStatusOffline},
			},
			Total:  2,
			Limit:  100,
			Offset: 0,
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

	result, err := client.Agents().List(context.Background(), nil)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if result.Total != 2 {
		t.Errorf("List() Total = %v, want %v", result.Total, 2)
	}
	if len(result.Agents) != 2 {
		t.Errorf("List() Agents length = %v, want %v", len(result.Agents), 2)
	}
}

func TestAgentsAPIUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_123" {
			t.Errorf("Expected /v1/agents/agent_123 path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Agent{
			ID:     "agent_123",
			Name:   "updated-agent",
			Status: AgentStatusOnline,
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

	agent, err := client.Agents().Update(context.Background(), "agent_123", &AgentUpdate{
		Name: "updated-agent",
	})
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if agent.Name != "updated-agent" {
		t.Errorf("Update() Name = %v, want %v", agent.Name, "updated-agent")
	}
}

func TestAgentsAPIDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_123" {
			t.Errorf("Expected /v1/agents/agent_123 path, got %s", r.URL.Path)
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

	err = client.Agents().Delete(context.Background(), "agent_123")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestAgentsAPIQueryByCapability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/query" {
			t.Errorf("Expected /v1/agents/query path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agents": []interface{}{
				map[string]interface{}{
					"id":           "agent_1",
					"name":         "code-agent",
					"capabilities": []string{"code-generation", "python"},
					"status":       "online",
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

	agents, err := client.Agents().QueryByCapability(context.Background(), &CapabilityQuery{
		Capabilities: []string{"code-generation"},
		MatchAll:     false,
	})
	if err != nil {
		t.Errorf("QueryByCapability() error = %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("QueryByCapability() length = %v, want %v", len(agents), 1)
	}
}

func TestAgentsAPIGetHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_123/health" {
			t.Errorf("Expected /v1/agents/agent_123/health path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthStatus{
			AgentID:     "agent_123",
			Status:      AgentStatusOnline,
			HealthScore: 95.5,
			Metrics:     []HealthMetric{},
			Issues:      []string{},
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

	health, err := client.Agents().GetHealth(context.Background(), "agent_123")
	if err != nil {
		t.Errorf("GetHealth() error = %v", err)
	}
	if health.AgentID != "agent_123" {
		t.Errorf("GetHealth() AgentID = %v, want %v", health.AgentID, "agent_123")
	}
	if health.HealthScore != 95.5 {
		t.Errorf("GetHealth() HealthScore = %v, want %v", health.HealthScore, 95.5)
	}
}

func TestAgentsAPIHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_123/heartbeat" {
			t.Errorf("Expected /v1/agents/agent_123/heartbeat path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Agents().Heartbeat(context.Background(), "agent_123")
	if err != nil {
		t.Errorf("Heartbeat() error = %v", err)
	}
}
