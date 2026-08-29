// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type e2bClient struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
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
	APIKey  string
	Version string
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
		baseURL: parsed,
		apiKey:  strings.TrimSpace(config.APIKey),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "terraform-provider-e2b/" + config.Version,
	}, nil
}

type requestAuthMode string

const (
	requestAuthDefault requestAuthMode = "default"
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

		return fmt.Errorf("missing E2B API key")
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
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
