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

var _ datasource.DataSource = &TemplateDataSource{}
var _ datasource.DataSource = &TemplatesDataSource{}

func NewTemplateDataSource() datasource.DataSource {
	return &TemplateDataSource{}
}

func NewTemplatesDataSource() datasource.DataSource {
	return &TemplatesDataSource{}
}

type TemplateDataSource struct {
	client *e2bClient
}

type TemplateDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	BuildID       types.String `tfsdk:"build_id"`
	CPUCount      types.Int64  `tfsdk:"cpu_count"`
	MemoryMB      types.Int64  `tfsdk:"memory_mb"`
	DiskSizeMB    types.Int64  `tfsdk:"disk_size_mb"`
	Public        types.Bool   `tfsdk:"public"`
	Aliases       types.List   `tfsdk:"aliases"`
	Names         types.List   `tfsdk:"names"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	LastSpawnedAt types.String `tfsdk:"last_spawned_at"`
	SpawnCount    types.Int64  `tfsdk:"spawn_count"`
	BuildCount    types.Int64  `tfsdk:"build_count"`
	EnvdVersion   types.String `tfsdk:"envd_version"`
	BuildStatus   types.String `tfsdk:"build_status"`
}

type TemplatesDataSource struct {
	client *e2bClient
}

type TemplatesDataSourceModel struct {
	Templates []TemplateDataSourceModel `tfsdk:"templates"`
}

func (d *TemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (d *TemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads E2B template metadata by template ID, alias, or name.",
		Attributes:          templateAttributes(true),
	}
}

func (d *TemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := d.client.getTemplate(ctx, data.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "read template", err)
		return
	}

	resp.Diagnostics.Append(data.applyTemplate(ctx, template)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *TemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}

func (d *TemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B templates accessible to the configured API key.",
		Attributes: map[string]schema.Attribute{
			"templates": schema.ListNestedAttribute{
				MarkdownDescription: "Templates accessible to the configured API key.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: templateAttributes(false),
				},
			},
		},
	}
}

func (d *TemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TemplatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	templates, err := d.client.listTemplates(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list templates", err)
		return
	}

	data.Templates = make([]TemplateDataSourceModel, 0, len(templates))
	for _, template := range templates {
		model := TemplateDataSourceModel{}
		resp.Diagnostics.Append(model.applyTemplate(ctx, &template)...)
		data.Templates = append(data.Templates, model)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *TemplateDataSourceModel) applyTemplate(ctx context.Context, template *templateResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	aliases, listDiags := listStringValue(ctx, template.Aliases)
	diags.Append(listDiags...)
	names, listDiags := listStringValue(ctx, template.Names)
	diags.Append(listDiags...)

	m.ID = types.StringValue(template.TemplateID)
	m.BuildID = types.StringValue(template.BuildID)
	m.CPUCount = types.Int64Value(template.CPUCount)
	m.MemoryMB = types.Int64Value(template.MemoryMB)
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
	m.BuildStatus = types.StringValue(template.BuildStatus)

	return diags
}

func templateAttributes(idRequired bool) map[string]schema.Attribute {
	idAttribute := schema.StringAttribute{
		MarkdownDescription: "Template ID.",
		Computed:            true,
	}
	if idRequired {
		idAttribute = schema.StringAttribute{
			MarkdownDescription: "Template ID, alias, or name.",
			Required:            true,
		}
	}

	return map[string]schema.Attribute{
		"id": idAttribute,
		"build_id": schema.StringAttribute{
			MarkdownDescription: "Identifier of the latest successful template build.",
			Computed:            true,
		},
		"cpu_count": schema.Int64Attribute{
			MarkdownDescription: "Number of CPUs configured for the template.",
			Computed:            true,
		},
		"memory_mb": schema.Int64Attribute{
			MarkdownDescription: "Memory configured for the template, in MB.",
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
	}
}
