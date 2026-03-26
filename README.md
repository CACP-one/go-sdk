# CACP Go SDK

Official Go SDK for [CACP](https://cacp.io) - the universal messaging and RPC layer for AI agent interoperability.

## Installation

```bash
go get github.com/cacp/cacp/sdks/go/cacp
```

## Quick Start

### Initialize the Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cacp/cacp/sdks/go/cacp"
)

func main() {
    // Initialize with API key
    client, err := cacp.NewClient(&cacp.Config{
        BaseURL: "https://api.cacp.io",
        APIKey:  "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
}
```

### Register an Agent

```go
agent, err := client.Agents.Register(context.Background(), &cacp.AgentRegistration{
    Name:         "my-assistant",
    Description:  "A helpful AI assistant",
    Capabilities: []string{"chat", "code-generation", "analysis"},
    Metadata: map[string]interface{}{
        "model":   "gpt-4",
        "version": "1.0",
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Registered agent: %s\n", agent.ID)
```

### Send Messages

```go
// Send a direct message
message, err := client.Messaging.Send(context.Background(), &cacp.MessageSend{
    ToAgent:  "target-agent-id",
    Content:  map[string]interface{}{"text": "Hello from Go SDK!"},
    MessageType: "request",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Message sent: %s\n", message.ID)

// RPC call with response
response, err := client.Messaging.RPCCall(context.Background(), &cacp.RPCRequest{
    ToAgent: "target-agent-id",
    Method:  "process_data",
    Params:  map[string]interface{}{"input": "some data"},
    Timeout: 30,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("RPC response: %v\n", response.Result)
```

### WebSocket Real-Time Communication

```go
ws, err := client.WebSocket.Connect(context.Background())
if err != nil {
    log.Fatal(err)
}
defer ws.Close()

// Subscribe to messages
err = ws.Subscribe(context.Background(), "my-agent-id")
if err != nil {
    log.Fatal(err)
}

// Listen for messages
go func() {
    for msg := range ws.Messages() {
        fmt.Printf("Received: %v\n", msg)
    }
}()

// Send a message
err = ws.Send(context.Background(), &cacp.WebSocketMessage{
    ToAgent: "other-agent-id",
    Content: map[string]interface{}{"response": "Got it!"},
})
```

### Query Agents by Capability

```go
agents, err := client.Agents.QueryByCapability(context.Background(), &cacp.CapabilityQuery{
    Capabilities: []string{"code-generation", "python"},
    MatchAll:     false,
})
if err != nil {
    log.Fatal(err)
}

for _, agent := range agents {
    fmt.Printf("Found: %s - %v\n", agent.Name, agent.Capabilities)
}
```

## Features

- **Idiomatic Go** - Context support, proper error handling
- **Goroutine-safe** - Thread-safe client for concurrent use
- **WebSocket support** - Phoenix Channels protocol for real-time bidirectional communication
- **Automatic retries** - Configurable retry logic with exponential backoff
- **Comprehensive error handling** - Typed errors with broker error code mapping
- **Observability** - Built-in logging and request/response callbacks
- **Request tracking** - Automatic request ID generation for debugging

## API Reference

### Client Configuration

```go
type Config struct {
    BaseURL     string                    // Broker API URL
    APIKey      string                    // API key for authentication
    JWTToken    string                    // JWT token for authentication
    Timeout     time.Duration             // Request timeout (default: 30s)
    MaxRetries  int                       // Maximum retry attempts (default: 3)
    RetryDelay  time.Duration             // Initial retry delay (default: 1s)
    Logger      Logger                    // Logger implementation (default: DefaultLogger)
    OnRequest   OnRequestCallback         // Callback before each request
    OnResponse  OnResponseCallback        // Callback after each response
}
```

### Agents API

```go
// Register a new agent
client.Agents.Register(ctx, &AgentRegistration{...}) (*Agent, error)

// Get agent by ID
client.Agents.Get(ctx, agentID) (*Agent, error)

// List all agents
client.Agents.List(ctx, &AgentListOptions{...}) (*AgentList, error)

// Update agent
client.Agents.Update(ctx, agentID, &AgentUpdate{...}) (*Agent, error)

// Delete agent
client.Agents.Delete(ctx, agentID) error

// Query by capability
client.Agents.QueryByCapability(ctx, &CapabilityQuery{...}) ([]*Agent, error)

// Semantic search
client.Agents.SemanticSearch(ctx, &SemanticSearchQuery{...}) ([]*Agent, error)
```

### Messaging API

```go
// Send a message
client.Messaging.Send(ctx, &MessageSend{...}) (*Message, error)

// Get message status
client.Messaging.Get(ctx, messageID) (*Message, error)

// RPC call with response
client.Messaging.RPCCall(ctx, &RPCRequest{...}) (*RPCResponse, error)

// Broadcast to all agents
client.Messaging.Broadcast(ctx, &BroadcastOptions{...}) ([]*Message, error)
```

### WebSocket API

```go
// Connect
ws, _ := client.WebSocket.Connect(ctx)

// Subscribe to messages
ws.Subscribe(ctx, agentID)

// Listen for messages
for msg := range ws.Messages() { ... }

// Send message
ws.Send(ctx, &WebSocketMessage{...})

// Close connection
ws.Close()
```

## Error Handling

```go
import "errors"

agent, err := client.Agents.Get(ctx, "non-existent-id")
if err != nil {
    var notFoundErr *cacp.AgentNotFoundError
    var authErr *cacp.AuthenticationError
    var validationErr *cacp.ValidationError
    var rateLimitErr *cacp.RateLimitError

    switch {
    case errors.As(err, &notFoundErr):
        fmt.Println("Agent not found:", notFoundErr.AgentID)
    case errors.As(err, &authErr):
        fmt.Println("Invalid credentials:", authErr.Error())
    case errors.As(err, &validationErr):
        fmt.Println("Validation failed:", validationErr.Field, "-", validationErr.Message)
    case errors.As(err, &rateLimitErr):
        fmt.Printf("Rate limited. Retry after %v\n", rateLimitErr.RetryAfter)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Broker Error Codes

The SDK maps broker error codes to specific error types:

- **5001-5010**: Authentication errors (AuthenticationError, ForbiddenError)
- **6001-6010**: Task errors (TaskError, TaskNotFoundError, TaskStateError)
- **7001-7010**: Group errors (GroupError, GroupNotFoundError, MemberError)
- **1001-1999**: Validation errors (ValidationError)
- **2001-2999**: Agent errors (AgentNotFoundError)
- **3001-3999**: Message errors (MessageError)

All errors include the request ID for debugging:

```go
var err *cacp.PeerError
if errors.As(err, &err) {
    fmt.Printf("Request ID: %s\n", err.RequestID)
}
```

## Observability

The SDK provides built-in observability through logging and callbacks.

### Custom Logger

```go
type MyLogger struct{}

func (l *MyLogger) Debug(msg string, fields ...map[string]interface{}) {
    // Custom debug logging
}

func (l *MyLogger) Info(msg string, fields ...map[string]interface{}) {
    // Custom info logging
}

func (l *MyLogger) Warn(msg string, fields ...map[string]interface{}) {
    // Custom warning logging
}

func (l *MyLogger) Error(msg string, fields ...map[string]interface{}) {
    // Custom error logging
}

client, err := cacp.NewClient(&cacp.Config{
    BaseURL: "https://api.cacp.io",
    APIKey:  "your-api-key",
    Logger:  &MyLogger{},
})
```

### Request/Response Callbacks

```go
client, err := cacp.NewClient(&cacp.Config{
    BaseURL: "https://api.cacp.io",
    APIKey:  "your-api-key",
    OnRequest: func(ctx context.Context, method, path string, headers map[string]string, body interface{}) {
        fmt.Printf(">> %s %s\n", method, path)
    },
    OnResponse: func(ctx context.Context, method, path string, statusCode int, headers map[string]string, body []byte, err error) {
        fmt.Printf("<< %s %s - %d\n", method, path, statusCode)
    },
})
```

### Request ID Tracking

Every request includes an `X-Request-ID` header for debugging and tracing. Errors include the request ID for correlation:

```go
err = client.Messaging.Send(ctx, &cacp.MessageSend{...})
if err != nil {
    fmt.Printf("Error (request_id: %s): %v\n", err.RequestID, err)
}
```

## License

MIT License - see [LICENSE](LICENSE) for details.
# go-sdk
