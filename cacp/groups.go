package cacp

import (
	"context"
	"time"
)

// GroupMember represents a member of a group.
type GroupMember struct {
	AgentID  string    `json:"agent_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// Group represents an agent group.
type Group struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	LeaderAgentID string                 `json:"leader_agent_id,omitempty"`
	Capabilities  []string               `json:"capabilities"`
	Members       []GroupMember          `json:"members"`
	MemberCount   int                    `json:"member_count"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	OrganizationID string                `json:"organization_id,omitempty"`
	CreatedAt     *time.Time             `json:"created_at,omitempty"`
	UpdatedAt     *time.Time             `json:"updated_at,omitempty"`
}

// GroupCreate represents a request to create a group.
type GroupCreate struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	LeaderAgentID string                 `json:"leader_agent_id,omitempty"`
	Capabilities  []string               `json:"capabilities,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GroupUpdate represents a request to update a group.
type GroupUpdate struct {
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	LeaderAgentID string                 `json:"leader_agent_id,omitempty"`
	Capabilities  []string               `json:"capabilities,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GroupListOptions represents options for listing groups.
type GroupListOptions struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// GroupList represents a paginated list of groups.
type GroupList struct {
	Groups []*Group `json:"groups"`
	Total  int      `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
}

// MemberAdd represents a request to add a member to a group.
type MemberAdd struct {
	AgentID string `json:"agent_id"`
	Role    string `json:"role,omitempty"`
}

// BroadcastResult represents the result of a broadcast operation.
type BroadcastResult struct {
	Status         string   `json:"status"`
	GroupID        string   `json:"group_id"`
	Recipients     []string `json:"recipients"`
	RecipientCount int      `json:"recipient_count"`
}

// GroupsAPI provides access to group-related operations.
type GroupsAPI struct {
	client *Client
}

func newGroupsAPI(client *Client) *GroupsAPI {
	return &GroupsAPI{client: client}
}

// Create creates a new group.
func (g *GroupsAPI) Create(ctx context.Context, create *GroupCreate) (*Group, error) {
	if create == nil {
		return nil, &ValidationError{Message: "create is required", Field: "create"}
	}
	if create.Name == "" {
		return nil, &ValidationError{Message: "name is required", Field: "name"}
	}

	var group Group
	err := g.client.post(ctx, "/v1/groups", create, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// List retrieves a list of groups.
func (g *GroupsAPI) List(ctx context.Context, options *GroupListOptions) (*GroupList, error) {
	var result struct {
		Groups []Group `json:"groups"`
		Count  int     `json:"count"`
	}
	
	err := g.client.get(ctx, "/v1/groups", &result)
	if err != nil {
		return nil, err
	}

	groupList := &GroupList{
		Groups: make([]*Group, 0, len(result.Groups)),
		Total:  result.Count,
	}
	
	for i := range result.Groups {
		groupList.Groups = append(groupList.Groups, &result.Groups[i])
	}
	
	if options != nil {
		groupList.Limit = options.Limit
		groupList.Offset = options.Offset
	} else {
		groupList.Limit = 100
		groupList.Offset = 0
	}
	
	return groupList, nil
}

// Get retrieves a group by ID.
func (g *GroupsAPI) Get(ctx context.Context, groupID string) (*Group, error) {
	if groupID == "" {
		return nil, &ValidationError{Message: "group ID is required", Field: "group_id"}
	}

	var group Group
	err := g.client.get(ctx, "/v1/groups/"+groupID, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Update updates an existing group.
func (g *GroupsAPI) Update(ctx context.Context, groupID string, update *GroupUpdate) (*Group, error) {
	if groupID == "" {
		return nil, &ValidationError{Message: "group ID is required", Field: "group_id"}
	}

	var group Group
	err := g.client.put(ctx, "/v1/groups/"+groupID, update, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Delete deletes a group.
func (g *GroupsAPI) Delete(ctx context.Context, groupID string) error {
	if groupID == "" {
		return &ValidationError{Message: "group ID is required", Field: "group_id"}
	}
	return g.client.delete(ctx, "/v1/groups/"+groupID)
}

// AddMember adds a member to a group.
func (g *GroupsAPI) AddMember(ctx context.Context, groupID string, add *MemberAdd) (*GroupMember, error) {
	if groupID == "" {
		return nil, &ValidationError{Message: "group ID is required", Field: "group_id"}
	}
	if add == nil || add.AgentID == "" {
		return nil, &ValidationError{Message: "agent_id is required", Field: "agent_id"}
	}

	var member GroupMember
	err := g.client.post(ctx, "/v1/groups/"+groupID+"/members", add, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// RemoveMember removes a member from a group.
func (g *GroupsAPI) RemoveMember(ctx context.Context, groupID, agentID string) error {
	if groupID == "" {
		return &ValidationError{Message: "group ID is required", Field: "group_id"}
	}
	if agentID == "" {
		return &ValidationError{Message: "agent ID is required", Field: "agent_id"}
	}
	return g.client.delete(ctx, "/v1/groups/"+groupID+"/members/"+agentID)
}

// Broadcast broadcasts a message to all members of a group.
func (g *GroupsAPI) Broadcast(ctx context.Context, groupID string, message map[string]interface{}, excludeSender bool) (*BroadcastResult, error) {
	if groupID == "" {
		return nil, &ValidationError{Message: "group ID is required", Field: "group_id"}
	}

	broadcastData := map[string]interface{}{
		"message":        message,
		"exclude_sender": excludeSender,
	}

	var result BroadcastResult
	err := g.client.post(ctx, "/v1/groups/"+groupID+"/message", broadcastData, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}