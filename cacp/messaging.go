package cacp

import (
	"context"
	"time"
)

// MessageType represents the type of a message.
type MessageType string

// Message type constants.
const (
	MessageTypeMessage      MessageType = "message"
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeBroadcast    MessageType = "broadcast"
	MessageTypeRPC          MessageType = "rpc"
)

// MessageStatus represents the status of a message.
type MessageStatus string

// Message status constants.
const (
	MessageStatusPending    MessageStatus = "pending"
	MessageStatusDelivered  MessageStatus = "delivered"
	MessageStatusProcessing MessageStatus = "processing"
	MessageStatusCompleted  MessageStatus = "completed"
	MessageStatusFailed     MessageStatus = "failed"
	MessageStatusTimeout    MessageStatus = "timeout"
)

// MessagePriority represents the priority of a message.
type MessagePriority string

// Message priority constants.
const (
	PriorityLow      MessagePriority = "low"
	PriorityNormal   MessagePriority = "normal"
	PriorityHigh     MessagePriority = "high"
	PriorityCritical MessagePriority = "critical"
)

// Message represents a message between agents.
type Message struct {
	ID           string                 `json:"id"`
	FromAgent    string                 `json:"from_agent,omitempty"`
	ToAgent      string                 `json:"to_agent,omitempty"`
	Content      map[string]interface{} `json:"content"`
	MessageType  MessageType            `json:"message_type"`
	Status       MessageStatus          `json:"status"`
	Priority     string                 `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    *time.Time             `json:"created_at,omitempty"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// MessageSend represents a request to send a message.
type MessageSend struct {
	ToAgent     string                 `json:"to_agent"`
	Content     map[string]interface{} `json:"content"`
	MessageType MessageType            `json:"message_type,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	TTL         int                    `json:"ttl,omitempty"`
}

// BroadcastOptions represents options for broadcasting a message.
type BroadcastOptions struct {
	Content          map[string]interface{} `json:"content"`
	MessageType      MessageType            `json:"message_type,omitempty"`
	CapabilityFilter []string               `json:"capability_filter,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RPCRequest represents an RPC request.
type RPCRequest struct {
	ToAgent string                 `json:"to_agent"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Timeout float64                `json:"timeout,omitempty"`
	ID      string                 `json:"id,omitempty"`
}

// RPCResponse represents an RPC response.
type RPCResponse struct {
	ID            string                 `json:"id"`
	Result        interface{}            `json:"result,omitempty"`
	Error         map[string]interface{} `json:"error,omitempty"`
	FromAgent     string                 `json:"from_agent"`
	ExecutionTime float64                `json:"execution_time,omitempty"`
}

// MessagingAPI provides access to messaging operations.
type MessagingAPI struct {
	client *Client
}

func newMessagingAPI(client *Client) *MessagingAPI {
	return &MessagingAPI{client: client}
}

// Send sends a message to another agent.
func (m *MessagingAPI) Send(ctx context.Context, options *MessageSend) (*Message, error) {
	if options == nil {
		return nil, &ValidationError{Message: "options are required", Field: "options"}
	}
	if options.ToAgent == "" {
		return nil, &ValidationError{Message: "to_agent is required", Field: "to_agent"}
	}

	var message Message
	err := m.client.post(ctx, "/v1/messages", options, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// Get retrieves a message by ID.
func (m *MessagingAPI) Get(ctx context.Context, messageID string) (*Message, error) {
	if messageID == "" {
		return nil, &ValidationError{Message: "message ID is required", Field: "message_id"}
	}

	var message Message
	err := m.client.get(ctx, "/v1/messages/"+messageID, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetStatus gets the status of a message.
func (m *MessagingAPI) GetStatus(ctx context.Context, messageID string) (MessageStatus, error) {
	if messageID == "" {
		return "", &ValidationError{Message: "message ID is required", Field: "message_id"}
	}

	message, err := m.Get(ctx, messageID)
	if err != nil {
		return "", err
	}
	return message.Status, nil
}

// RPCCall makes an RPC call to another agent.
func (m *MessagingAPI) RPCCall(ctx context.Context, request *RPCRequest) (*RPCResponse, error) {
	if request == nil {
		return nil, &ValidationError{Message: "request is required", Field: "request"}
	}
	if request.ToAgent == "" {
		return nil, &ValidationError{Message: "to_agent is required", Field: "to_agent"}
	}
	if request.Method == "" {
		return nil, &ValidationError{Message: "method is required", Field: "method"}
	}

	var response RPCResponse
	err := m.client.post(ctx, "/v1/messages/rpc", request, &response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		// Safely extract error message and code
		errMsg := "unknown error"
		if msg, ok := response.Error["message"].(string); ok {
			errMsg = msg
		}
		var errCode int
		if code, ok := response.Error["code"].(float64); ok {
			errCode = int(code)
		}
		return nil, &RPCError{
			Message: errMsg,
			Method:  request.Method,
			Code:    errCode,
		}
	}

	return &response, nil
}

// Broadcast broadcasts a message to all agents.
func (m *MessagingAPI) Broadcast(ctx context.Context, options *BroadcastOptions) ([]*Message, error) {
	if options == nil {
		return nil, &ValidationError{Message: "options are required", Field: "options"}
	}

	var result struct {
		Messages []*Message `json:"messages"`
	}
	err := m.client.post(ctx, "/v1/messages/broadcast", options, &result)
	if err != nil {
		return nil, err
	}
	return result.Messages, nil
}

// Cancel cancels a pending message.
func (m *MessagingAPI) Cancel(ctx context.Context, messageID string) error {
	if messageID == "" {
		return &ValidationError{Message: "message ID is required", Field: "message_id"}
	}
	return m.client.delete(ctx, "/v1/messages/"+messageID)
}

// Retry retries a failed message.
func (m *MessagingAPI) Retry(ctx context.Context, messageID string) (*Message, error) {
	if messageID == "" {
		return nil, &ValidationError{Message: "message ID is required", Field: "message_id"}
	}

	var message Message
	err := m.client.post(ctx, "/v1/messages/"+messageID+"/retry", nil, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}
