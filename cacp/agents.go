package cacp

import (
	"context"
	"time"
)

// AgentStatus represents the status of an agent.
type AgentStatus string

// Agent status constants.
const (
	AgentStatusOnline      AgentStatus = "online"
	AgentStatusOffline     AgentStatus = "offline"
	AgentStatusDegraded    AgentStatus = "degraded"
	AgentStatusError       AgentStatus = "error"
	AgentStatusMaintenance AgentStatus = "maintenance"
)

// Agent represents a registered agent.
type Agent struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Capabilities   []string               `json:"capabilities"`
	Status         AgentStatus            `json:"status"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	CreatedAt      *time.Time             `json:"created_at,omitempty"`
	UpdatedAt      *time.Time             `json:"updated_at,omitempty"`
	LastSeenAt     *time.Time             `json:"last_seen_at,omitempty"`
	MatchScore     float64                `json:"match_score,omitempty"`
}

// AgentRegistration represents a request to register an agent.
type AgentRegistration struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AgentUpdate represents a request to update an agent.
type AgentUpdate struct {
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Status       AgentStatus            `json:"status,omitempty"`
}

// AgentListOptions represents options for listing agents.
type AgentListOptions struct {
	Status AgentStatus `json:"status,omitempty"`
	Limit  int         `json:"limit,omitempty"`
	Offset int         `json:"offset,omitempty"`
}

// AgentList represents a paginated list of agents.
type AgentList struct {
	Agents []*Agent `json:"agents"`
	Total  int      `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
}

// CapabilityQuery represents a capability-based agent query.
type CapabilityQuery struct {
	Capabilities []string    `json:"capabilities"`
	MatchAll     bool        `json:"match_all,omitempty"`
	Status       AgentStatus `json:"status,omitempty"`
	Limit        int         `json:"limit,omitempty"`
}

// SemanticSearchQuery represents a semantic search query.
type SemanticSearchQuery struct {
	Query     string  `json:"query"`
	Limit     int     `json:"limit,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
}

// AgentsAPI provides access to agent-related operations.
type AgentsAPI struct {
	client *Client
}

func newAgentsAPI(client *Client) *AgentsAPI {
	return &AgentsAPI{client: client}
}

// Register registers a new agent.
func (a *AgentsAPI) Register(ctx context.Context, registration *AgentRegistration) (*Agent, error) {
	if registration == nil {
		return nil, &ValidationError{Message: "registration is required", Field: "registration"}
	}
	if registration.Name == "" {
		return nil, &ValidationError{Message: "name is required", Field: "name"}
	}
	if len(registration.Capabilities) == 0 {
		return nil, &ValidationError{Message: "at least one capability is required", Field: "capabilities"}
	}

	var agent Agent
	err := a.client.post(ctx, "/v1/agents", registration, &agent)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Get retrieves an agent by ID.
func (a *AgentsAPI) Get(ctx context.Context, agentID string) (*Agent, error) {
	if agentID == "" {
		return nil, &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}

	var agent Agent
	err := a.client.get(ctx, "/v1/agents/"+agentID, &agent)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// List retrieves a list of agents.
func (a *AgentsAPI) List(ctx context.Context, options *AgentListOptions) (*AgentList, error) {
	var result AgentList
	err := a.client.get(ctx, "/v1/agents", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an agent.
func (a *AgentsAPI) Update(ctx context.Context, agentID string, update *AgentUpdate) (*Agent, error) {
	if agentID == "" {
		return nil, &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}

	var agent Agent
	err := a.client.patch(ctx, "/v1/agents/"+agentID, update, &agent)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Delete deletes an agent.
func (a *AgentsAPI) Delete(ctx context.Context, agentID string) error {
	if agentID == "" {
		return &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}
	return a.client.delete(ctx, "/v1/agents/"+agentID)
}

// QueryByCapability queries agents by capability.
func (a *AgentsAPI) QueryByCapability(ctx context.Context, query *CapabilityQuery) ([]*Agent, error) {
	if query == nil || len(query.Capabilities) == 0 {
		return nil, &ValidationError{Message: "capabilities are required", Field: "capabilities"}
	}

	var result struct {
		Agents []*Agent `json:"agents"`
	}
	err := a.client.post(ctx, "/v1/agents/query", query, &result)
	if err != nil {
		return nil, err
	}
	return result.Agents, nil
}

// SemanticSearch performs a semantic search for agents.
func (a *AgentsAPI) SemanticSearch(ctx context.Context, query *SemanticSearchQuery) ([]*Agent, error) {
	if query == nil || query.Query == "" {
		return nil, &ValidationError{Message: "query is required", Field: "query"}
	}

	var result struct {
		Agents []*Agent `json:"agents"`
	}
	err := a.client.post(ctx, "/v1/agents/semantic-search", query, &result)
	if err != nil {
		return nil, err
	}
	return result.Agents, nil
}

// SetStatus sets an agent's status.
func (a *AgentsAPI) SetStatus(ctx context.Context, agentID string, status AgentStatus) (*Agent, error) {
	if agentID == "" {
		return nil, &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}
	return a.Update(ctx, agentID, &AgentUpdate{Status: status})
}

// Heartbeat sends a heartbeat for an agent.
func (a *AgentsAPI) Heartbeat(ctx context.Context, agentID string) error {
	if agentID == "" {
		return &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}
	return a.client.post(ctx, "/v1/agents/"+agentID+"/heartbeat", nil, nil)
}

// HealthMetric represents a health metric for an agent.
type HealthMetric struct {
	AgentID    string                 `json:"agent_id"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// HealthStatus represents the health status of an agent.
type HealthStatus struct {
	AgentID     string          `json:"agent_id"`
	Status      AgentStatus     `json:"status"`
	HealthScore float64         `json:"health_score"`
	Metrics     []HealthMetric  `json:"metrics"`
	LastCheck   time.Time       `json:"last_check"`
	Issues      []string        `json:"issues"`
}

// GetHealth retrieves the health status of an agent.
func (a *AgentsAPI) GetHealth(ctx context.Context, agentID string) (*HealthStatus, error) {
	if agentID == "" {
		return nil, &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}

	var health HealthStatus
	err := a.client.get(ctx, "/v1/agents/"+agentID+"/health", &health)
	if err != nil {
		return nil, err
	}
	return &health, nil
}

// DiscoverRequest represents a request to discover agents.
type DiscoverRequest struct {
	Query     string  `json:"query"`
	Threshold float64 `json:"threshold,omitempty"`
	Limit     int     `json:"limit,omitempty"`
}

// DiscoverResponse represents the response from a discover request.
type DiscoverResponse struct {
	Agents []*Agent `json:"agents"`
	Total  int      `json:"total"`
	Query  string   `json:"query"`
}

// Discover discovers agents using natural language semantic search.
func (a *AgentsAPI) Discover(ctx context.Context, req *DiscoverRequest) ([]*Agent, error) {
	if req == nil || req.Query == "" {
		return nil, &ValidationError{Message: "query is required", Field: "query"}
	}

	if req.Threshold == 0 {
		req.Threshold = 0.7
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	var result DiscoverResponse
	err := a.client.post(ctx, "/v1/agents/discover", req, &result)
	if err != nil {
		return nil, err
	}
	return result.Agents, nil
}
