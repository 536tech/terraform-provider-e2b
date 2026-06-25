// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SandboxDataSource{}
var _ datasource.DataSource = &SandboxesDataSource{}

func NewSandboxDataSource() datasource.DataSource {
	return &SandboxDataSource{}
}

func NewSandboxesDataSource() datasource.DataSource {
	return &SandboxesDataSource{}
}

type SandboxDataSource struct {
	client *e2bClient
}

type SandboxDataSourceModel struct {
	ID                        types.String              `tfsdk:"id"`
	TemplateID                types.String              `tfsdk:"template_id"`
	Alias                     types.String              `tfsdk:"alias"`
	ClientID                  types.String              `tfsdk:"client_id"`
	StartedAt                 types.String              `tfsdk:"started_at"`
	EndAt                     types.String              `tfsdk:"end_at"`
	EnvdVersion               types.String              `tfsdk:"envd_version"`
	EnvdAccessToken           types.String              `tfsdk:"envd_access_token"`
	TrafficAccessToken        types.String              `tfsdk:"traffic_access_token"`
	AllowInternetAccess       types.Bool                `tfsdk:"allow_internet_access"`
	NetworkAllowPublicTraffic types.Bool                `tfsdk:"network_allow_public_traffic"`
	NetworkMaskRequestHost    types.String              `tfsdk:"network_mask_request_host"`
	NetworkEgressProxy        *SandboxEgressProxyModel  `tfsdk:"network_egress_proxy"`
	NetworkRules              types.Map                 `tfsdk:"network_rules"`
	LifecycleAutoResume       types.Bool                `tfsdk:"lifecycle_auto_resume"`
	LifecycleOnTimeout        types.String              `tfsdk:"lifecycle_on_timeout"`
	CPUCount                  types.Int64               `tfsdk:"cpu_count"`
	MemoryMB                  types.Int64               `tfsdk:"memory_mb"`
	DiskSizeMB                types.Int64               `tfsdk:"disk_size_mb"`
	Metadata                  types.Map                 `tfsdk:"metadata"`
	NetworkAllowOut           types.Set                 `tfsdk:"network_allow_out"`
	NetworkDenyOut            types.Set                 `tfsdk:"network_deny_out"`
	VolumeMounts              []SandboxVolumeMountModel `tfsdk:"volume_mounts"`
	State                     types.String              `tfsdk:"state"`
}

type SandboxesDataSource struct {
	client *e2bClient
}

type SandboxesDataSourceModel struct {
	Metadata  types.Map            `tfsdk:"metadata"`
	States    types.Set            `tfsdk:"states"`
	Limit     types.Int64          `tfsdk:"limit"`
	Sandboxes []ListedSandboxModel `tfsdk:"sandboxes"`
}

type ListedSandboxModel struct {
	ID           types.String              `tfsdk:"id"`
	TemplateID   types.String              `tfsdk:"template_id"`
	Alias        types.String              `tfsdk:"alias"`
	ClientID     types.String              `tfsdk:"client_id"`
	StartedAt    types.String              `tfsdk:"started_at"`
	EndAt        types.String              `tfsdk:"end_at"`
	CPUCount     types.Int64               `tfsdk:"cpu_count"`
	MemoryMB     types.Int64               `tfsdk:"memory_mb"`
	DiskSizeMB   types.Int64               `tfsdk:"disk_size_mb"`
	Metadata     types.Map                 `tfsdk:"metadata"`
	State        types.String              `tfsdk:"state"`
	EnvdVersion  types.String              `tfsdk:"envd_version"`
	VolumeMounts []SandboxVolumeMountModel `tfsdk:"volume_mounts"`
}

func (d *SandboxDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox"
}

func (d *SandboxDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an E2B sandbox by ID.",
		Attributes:          sandboxDetailAttributes(true),
	}
}

func (d *SandboxDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SandboxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SandboxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := d.client.getSandbox(ctx, data.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "read sandbox", err)
		return
	}

	resp.Diagnostics.Append(data.applySandboxDetail(ctx, detail)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *SandboxesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandboxes"
}

func (d *SandboxesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B sandboxes.",
		Attributes: map[string]schema.Attribute{
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Optional metadata filter.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"states": schema.SetAttribute{
				MarkdownDescription: "Optional sandbox state filters, such as `running` or `paused`.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of sandboxes to return. E2B treats `0` as no explicit limit.",
				Optional:            true,
			},
			"sandboxes": schema.ListNestedAttribute{
				MarkdownDescription: "Matching sandboxes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: listedSandboxAttributes(),
				},
			},
		},
	}
}

