// Package cacp provides a Go SDK for CACP.
//
// CACP is a universal messaging and RPC layer for AI agent interoperability.
// This SDK provides a client for interacting with the broker's HTTP and WebSocket APIs.
package cacp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version is the SDK version.
const Version = "0.1.0"

// Default configuration values.
const (
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * time.Second
)

// Logger interface for observability.
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})
}

// DefaultLogger is a simple logger implementation.
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, fields ...map[string]interface{}) {
	log.Printf("[DEBUG] %s", msg)
}

func (l *DefaultLogger) Info(msg string, fields ...map[string]interface{}) {
	log.Printf("[INFO] %s", msg)
}

func (l *DefaultLogger) Warn(msg string, fields ...map[string]interface{}) {
	log.Printf("[WARN] %s", msg)
}

func (l *DefaultLogger) Error(msg string, fields ...map[string]interface{}) {
	log.Printf("[ERROR] %s", msg)
}

// OnRequestCallback is called before each request.
type OnRequestCallback func(ctx context.Context, method, path string, headers map[string]string, body interface{})

// OnResponseCallback is called after each response.
type OnResponseCallback func(ctx context.Context, method, path string, statusCode int, headers map[string]string, body []byte, err error)

// Config holds the client configuration.
type Config struct {
	BaseURL     string
	APIKey      string
	JWTToken    string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	HTTPClient  *http.Client
	UserAgent   string
	Logger      Logger
	OnRequest   OnRequestCallback
	OnResponse  OnResponseCallback
}

// Client is the main CACP client.
type Client struct {
	config     *Config
	baseURL    *url.URL
	httpClient *http.Client

	agents        *AgentsAPI
	messaging     *MessagingAPI
	tasks         *TasksAPI
	groups        *GroupsAPI
	auth          *AuthAPI
	apiKeys       *APIKeysAPI
	websocket     *WebSocketClient

	phoenixChannel *PhoenixChannelClient
}

// NewClient creates a new CACP client.
func NewClient(cfg *Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Set defaults
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = DefaultRetryDelay
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = "cacp-sdk-go/" + Version
	}

	logger := cfg.Logger
	if logger == nil {
		logger = &DefaultLogger{}
	}

	client := &Client{
		config: &Config{
			BaseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
			APIKey:     cfg.APIKey,
			JWTToken:   cfg.JWTToken,
			Timeout:    timeout,
			MaxRetries: maxRetries,
			RetryDelay: retryDelay,
			UserAgent:  userAgent,
			Logger:     logger,
			OnRequest:  cfg.OnRequest,
			OnResponse: cfg.OnResponse,
		},
		baseURL:        baseURL,
		httpClient:     httpClient,
		phoenixChannel: newPhoenixChannelClient(client),
	}

	client.agents = newAgentsAPI(client)
	client.messaging = newMessagingAPI(client)
	client.tasks = newTasksAPI(client)
	client.groups = newGroupsAPI(client)
	client.auth = newAuthAPI(client)
	client.apiKeys = newAPIKeysAPI(client)
	client.websocket = newWebSocketClient(client)

	logger.Debug("CACP Go client initialized", map[string]interface{}{
		"version":  Version,
		"base_url": baseURL.String(),
	})

	return client, nil
}

// Close closes the client and releases resources.
func (c *Client) Close() error {
	c.websocket.close()
	return nil
}

// Agents returns the agents API.
func (c *Client) Agents() *AgentsAPI {
	return c.agents
}

// Messaging returns the messaging API.
func (c *Client) Messaging() *MessagingAPI {
	return c.messaging
}

// Tasks returns the tasks API.
func (c *Client) Tasks() *TasksAPI {
	return c.tasks
}

// Groups returns the groups API.
func (c *Client) Groups() *GroupsAPI {
	return c.groups
}

// Auth returns the auth API.
func (c *Client) Auth() *AuthAPI {
	return c.auth
}

// APIKeys returns the API keys API.
func (c *Client) APIKeys() *APIKeysAPI {
	return c.apiKeys
}

// WebSocket returns the WebSocket client.
func (c *Client) WebSocket() *WebSocketClient {
	return c.websocket
}

// BaseURL returns the base URL.
func (c *Client) BaseURL() string {
	return c.config.BaseURL
}

// WebSocketURL returns the WebSocket URL.
func (c *Client) WebSocketURL() string {
	return strings.Replace(c.config.BaseURL, "http://", "ws://", 1) + "/ws/v1"
}

func (c *Client) getLogger() Logger {
	return c.config.Logger
}

func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(b)
}

