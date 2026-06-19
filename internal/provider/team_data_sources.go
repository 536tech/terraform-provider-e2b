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

var _ datasource.DataSource = &TeamsDataSource{}
var _ datasource.DataSource = &TeamMetricsDataSource{}
var _ datasource.DataSource = &TeamMetricMaxDataSource{}

func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

func NewTeamMetricsDataSource() datasource.DataSource {
	return &TeamMetricsDataSource{}
}

func NewTeamMetricMaxDataSource() datasource.DataSource {
	return &TeamMetricMaxDataSource{}
}

type TeamsDataSource struct {
	client *e2bClient
}

type TeamModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	APIKey    types.String `tfsdk:"api_key"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

type TeamsDataSourceModel struct {
	Teams []TeamModel `tfsdk:"teams"`
}

type TeamMetricsDataSource struct {
	client *e2bClient
}

type TeamMetricModel struct {
	Timestamp           types.String  `tfsdk:"timestamp"`
	TimestampUnix       types.Int64   `tfsdk:"timestamp_unix"`
	ConcurrentSandboxes types.Int64   `tfsdk:"concurrent_sandboxes"`
	SandboxStartRate    types.Float64 `tfsdk:"sandbox_start_rate"`
}

type TeamMetricsDataSourceModel struct {
	TeamID  types.String      `tfsdk:"team_id"`
	Start   types.Int64       `tfsdk:"start"`
	End     types.Int64       `tfsdk:"end"`
	Metrics []TeamMetricModel `tfsdk:"metrics"`
}

type TeamMetricMaxDataSource struct {
	client *e2bClient
}

type TeamMetricMaxDataSourceModel struct {
	TeamID        types.String  `tfsdk:"team_id"`
	Metric        types.String  `tfsdk:"metric"`
	Start         types.Int64   `tfsdk:"start"`
	End           types.Int64   `tfsdk:"end"`
	Timestamp     types.String  `tfsdk:"timestamp"`
	TimestampUnix types.Int64   `tfsdk:"timestamp_unix"`
	Value         types.Float64 `tfsdk:"value"`
}

func (d *TeamsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *TeamsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B teams visible to the configured access token. E2B exposes teams as read-only identity context in the public API and does not expose team creation through this API. This route requires `access_token` provider configuration.",
		Attributes: map[string]schema.Attribute{
			"teams": schema.ListNestedAttribute{
				MarkdownDescription: "E2B teams.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Team ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Team name.",
							Computed:            true,
						},
						"api_key": schema.StringAttribute{
							MarkdownDescription: "Team API key returned by E2B's team listing route.",
							Computed:            true,
							Sensitive:           true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default team.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TeamsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teams, err := d.client.listTeams(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list teams", err)
		return
	}

	data.Teams = make([]TeamModel, 0, len(teams))
	for _, team := range teams {
		data.Teams = append(data.Teams, TeamModel{
			ID:        types.StringValue(team.TeamID),
			Name:      types.StringValue(team.Name),
			APIKey:    types.StringValue(team.APIKey),
			IsDefault: types.BoolValue(team.IsDefault),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *TeamMetricsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_metrics"
}

func (d *TeamMetricsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads E2B team concurrency and sandbox-start-rate metrics. E2B exposes team metrics as read-only telemetry and does not expose metric creation through the API.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				MarkdownDescription: "E2B team ID.",
				Required:            true,
			},
			"start": schema.Int64Attribute{
				MarkdownDescription: "Optional Unix timestamp for the start of the interval, in seconds.",
				Optional:            true,
			},
			"end": schema.Int64Attribute{
				MarkdownDescription: "Optional Unix timestamp for the end of the interval, in seconds.",
				Optional:            true,
			},
			"metrics": schema.ListNestedAttribute{
				MarkdownDescription: "Metric points returned by E2B.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"timestamp":            schema.StringAttribute{MarkdownDescription: "Deprecated RFC3339 timestamp returned by E2B.", Computed: true},
						"timestamp_unix":       schema.Int64Attribute{MarkdownDescription: "Unix timestamp for the metric point.", Computed: true},
						"concurrent_sandboxes": schema.Int64Attribute{MarkdownDescription: "Concurrent sandboxes for the team.", Computed: true},
						"sandbox_start_rate":   schema.Float64Attribute{MarkdownDescription: "Sandboxes started per second.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *TeamMetricsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamMetricsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metrics, err := d.client.getTeamMetrics(ctx, data.TeamID.ValueString(), optionalInt64Value(data.Start), optionalInt64Value(data.End))
	if err != nil {
		addClientError(&resp.Diagnostics, "read team metrics", err)
		return
	}

	data.Metrics = make([]TeamMetricModel, 0, len(metrics))
	for _, metric := range metrics {
		data.Metrics = append(data.Metrics, TeamMetricModel{
			Timestamp:           types.StringValue(metric.Timestamp),
			TimestampUnix:       types.Int64Value(metric.TimestampUnix),
			ConcurrentSandboxes: types.Int64Value(metric.ConcurrentSandboxes),
			SandboxStartRate:    types.Float64Value(metric.SandboxStartRate),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *TeamMetricMaxDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_metric_max"
}

func (d *TeamMetricMaxDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the maximum E2B team metric value for an interval. E2B exposes team metrics as read-only telemetry and does not expose metric creation through the API.",
		Attributes: map[string]schema.Attribute{
			"team_id":        schema.StringAttribute{MarkdownDescription: "E2B team ID.", Required: true},
			"metric":         schema.StringAttribute{MarkdownDescription: "Metric to evaluate. E2B supports `concurrent_sandboxes` and `sandbox_start_rate`.", Required: true},
			"start":          schema.Int64Attribute{MarkdownDescription: "Optional Unix timestamp for the start of the interval, in seconds.", Optional: true},
			"end":            schema.Int64Attribute{MarkdownDescription: "Optional Unix timestamp for the end of the interval, in seconds.", Optional: true},
			"timestamp":      schema.StringAttribute{MarkdownDescription: "Deprecated RFC3339 timestamp returned by E2B.", Computed: true},
			"timestamp_unix": schema.Int64Attribute{MarkdownDescription: "Unix timestamp for the max metric point.", Computed: true},
			"value":          schema.Float64Attribute{MarkdownDescription: "Maximum metric value.", Computed: true},
		},
	}
}

func (d *TeamMetricMaxDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamMetricMaxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamMetricMaxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metric, err := d.client.getTeamMetricMax(ctx, data.TeamID.ValueString(), data.Metric.ValueString(), optionalInt64Value(data.Start), optionalInt64Value(data.End))
	if err != nil {
		addClientError(&resp.Diagnostics, "read team metric max", err)
		return
	}

	data.Timestamp = types.StringValue(metric.Timestamp)
	data.TimestampUnix = types.Int64Value(metric.TimestampUnix)
	data.Value = types.Float64Value(metric.Value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
