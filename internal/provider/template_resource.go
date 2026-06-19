// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const templateBuildTimeout = 15 * time.Minute

var _ resource.Resource = &TemplateResource{}
var _ resource.ResourceWithImportState = &TemplateResource{}

func NewTemplateResource() resource.Resource {
	return &TemplateResource{}
}

type TemplateResource struct {
	client *e2bClient
}

type TemplateResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	FromImage  types.String `tfsdk:"from_image"`
	StartCmd   types.String `tfsdk:"start_cmd"`
	ReadyCmd   types.String `tfsdk:"ready_cmd"`
	CPUCount   types.Int64  `tfsdk:"cpu_count"`
	MemoryMB   types.Int64  `tfsdk:"memory_mb"`
	Force      types.Bool   `tfsdk:"force"`
	BuildID    types.String `tfsdk:"build_id"`
	DiskSizeMB types.Int64  `tfsdk:"disk_size_mb"`
	Public     types.Bool   `tfsdk:"public"`
	Aliases    types.List   `tfsdk:"aliases"`
	Names      types.List   `tfsdk:"names"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`

	LastSpawnedAt types.String `tfsdk:"last_spawned_at"`
	SpawnCount    types.Int64  `tfsdk:"spawn_count"`
	BuildCount    types.Int64  `tfsdk:"build_count"`
	EnvdVersion   types.String `tfsdk:"envd_version"`
	BuildStatus   types.String `tfsdk:"build_status"`
}

func (r *TemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *TemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an E2B template built from a container image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Template ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name. The name can include a tag separated by a colon, such as `my-template:v1`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"from_image": schema.StringAttribute{
				MarkdownDescription: "Container image used as the template build base. Required when creating a template.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_cmd": schema.StringAttribute{
				MarkdownDescription: "Command E2B runs when a sandbox starts from the template.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ready_cmd": schema.StringAttribute{
				MarkdownDescription: "Command E2B runs to determine when sandboxes from the template are ready.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cpu_count": schema.Int64Attribute{
				MarkdownDescription: "Number of CPUs configured for the template.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"memory_mb": schema.Int64Attribute{
				MarkdownDescription: "Memory configured for the template, in MB.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Whether E2B should force the build to run regardless of cache.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"build_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the latest successful template build.",
				Computed:            true,
			},
			"disk_size_mb": schema.Int64Attribute{
				MarkdownDescription: "Disk size configured for the template, in MB.",
				Computed:            true,
			},
			"public": schema.BoolAttribute{
				MarkdownDescription: "Whether the template is public.",
				Computed:            true,
			},
			"aliases": schema.ListAttribute{
				MarkdownDescription: "Template aliases.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"names": schema.ListAttribute{
				MarkdownDescription: "Fully qualified template names.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the template was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the template was last updated.",
				Computed:            true,
			},
			"last_spawned_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the template was last spawned.",
				Computed:            true,
			},
			"spawn_count": schema.Int64Attribute{
				MarkdownDescription: "Number of sandboxes created from the template.",
				Computed:            true,
			},
			"build_count": schema.Int64Attribute{
				MarkdownDescription: "Number of template builds.",
				Computed:            true,
			},
			"envd_version": schema.StringAttribute{
				MarkdownDescription: "Template envd version.",
				Computed:            true,
			},
			"build_status": schema.StringAttribute{
				MarkdownDescription: "Current build status for the template.",
				Computed:            true,
			},
		},
	}
}

