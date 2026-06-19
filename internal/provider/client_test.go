// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
