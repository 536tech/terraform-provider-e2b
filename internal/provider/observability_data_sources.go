// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SandboxMetricsDataSource{}
var _ datasource.DataSource = &SandboxMetricHistoryDataSource{}
var _ datasource.DataSource = &SandboxLogsDataSource{}

func NewSandboxMetricsDataSource() datasource.DataSource {
	return &SandboxMetricsDataSource{}
}

func NewSandboxMetricHistoryDataSource() datasource.DataSource {
	return &SandboxMetricHistoryDataSource{}
}

func NewSandboxLogsDataSource() datasource.DataSource {
	return &SandboxLogsDataSource{}
}

type SandboxMetricsDataSource struct {
	client *e2bClient
}

type SandboxMetricsDataSourceModel struct {
	SandboxIDs types.Set          `tfsdk:"sandbox_ids"`
	Sandboxes  []SandboxMetricRow `tfsdk:"sandboxes"`
}

type SandboxMetricHistoryDataSource struct {
	client *e2bClient
}

type SandboxMetricHistoryDataSourceModel struct {
	SandboxID types.String         `tfsdk:"sandbox_id"`
	Start     types.Int64          `tfsdk:"start"`
	End       types.Int64          `tfsdk:"end"`
	Metrics   []SandboxMetricModel `tfsdk:"metrics"`
}

type SandboxMetricRow struct {
	SandboxID     types.String  `tfsdk:"sandbox_id"`
	Timestamp     types.String  `tfsdk:"timestamp"`
	TimestampUnix types.Int64   `tfsdk:"timestamp_unix"`
	CPUCount      types.Int64   `tfsdk:"cpu_count"`
	CPUUsedPct    types.Float64 `tfsdk:"cpu_used_pct"`
	MemUsed       types.Int64   `tfsdk:"mem_used"`
	MemTotal      types.Int64   `tfsdk:"mem_total"`
	MemCache      types.Int64   `tfsdk:"mem_cache"`
	DiskUsed      types.Int64   `tfsdk:"disk_used"`
	DiskTotal     types.Int64   `tfsdk:"disk_total"`
}

type SandboxMetricModel struct {
	Timestamp     types.String  `tfsdk:"timestamp"`
	TimestampUnix types.Int64   `tfsdk:"timestamp_unix"`
	CPUCount      types.Int64   `tfsdk:"cpu_count"`
	CPUUsedPct    types.Float64 `tfsdk:"cpu_used_pct"`
	MemUsed       types.Int64   `tfsdk:"mem_used"`
	MemTotal      types.Int64   `tfsdk:"mem_total"`
	MemCache      types.Int64   `tfsdk:"mem_cache"`
	DiskUsed      types.Int64   `tfsdk:"disk_used"`
	DiskTotal     types.Int64   `tfsdk:"disk_total"`
}

type SandboxLogsDataSource struct {
	client *e2bClient
}

type SandboxLogsDataSourceModel struct {
	SandboxID types.String      `tfsdk:"sandbox_id"`
	Cursor    types.Int64       `tfsdk:"cursor"`
	Limit     types.Int64       `tfsdk:"limit"`
	Direction types.String      `tfsdk:"direction"`
	Level     types.String      `tfsdk:"level"`
	Search    types.String      `tfsdk:"search"`
	Logs      []SandboxLogModel `tfsdk:"logs"`
}

type SandboxLogModel struct {
	Timestamp types.String `tfsdk:"timestamp"`
	Level     types.String `tfsdk:"level"`
	Message   types.String `tfsdk:"message"`
	Fields    types.Map    `tfsdk:"fields"`
}

func (d *SandboxMetricsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_metrics"
}

func (d *SandboxMetricsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the latest E2B metrics for one or more running sandboxes.",
		Attributes: map[string]schema.Attribute{
			"sandbox_ids": schema.SetAttribute{
				MarkdownDescription: "Sandbox IDs to read metrics for. E2B allows up to 100 IDs.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"sandboxes": schema.ListNestedAttribute{
				MarkdownDescription: "Latest metrics keyed by sandbox.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: sandboxMetricAttributes(true),
				},
			},
		},
	}
}

