# CACP Go SDK

Official Go SDK for [CACP](https://cacp.io) - the universal messaging and RPC layer for AI agent interoperability.

CACP enables AI agents built on different frameworks (LangChain, AutoGen, CrewAI, Vertex AI, Claude Agents, etc.) to communicate using a single open standard protocol.

## Features

- ✅ **Complete API Coverage** - All broker endpoints (Agents, Messaging, Tasks, Groups, Auth, API Keys)
- 🔄 **Context-based Operations** - Full support for Go context cancellation and timeouts
- 🌐 **WebSocket Support** - Real-time communication using Phoenix Channels protocol
- 🔐 **Flexible Authentication** - API keys and JWT token support with user registration
- 📦 **Type Safety** - Statically typed structs with proper JSON tags
- 🛡️ **Error Handling** - Comprehensive error types with broker error codes (1001-6002)
- 🔄 **Auto Retry** - Configurable retry logic with exponential backoff
- 📊 **Rate Limiting** - Built-in rate limit handling
- 🔍 **Semantic Discovery** - Natural language agent discovery
- 👥 **Team Management** - Groups API for agent teams
- 📋 **Task Tracking** - Long-running asynchronous task management
- 🎯 **Production Ready** - Idiomatic Go code with proper error handling

## Installation

```bash
go get github.com/cacp/cacp/sdks/go/cacp
```

**Minimum Go version:** 1.21+

**Dependencies:**
- `github.com/gorilla/websocket` - WebSocket support

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cacp/cacp/sdks/go/cacp"
)

