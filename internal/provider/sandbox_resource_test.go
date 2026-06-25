// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSandboxNetworkRulesTerraformRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value := testSandboxNetworkRulesValue(t, ctx)

	rules, diags := sandboxNetworkRulesFromTerraform(ctx, value)
	if diags.HasError() {
		t.Fatalf("convert network rules from Terraform: %#v", diags)
	}

	if len(rules["api.example.com"]) != 1 {
		t.Fatalf("unexpected network rules: %#v", rules)
	}
	if got := rules["api.example.com"][0].Transform.Headers["X-E2B-Policy"]; got != "terraform" {
		t.Fatalf("unexpected header value: %q", got)
	}

	roundTrip, diags := sandboxNetworkRulesValue(ctx, rules)
	if diags.HasError() {
		t.Fatalf("convert network rules to Terraform: %#v", diags)
	}
	if !roundTrip.Equal(value) {
		t.Fatalf("unexpected round trip value: %#v", roundTrip)
	}
}

func TestSandboxApplyDetailPreservesNetworkRulesWhenOmitted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value := testSandboxNetworkRulesValue(t, ctx)
	model := SandboxResourceModel{
		NetworkRules: value,
	}

	diags := model.applySandboxDetail(ctx, &sandboxDetailResponse{
		Network: &sandboxNetworkConfig{},
	})
	if diags.HasError() {
		t.Fatalf("apply sandbox detail: %#v", diags)
	}

	if !model.NetworkRules.Equal(value) {
		t.Fatalf("expected network rules to be preserved, got: %#v", model.NetworkRules)
	}
}

func TestFlattenSandboxEgressProxyPreservesSensitiveValuesWhenOmitted(t *testing.T) {
	t.Parallel()

	prior := &SandboxEgressProxyModel{
		Address:  types.StringValue("proxy.example.com:1080"),
		Username: types.StringValue("agent"),
		Password: types.StringValue("secret"),
	}

	got := flattenSandboxEgressProxy(&sandboxEgressProxyConfig{
		Address: "proxy.example.com:1080",
	}, prior)
	if got == nil {
		t.Fatal("expected egress proxy")
	}
	if got.Username.ValueString() != "agent" {
		t.Fatalf("unexpected username: %#v", got.Username)
	}
	if got.Password.ValueString() != "secret" {
		t.Fatalf("unexpected password: %#v", got.Password)
	}
}

func testSandboxNetworkRulesValue(t *testing.T, ctx context.Context) types.Map {
	t.Helper()

	headers, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"X-E2B-Policy": "terraform",
	})
	if diags.HasError() {
		t.Fatalf("build headers value: %#v", diags)
	}

	value, diags := types.MapValueFrom(ctx, sandboxNetworkRuleListType(), map[string][]SandboxNetworkRuleModel{
		"api.example.com": {
			{Headers: headers},
		},
	})
	if diags.HasError() {
		t.Fatalf("build network rules value: %#v", diags)
	}

	return value
}