func (d *SandboxMetricsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	d.client = client
}

func (d *SandboxMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SandboxMetricsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sandboxIDs, diags := stringSetFromTerraform(ctx, data.SandboxIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(sandboxIDs) == 0 {
		resp.Diagnostics.AddError("Missing sandbox IDs", "Set at least one sandbox ID.")
		return
	}

	metrics, err := d.client.listSandboxesMetrics(ctx, sandboxIDs)
	if err != nil {
		addClientError(&resp.Diagnostics, "read sandbox metrics", err)
		return
	}

	keys := make([]string, 0, len(metrics))
	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data.Sandboxes = make([]SandboxMetricRow, 0, len(keys))
	for _, sandboxID := range keys {
		data.Sandboxes = append(data.Sandboxes, flattenSandboxMetricRow(sandboxID, metrics[sandboxID]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *SandboxMetricHistoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_metric_history"
}

func (d *SandboxMetricHistoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads E2B metric history for one sandbox.",
		Attributes: map[string]schema.Attribute{
			"sandbox_id": schema.StringAttribute{MarkdownDescription: "Sandbox ID.", Required: true},
			"start":      schema.Int64Attribute{MarkdownDescription: "Optional Unix timestamp for the start of the interval, in seconds.", Optional: true},
			"end":        schema.Int64Attribute{MarkdownDescription: "Optional Unix timestamp for the end of the interval, in seconds.", Optional: true},
			"metrics": schema.ListNestedAttribute{
				MarkdownDescription: "Metric points returned by E2B.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: sandboxMetricAttributes(false),
				},
			},
		},
	}
}

func (d *SandboxMetricHistoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	d.client = client
}

func (d *SandboxMetricHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SandboxMetricHistoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metrics, err := d.client.getSandboxMetrics(ctx, data.SandboxID.ValueString(), optionalInt64Value(data.Start), optionalInt64Value(data.End))
	if err != nil {
		addClientError(&resp.Diagnostics, "read sandbox metric history", err)
		return
	}

	data.Metrics = make([]SandboxMetricModel, 0, len(metrics))
	for _, metric := range metrics {
		data.Metrics = append(data.Metrics, flattenSandboxMetric(metric))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *SandboxLogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_logs"
}

func (d *SandboxLogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads structured logs from a running E2B sandbox.",
		Attributes: map[string]schema.Attribute{
			"sandbox_id": schema.StringAttribute{MarkdownDescription: "Sandbox ID.", Required: true},
			"cursor":     schema.Int64Attribute{MarkdownDescription: "Optional starting log timestamp in milliseconds.", Optional: true},
			"limit":      schema.Int64Attribute{MarkdownDescription: "Optional maximum number of logs to return, up to 1000.", Optional: true},
			"direction":  schema.StringAttribute{MarkdownDescription: "Optional log direction: `forward` or `backward`.", Optional: true},
			"level":      schema.StringAttribute{MarkdownDescription: "Optional minimum log level.", Optional: true},
			"search":     schema.StringAttribute{MarkdownDescription: "Optional case-sensitive message substring search.", Optional: true},
			"logs": schema.ListNestedAttribute{
				MarkdownDescription: "Structured sandbox logs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"timestamp": schema.StringAttribute{MarkdownDescription: "Log timestamp.", Computed: true},
						"level":     schema.StringAttribute{MarkdownDescription: "Log level.", Computed: true},
						"message":   schema.StringAttribute{MarkdownDescription: "Log message.", Computed: true},
						"fields":    schema.MapAttribute{MarkdownDescription: "Structured log fields.", ElementType: types.StringType, Computed: true},
					},
				},
			},
		},
	}
}

func (d *SandboxLogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	d.client = client
}

