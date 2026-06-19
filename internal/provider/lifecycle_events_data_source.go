// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &LifecycleEventsDataSource{}
var _ datasource.DataSource = &SnapshotsDataSource{}

func NewLifecycleEventsDataSource() datasource.DataSource {
	return &LifecycleEventsDataSource{}
}

func NewSnapshotsDataSource() datasource.DataSource {
	return &SnapshotsDataSource{}
}

type LifecycleEventsDataSource struct {
	client *e2bClient
}

type LifecycleEventsDataSourceModel struct {
	SandboxID types.String          `tfsdk:"sandbox_id"`
	Types     types.Set             `tfsdk:"types"`
	Offset    types.Int64           `tfsdk:"offset"`
	Limit     types.Int64           `tfsdk:"limit"`
	OrderAsc  types.Bool            `tfsdk:"order_asc"`
	Events    []LifecycleEventModel `tfsdk:"events"`
}

type LifecycleEventModel struct {
	Version            types.String `tfsdk:"version"`
	ID                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	EventDataJSON      types.String `tfsdk:"event_data_json"`
	SandboxBuildID     types.String `tfsdk:"sandbox_build_id"`
	SandboxExecutionID types.String `tfsdk:"sandbox_execution_id"`
	SandboxID          types.String `tfsdk:"sandbox_id"`
	SandboxTeamID      types.String `tfsdk:"sandbox_team_id"`
	SandboxTemplateID  types.String `tfsdk:"sandbox_template_id"`
	Timestamp          types.String `tfsdk:"timestamp"`
}

type SnapshotsDataSource struct {
	client *e2bClient
}

type SnapshotsDataSourceModel struct {
	SandboxID types.String    `tfsdk:"sandbox_id"`
	Limit     types.Int64     `tfsdk:"limit"`
	NextToken types.String    `tfsdk:"next_token"`
	Snapshots []SnapshotModel `tfsdk:"snapshots"`
}

type SnapshotModel struct {
	ID    types.String `tfsdk:"id"`
	Names types.List   `tfsdk:"names"`
}

func (d *LifecycleEventsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_events"
}

func (d *LifecycleEventsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B sandbox lifecycle events. E2B exposes lifecycle events as read-only event history and does not expose an API to create events directly.",
		Attributes: map[string]schema.Attribute{
			"sandbox_id": schema.StringAttribute{
				MarkdownDescription: "Optional sandbox ID. When omitted, E2B returns events across sandboxes visible to the configured team.",
				Optional:            true,
			},
			"types": schema.SetAttribute{
				MarkdownDescription: "Optional lifecycle event type filters.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Optional pagination offset.",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Optional maximum number of events to return.",
				Optional:            true,
			},
			"order_asc": schema.BoolAttribute{
				MarkdownDescription: "Whether to sort events ascending by timestamp.",
				Optional:            true,
			},
			"events": schema.ListNestedAttribute{
				MarkdownDescription: "Lifecycle events returned by E2B.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							MarkdownDescription: "Lifecycle event schema version.",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Lifecycle event ID.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Lifecycle event type.",
							Computed:            true,
						},
						"event_data_json": schema.StringAttribute{
							MarkdownDescription: "Lifecycle event data encoded as JSON.",
							Computed:            true,
						},
						"sandbox_build_id": schema.StringAttribute{
							MarkdownDescription: "Sandbox build ID associated with the event.",
							Computed:            true,
						},
						"sandbox_execution_id": schema.StringAttribute{
							MarkdownDescription: "Sandbox execution ID associated with the event.",
							Computed:            true,
						},
						"sandbox_id": schema.StringAttribute{
							MarkdownDescription: "Sandbox ID associated with the event.",
							Computed:            true,
						},
						"sandbox_team_id": schema.StringAttribute{
							MarkdownDescription: "Team ID associated with the event.",
							Computed:            true,
						},
						"sandbox_template_id": schema.StringAttribute{
							MarkdownDescription: "Template ID associated with the event.",
							Computed:            true,
						},
						"timestamp": schema.StringAttribute{
							MarkdownDescription: "Lifecycle event timestamp.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *LifecycleEventsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *LifecycleEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LifecycleEventsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eventTypes, diags := stringSetFromTerraform(ctx, data.Types)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sandboxID string
	if !data.SandboxID.IsNull() && !data.SandboxID.IsUnknown() {
		sandboxID = data.SandboxID.ValueString()
	}

	events, err := d.client.listLifecycleEvents(ctx, sandboxID, eventTypes, optionalInt64Value(data.Offset), optionalInt64Value(data.Limit), optionalBoolPointer(data.OrderAsc))
	if err != nil {
		addClientError(&resp.Diagnostics, "list lifecycle events", err)
		return
	}

	data.Events = make([]LifecycleEventModel, 0, len(events))
	for _, event := range events {
		eventData := types.StringNull()
		if len(event.EventData) > 0 {
			eventData = types.StringValue(string(event.EventData))
		}

		data.Events = append(data.Events, LifecycleEventModel{
			Version:            types.StringValue(event.Version),
			ID:                 types.StringValue(event.ID),
			Type:               types.StringValue(event.Type),
			EventDataJSON:      eventData,
			SandboxBuildID:     types.StringValue(event.SandboxBuildID),
			SandboxExecutionID: types.StringValue(event.SandboxExecutionID),
			SandboxID:          types.StringValue(event.SandboxID),
			SandboxTeamID:      types.StringValue(event.SandboxTeamID),
			SandboxTemplateID:  types.StringValue(event.SandboxTemplateID),
			Timestamp:          types.StringValue(event.Timestamp),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *SnapshotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshots"
}

func (d *SnapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B snapshots.",
		Attributes: map[string]schema.Attribute{
			"sandbox_id": schema.StringAttribute{
				MarkdownDescription: "Optional sandbox ID filter.",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Optional maximum number of snapshots to return.",
				Optional:            true,
			},
			"next_token": schema.StringAttribute{
				MarkdownDescription: "Optional E2B pagination token.",
				Optional:            true,
			},
			"snapshots": schema.ListNestedAttribute{
				MarkdownDescription: "Snapshots returned by E2B.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Snapshot ID.",
							Computed:            true,
						},
						"names": schema.ListAttribute{
							MarkdownDescription: "Names associated with the snapshot.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SnapshotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *SnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapshotsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sandboxID string
	if !data.SandboxID.IsNull() && !data.SandboxID.IsUnknown() {
		sandboxID = data.SandboxID.ValueString()
	}

	var nextToken string
	if !data.NextToken.IsNull() && !data.NextToken.IsUnknown() {
		nextToken = data.NextToken.ValueString()
	}

	snapshots, err := d.client.listSnapshots(ctx, sandboxID, optionalInt64Value(data.Limit), nextToken)
	if err != nil {
		addClientError(&resp.Diagnostics, "list snapshots", err)
		return
	}

	data.Snapshots = make([]SnapshotModel, 0, len(snapshots))
	for _, snapshot := range snapshots {
		names, diags := listStringValue(ctx, snapshot.Names)
		resp.Diagnostics.Append(diags...)
		data.Snapshots = append(data.Snapshots, SnapshotModel{
			ID:    types.StringValue(snapshot.SnapshotID),
			Names: names,
		})
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
