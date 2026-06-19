// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type e2bClient struct {
	baseURL     *url.URL
	apiKey      string
	accessToken string
	teamID      string
	httpClient  *http.Client
	userAgent   string
}

type apiError struct {
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("E2B API returned status %d", e.StatusCode)
	}

	return fmt.Sprintf("E2B API returned status %d: %s", e.StatusCode, e.Body)
}

type e2bClientConfig struct {
	APIKey      string
	AccessToken string
	TeamID      string
	Version     string
}

func newE2BClientWithConfig(rawBaseURL string, config e2bClientConfig) (*e2bClient, error) {
	parsed, err := url.Parse(strings.TrimRight(rawBaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse api_url: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("api_url must be an absolute URL")
	}

	if config.Version == "" {
		config.Version = "dev"
	}

	return &e2bClient{
		baseURL:     parsed,
		apiKey:      strings.TrimSpace(config.APIKey),
		accessToken: strings.TrimSpace(config.AccessToken),
		teamID:      strings.TrimSpace(config.TeamID),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "terraform-provider-e2b/" + config.Version,
	}, nil
}

type requestAuthMode string

const (
	requestAuthDefault    requestAuthMode = "default"
	requestAuthBearer     requestAuthMode = "bearer"
	requestAuthTeamBearer requestAuthMode = "team_bearer"
)

func (c *e2bClient) newRequestWithAuth(ctx context.Context, method string, path string, query url.Values, body any, authMode requestAuthMode) (*http.Request, error) {
	u := *c.baseURL
	u.Path = strings.TrimRight(c.baseURL.Path, "/") + path
	u.RawQuery = query.Encode()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if err := c.setAuthHeaders(req, authMode); err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *e2bClient) setAuthHeaders(req *http.Request, authMode requestAuthMode) error {
	switch authMode {
	case requestAuthDefault:
		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
			return nil
		}
		if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
			if c.teamID != "" {
				req.Header.Set("X-Team-ID", c.teamID)
			}
			return nil
		}

		return fmt.Errorf("missing E2B authentication token")
	case requestAuthBearer:
		if c.accessToken == "" {
			return fmt.Errorf("missing E2B access token; set access_token or E2B_ACCESS_TOKEN")
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		return nil
	case requestAuthTeamBearer:
		if c.accessToken == "" {
			return fmt.Errorf("missing E2B access token; set access_token or E2B_ACCESS_TOKEN")
		}
		if c.teamID == "" {
			return fmt.Errorf("missing E2B team ID; set team_id or E2B_TEAM_ID")
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		req.Header.Set("X-Team-ID", c.teamID)
		return nil
	default:
		return fmt.Errorf("unknown E2B request auth mode %q", authMode)
	}
}

func (c *e2bClient) do(ctx context.Context, method string, path string, query url.Values, body any, target any) error {
	return c.doWithAuth(ctx, method, path, query, body, target, requestAuthDefault)
}

func (c *e2bClient) doWithAuth(ctx context.Context, method string, path string, query url.Values, body any, target any, authMode requestAuthMode) error {
	req, err := c.newRequestWithAuth(ctx, method, path, query, body, authMode)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &apiError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(respBody)),
		}
	}

	if target == nil || len(respBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func isNotFound(err error) bool {
	apiErr, ok := err.(*apiError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}
