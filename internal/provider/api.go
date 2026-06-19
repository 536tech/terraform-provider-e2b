// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
