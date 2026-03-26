package cacp

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a WebSocket message.
type WebSocketMessage struct {
	Type        string                 `json:"type"`
	ToAgent     string                 `json:"to_agent,omitempty"`
	FromAgent   string                 `json:"from_agent,omitempty"`
	Content     map[string]interface{} `json:"content,omitempty"`
	MessageType string                 `json:"message_type,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	Error       interface{}            `json:"error,omitempty"`
}

// WebSocketClient provides WebSocket communication.
type WebSocketClient struct {
	client         *Client
	phoenixClient *PhoenixChannelClient
	mu            sync.RWMutex
}

func newWebSocketClient(client *Client) *WebSocketClient {
	return &WebSocketClient{
		client:         client,
		phoenixClient: newPhoenixChannelClient(client),
	}
}

// Connect establishes a WebSocket connection.
func (w *WebSocketClient) Connect(ctx context.Context) error {
	return w.phoenixClient.Connect(ctx)
}

// Close closes the WebSocket connection.
func (w *WebSocketClient) Close() error {
	return w.phoenixClient.Close()
}

// IsConnected returns whether the WebSocket is connected.
func (w *WebSocketClient) IsConnected() bool {
	return w.phoenixClient.IsConnected()
}

// Subscribe subscribes to messages for an agent.
func (w *WebSocketClient) Subscribe(ctx context.Context, agentID string) error {
	w.phoenixClient.Subscribe(agentID)
	return nil
}

// Unsubscribe unsubscribes from messages for an agent.
func (w *WebSocketClient) Unsubscribe(ctx context.Context, agentID string) error {
	return w.phoenixClient.Unsubscribe(agentID)
}

// Send sends a message through WebSocket.
func (w *WebSocketClient) Send(ctx context.Context, msg *WebSocketMessage) error {
	return w.phoenixClient.Send(
		ctx,
		msg.ToAgent,
		msg.Content,
		msg.MessageType,
		msg.FromAgent,
		nil,
	)
}

// SendRPC sends an RPC request through WebSocket.
func (w *WebSocketClient) SendRPC(ctx context.Context, toAgent, method string, params map[string]interface{}, requestID, fromAgent string) error {
	return w.phoenixClient.SendRPC(ctx, toAgent, method, params, requestID, fromAgent)
}

// OnMessage sets a handler for incoming messages.
func (w *WebSocketClient) OnMessage(handler func(map[string]interface{})) {
	w.phoenixClient.OnGlobalMessage(func(msg *PhoenixMessage) {
		if message, ok := msg.Payload["message"].(map[string]interface{}); ok {
			handler(message)
		}
	})
}

// Messages returns a channel for receiving messages.
func (w *WebSocketClient) Messages() <-chan map[string]interface{} {
	return w.phoenixClient.Messages()
}

// PhoenixClient returns the underlying PhoenixChannelClient.
func (w *WebSocketClient) PhoenixClient() *PhoenixChannelClient {
	return w.phoenixClient
}