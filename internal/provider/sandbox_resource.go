// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SandboxResource{}
var _ resource.ResourceWithImportState = &SandboxResource{}

func NewSandboxResource() resource.Resource {
	return &SandboxResource{}
}

type SandboxResource struct {
	client *e2bClient
}

type SandboxResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	TemplateID          types.String `tfsdk:"template_id"`
	Timeout             types.Int64  `tfsdk:"timeout"`
	AutoPause           types.Bool   `tfsdk:"auto_pause"`
	Secure              types.Bool   `tfsdk:"secure"`
	AllowInternetAccess types.Bool   `tfsdk:"allow_internet_access"`
	Metadata            types.Map    `tfsdk:"metadata"`
	EnvVars             types.Map    `tfsdk:"env_vars"`
	Alias               types.String `tfsdk:"alias"`
	ClientID            types.String `tfsdk:"client_id"`
	EnvdVersion         types.String `tfsdk:"envd_version"`
	EnvdAccessToken     types.String `tfsdk:"envd_access_token"`
	TrafficAccessToken  types.String `tfsdk:"traffic_access_token"`
	State               types.String `tfsdk:"state"`
	StartedAt           types.String `tfsdk:"started_at"`
	EndAt               types.String `tfsdk:"end_at"`
	CPUCount            types.Int64  `tfsdk:"cpu_count"`
	MemoryMB            types.Int64  `tfsdk:"memory_mb"`
	DiskSizeMB          types.Int64  `tfsdk:"disk_size_mb"`
}

func (r *SandboxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox"
}

func (r *SandboxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a running E2B sandbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Sandbox ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Template ID, alias, or name used to create the sandbox. For example, `base`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Sandbox time to live in seconds.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"auto_pause": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox should pause instead of being killed when the timeout is reached.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"secure": schema.BoolAttribute{
				MarkdownDescription: "Whether E2B should secure system communication with the sandbox and return access tokens.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"allow_internet_access": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox can access the internet.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Metadata assigned to the sandbox.",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"env_vars": schema.MapAttribute{
				MarkdownDescription: "Environment variables injected into the sandbox.",
				ElementType:         types.StringType,
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias of the template used to create the sandbox.",
				Computed:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "E2B client ID for the sandbox.",
				Computed:            true,
			},
			"envd_version": schema.StringAttribute{
				MarkdownDescription: "Version of envd running in the sandbox.",
				Computed:            true,
			},
			"envd_access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authenticated envd requests when secure sandbox mode is enabled.",
				Computed:            true,
				Sensitive:           true,
			},
			"traffic_access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authenticated sandbox traffic when secure sandbox mode is enabled.",
				Computed:            true,
				Sensitive:           true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current sandbox state.",
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
		},
	}
}

func (r *SandboxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*e2bClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *e2bClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SandboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, diags := mapStringFromTerraform(ctx, data.Metadata)
	resp.Diagnostics.Append(diags...)
	envVars, diags := mapStringFromTerraform(ctx, data.EnvVars)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.createSandbox(ctx, sandboxCreateRequest{
		TemplateID:           data.TemplateID.ValueString(),
		Timeout:              optionalInt64Pointer(data.Timeout),
		AutoPause:            optionalBoolPointer(data.AutoPause),
		Secure:               optionalBoolPointer(data.Secure),
		AllowInternetAccess:  optionalBoolPointer(data.AllowInternetAccess),
		Metadata:             metadata,
		EnvironmentVariables: envVars,
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "create sandbox", err)
		return
	}

	data.ID = types.StringValue(created.SandboxID)
	data.Alias = types.StringValue(created.Alias)
	data.ClientID = types.StringValue(created.ClientID)
	data.EnvdVersion = types.StringValue(created.EnvdVersion)
	data.EnvdAccessToken = stringPointerValue(created.EnvdAccessToken)
	data.TrafficAccessToken = stringPointerValue(created.TrafficAccessToken)

	if detail, err := r.client.getSandbox(ctx, created.SandboxID); err == nil {
		resp.Diagnostics.Append(data.applySandboxDetail(ctx, detail)...)
	} else if !isNotFound(err) {
		addClientError(&resp.Diagnostics, "read created sandbox", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SandboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.getSandbox(ctx, data.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
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

func (r *SandboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Sandbox updates require replacement",
		"The E2B sandbox resource marks configurable attributes as ForceNew. Terraform should replace the sandbox instead of updating it in place.",
	)
}

func (r *SandboxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteSandbox(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete sandbox", err)
	}
}

func (r *SandboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *SandboxResourceModel) applySandboxDetail(ctx context.Context, detail *sandboxDetailResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(detail.SandboxID)
	if m.TemplateID.IsNull() || m.TemplateID.IsUnknown() || m.TemplateID.ValueString() == "" {
		m.TemplateID = types.StringValue(detail.TemplateID)
	}
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

	metadata, mapDiags := mapStringValue(ctx, detail.Metadata)
	diags.Append(mapDiags...)
	if !m.Metadata.IsNull() || len(detail.Metadata) > 0 {
		m.Metadata = metadata
	}

	return diags
}