func (d *SandboxesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SandboxesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SandboxesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, diags := mapStringFromTerraform(ctx, data.Metadata)
	resp.Diagnostics.Append(diags...)
	states, diags := stringSetFromTerraform(ctx, data.States)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var limit int64
	if !data.Limit.IsNull() && !data.Limit.IsUnknown() {
		limit = data.Limit.ValueInt64()
	}

	sandboxes, err := d.client.listSandboxes(ctx, metadata, states, limit)
	if err != nil {
		addClientError(&resp.Diagnostics, "list sandboxes", err)
		return
	}

	data.Sandboxes = make([]ListedSandboxModel, 0, len(sandboxes))
	for _, sandbox := range sandboxes {
		model, diags := flattenListedSandbox(ctx, sandbox)
		resp.Diagnostics.Append(diags...)
		data.Sandboxes = append(data.Sandboxes, model)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *SandboxDataSourceModel) applySandboxDetail(ctx context.Context, detail *sandboxDetailResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(detail.SandboxID)
	m.TemplateID = types.StringValue(detail.TemplateID)
	m.Alias = types.StringValue(detail.Alias)
	m.ClientID = types.StringValue(detail.ClientID)
	m.StartedAt = types.StringValue(detail.StartedAt)
	m.EndAt = types.StringValue(detail.EndAt)
	m.EnvdVersion = types.StringValue(detail.EnvdVersion)
	m.EnvdAccessToken = stringPointerValue(detail.EnvdAccessToken)
	m.TrafficAccessToken = stringPointerValue(detail.TrafficAccessToken)
	m.AllowInternetAccess = boolPointerValue(detail.AllowInternetAccess)
	m.CPUCount = int64PointerValue(detail.CPUCount)
	m.MemoryMB = int64PointerValue(detail.MemoryMB)
	m.DiskSizeMB = int64PointerValue(detail.DiskSizeMB)
	m.State = types.StringValue(detail.State)
	if detail.Lifecycle != nil {
		m.LifecycleAutoResume = types.BoolValue(detail.Lifecycle.AutoResume)
		m.LifecycleOnTimeout = types.StringValue(detail.Lifecycle.OnTimeout)
	} else {
		m.LifecycleAutoResume = types.BoolNull()
		m.LifecycleOnTimeout = types.StringNull()
	}
	m.VolumeMounts = flattenSandboxVolumeMounts(detail.VolumeMounts)

	metadata, mapDiags := mapStringValue(ctx, detail.Metadata)
	diags.Append(mapDiags...)
	m.Metadata = metadata

	if detail.Network == nil {
		m.NetworkAllowPublicTraffic = types.BoolNull()
		m.NetworkMaskRequestHost = types.StringNull()
		m.NetworkEgressProxy = nil
		m.NetworkRules = types.MapNull(sandboxNetworkRuleListType())
		m.NetworkAllowOut = types.SetNull(types.StringType)
		m.NetworkDenyOut = types.SetNull(types.StringType)
		return diags
	}

	m.NetworkAllowPublicTraffic = boolPointerValue(detail.Network.AllowPublicTraffic)
	if detail.Network.MaskRequestHost == "" {
		m.NetworkMaskRequestHost = types.StringNull()
	} else {
		m.NetworkMaskRequestHost = types.StringValue(detail.Network.MaskRequestHost)
	}

	allowOut, setDiags := setStringValue(ctx, detail.Network.AllowOut)
	diags.Append(setDiags...)
	m.NetworkAllowOut = allowOut

	denyOut, setDiags := setStringValue(ctx, detail.Network.DenyOut)
	diags.Append(setDiags...)
	m.NetworkDenyOut = denyOut
	m.NetworkEgressProxy = flattenSandboxEgressProxy(detail.Network.EgressProxy, nil)
	rules, rulesDiags := sandboxNetworkRulesValue(ctx, detail.Network.Rules)
	diags.Append(rulesDiags...)
	m.NetworkRules = rules

	return diags
}

func flattenListedSandbox(ctx context.Context, sandbox listedSandboxResponse) (ListedSandboxModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	metadata, mapDiags := mapStringValue(ctx, sandbox.Metadata)
	diags.Append(mapDiags...)

	return ListedSandboxModel{
		ID:           types.StringValue(sandbox.SandboxID),
		TemplateID:   types.StringValue(sandbox.TemplateID),
		Alias:        types.StringValue(sandbox.Alias),
		ClientID:     types.StringValue(sandbox.ClientID),
		StartedAt:    types.StringValue(sandbox.StartedAt),
		EndAt:        types.StringValue(sandbox.EndAt),
		CPUCount:     int64PointerValue(sandbox.CPUCount),
		MemoryMB:     int64PointerValue(sandbox.MemoryMB),
		DiskSizeMB:   int64PointerValue(sandbox.DiskSizeMB),
		Metadata:     metadata,
		State:        types.StringValue(sandbox.State),
		EnvdVersion:  types.StringValue(sandbox.EnvdVersion),
		VolumeMounts: flattenSandboxVolumeMounts(sandbox.VolumeMounts),
	}, diags
}

func sandboxDetailAttributes(idRequired bool) map[string]schema.Attribute {
	idAttribute := schema.StringAttribute{
		MarkdownDescription: "Sandbox ID.",
		Computed:            true,
	}
	if idRequired {
		idAttribute = schema.StringAttribute{
			MarkdownDescription: "Sandbox ID.",
			Required:            true,
		}
	}

	attributes := listedSandboxAttributes()
	attributes["id"] = idAttribute
	attributes["envd_access_token"] = schema.StringAttribute{
		MarkdownDescription: "Access token for authenticated envd requests when secure sandbox mode is enabled.",
		Computed:            true,
		Sensitive:           true,
	}
	attributes["traffic_access_token"] = schema.StringAttribute{
		MarkdownDescription: "Access token for authenticated sandbox traffic when secure sandbox mode is enabled.",
		Computed:            true,
		Sensitive:           true,
	}
	attributes["allow_internet_access"] = schema.BoolAttribute{
		MarkdownDescription: "Whether internet access was explicitly enabled or disabled for the sandbox.",
		Computed:            true,
	}
	attributes["network_allow_public_traffic"] = schema.BoolAttribute{
		MarkdownDescription: "Whether the sandbox may receive public traffic.",
		Computed:            true,
	}
	attributes["network_mask_request_host"] = schema.StringAttribute{
		MarkdownDescription: "Host value E2B masks on incoming sandbox requests.",
		Computed:            true,
	}
	attributes["network_allow_out"] = schema.SetAttribute{
		MarkdownDescription: "Destinations that sandbox egress traffic is allowed to reach.",
		ElementType:         types.StringType,
		Computed:            true,
	}
	attributes["network_deny_out"] = schema.SetAttribute{
		MarkdownDescription: "CIDR blocks, IP addresses, or `ALL_TRAFFIC` entries that sandbox egress traffic is denied from reaching.",
		ElementType:         types.StringType,
		Computed:            true,
	}
	attributes["network_egress_proxy"] = schema.SingleNestedAttribute{
		MarkdownDescription: "SOCKS5 proxy configured for sandbox egress traffic.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "SOCKS5 proxy address in `host:port` format.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "SOCKS5 username.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SOCKS5 password, when returned by E2B.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
	attributes["network_rules"] = schema.MapAttribute{
		MarkdownDescription: "Per-domain egress HTTP/HTTPS transform rules.",
		ElementType:         sandboxNetworkRuleListType(),
		Computed:            true,
	}
	attributes["lifecycle_auto_resume"] = schema.BoolAttribute{
		MarkdownDescription: "Whether the sandbox is configured to auto-resume.",
		Computed:            true,
	}
	attributes["lifecycle_on_timeout"] = schema.StringAttribute{
		MarkdownDescription: "Lifecycle action E2B applies when the sandbox timeout is reached.",
		Computed:            true,
	}

	return attributes
}

func listedSandboxAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Sandbox ID.",
			Computed:            true,
		},
		"template_id": schema.StringAttribute{
			MarkdownDescription: "Template ID used to create the sandbox.",
			Computed:            true,
		},
		"alias": schema.StringAttribute{
			MarkdownDescription: "Template alias used to create the sandbox.",
			Computed:            true,
		},
		"client_id": schema.StringAttribute{
			MarkdownDescription: "E2B client ID for the sandbox.",
			Computed:            true,
		},
		"started_at": schema.StringAttribute{
			MarkdownDescription: "Timestamp when the sandbox started.",
			Computed:            true,
		},
		"end_at": schema.StringAttribute{
			MarkdownDescription: "Timestamp when the sandbox will expire.",
			Computed:            true,
		},
		"cpu_count": schema.Int64Attribute{
			MarkdownDescription: "Number of CPU cores allocated to the sandbox.",
			Computed:            true,
		},
		"memory_mb": schema.Int64Attribute{
			MarkdownDescription: "Memory allocated to the sandbox, in MB.",
			Computed:            true,
		},
		"disk_size_mb": schema.Int64Attribute{
			MarkdownDescription: "Disk allocated to the sandbox, in MB.",
			Computed:            true,
		},
		"metadata": schema.MapAttribute{
			MarkdownDescription: "Sandbox metadata.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "Sandbox state.",
			Computed:            true,
		},
		"envd_version": schema.StringAttribute{
			MarkdownDescription: "Version of envd running in the sandbox.",
			Computed:            true,
		},
		"volume_mounts": schema.ListNestedAttribute{
			MarkdownDescription: "Volumes mounted into the sandbox.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "E2B volume name.",
						Computed:            true,
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Path where the volume is mounted in the sandbox.",
						Computed:            true,
					},
				},
			},
		},
	}
}
