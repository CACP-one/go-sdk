package cacp

import (
	"context"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Role  string `json:"role,omitempty"`
}

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	PlanType string `json:"plan_type,omitempty"`
}

type AuthRegisterResponse struct {
	Token        string       `json:"token"`
	User         User         `json:"user"`
	Organization Organization `json:"organization"`
}

type AuthLoginResponse struct {
	Token          string `json:"token"`
	User           User   `json:"user"`
	OrganizationID string `json:"organization_id"`
}

type AuthTokenResponse struct {
	Token          string `json:"token"`
	User           *User  `json:"user,omitempty"`
	AgentID        string `json:"agent_id,omitempty"`
	OrganizationID string `json:"organization_id"`
	TokenType      string `json:"token_type,omitempty"`
}

type AuthRefreshResponse struct {
	Token          string `json:"token"`
	TokenType      string `json:"token_type"`
	User           *User  `json:"user,omitempty"`
	AgentID        string `json:"agent_id,omitempty"`
	OrganizationID string `json:"organization_id"`
}

type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	UserName         string `json:"user_name"`
	OrganizationName string `json:"organization_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenRequest struct {
	APIKey string `json:"api_key"`
}

type AuthAPI struct {
	client *Client
}

func newAuthAPI(client *Client) *AuthAPI {
	return &AuthAPI{client: client}
}

func (a *AuthAPI) Register(ctx context.Context, req RegisterRequest) (*AuthRegisterResponse, error) {
	var result AuthRegisterResponse
	if err := a.client.post(ctx, "/v1/auth/register", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AuthAPI) Login(ctx context.Context, req LoginRequest) (*AuthLoginResponse, error) {
	var result AuthLoginResponse
	if err := a.client.post(ctx, "/v1/auth/login", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AuthAPI) GetToken(ctx context.Context, req TokenRequest) (*AuthTokenResponse, error) {
	var result AuthTokenResponse
	if err := a.client.post(ctx, "/v1/auth/token", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AuthAPI) RefreshToken(ctx context.Context) (*AuthRefreshResponse, error) {
	var result AuthRefreshResponse
	if err := a.client.post(ctx, "/v1/auth/refresh", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}