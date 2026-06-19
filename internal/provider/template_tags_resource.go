// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TemplateTagsResource{}
var _ datasource.DataSource = &TemplateTagsDataSource{}

func NewTemplateTagsResource() resource.Resource {
	return &TemplateTagsResource{}
}

func NewTemplateTagsDataSource() datasource.DataSource {
	return &TemplateTagsDataSource{}
}

type TemplateTagsResource struct {
	client *e2bClient
}

type TemplateTagsResourceModel struct {
	ID         types.String `tfsdk:"id"`
	TemplateID types.String `tfsdk:"template_id"`
	Target     types.String `tfsdk:"target"`
	Name       types.String `tfsdk:"name"`
	Tags       types.Set    `tfsdk:"tags"`
	BuildID    types.String `tfsdk:"build_id"`
}

type TemplateTagsDataSource struct {
	client *e2bClient
}

type TemplateTagsDataSourceModel struct {
	TemplateID types.String       `tfsdk:"template_id"`
	Tags       []TemplateTagModel `tfsdk:"tags"`
}

type TemplateTagModel struct {
	Tag       types.String `tfsdk:"tag"`
	BuildID   types.String `tfsdk:"build_id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *TemplateTagsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_tags"
}

func (r *TemplateTagsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Assigns tag aliases to an E2B template build.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "Template ID used for tag readback.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": resourceschema.StringAttribute{
				MarkdownDescription: "Template ID used to list assigned tags.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": resourceschema.StringAttribute{
				MarkdownDescription: "Template build target to tag, usually a fully qualified template name such as `my-template:build-tag`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "Template name to use when deleting tags. Defaults to the part of `target` before the last colon.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": resourceschema.SetAttribute{
				MarkdownDescription: "Tag aliases to assign to the template build.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"build_id": resourceschema.StringAttribute{
				MarkdownDescription: "Build ID returned by E2B when tags were assigned.",
				Computed:            true,
			},
		},
	}
}

func (r *TemplateTagsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TemplateTagsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TemplateTagsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, diags := stringSetFromTerraform(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assigned, err := r.client.assignTemplateTags(ctx, data.Target.ValueString(), tags)
	if err != nil {
		addClientError(&resp.Diagnostics, "assign template tags", err)
		return
	}

	data.ID = data.TemplateID
	data.BuildID = types.StringValue(assigned.BuildID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateTagsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TemplateTagsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readTemplateTagsState(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, diags := stringSetFromTerraform(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(tags) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateTagsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TemplateTagsResourceModel
	var state TemplateTagsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planTags, diags := stringSetFromTerraform(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	stateTags, diags := stringSetFromTerraform(ctx, state.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	added := stringDifference(planTags, stateTags)
	if len(added) > 0 {
		assigned, err := r.client.assignTemplateTags(ctx, plan.Target.ValueString(), added)
		if err != nil {
			addClientError(&resp.Diagnostics, "assign template tags", err)
			return
		}
		plan.BuildID = types.StringValue(assigned.BuildID)
	}

	removed := stringDifference(stateTags, planTags)
	if len(removed) > 0 {
		if err := r.client.deleteTemplateTags(ctx, templateTagsDeleteName(state), removed); err != nil {
			addClientError(&resp.Diagnostics, "delete template tags", err)
			return
		}
	}

	plan.ID = plan.TemplateID
	resp.Diagnostics.Append(r.readTemplateTagsState(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateTagsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TemplateTagsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, diags := stringSetFromTerraform(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(tags) == 0 {
		return
	}

	if err := r.client.deleteTemplateTags(ctx, templateTagsDeleteName(data), tags); err != nil {
		addClientError(&resp.Diagnostics, "delete template tags", err)
	}
}

func (r *TemplateTagsResource) readTemplateTagsState(ctx context.Context, data *TemplateTagsResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	current, err := r.client.listTemplateTags(ctx, data.TemplateID.ValueString())
	if err != nil {
		addClientError(&diags, "list template tags", err)
		return diags
	}

	wanted, setDiags := stringSetFromTerraform(ctx, data.Tags)
	diags.Append(setDiags...)
	if diags.HasError() {
		return diags
	}

	wantedSet := map[string]struct{}{}
	for _, tag := range wanted {
		wantedSet[tag] = struct{}{}
	}

	present := make([]string, 0, len(wanted))
	for _, tag := range current {
		if _, ok := wantedSet[tag.Tag]; ok {
			present = append(present, tag.Tag)
			if tag.BuildID != "" {
				data.BuildID = types.StringValue(tag.BuildID)
			}
		}
	}
	sort.Strings(present)

	tagsValue, setDiags := setStringValue(ctx, present)
	diags.Append(setDiags...)
	data.Tags = tagsValue
	data.ID = data.TemplateID

	return diags
}

func (d *TemplateTagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_tags"
}

func (d *TemplateTagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists tags assigned to an E2B template.",
		Attributes: map[string]datasourceschema.Attribute{
			"template_id": datasourceschema.StringAttribute{
				MarkdownDescription: "Template ID.",
				Required:            true,
			},
			"tags": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Tags assigned to the template.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: map[string]datasourceschema.Attribute{
						"tag": datasourceschema.StringAttribute{
							MarkdownDescription: "Template tag alias.",
							Computed:            true,
						},
						"build_id": datasourceschema.StringAttribute{
							MarkdownDescription: "Template build ID associated with the tag.",
							Computed:            true,
						},
						"created_at": datasourceschema.StringAttribute{
							MarkdownDescription: "Timestamp when the tag was created.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TemplateTagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TemplateTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TemplateTagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := d.client.listTemplateTags(ctx, data.TemplateID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "list template tags", err)
		return
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Tag < tags[j].Tag
	})

	data.Tags = make([]TemplateTagModel, 0, len(tags))
	for _, tag := range tags {
		data.Tags = append(data.Tags, TemplateTagModel{
			Tag:       types.StringValue(tag.Tag),
			BuildID:   types.StringValue(tag.BuildID),
			CreatedAt: types.StringValue(tag.CreatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func templateTagsDeleteName(data TemplateTagsResourceModel) string {
	if !data.Name.IsNull() && !data.Name.IsUnknown() && data.Name.ValueString() != "" {
		return data.Name.ValueString()
	}

	target := data.Target.ValueString()
	if index := strings.LastIndex(target, ":"); index > 0 {
		return target[:index]
	}

	return target
}

func stringDifference(left []string, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, value := range right {
		rightSet[value] = struct{}{}
	}

	result := make([]string, 0, len(left))
	for _, value := range left {
		if _, ok := rightSet[value]; !ok {
			result = append(result, value)
		}
	}
	sort.Strings(result)

	return result
}