// request makes an HTTP request.
func (c *Client) request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBodyData []byte
	if body != nil {
		var err error
		reqBodyData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	reqURL := c.config.BaseURL + path
	requestID := generateRequestID()

	var lastErr error
	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		// Create a new reader for each attempt
		var reqBody io.Reader
		if reqBodyData != nil {
			reqBody = strings.NewReader(string(reqBodyData))
		}

		startTime := time.Now()

		req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.config.UserAgent)
		req.Header.Set("X-Request-ID", string(requestID))

		// Set authentication
		if c.config.APIKey != "" {
			req.Header.Set("X-API-Key", c.config.APIKey)
		} else if c.config.JWTToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.JWTToken)
		}

		headers := make(map[string]string)
		for k, v := range req.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// Call OnRequest callback
		if c.config.OnRequest != nil {
			c.config.OnRequest(ctx, method, path, headers, body)
		}

		c.getLogger().Debug(fmt.Sprintf("HTTP request: %s %s", method, path), map[string]interface{}{
			"request_id": requestID,
			"attempt":    attempt,
		})

		resp, err := c.httpClient.Do(req)
		duration := time.Since(startTime)

		// Call OnResponse callback
		if c.config.OnResponse != nil {
			respBody := []byte{}
			if err == nil {
				respBody, _ = io.ReadAll(resp.Body)
			}
			statusCode := 0
			if resp != nil {
				statusCode = resp.StatusCode
			}
			c.config.OnResponse(ctx, method, path, statusCode, headers, respBody, err)
		}

		if err != nil {
			lastErr = &ConnectionError{Message: err.Error()}
			c.getLogger().Error(fmt.Sprintf("HTTP connection error: %s %s", method, path), map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
				"duration":   duration,
			})
			if attempt < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay * time.Duration(attempt))
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.getLogger().Error(fmt.Sprintf("Failed to read response: %s %s", method, path), map[string]interface{}{
				"request_id": requestID,
				"error":      err.Error(),
			})
			return fmt.Errorf("failed to read response body: %w", err)
		}

		respHeaders := make(map[string]string)
		for k, v := range resp.Header {
			if len(v) > 0 {
				respHeaders[k] = v[0]
			}
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.getLogger().Debug(fmt.Sprintf("HTTP response: %s %s - %d", method, path, resp.StatusCode), map[string]interface{}{
				"request_id": requestID,
				"duration":   duration,
			})
			if result != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, result); err != nil {
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}
			}
			return nil
		}

		// Handle error responses using broker error code mapping
		apiErr := &ErrorResponse{}
		if len(respBody) > 0 {
			_ = json.Unmarshal(respBody, apiErr)
		}

		c.getLogger().Error(fmt.Sprintf("HTTP error: %s %s - %d", method, path, resp.StatusCode), map[string]interface{}{
			"request_id": requestID,
			"error_code": apiErr.Code,
			"error":      apiErr.Error,
			"duration":   duration,
		})

		// Use ErrorFromResponse for proper error code mapping
		brokerErr := ErrorFromResponse(apiErr, requestID)
		if brokerErr != nil {
			// If error comes from broker code mapping
			return brokerErr
		}

		// Fallback to status code handling
		switch resp.StatusCode {
		case 401:
			return &AuthenticationError{
				Message:   apiErr.Error,
				RequestID: requestID,
			}
		case 400:
			var field string
			if apiErr.Details != nil {
				if f, ok := apiErr.Details["field"].(string); ok {
					field = f
				}
			}
			return &ValidationError{
				Message:   apiErr.Error,
				Field:     field,
				RequestID: requestID,
			}
		case 404:
			return &NotFoundError{
				Message:   apiErr.Error,
				Code:      apiErr.Code,
				RequestID: requestID,
			}
		case 429:
			retryAfter := 0 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if d, err := time.ParseDuration(ra + "s"); err == nil {
					retryAfter = d
				}
			}
			return &RateLimitError{
				Message:    apiErr.Error,
				RetryAfter: retryAfter,
				RequestID:  requestID,
			}
		default:
			if resp.StatusCode >= 500 && attempt < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay * time.Duration(attempt))
				continue
			}
			return &APIError{
				Message:    apiErr.Error,
				Code:       apiErr.Code,
				StatusCode: resp.StatusCode,
				RequestID:  requestID,
			}
		}
	}

	return lastErr
}

// get makes a GET request.
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodGet, path, nil, result)
}

// post makes a POST request.
func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.request(ctx, http.MethodPost, path, body, result)
}

// put makes a PUT request.
func (c *Client) put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.request(ctx, http.MethodPut, path, body, result)
}

// patch makes a PATCH request.
func (c *Client) patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.request(ctx, http.MethodPatch, path, body, result)
}

// delete makes a DELETE request.
func (c *Client) delete(ctx context.Context, path string) error {
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}
