// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/url"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringPointerValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(*value)
}

func int64PointerValue(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*value)
}

func boolPointerValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}

	return types.BoolValue(*value)
}

func optionalInt64Pointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	v := value.ValueInt64()
	return &v
}

func optionalBoolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	v := value.ValueBool()
	return &v
}

func mapStringFromTerraform(ctx context.Context, value types.Map) (map[string]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	result := map[string]string{}
	diags := value.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, diags
	}

	return result, diags
}

func mapStringValue(ctx context.Context, value map[string]string) (types.Map, diag.Diagnostics) {
	if len(value) == 0 {
		return types.MapNull(types.StringType), nil
	}

	return types.MapValueFrom(ctx, types.StringType, value)
}

func listStringValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	return types.ListValueFrom(ctx, types.StringType, values)
}

func setStringValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	return types.SetValueFrom(ctx, types.StringType, values)
}

func stringSetFromTerraform(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	result := []string{}
	diags := value.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, diags
	}

	sort.Strings(result)
	return result, diags
}

func encodeMetadataQuery(metadata map[string]string) string {
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	values := url.Values{}
	for _, key := range keys {
		values.Set(key, metadata[key])
	}

	return values.Encode()
}

func addClientError(diags *diag.Diagnostics, action string, err error) {
	diags.AddError(
		fmt.Sprintf("Unable to %s", action),
		err.Error(),
	)
}
