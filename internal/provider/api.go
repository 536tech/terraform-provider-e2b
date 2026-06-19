// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *e2bClient) createSandbox(ctx context.Context, request sandboxCreateRequest) (*sandboxResponse, error) {
	var result sandboxResponse
	if err := c.do(ctx, http.MethodPost, "/sandboxes", nil, request, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) getSandbox(ctx context.Context, sandboxID string) (*sandboxDetailResponse, error) {
	var result sandboxDetailResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/sandboxes/%s", url.PathEscape(sandboxID)), nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) listSandboxes(ctx context.Context, metadata map[string]string, states []string, limit int64) ([]listedSandboxResponse, error) {
	query := url.Values{}
	if len(metadata) > 0 {
		query.Set("metadata", encodeMetadataQuery(metadata))
	}
	if len(states) > 0 {
		query.Set("state", strings.Join(states, ","))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}

	var result []listedSandboxResponse
	if err := c.do(ctx, http.MethodGet, "/v2/sandboxes", query, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) deleteSandbox(ctx context.Context, sandboxID string) error {
	err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/sandboxes/%s", url.PathEscape(sandboxID)), nil, nil, nil)
	if isNotFound(err) {
		return nil
	}

	return err
}

func (c *e2bClient) setSandboxTimeout(ctx context.Context, sandboxID string, timeout int64) error {
	request := map[string]int64{"timeout": timeout}
	return c.do(ctx, http.MethodPost, fmt.Sprintf("/sandboxes/%s/timeout", url.PathEscape(sandboxID)), nil, request, nil)
}

func (c *e2bClient) updateSandboxNetwork(ctx context.Context, sandboxID string, network sandboxNetworkConfig) error {
	return c.do(ctx, http.MethodPut, fmt.Sprintf("/sandboxes/%s/network", url.PathEscape(sandboxID)), nil, network, nil)
}

func (c *e2bClient) createVolume(ctx context.Context, name string) (*volumeResponse, error) {
	var result volumeResponse
	if err := c.do(ctx, http.MethodPost, "/volumes", nil, volumeRequest{Name: name}, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) getVolume(ctx context.Context, volumeID string) (*volumeResponse, error) {
	var result volumeResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/volumes/%s", url.PathEscape(volumeID)), nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) listVolumes(ctx context.Context) ([]volumeResponse, error) {
	var result []volumeResponse
	if err := c.do(ctx, http.MethodGet, "/volumes", nil, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) deleteVolume(ctx context.Context, volumeID string) error {
	err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/volumes/%s", url.PathEscape(volumeID)), nil, nil, nil)
	if isNotFound(err) {
		return nil
	}

	return err
}

func (c *e2bClient) getTemplate(ctx context.Context, templateID string) (*templateResponse, error) {
	var result templateResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/templates/%s", url.PathEscape(templateID)), nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) listTemplates(ctx context.Context) ([]templateResponse, error) {
	var result []templateResponse
	if err := c.do(ctx, http.MethodGet, "/templates", nil, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) createTemplate(ctx context.Context, request templateCreateRequest) (*templateCreateResponse, error) {
	var result templateCreateResponse
	if err := c.do(ctx, http.MethodPost, "/v3/templates", nil, request, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) startTemplateBuild(ctx context.Context, templateID string, buildID string, request templateBuildStartRequest) error {
	path := fmt.Sprintf("/v2/templates/%s/builds/%s", url.PathEscape(templateID), url.PathEscape(buildID))
	return c.do(ctx, http.MethodPost, path, nil, request, nil)
}

func (c *e2bClient) getTemplateBuildStatus(ctx context.Context, templateID string, buildID string) (*templateBuildStatusResponse, error) {
	var result templateBuildStatusResponse
	path := fmt.Sprintf("/templates/%s/builds/%s/status", url.PathEscape(templateID), url.PathEscape(buildID))
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) waitForTemplateBuild(ctx context.Context, templateID string, buildID string, interval time.Duration) (*templateBuildStatusResponse, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		status, err := c.getTemplateBuildStatus(ctx, templateID, buildID)
		if err != nil {
			return nil, err
		}

		switch strings.ToLower(status.Status) {
		case "ready":
			return status, nil
		case "error":
			return nil, fmt.Errorf("template build %s for template %s failed: %s", buildID, templateID, templateBuildFailureMessage(status))
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for template build %s for template %s: %w", buildID, templateID, ctx.Err())
		case <-ticker.C:
		}
	}
}

func (c *e2bClient) deleteTemplate(ctx context.Context, templateID string) error {
	err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/templates/%s", url.PathEscape(templateID)), nil, nil, nil)
	if isNotFound(err) {
		return nil
	}

	return err
}

func templateBuildFailureMessage(status *templateBuildStatusResponse) string {
	if status.Reason != nil && status.Reason.Message != "" {
		if status.Reason.Step != "" {
			return fmt.Sprintf("%s: %s", status.Reason.Step, status.Reason.Message)
		}

		return status.Reason.Message
	}

	if len(status.Logs) > 0 {
		return status.Logs[len(status.Logs)-1]
	}

	if len(status.LogEntries) > 0 {
		return status.LogEntries[len(status.LogEntries)-1].Message
	}

	return "E2B did not return a failure reason"
}
