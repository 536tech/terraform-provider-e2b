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

func newTestAPIKeyClient(t *testing.T, rawBaseURL string) *e2bClient {
	t.Helper()

	client, err := newE2BClientWithConfig(rawBaseURL, e2bClientConfig{
		APIKey:  "test-key",
		Version: "test",
	})
	if err != nil {
		t.Fatalf("new client: %s", err)
	}

	return client
}

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

	client := newTestAPIKeyClient(t, server.URL)

	timeout := int64(60)
	autoPause := true
	autoResume := true
	allowPublicTraffic := true
	created, err := client.createSandbox(context.Background(), sandboxCreateRequest{
		TemplateID: "base",
		Timeout:    &timeout,
		AutoPause:  &autoPause,
		AutoResume: &sandboxAutoResume{Enabled: autoResume},
		Metadata: map[string]string{
			"managed_by": "terraform",
		},
		Network: &sandboxNetworkConfig{
			AllowPublicTraffic: &allowPublicTraffic,
			MaskRequestHost:    "sandbox.example.com",
		},
		VolumeMounts: []sandboxVolumeMount{
			{Name: "cache", Path: "/mnt/cache"},
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
	if gotRequest.AutoPause == nil || !*gotRequest.AutoPause {
		t.Fatalf("unexpected auto pause: %#v", gotRequest.AutoPause)
	}
	if gotRequest.AutoResume == nil || !gotRequest.AutoResume.Enabled {
		t.Fatalf("unexpected auto resume: %#v", gotRequest.AutoResume)
	}
	if gotRequest.Network == nil || gotRequest.Network.AllowPublicTraffic == nil || !*gotRequest.Network.AllowPublicTraffic {
		t.Fatalf("unexpected public traffic config: %#v", gotRequest.Network)
	}
	if gotRequest.Network == nil || gotRequest.Network.MaskRequestHost != "sandbox.example.com" {
		t.Fatalf("unexpected mask request host: %#v", gotRequest.Network)
	}
	if len(gotRequest.VolumeMounts) != 1 || gotRequest.VolumeMounts[0].Name != "cache" || gotRequest.VolumeMounts[0].Path != "/mnt/cache" {
		t.Fatalf("unexpected volume mounts: %#v", gotRequest.VolumeMounts)
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

	client := newTestAPIKeyClient(t, server.URL)

	err := client.updateSandboxNetwork(context.Background(), "sbx_123", sandboxNetworkConfig{
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

	client := newTestAPIKeyClient(t, server.URL)

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

	client := newTestAPIKeyClient(t, server.URL)

	_, err := client.listSandboxes(context.Background(), map[string]string{
		"user": "abc",
		"app":  "prod",
	}, []string{"paused", "running"}, 10)
	if err != nil {
		t.Fatalf("list sandboxes: %s", err)
	}
}

func TestClientSandboxObservability(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sandboxes/metrics":
			if got := r.URL.Query().Get("sandbox_ids"); got != "sbx_1,sbx_2" {
				t.Fatalf("unexpected sandbox_ids: %q", got)
			}
			_, _ = w.Write([]byte(`{"sandboxes":{"sbx_1":{"timestamp":"2026-06-24T00:00:00Z","timestampUnix":1782259200,"cpuCount":2,"cpuUsedPct":12.5,"memUsed":1024,"memTotal":2048,"memCache":256,"diskUsed":4096,"diskTotal":8192}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/sandboxes/sbx_1/metrics":
			if got := r.URL.Query().Get("start"); got != "10" {
				t.Fatalf("unexpected start: %q", got)
			}
			if got := r.URL.Query().Get("end"); got != "20" {
				t.Fatalf("unexpected end: %q", got)
			}
			_, _ = w.Write([]byte(`[{"timestamp":"2026-06-24T00:00:00Z","timestampUnix":1782259200,"cpuCount":2,"cpuUsedPct":12.5,"memUsed":1024,"memTotal":2048,"memCache":256,"diskUsed":4096,"diskTotal":8192}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/sandboxes/sbx_1/logs":
			if got := r.URL.Query().Get("cursor"); got != "100" {
				t.Fatalf("unexpected cursor: %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "50" {
				t.Fatalf("unexpected limit: %q", got)
			}
			if got := r.URL.Query().Get("direction"); got != "backward" {
				t.Fatalf("unexpected direction: %q", got)
			}
			if got := r.URL.Query().Get("level"); got != "warn" {
				t.Fatalf("unexpected level: %q", got)
			}
			if got := r.URL.Query().Get("search"); got != "ready" {
				t.Fatalf("unexpected search: %q", got)
			}
			_, _ = w.Write([]byte(`{"logs":[{"timestamp":"2026-06-24T00:00:01Z","level":"info","message":"ready","fields":{"step":"boot"}}]}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestAPIKeyClient(t, server.URL)

	latest, err := client.listSandboxesMetrics(context.Background(), []string{"sbx_1", "sbx_2"})
	if err != nil {
		t.Fatalf("list sandbox metrics: %s", err)
	}
	history, err := client.getSandboxMetrics(context.Background(), "sbx_1", 10, 20)
	if err != nil {
		t.Fatalf("get sandbox metrics: %s", err)
	}
	logs, err := client.getSandboxLogs(context.Background(), "sbx_1", 100, 50, "backward", "warn", "ready")
	if err != nil {
		t.Fatalf("get sandbox logs: %s", err)
	}

	if latest["sbx_1"].CPUUsedPct != 12.5 {
		t.Fatalf("unexpected latest metrics: %#v", latest)
	}
	if len(history) != 1 || history[0].DiskTotal != 8192 {
		t.Fatalf("unexpected metric history: %#v", history)
	}
	if len(logs) != 1 || logs[0].Fields["step"] != "boot" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestClientGetTemplateAlias(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/aliases/base" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		_, _ = w.Write([]byte(`{"templateID":"tpl_123","public":true}`))
	}))
	defer server.Close()

	client := newTestAPIKeyClient(t, server.URL)

	alias, err := client.getTemplateAlias(context.Background(), "base")
	if err != nil {
		t.Fatalf("get template alias: %s", err)
	}
	if alias.TemplateID != "tpl_123" || !alias.Public {
		t.Fatalf("unexpected alias response: %#v", alias)
	}
}

func TestClientTemplateTags(t *testing.T) {
	t.Parallel()

	var gotAssign templateTagsAssignRequest
	var gotDelete templateTagsDeleteRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/templates/tags":
			if err := json.NewDecoder(r.Body).Decode(&gotAssign); err != nil {
				t.Fatalf("decode assign request: %s", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tags":["prod"],"buildID":"bld_123"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/templates/tpl_123/tags":
			_, _ = w.Write([]byte(`[{"tag":"prod","buildID":"bld_123","createdAt":"2026-06-18T00:00:00Z"}]`))
		case r.Method == http.MethodDelete && r.URL.Path == "/templates/tags":
			if err := json.NewDecoder(r.Body).Decode(&gotDelete); err != nil {
				t.Fatalf("decode delete request: %s", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestAPIKeyClient(t, server.URL)

	assigned, err := client.assignTemplateTags(context.Background(), "example:build", []string{"prod"})
	if err != nil {
		t.Fatalf("assign template tags: %s", err)
	}
	tags, err := client.listTemplateTags(context.Background(), "tpl_123")
	if err != nil {
		t.Fatalf("list template tags: %s", err)
	}
	if err := client.deleteTemplateTags(context.Background(), "example", []string{"prod"}); err != nil {
		t.Fatalf("delete template tags: %s", err)
	}

	if gotAssign.Target != "example:build" || len(gotAssign.Tags) != 1 || gotAssign.Tags[0] != "prod" {
		t.Fatalf("unexpected assign request: %#v", gotAssign)
	}
	if assigned.BuildID != "bld_123" {
		t.Fatalf("unexpected assign response: %#v", assigned)
	}
	if len(tags) != 1 || tags[0].Tag != "prod" {
		t.Fatalf("unexpected template tags: %#v", tags)
	}
	if gotDelete.Name != "example" || len(gotDelete.Tags) != 1 || gotDelete.Tags[0] != "prod" {
		t.Fatalf("unexpected delete request: %#v", gotDelete)
	}
}

func TestClientLifecycleEventsAndSnapshots(t *testing.T) {
	t.Parallel()

	var gotSnapshotRequest snapshotRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/events/sandboxes/sbx_123":
			if got := r.URL.Query()["types"]; len(got) != 2 || got[0] != "sandbox.started" || got[1] != "sandbox.paused" {
				t.Fatalf("unexpected event types: %#v", got)
			}
			if got := r.URL.Query().Get("offset"); got != "2" {
				t.Fatalf("unexpected offset: %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "5" {
				t.Fatalf("unexpected limit: %q", got)
			}
			if got := r.URL.Query().Get("orderAsc"); got != "true" {
				t.Fatalf("unexpected orderAsc: %q", got)
			}
			_, _ = w.Write([]byte(`[{"version":"v1","id":"evt_123","type":"sandbox.started","eventData":{"state":"running"},"sandboxBuildId":"bld_123","sandboxExecutionId":"exec_123","sandboxId":"sbx_123","sandboxTeamId":"team_123","sandboxTemplateId":"tpl_123","timestamp":"2026-06-18T00:00:00Z"}]`))
		case r.Method == http.MethodPost && r.URL.Path == "/sandboxes/sbx_123/snapshots":
			if err := json.NewDecoder(r.Body).Decode(&gotSnapshotRequest); err != nil {
				t.Fatalf("decode snapshot request: %s", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"snapshotID":"snap_123","names":["checkpoint"]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/snapshots":
			if got := r.URL.Query().Get("sandboxID"); got != "sbx_123" {
				t.Fatalf("unexpected sandboxID: %q", got)
			}
			if got := r.URL.Query().Get("limit"); got != "10" {
				t.Fatalf("unexpected limit: %q", got)
			}
			if got := r.URL.Query().Get("nextToken"); got != "next" {
				t.Fatalf("unexpected nextToken: %q", got)
			}
			_, _ = w.Write([]byte(`[{"snapshotID":"snap_123","names":["checkpoint"]}]`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestAPIKeyClient(t, server.URL)

	orderAsc := true
	events, err := client.listLifecycleEvents(context.Background(), "sbx_123", []string{"sandbox.started", "sandbox.paused"}, 2, 5, &orderAsc)
	if err != nil {
		t.Fatalf("list lifecycle events: %s", err)
	}
	created, err := client.createSnapshot(context.Background(), "sbx_123", "checkpoint")
	if err != nil {
		t.Fatalf("create snapshot: %s", err)
	}
	snapshots, err := client.listSnapshots(context.Background(), "sbx_123", 10, "next")
	if err != nil {
		t.Fatalf("list snapshots: %s", err)
	}

	if len(events) != 1 || events[0].ID != "evt_123" || string(events[0].EventData) != `{"state":"running"}` {
		t.Fatalf("unexpected lifecycle events: %#v", events)
	}
	if gotSnapshotRequest.Name != "checkpoint" {
		t.Fatalf("unexpected snapshot request: %#v", gotSnapshotRequest)
	}
	if created.SnapshotID != "snap_123" || len(created.Names) != 1 || created.Names[0] != "checkpoint" {
		t.Fatalf("unexpected created snapshot: %#v", created)
	}
	if len(snapshots) != 1 || snapshots[0].SnapshotID != "snap_123" || len(snapshots[0].Names) != 1 || snapshots[0].Names[0] != "checkpoint" {
		t.Fatalf("unexpected snapshots: %#v", snapshots)
	}
}

func TestClientLifecycleWebhookLifecycle(t *testing.T) {
	t.Parallel()

	var gotCreate lifecycleWebhookRequest
	var gotUpdate lifecycleWebhookRequest
	var deleted bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/events/webhooks":
			if err := json.NewDecoder(r.Body).Decode(&gotCreate); err != nil {
				t.Fatalf("decode create request: %s", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"wh_123","teamId":"team_123","name":"audit","createdAt":"2026-06-18T00:00:00Z","enabled":true,"url":"https://example.com/e2b","events":["sandbox.started"]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/events/webhooks/wh_123":
			_, _ = w.Write([]byte(`{"id":"wh_123","teamId":"team_123","name":"audit","createdAt":"2026-06-18T00:00:00Z","enabled":true,"url":"https://example.com/e2b","events":["sandbox.started"]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/events/webhooks":
			_, _ = w.Write([]byte(`[{"id":"wh_123","teamId":"team_123","name":"audit","createdAt":"2026-06-18T00:00:00Z","enabled":true,"url":"https://example.com/e2b","events":["sandbox.started"]}]`))
		case r.Method == http.MethodPatch && r.URL.Path == "/events/webhooks/wh_123":
			if err := json.NewDecoder(r.Body).Decode(&gotUpdate); err != nil {
				t.Fatalf("decode update request: %s", err)
			}
			_, _ = w.Write([]byte(`{"id":"wh_123","teamId":"team_123","name":"audit","createdAt":"2026-06-18T00:00:00Z","enabled":false,"url":"https://example.com/e2b/v2","events":["sandbox.paused"]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/events/webhooks/wh_123":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestAPIKeyClient(t, server.URL)

	enabled := true
	created, err := client.createLifecycleWebhook(context.Background(), lifecycleWebhookRequest{
		Name:            "audit",
		URL:             "https://example.com/e2b",
		Enabled:         &enabled,
		Events:          []string{"sandbox.started"},
		SignatureSecret: "secret",
	})
	if err != nil {
		t.Fatalf("create lifecycle webhook: %s", err)
	}
	read, err := client.getLifecycleWebhook(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("read lifecycle webhook: %s", err)
	}
	webhooks, err := client.listLifecycleWebhooks(context.Background())
	if err != nil {
		t.Fatalf("list lifecycle webhooks: %s", err)
	}
	enabled = false
	updated, err := client.updateLifecycleWebhook(context.Background(), created.ID, lifecycleWebhookRequest{
		URL:     "https://example.com/e2b/v2",
		Enabled: &enabled,
		Events:  []string{"sandbox.paused"},
	})
	if err != nil {
		t.Fatalf("update lifecycle webhook: %s", err)
	}
	if err := client.deleteLifecycleWebhook(context.Background(), created.ID); err != nil {
		t.Fatalf("delete lifecycle webhook: %s", err)
	}

	if gotCreate.Name != "audit" || gotCreate.URL != "https://example.com/e2b" || gotCreate.SignatureSecret != "secret" {
		t.Fatalf("unexpected create request: %#v", gotCreate)
	}
	if gotCreate.Enabled == nil || !*gotCreate.Enabled || len(gotCreate.Events) != 1 || gotCreate.Events[0] != "sandbox.started" {
		t.Fatalf("unexpected create event settings: %#v", gotCreate)
	}
	if read.ID != "wh_123" || len(webhooks) != 1 || updated.URL != "https://example.com/e2b/v2" || updated.Enabled {
		t.Fatalf("unexpected webhook responses: created=%#v read=%#v webhooks=%#v updated=%#v", created, read, webhooks, updated)
	}
	if gotUpdate.Enabled == nil || *gotUpdate.Enabled || len(gotUpdate.Events) != 1 || gotUpdate.Events[0] != "sandbox.paused" {
		t.Fatalf("unexpected update request: %#v", gotUpdate)
	}
	if !deleted {
		t.Fatalf("expected webhook to be deleted")
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

	client := newTestAPIKeyClient(t, server.URL)

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