func (r *TemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.FromImage.IsNull() || data.FromImage.IsUnknown() || strings.TrimSpace(data.FromImage.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("from_image"),
			"Missing Template Source Image",
			"Set from_image to the container image E2B should use as the template build base.",
		)
		return
	}

	created, err := r.client.createTemplate(ctx, templateCreateRequest{
		Name:     data.Name.ValueString(),
		CPUCount: optionalInt64Pointer(data.CPUCount),
		MemoryMB: optionalInt64Pointer(data.MemoryMB),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "create template", err)
		return
	}

	data.ID = types.StringValue(created.TemplateID)
	data.BuildID = types.StringValue(created.BuildID)
	data.Public = types.BoolValue(created.Public)
	resp.Diagnostics.Append(data.applyCreatedTemplate(ctx, created)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildRequest := templateBuildStartRequest{
		FromImage: strings.TrimSpace(data.FromImage.ValueString()),
		Force:     optionalBoolPointer(data.Force),
	}
	if !data.StartCmd.IsNull() && !data.StartCmd.IsUnknown() {
		buildRequest.StartCmd = data.StartCmd.ValueString()
	}
	if !data.ReadyCmd.IsNull() && !data.ReadyCmd.IsUnknown() {
		buildRequest.ReadyCmd = data.ReadyCmd.ValueString()
	}

	if err := r.client.startTemplateBuild(ctx, created.TemplateID, created.BuildID, buildRequest); err != nil {
		addClientError(&resp.Diagnostics, "start template build", err)
		return
	}

	waitCtx, cancel := context.WithTimeout(ctx, templateBuildTimeout)
	defer cancel()

	buildStatus, err := r.client.waitForTemplateBuild(waitCtx, created.TemplateID, created.BuildID, 5*time.Second)
	if err != nil {
		addClientError(&resp.Diagnostics, "wait for template build", err)
		return
	}

	template, err := r.client.getTemplate(ctx, created.TemplateID)
	if err != nil {
		addClientError(&resp.Diagnostics, "read created template", err)
		return
	}

	resp.Diagnostics.Append(data.applyTemplate(ctx, template)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if buildStatus.Status != "" {
		data.BuildStatus = types.StringValue(buildStatus.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := r.client.getTemplate(ctx, data.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "read template", err)
		return
	}

	resp.Diagnostics.Append(data.applyTemplate(ctx, template)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refreshTemplateBuildStatus(ctx, &data); err != nil {
		addClientError(&resp.Diagnostics, "read template build status", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Template updates require replacement",
		"The E2B template resource marks configurable attributes as ForceNew. Terraform should replace the template instead of updating it in place.",
	)
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteTemplate(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete template", err)
	}
}

func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TemplateResource) refreshTemplateBuildStatus(ctx context.Context, data *TemplateResourceModel) error {
	if data.ID.IsNull() || data.ID.IsUnknown() || data.BuildID.IsNull() || data.BuildID.IsUnknown() {
		return nil
	}

	templateID := data.ID.ValueString()
	buildID := data.BuildID.ValueString()
	if buildID == "" || buildID == "00000000-0000-0000-0000-000000000000" {
		return nil
	}

	status, err := r.client.getTemplateBuildStatus(ctx, templateID, buildID)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if status.Status != "" {
		data.BuildStatus = types.StringValue(status.Status)
	}

	return nil
}

func (m *TemplateResourceModel) applyCreatedTemplate(ctx context.Context, template *templateCreateResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	aliases, listDiags := listStringValue(ctx, template.Aliases)
	diags.Append(listDiags...)
	names, listDiags := listStringValue(ctx, template.Names)
	diags.Append(listDiags...)

	m.ID = types.StringValue(template.TemplateID)
	m.BuildID = types.StringValue(template.BuildID)
	m.Public = types.BoolValue(template.Public)
	m.Aliases = aliases
	m.Names = names

	return diags
}

func (m *TemplateResourceModel) applyTemplate(ctx context.Context, template *templateResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	aliases, listDiags := listStringValue(ctx, template.Aliases)
	diags.Append(listDiags...)
	names, listDiags := listStringValue(ctx, template.Names)
	diags.Append(listDiags...)

	m.ID = types.StringValue(template.TemplateID)
	if m.Name.IsNull() || m.Name.IsUnknown() || m.Name.ValueString() == "" {
		m.Name = types.StringValue(inferTemplateName(template))
	}
	m.BuildID = types.StringValue(template.BuildID)
	if template.CPUCount > 0 || m.CPUCount.IsNull() || m.CPUCount.IsUnknown() {
		m.CPUCount = types.Int64Value(template.CPUCount)
	}
	if template.MemoryMB > 0 || m.MemoryMB.IsNull() || m.MemoryMB.IsUnknown() {
		m.MemoryMB = types.Int64Value(template.MemoryMB)
	}
	m.DiskSizeMB = types.Int64Value(template.DiskSizeMB)
	m.Public = types.BoolValue(template.Public)
	m.Aliases = aliases
	m.Names = names
	m.CreatedAt = types.StringValue(template.CreatedAt)
	m.UpdatedAt = types.StringValue(template.UpdatedAt)
	m.LastSpawnedAt = stringPointerValue(template.LastSpawnedAt)
	m.SpawnCount = types.Int64Value(template.SpawnCount)
	m.BuildCount = types.Int64Value(template.BuildCount)
	m.EnvdVersion = types.StringValue(template.EnvdVersion)
	if template.BuildStatus != "" || m.BuildStatus.IsNull() || m.BuildStatus.IsUnknown() {
		m.BuildStatus = types.StringValue(template.BuildStatus)
	}

	return diags
}

func inferTemplateName(template *templateResponse) string {
	if len(template.Aliases) > 0 {
		return template.Aliases[0]
	}

	if len(template.Names) == 0 {
		return template.TemplateID
	}

	name := template.Names[0]
	if afterSlash := strings.LastIndex(name, "/"); afterSlash >= 0 && afterSlash < len(name)-1 {
		return name[afterSlash+1:]
	}

	return name
}
