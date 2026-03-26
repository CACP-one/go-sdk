package cacp

import (
	"context"
	"time"
)

type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

type APIKeyCreated struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	OwnerType  string     `json:"owner_type"`
	OwnerID    string     `json:"owner_id"`
	Warning    string     `json:"warning"`
}

type APIKeyListResponse struct {
	APIKeys []APIKey `json:"api_keys"`
	Total   int      `json:"total"`
}

type CreateAPIKeyRequest struct {
	Name          string   `json:"name,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
}

type APIKeysAPI struct {
	client *Client
}

func newAPIKeysAPI(client *Client) *APIKeysAPI {
	return &APIKeysAPI{client: client}
}

func (a *APIKeysAPI) Create(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyCreated, error) {
	var result APIKeyCreated
	if err := a.client.post(ctx, "/v1/api-keys", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *APIKeysAPI) List(ctx context.Context) (*APIKeyListResponse, error) {
	var result APIKeyListResponse
	if err := a.client.get(ctx, "/v1/api-keys", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *APIKeysAPI) Get(ctx context.Context, keyID string) (*APIKey, error) {
	var result APIKey
	if err := a.client.get(ctx, "/v1/api-keys/"+keyID, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *APIKeysAPI) Delete(ctx context.Context, keyID string) error {
	return a.client.delete(ctx, "/v1/api-keys/"+keyID)
}