func main() {
    // Initialize client with API key
    cfg := &cacp.Config{
        BaseURL: "http://localhost:4001",
        APIKey:  "your-api-key",
    }

    client, err := cacp.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Register an agent
    agent, err := client.Agents().Register(ctx, &cacp.AgentCreate{
        Name:         "my-agent",
        Capabilities: []string{"chat", "analysis"},
        Metadata:     map[string]interface{}{"version": "1.0.0"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Registered agent: %s\n", agent.AgentID)

    // Send a message
    message, err := client.Messaging().Send(ctx, &cacp.MessageSend{
        SenderID:     agent.AgentID,
        RecipientID:  "other-agent-id",
        MessageType:  "chat",
        Payload:      map[string]interface{}{"text": "Hello, world!"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Sent message: %s\n", message.MessageID)
}
```

## Authentication

### Using API Key (Recommended for Services)

```go
cfg := &cacp.Config{
    BaseURL: "http://localhost:4001",
    APIKey:  "your-api-key-here",
}

client, err := cacp.NewClient(cfg)
```

### Using JWT Token (For Users)

```go
cfg := &cacp.Config{
    BaseURL:  "http://localhost:4001",
    JWTToken: "your-jwt-token-here",
}

client, err := cacp.NewClient(cfg)
```

### User Registration & Login

```go
cfg := &cacp.Config{
    BaseURL: "http://localhost:4001",
}

client, err := cacp.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// Register a new user
registerResp, err := client.Auth().Register(ctx, &cacp.AuthRegisterRequest{
    UserName: "john_doe",
    Email:    "john@example.com",
    Password: "secure-password",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("✓ User registered: %s\n", registerResp.User.UserID)

// Login to get JWT token
loginResp, err := client.Auth().Login(ctx, &cacp.AuthLoginRequest{
    Email:    "john@example.com",
    Password: "secure-password",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("✓ JWT token: %s\n", loginResp.AccessToken)

// Use the token
authCfg := &cacp.Config{
    BaseURL:  "http://localhost:4001",
    JWTToken: loginResp.AccessToken,
}
authenticatedClient, err := cacp.NewClient(authCfg)

// Refresh token when expired
refreshResp, err := client.Auth().RefreshToken(ctx, &cacp.AuthRefreshRequest{
    RefreshToken: loginResp.RefreshToken,
})
```

## API Modules

### 1. Agents API

Manage agent registration, discovery, and health.

```go
// Register an agent
agent, err := client.Agents().Register(ctx, &cacp.AgentCreate{
    Name:         "analysis-agent",
    Capabilities: []string{"analysis", "reporting", "financial"},
    Metadata: map[string]interface{}{
        "version":     "1.0.0",
        "description": "Financial analysis agent",
        "model":       "gpt-4",
    },
})

// List all agents
agents, err := client.Agents().List(ctx)
for _, agent := range agents.Agents {
    fmt.Printf("%s: %v\n", agent.Name, agent.Capabilities)
}

// Get agent by ID
agent, err := client.Agents().Get(ctx, "agent-123")

// Update agent
updated, err := client.Agents().Update(ctx, "agent-123", &cacp.AgentUpdate{
    Capabilities: []string{"analysis", "reporting", "visualization"},
})

// Delete agent
err = client.Agents().Delete(ctx, "agent-123")

// Query agents by capability
agents, err := client.Agents().Query(ctx, "financial")

// Semantic agent discovery (NEW!)
discovered, err := client.Agents().Discover(ctx, &cacp.DiscoverRequest{
    Query: "Find agents that can analyze financial data and generate reports",
    Limit: 5,
})

for _, agent := range discovered.Agents {
    fmt.Printf("✓ Found: %s\n", agent.Name)
    fmt.Printf("  Capabilities: %v\n", agent.Capabilities)
    fmt.Printf("  Match score: %.2f\n", agent.MatchScore)
}
```

### 2. Messaging API

Send and receive messages between agents.

```go
// Send a message
message, err := client.Messaging().Send(ctx, &cacp.MessageSend{
    SenderID:    "agent-123",
    RecipientID: "agent-456",
    MessageType: "task",
    Payload: map[string]interface{}{
        "task_id":     "task-789",
        "description": "Analyze Q4 financial data",
        "data": map[string]interface{}{
            "dataset": "q4_data.csv",
            "metrics": []string{"revenue", "profit", "growth"},
        },
    },
})
fmt.Printf("✓ Message sent: %s\n", message.MessageID)

// RPC call (method invocation)
response, err := client.Messaging().RPCCall(ctx, &cacp.RPCRequest{
    SenderID:    "agent-123",
    RecipientID: "agent-456",
    Method:      "process_data",
    Params:      map[string]interface{}{"input": "some data"},
    Timeout:     30,
})
fmt.Printf("✓ RPC response: %v\n", response.Result)

// Get message status
status, err := client.Messaging().GetStatus(ctx, "msg-123")
fmt.Printf("Status: %s\n", status.Status)

// Get message details
message, err := client.Messaging().Get(ctx, "msg-123")
```

### 3. Tasks API (NEW!)

Manage long-running asynchronous tasks.

```go
// Create a task
task, err := client.Tasks().Create(ctx, &cacp.TaskCreate{
    AgentID: "agent-123",
    Operation: "data-analysis",
    InputData: map[string]interface{}{
        "dataset":      "financial_data.csv",
        "analysis_type": "time_series",
    },
    Metadata: map[string]interface{}{"priority": "high"},
})
fmt.Printf("✓ Task created: %s\n", task.TaskID)

// List tasks
tasks, err := client.Tasks().List(ctx, &cacp.TaskListOptions{
    AgentID: "agent-123",
    Status:  "running",
})

// Get task details
task, err := client.Tasks().Get(ctx, "task-123")
fmt.Printf("Task status: %s\n", task.Status)

// Cancel a task
err = client.Tasks().Cancel(ctx, "task-123")

// Retry a failed task
task, err = client.Tasks().Retry(ctx, "task-123")

// Poll task status
for task.Status == "pending" || task.Status == "running" {
    task, err = client.Tasks().Get(ctx, task.TaskID)
    if err != nil {
        log.Fatal(err)
    }
    time.Sleep(1 * time.Second)
}
fmt.Printf("✓ Task completed: %s, output: %v\n", task.Status, task.Output)
```

### 4. Groups API (NEW!)

Manage agent groups for team communication and broadcasting.

```go
// Create a group
group, err := client.Groups().Create(ctx, &cacp.GroupCreate{
    Name:        "data-science-team",
    Description: "Team for data science tasks",
})
fmt.Printf("✓ Group created: %s\n", group.ID)

// Add members to group
err = client.Groups().AddMember(ctx, "group-123", "agent-456")
err = client.Groups().AddMember(ctx, "group-123", "agent-789")

// List groups
groups, err := client.Groups().List(ctx)
for _, group := range groups.Groups {
    fmt.Printf("%s: %d members\n", group.Name, len(group.Members))
}

// Get group details
group, err := client.Groups().Get(ctx, "group-123")

// Update group
updated, err := client.Groups().Update(ctx, "group-123", &cacp.GroupUpdate{
    Description: "Updated description",
})

// Broadcast message to group
result, err := client.Groups().Broadcast(ctx, &cacp.GroupBroadcast{
    GroupID:     "group-123",
    SenderID:    "agent-123",
    MessageType: "task",
    Payload:     map[string]interface{}{"task": "Please analyze Q4 data"},
})
fmt.Printf("✓ Broadcast to %d agents\n", len(result.DeliveredTo))

// Remove member from group
err = client.Groups().RemoveMember(ctx, "group-123", "agent-456")

// Delete group
err = client.Groups().Delete(ctx, "group-123")
```

### 5. API Keys API (NEW!)

Manage API keys for your account.

```go
// Create an API key
keyResp, err := client.APIKeys().Create(ctx, &cacp.APIKeyCreate{
    Name:   "production-key",
    Scopes: []string{"agents:read", "agents:write", "messaging:send"},
})
apiKey := keyResp.APIKey
fmt.Printf("✓ API key created: %s\n", keyResp.KeyID)
fmt.Printf("  Key (save this!): %s\n", apiKey)

// List API keys
keys, err := client.APIKeys().List(ctx)
for _, key := range keys.APIKeys {
    fmt.Printf("%s: %s\n", key.Name, key.KeyID)
}

// Get API key details
key, err := client.APIKeys().Get(ctx, "key-123")

// Delete an API key
err = client.APIKeys().Delete(ctx, "key-123")
```

### 6. WebSocket Support

Real-time communication using Phoenix Channels protocol.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/cacp/cacp/sdks/go/cacp"
)

func main() {
    cfg := &cacp.Config{
        BaseURL:  "http://localhost:4001",
        JWTToken: "your-jwt-token",
    }

    client, err := cacp.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Connect to WebSocket
    if err := client.WebSocket().Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Join agent channel
    if err := client.WebSocket().JoinAgentChannel(ctx, "agent-123"); err != nil {
        log.Fatal(err)
    }

    // Define message handler
    handleMessage := func(msg *cacp.PhoenixMessage) {
        fmt.Printf("📨 Received: %v\n", msg.Payload)
        if msg.Event == "message" {
            // Handle incoming messages
        } else if msg.Event == "rpc_response" {
            // Handle RPC responses
        }
    }

    // Subscribe to messages
    if err := client.WebSocket().Subscribe(handleMessage); err != nil {
        log.Fatal(err)
    }

    // Send message via WebSocket
    if err := client.WebSocket().Send(ctx, &cacp.MessageSend{
        RecipientID: "agent-456",
        MessageType: "chat",
        Payload:     map[string]interface{}{"text": "Hello via WebSocket!"},
    }); err != nil {
        log.Fatal(err)
    }

    // Keep connection alive
    time.Sleep(60 * time.Second)

    // Disconnect
    client.WebSocket().Close()
}
```

## Complete Tutorial: Building a Multi-Agent System

Let's build a complete multi-agent system with agent teams, tasks, and real-time communication.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/cacp/cacp/sdks/go/cacp"
)

func buildMultiAgentSystem() error {
    cfg := &cacp.Config{
        BaseURL: "http://localhost:4001",
        APIKey:  "your-api-key",
    }

    client, err := cacp.NewClient(cfg)
    if err != nil {
        return err
    }
    defer client.Close()

    ctx := context.Background()

    // ========================================
    // Step 1: Register specialized agents
    // ========================================
    fmt.Println("Step 1: Registering agents...")

    dataAgent, err := client.Agents().Register(ctx, &cacp.AgentCreate{
        Name:         "data-analyst",
        Capabilities: []string{"data-analysis", "statistics", "visualization"},
        Metadata:     map[string]interface{}{"specialty": "financial data"},
    })
    if err != nil {
        return err
    }
    fmt.Printf("  ✓ Registered: %s\n", dataAgent.Name)

    reportAgent, err := client.Agents().Register(ctx, &cacp.AgentCreate{
        Name:         "report-writer",
        Capabilities: []string{"writing", "reporting", "formatting"},
        Metadata:     map[string]interface{}{"specialty": "business reports"},
    })
    if err != nil {
        return err
    }
    fmt.Printf("  ✓ Registered: %s\n", reportAgent.Name)

    // ========================================
    // Step 2: Create an agent team (group)
    // ========================================
    fmt.Println("\nStep 2: Creating agent team...")

    team, err := client.Groups().Create(ctx, &cacp.GroupCreate{
        Name:        "financial-analysis-team",
        Description: "Team for financial data analysis and reporting",
    })
    if err != nil {
        return err
    }
    fmt.Printf("  ✓ Created group: %s\n", team.ID)

    err = client.Groups().AddMember(ctx, team.ID, dataAgent.AgentID)
    if err != nil {
        return err
    }
    err = client.Groups().AddMember(ctx, team.ID, reportAgent.AgentID)
    if err != nil {
        return err
    }
    fmt.Println("  ✓ Added 2 members to team")

    // ========================================
    // Step 3: Create background tasks
    // ========================================
    fmt.Println("\nStep 3: Creating tasks...")

    analysisTask, err := client.Tasks().Create(ctx, &cacp.TaskCreate{
        AgentID:   dataAgent.AgentID,
        Operation: "analyze_financial_data",
        InputData: map[string]interface{}{
            "dataset": "q4_2024.csv",
            "metrics": []string{"revenue", "profit", "growth"},
        },
        Metadata: map[string]interface{}{"priority": "high"},
    })
    if err != nil {
        return err
    }
    fmt.Printf("  ✓ Created analysis task: %s\n", analysisTask.TaskID)

    // ========================================
    // Step 4: Broadcast work to team
    // ========================================
    fmt.Println("\nStep 4: Broadcasting task to team...")

    result, err := client.Groups().Broadcast(ctx, &cacp.GroupBroadcast{
        GroupID:     team.ID,
        SenderID:    dataAgent.AgentID,
        MessageType: "task",
        Payload: map[string]interface{}{
            "action": "prepare_report",
            "data":   analysisTask.TaskID,
        },
    })
    if err != nil {
        return err
    }
    fmt.Printf("  ✓ Broadcast to %d agents\n", len(result.DeliveredTo))

    // ========================================
    // Step 5: Monitor task completion
    // ========================================
    fmt.Println("\nStep 5: Monitoring task completion...")

    task := analysisTask
    for task.Status == "pending" || task.Status == "running" {
        task, err = client.Tasks().Get(ctx, task.TaskID)
        if err != nil {
            return err
        }
        time.Sleep(1 * time.Second)
    }

    if task.Status == "completed" {
        fmt.Println("  ✓ Task completed successfully")
        fmt.Printf("  Output: %v\n", task.Output)
    } else {
        fmt.Printf("  ✗ Task failed: %v\n", task.Error)
    }

    // ========================================
    // Step 6: Semantic discovery for help
    // ========================================
    fmt.Println("\nStep 6: Discovering agents for next task...")

    discovered, err := client.Agents().Discover(ctx, &cacp.DiscoverRequest{
        Query: "Find agents that can create charts and visualize data",
        Limit: 3,
    })
    if err != nil {
        return err
    }

    fmt.Printf("  Found %d agents:\n", len(discovered.Agents))
    for _, agent := range discovered.Agents {
        fmt.Printf("    - %s (score: %.2f)\n", agent.Name, agent.MatchScore)
    }

    // ========================================
    // Step 7: Cleanup
    // ========================================
    fmt.Println("\nStep 7: Cleaning up...")

    err = client.Groups().Delete(ctx, team.ID)
    if err != nil {
        return err
    }
    err = client.Agents().Delete(ctx, dataAgent.AgentID)
    if err != nil {
        return err
    }
    err = client.Agents().Delete(ctx, reportAgent.AgentID)
    if err != nil {
        return err
    }
    fmt.Println("  ✓ Cleanup complete")

    return nil
}

func main() {
    if err := buildMultiAgentSystem(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

### Basic Configuration

```go
import (
    "time"
    "github.com/cacp/cacp/sdks/go/cacp"
)

cfg := &cacp.Config{
    BaseURL:    "http://localhost:4001",
    APIKey:     "your-api-key",
    Timeout:    30 * time.Second,   // Request timeout
    MaxRetries: 3,                  // Maximum retry attempts
    RetryDelay: 1 * time.Second,    // Initial retry delay
}

client, err := cacp.NewClient(cfg)
```

### Custom Logger

```go
import (
    "log"
    "github.com/cacp/cacp/sdks/go/cacp"
)

type StructuredLogger struct{}

func (l *StructuredLogger) Debug(msg string, fields ...map[string]interface{}) {
    log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *StructuredLogger) Info(msg string, fields ...map[string]interface{}) {
    log.Printf("[INFO] %s %v", msg, fields)
}

func (l *StructuredLogger) Warn(msg string, fields ...map[string]interface{}) {
    log.Printf("[WARN] %s %v", msg, fields)
}

func (l *StructuredLogger) Error(msg string, fields ...map[string]interface{}) {
    log.Printf("[ERROR] %s %v", msg, fields)
}

cfg := &cacp.Config{
    BaseURL: "http://localhost:4001",
    APIKey:  "your-api-key",
    Logger:  &StructuredLogger{},
}

client, err := cacp.NewClient(cfg)
```

### Request/Response Callbacks

```go
import (
    "context"
    "net/http"
    "github.com/cacp/cacp/sdks/go/cacp"
)

cfg := &cacp.Config{
    BaseURL: "http://localhost:4001",
    APIKey:  "your-api-key",
    OnRequest: func(ctx context.Context, method, path string, headers map[string]string, body interface{}) {
        log.Printf("→ Request: %s %s", method, path)
    },
    OnResponse: func(ctx context.Context, method, path string, statusCode int, headers map[string]string, body []byte, err error) {
        log.Printf("← Response: %d for %s", statusCode, path)
    },
}

client, err := cacp.NewClient(cfg)
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
    "github.com/cacp/cacp/sdks/go/cacp"
)

httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

cfg := &cacp.Config{
    BaseURL:    "http://localhost:4001",
    APIKey:     "your-api-key",
    HTTPClient: httpClient,
}

client, err := cacp.NewClient(cfg)
```

## Error Handling

Comprehensive error handling with specific error types:

```go
import (
    "errors"
    "fmt"
    "log"

    cacperrors "github.com/cacp/cacp/sdks/go/cacp/errors"
    "github.com/cacp/cacp/sdks/go/cacp"
)

agent, err := client.Agents().Get(ctx, "non-existent")
if err != nil {
    var agentNotFoundErr *cacperrors.AgentNotFoundError
    var authErr *cacperrors.AuthenticationError
    var validationErr *cacperrors.ValidationError
    var rateLimitErr *cacperrors.RateLimitError

    switch {
    case errors.As(err, &agentNotFoundErr):
        fmt.Printf("✗ Agent not found: %s\n", agentNotFoundErr.Message)
        fmt.Printf("  Error code: %s\n", agentNotFoundErr.Code)
        fmt.Printf("  Request ID: %s\n", agentNotFoundErr.RequestID)
    case errors.As(err, &authErr):
        fmt.Printf("✗ Authentication failed: %s\n", authErr.Message)
    case errors.As(err, &validationErr):
        fmt.Printf("✗ Validation failed: %s - %s\n", validationErr.Field, validationErr.Message)
    case errors.As(err, &rateLimitErr):
        fmt.Printf("✗ Rate limited, retry after %v\n", rateLimitErr.RetryAfter)
    case errors.As(err, &cacperrors.APIError{}):
        apiErr := err.(*cacperrors.APIError)
        fmt.Printf("✗ CACP error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
    default:
        log.Fatal(err)
    }
}
```

### Common Error Codes

| Error Code | Error Type | Description |
|------------|------------|-------------|
| 1001 | InvalidCredentialsError | Invalid credentials |
| 1002 | AccountDisabledError | Account disabled |
| 1003 | InvalidTokenError | Invalid token |
| 1004 | QuotaExceededError | Quota exceeded |
| 2001 | AgentNotFoundError | Agent not found |
| 2002 | DuplicateAgentError | Agent already exists |
| 2003 | AgentNotInGroupError | Agent not in group |
| 3001 | MessageNotFoundError | Message not found |
| 3002 | MessageError | Invalid message format |
| 5001 | ValidationError | Validation error |
| 5002 | RateLimitError | Rate limit exceeded |
| 6001 | TaskNotFoundError | Task not found |
| 6002 | TaskStateError | Invalid task operation |
| 7001 | GroupNotFoundError | Group not found |

## Development

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run specific test
go test -v -run TestAgentRegistration
```

## Best Practices

### 1. Always Close the Client

```go
client, err := cacp.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Always defer close
```

### 2. Use Context Properly

```go
// For long-running operations, use context with timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

agent, err := client.Agents().Register(ctx, req)
```

### 3. Handle Errors Gracefully

```go
agent, err := client.Agents().Register(ctx, req)
if err != nil {
    // Check for specific error types
    if errors.Is(err, &cacperrors.ValidationError{}) {
        // Handle validation error
    } else if errors.Is(err, &cacperrors.RateLimitError{}) {
        // Implement retry logic
    }
    return err
}
```

### 4. Use Structured Logging

```go
type StructuredLogger struct{}

func (l *StructuredLogger) Debug(msg string, fields ...map[string]interface{}) {
    log.Printf("[DEBUG] %s %+v", msg, fields)
}

func (l *StructuredLogger) Info(msg string, fields ...map[string]interface{}) {
    log.Printf("[INFO] %s %+v", msg, fields)
}

func (l *StructuredLogger) Warn(msg string, fields ...map[string]interface{}) {
    log.Printf("[WARN] %s %+v", msg, fields)
}

func (l *StructuredLogger) Error(msg string, fields ...map[string]interface{}) {
    log.Printf("[ERROR] %s %+v", msg, fields)
}
```

## Links

- 📚 [Documentation](https://docs.cacp.io)
- 🐙 [GitHub Repository](https://github.com/cacp/cacp)
- 📖 [Protocol Specification](https://github.com/cacp/cacp/blob/main/spec/README.md)
- 🐛 [Issue Tracker](https://github.com/cacp/cacp/issues)
- 💬 [Discord Community](https://discord.gg/cacp)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

For questions and support:
- 📧 Email: support@cacp.io
- 💬 Discord: https://discord.gg/cacp
- 📖 Docs: https://docs.cacp.io