func (d *SandboxLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SandboxLogsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logs, err := d.client.getSandboxLogs(ctx, data.SandboxID.ValueString(), optionalInt64Value(data.Cursor), optionalInt64Value(data.Limit), optionalStringValue(data.Direction), optionalStringValue(data.Level), optionalStringValue(data.Search))
	if err != nil {
		addClientError(&resp.Diagnostics, "read sandbox logs", err)
		return
	}

	data.Logs = make([]SandboxLogModel, 0, len(logs))
	for _, log := range logs {
		model, diags := flattenSandboxLog(ctx, log)
		resp.Diagnostics.Append(diags...)
		data.Logs = append(data.Logs, model)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func sandboxMetricAttributes(includeSandboxID bool) map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
		"timestamp":      schema.StringAttribute{MarkdownDescription: "Deprecated RFC3339 timestamp returned by E2B.", Computed: true},
		"timestamp_unix": schema.Int64Attribute{MarkdownDescription: "Unix timestamp for the metric point.", Computed: true},
		"cpu_count":      schema.Int64Attribute{MarkdownDescription: "Number of CPU cores.", Computed: true},
		"cpu_used_pct":   schema.Float64Attribute{MarkdownDescription: "CPU usage percentage.", Computed: true},
		"mem_used":       schema.Int64Attribute{MarkdownDescription: "Memory used in bytes.", Computed: true},
		"mem_total":      schema.Int64Attribute{MarkdownDescription: "Total memory in bytes.", Computed: true},
		"mem_cache":      schema.Int64Attribute{MarkdownDescription: "Cached memory in bytes.", Computed: true},
		"disk_used":      schema.Int64Attribute{MarkdownDescription: "Disk used in bytes.", Computed: true},
		"disk_total":     schema.Int64Attribute{MarkdownDescription: "Total disk space in bytes.", Computed: true},
	}
	if includeSandboxID {
		attributes["sandbox_id"] = schema.StringAttribute{MarkdownDescription: "Sandbox ID.", Computed: true}
	}

	return attributes
}

func flattenSandboxMetric(metric sandboxMetricResponse) SandboxMetricModel {
	return SandboxMetricModel{
		Timestamp:     types.StringValue(metric.Timestamp),
		TimestampUnix: types.Int64Value(metric.TimestampUnix),
		CPUCount:      types.Int64Value(metric.CPUCount),
		CPUUsedPct:    types.Float64Value(metric.CPUUsedPct),
		MemUsed:       types.Int64Value(metric.MemUsed),
		MemTotal:      types.Int64Value(metric.MemTotal),
		MemCache:      types.Int64Value(metric.MemCache),
		DiskUsed:      types.Int64Value(metric.DiskUsed),
		DiskTotal:     types.Int64Value(metric.DiskTotal),
	}
}

func flattenSandboxMetricRow(sandboxID string, metric sandboxMetricResponse) SandboxMetricRow {
	return SandboxMetricRow{
		SandboxID:     types.StringValue(sandboxID),
		Timestamp:     types.StringValue(metric.Timestamp),
		TimestampUnix: types.Int64Value(metric.TimestampUnix),
		CPUCount:      types.Int64Value(metric.CPUCount),
		CPUUsedPct:    types.Float64Value(metric.CPUUsedPct),
		MemUsed:       types.Int64Value(metric.MemUsed),
		MemTotal:      types.Int64Value(metric.MemTotal),
		MemCache:      types.Int64Value(metric.MemCache),
		DiskUsed:      types.Int64Value(metric.DiskUsed),
		DiskTotal:     types.Int64Value(metric.DiskTotal),
	}
}

func flattenSandboxLog(ctx context.Context, log sandboxLogEntryResponse) (SandboxLogModel, diag.Diagnostics) {
	fields, diags := mapStringValue(ctx, log.Fields)

	return SandboxLogModel{
		Timestamp: types.StringValue(log.Timestamp),
		Level:     types.StringValue(log.Level),
		Message:   types.StringValue(log.Message),
		Fields:    fields,
	}, diags
}
