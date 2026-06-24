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

func (c *e2bClient) listSandboxesMetrics(ctx context.Context, sandboxIDs []string) (map[string]sandboxMetricResponse, error) {
	query := url.Values{}
	query.Set("sandbox_ids", strings.Join(sandboxIDs, ","))

	var result sandboxesMetricsResponse
	if err := c.do(ctx, http.MethodGet, "/sandboxes/metrics", query, nil, &result); err != nil {
		return nil, err
	}

	return result.Sandboxes, nil
}

func (c *e2bClient) getSandboxMetrics(ctx context.Context, sandboxID string, start int64, end int64) ([]sandboxMetricResponse, error) {
	query := url.Values{}
	if start > 0 {
		query.Set("start", fmt.Sprintf("%d", start))
	}
	if end > 0 {
		query.Set("end", fmt.Sprintf("%d", end))
	}

	var result []sandboxMetricResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/sandboxes/%s/metrics", url.PathEscape(sandboxID)), query, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) getSandboxLogs(ctx context.Context, sandboxID string, cursor int64, limit int64, direction string, level string, search string) ([]sandboxLogEntryResponse, error) {
	query := url.Values{}
	if cursor > 0 {
		query.Set("cursor", fmt.Sprintf("%d", cursor))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if direction != "" {
		query.Set("direction", direction)
	}
	if level != "" {
		query.Set("level", level)
	}
	if search != "" {
		query.Set("search", search)
	}

	var result sandboxLogsResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/v2/sandboxes/%s/logs", url.PathEscape(sandboxID)), query, nil, &result); err != nil {
		return nil, err
	}

	return result.Logs, nil
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

func (c *e2bClient) getTemplateAlias(ctx context.Context, alias string) (*templateAliasResponse, error) {
	var result templateAliasResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/templates/aliases/%s", url.PathEscape(alias)), nil, nil, &result); err != nil {
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

func (c *e2bClient) getTeamMetrics(ctx context.Context, teamID string, start int64, end int64) ([]teamMetricResponse, error) {
	query := url.Values{}
	if start > 0 {
		query.Set("start", fmt.Sprintf("%d", start))
	}
	if end > 0 {
		query.Set("end", fmt.Sprintf("%d", end))
	}

	var result []teamMetricResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/teams/%s/metrics", url.PathEscape(teamID)), query, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) getTeamMetricMax(ctx context.Context, teamID string, metric string, start int64, end int64) (*maxTeamMetricResponse, error) {
	query := url.Values{}
	query.Set("metric", metric)
	if start > 0 {
		query.Set("start", fmt.Sprintf("%d", start))
	}
	if end > 0 {
		query.Set("end", fmt.Sprintf("%d", end))
	}

	var result maxTeamMetricResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/teams/%s/metrics/max", url.PathEscape(teamID)), query, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) assignTemplateTags(ctx context.Context, target string, tags []string) (*templateTagsAssignResponse, error) {
	var result templateTagsAssignResponse
	if err := c.do(ctx, http.MethodPost, "/templates/tags", nil, templateTagsAssignRequest{Target: target, Tags: tags}, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) deleteTemplateTags(ctx context.Context, name string, tags []string) error {
	return c.do(ctx, http.MethodDelete, "/templates/tags", nil, templateTagsDeleteRequest{Name: name, Tags: tags}, nil)
}

func (c *e2bClient) listTemplateTags(ctx context.Context, templateID string) ([]templateTagResponse, error) {
	var result []templateTagResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/templates/%s/tags", url.PathEscape(templateID)), nil, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) createLifecycleWebhook(ctx context.Context, request lifecycleWebhookRequest) (*lifecycleWebhookResponse, error) {
	var result lifecycleWebhookResponse
	if err := c.do(ctx, http.MethodPost, "/events/webhooks", nil, request, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) getLifecycleWebhook(ctx context.Context, webhookID string) (*lifecycleWebhookResponse, error) {
	var result lifecycleWebhookResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/events/webhooks/%s", url.PathEscape(webhookID)), nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) listLifecycleWebhooks(ctx context.Context) ([]lifecycleWebhookResponse, error) {
	var result []lifecycleWebhookResponse
	if err := c.do(ctx, http.MethodGet, "/events/webhooks", nil, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) updateLifecycleWebhook(ctx context.Context, webhookID string, request lifecycleWebhookRequest) (*lifecycleWebhookResponse, error) {
	var result lifecycleWebhookResponse
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/events/webhooks/%s", url.PathEscape(webhookID)), nil, request, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) deleteLifecycleWebhook(ctx context.Context, webhookID string) error {
	err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/events/webhooks/%s", url.PathEscape(webhookID)), nil, nil, nil)
	if isNotFound(err) {
		return nil
	}

	return err
}

func (c *e2bClient) listLifecycleEvents(ctx context.Context, sandboxID string, eventTypes []string, offset int64, limit int64, orderAsc *bool) ([]lifecycleEventResponse, error) {
	query := url.Values{}
	for _, eventType := range eventTypes {
		query.Add("types", eventType)
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if orderAsc != nil {
		query.Set("orderAsc", fmt.Sprintf("%t", *orderAsc))
	}

	path := "/events/sandboxes"
	if sandboxID != "" {
		path = fmt.Sprintf("/events/sandboxes/%s", url.PathEscape(sandboxID))
	}

	var result []lifecycleEventResponse
	if err := c.do(ctx, http.MethodGet, path, query, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *e2bClient) createSnapshot(ctx context.Context, sandboxID string, name string) (*snapshotResponse, error) {
	var result snapshotResponse
	path := fmt.Sprintf("/sandboxes/%s/snapshots", url.PathEscape(sandboxID))
	if err := c.do(ctx, http.MethodPost, path, nil, snapshotRequest{Name: name}, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *e2bClient) listSnapshots(ctx context.Context, sandboxID string, limit int64, nextToken string) ([]snapshotResponse, error) {
	query := url.Values{}
	if sandboxID != "" {
		query.Set("sandboxID", sandboxID)
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if nextToken != "" {
		query.Set("nextToken", nextToken)
	}

	var result []snapshotResponse
	if err := c.do(ctx, http.MethodGet, "/snapshots", query, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
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
