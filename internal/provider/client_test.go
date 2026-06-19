// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientCreateSandbox(t *testing.T) {
	t.Parallel()

	var gotAPIKey string
	var gotRequest sandboxCreateRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sandboxes" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		gotAPIKey = r.Header.Get("X-API-Key")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %s", err)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"templateID":"rki5","sandboxID":"sbx_123","alias":"base","clientID":"client","envdVersion":"0.6.4"}`))
	}))
	defer server.Close()

	client, err := newE2BClient(server.URL, "test-key", "test")
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	timeout := int64(60)
	created, err := client.createSandbox(context.Background(), sandboxCreateRequest{
		TemplateID: "base",
		Timeout:    &timeout,
		Metadata: map[string]string{
			"managed_by": "terraform",
		},
	})
	if err != nil {
		t.Fatalf("create sandbox: %s", err)
	}

	if gotAPIKey != "test-key" {
		t.Fatalf("unexpected API key header: %q", gotAPIKey)
	}
	if gotRequest.TemplateID != "base" {
		t.Fatalf("unexpected template id: %q", gotRequest.TemplateID)
	}
	if gotRequest.Timeout == nil || *gotRequest.Timeout != 60 {
		t.Fatalf("unexpected timeout: %#v", gotRequest.Timeout)
	}
	if created.SandboxID != "sbx_123" {
		t.Fatalf("unexpected sandbox id: %q", created.SandboxID)
	}
}

func TestClientUpdateSandboxNetwork(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sandboxes/sbx_123/network" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var gotRequest sandboxNetworkConfig
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %s", err)
		}

		if len(gotRequest.AllowOut) != 1 || gotRequest.AllowOut[0] != "8.8.8.8/32" {
			t.Fatalf("unexpected allow out: %#v", gotRequest.AllowOut)
		}
		if len(gotRequest.DenyOut) != 1 || gotRequest.DenyOut[0] != "203.0.113.0/24" {
			t.Fatalf("unexpected deny out: %#v", gotRequest.DenyOut)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := newE2BClient(server.URL, "test-key", "test")
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	err = client.updateSandboxNetwork(context.Background(), "sbx_123", sandboxNetworkConfig{
		AllowOut: []string{"8.8.8.8/32"},
		DenyOut:  []string{"203.0.113.0/24"},
	})
	if err != nil {
		t.Fatalf("update sandbox network: %s", err)
	}
}

func TestClientSetSandboxTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sandboxes/sbx_123/timeout" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var gotRequest map[string]int64
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %s", err)
		}
		if gotRequest["timeout"] != 120 {
			t.Fatalf("unexpected timeout: %#v", gotRequest)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := newE2BClient(server.URL, "test-key", "test")
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	if err := client.setSandboxTimeout(context.Background(), "sbx_123", 120); err != nil {
		t.Fatalf("set sandbox timeout: %s", err)
	}
}

func TestClientListSandboxesEncodesFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sandboxes" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("metadata"); got != "app=prod&user=abc" {
			t.Fatalf("unexpected metadata query: %q", got)
		}
		if got := r.URL.Query().Get("state"); got != "paused,running" {
			t.Fatalf("unexpected state query: %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Fatalf("unexpected limit query: %q", got)
		}

		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := newE2BClient(server.URL, "test-key", "test")
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	_, err = client.listSandboxes(context.Background(), map[string]string{
		"user": "abc",
		"app":  "prod",
	}, []string{"paused", "running"}, 10)
	if err != nil {
		t.Fatalf("list sandboxes: %s", err)
	}
}

func TestClientTemplateBuildLifecycle(t *testing.T) {
	t.Parallel()

	var gotCreate templateCreateRequest
	var gotBuild templateBuildStartRequest
	var deleted bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/templates":
			if err := json.NewDecoder(r.Body).Decode(&gotCreate); err != nil {
				t.Fatalf("decode create request: %s", err)
			}

			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"templateID":"tpl_123","buildID":"bld_123","public":false,"aliases":["tf-acc"],"names":["team/tf-acc"],"tags":["default"]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/templates/tpl_123/builds/bld_123":
			if err := json.NewDecoder(r.Body).Decode(&gotBuild); err != nil {
				t.Fatalf("decode build request: %s", err)
			}

			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/templates/tpl_123/builds/bld_123/status":
			_, _ = w.Write([]byte(`{"templateID":"tpl_123","buildID":"bld_123","status":"ready","logs":[],"logEntries":[]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/templates/tpl_123":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := newE2BClient(server.URL, "test-key", "test")
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	cpuCount := int64(2)
	memoryMB := int64(512)
	created, err := client.createTemplate(context.Background(), templateCreateRequest{
		Name:     "tf-acc",
		CPUCount: &cpuCount,
		MemoryMB: &memoryMB,
	})
	if err != nil {
		t.Fatalf("create template: %s", err)
	}

	force := true
	if err := client.startTemplateBuild(context.Background(), created.TemplateID, created.BuildID, templateBuildStartRequest{
		FromImage: "e2bdev/base:latest",
		Force:     &force,
		StartCmd:  `sh -c "sleep 3600"`,
		ReadyCmd:  "true",
	}); err != nil {
		t.Fatalf("start template build: %s", err)
	}

	status, err := client.waitForTemplateBuild(context.Background(), created.TemplateID, created.BuildID, time.Millisecond)
	if err != nil {
		t.Fatalf("wait for template build: %s", err)
	}

	if err := client.deleteTemplate(context.Background(), created.TemplateID); err != nil {
		t.Fatalf("delete template: %s", err)
	}

	if gotCreate.Name != "tf-acc" {
		t.Fatalf("unexpected create name: %q", gotCreate.Name)
	}
	if gotCreate.CPUCount == nil || *gotCreate.CPUCount != 2 {
		t.Fatalf("unexpected create CPU count: %#v", gotCreate.CPUCount)
	}
	if gotCreate.MemoryMB == nil || *gotCreate.MemoryMB != 512 {
		t.Fatalf("unexpected create memory: %#v", gotCreate.MemoryMB)
	}
	if gotBuild.FromImage != "e2bdev/base:latest" {
		t.Fatalf("unexpected from image: %q", gotBuild.FromImage)
	}
	if gotBuild.Force == nil || !*gotBuild.Force {
		t.Fatalf("unexpected force: %#v", gotBuild.Force)
	}
	if gotBuild.StartCmd != `sh -c "sleep 3600"` {
		t.Fatalf("unexpected start command: %q", gotBuild.StartCmd)
	}
	if gotBuild.ReadyCmd != "true" {
		t.Fatalf("unexpected ready command: %q", gotBuild.ReadyCmd)
	}
	if status.Status != "ready" {
		t.Fatalf("unexpected build status: %q", status.Status)
	}
	if !deleted {
		t.Fatalf("expected template to be deleted")
	}
}
