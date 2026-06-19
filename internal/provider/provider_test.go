// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"e2b": providerserver.NewProtocol6WithError(New("test")()),
}

func TestDataSourcesHaveResourceCounterpartOrReadOnlyAPI(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := &E2BProvider{version: "test"}

	resources := map[string]struct{}{}
	for _, newResource := range provider.Resources(ctx) {
		var resp resource.MetadataResponse
		newResource().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "e2b"}, &resp)
		resources[resp.TypeName] = struct{}{}
	}

	counterparts := map[string]string{
		"e2b_api_key":            "e2b_api_key",
		"e2b_api_keys":           "e2b_api_key",
		"e2b_lifecycle_webhook":  "e2b_lifecycle_webhook",
		"e2b_lifecycle_webhooks": "e2b_lifecycle_webhook",
		"e2b_sandbox":            "e2b_sandbox",
		"e2b_sandboxes":          "e2b_sandbox",
		"e2b_snapshots":          "e2b_snapshot",
		"e2b_template":           "e2b_template",
		"e2b_template_tags":      "e2b_template_tags",
		"e2b_templates":          "e2b_template",
		"e2b_volume":             "e2b_volume",
		"e2b_volumes":            "e2b_volume",
	}

	readOnlyAPIDataSources := map[string]string{
		"e2b_lifecycle_events": "E2B exposes lifecycle events as GET-only event history.",
		"e2b_team_metric_max":  "E2B exposes team metric maximums as GET-only telemetry.",
		"e2b_team_metrics":     "E2B exposes team metrics as GET-only telemetry.",
		"e2b_teams":            "E2B exposes teams as GET-only identity context.",
	}

	for _, newDataSource := range provider.DataSources(ctx) {
		var resp datasource.MetadataResponse
		newDataSource().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "e2b"}, &resp)

		if counterpart, ok := counterparts[resp.TypeName]; ok {
			if _, exists := resources[counterpart]; !exists {
				t.Fatalf("data source %s expects resource %s, but it is not registered", resp.TypeName, counterpart)
			}
			continue
		}

		if reason, ok := readOnlyAPIDataSources[resp.TypeName]; ok && reason != "" {
			continue
		}

		t.Fatalf("data source %s must have a resource counterpart or an explicit read-only API exception", resp.TypeName)
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("E2B_API_KEY") == "" {
		loadDotEnv(t)
	}

	if os.Getenv("E2B_API_KEY") == "" {
		t.Fatal("E2B_API_KEY must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	return `provider "e2b" {}`
}

func loadDotEnv(t *testing.T) {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("read home directory: %s", err)
	}

	content, err := os.ReadFile(filepath.Join(home, ".env"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("read ~/.env: %s", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "E2B_API_KEY" && os.Getenv("E2B_API_KEY") == "" {
			t.Setenv("E2B_API_KEY", value)
		}
	}
}
