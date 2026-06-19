// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &LifecycleWebhookResource{}
var _ datasource.DataSource = &LifecycleWebhookDataSource{}
var _ datasource.DataSource = &LifecycleWebhooksDataSource{}

func NewLifecycleWebhookResource() resource.Resource {
	return &LifecycleWebhookResource{}
}

func NewLifecycleWebhookDataSource() datasource.DataSource {
	return &LifecycleWebhookDataSource{}
}

func NewLifecycleWebhooksDataSource() datasource.DataSource {
	return &LifecycleWebhooksDataSource{}
}

type LifecycleWebhookResource struct {
	client *e2bClient
}

type LifecycleWebhookResourceModel struct {
	ID              types.String `tfsdk:"id"`
	TeamID          types.String `tfsdk:"team_id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Events          types.Set    `tfsdk:"events"`
	SignatureSecret types.String `tfsdk:"signature_secret"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

type LifecycleWebhookDataSource struct {
	client *e2bClient
}

type LifecycleWebhookDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	TeamID    types.String `tfsdk:"team_id"`
	Name      types.String `tfsdk:"name"`
	URL       types.String `tfsdk:"url"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Events    types.Set    `tfsdk:"events"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type LifecycleWebhooksDataSource struct {
	client *e2bClient
}

type LifecycleWebhooksDataSourceModel struct {
	Webhooks []LifecycleWebhookDataSourceModel `tfsdk:"webhooks"`
}

func (r *LifecycleWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_webhook"
}

func (r *LifecycleWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Manages an E2B sandbox lifecycle webhook.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "Lifecycle webhook ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": resourceschema.StringAttribute{
				MarkdownDescription: "E2B team ID that owns the webhook.",
				Computed:            true,
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "Lifecycle webhook name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": resourceschema.StringAttribute{
				MarkdownDescription: "HTTPS endpoint that receives lifecycle events.",
				Required:            true,
			},
			"enabled": resourceschema.BoolAttribute{
				MarkdownDescription: "Whether E2B should deliver events to the webhook.",
				Optional:            true,
				Computed:            true,
			},
			"events": resourceschema.SetAttribute{
				MarkdownDescription: "Lifecycle event names delivered to the webhook.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"signature_secret": resourceschema.StringAttribute{
				MarkdownDescription: "Secret E2B uses to sign webhook payloads.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Timestamp when the webhook was created.",
				Computed:            true,
			},
		},
	}
}

func (r *LifecycleWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LifecycleWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LifecycleWebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	events, diags := stringSetFromTerraform(ctx, data.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.createLifecycleWebhook(ctx, lifecycleWebhookRequest{
		Name:            data.Name.ValueString(),
		URL:             data.URL.ValueString(),
		Enabled:         optionalBoolPointer(data.Enabled),
		Events:          events,
		SignatureSecret: data.SignatureSecret.ValueString(),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "create lifecycle webhook", err)
		return
	}

	existingSecret := data.SignatureSecret
	resp.Diagnostics.Append(data.applyLifecycleWebhook(ctx, created)...)
	data.SignatureSecret = existingSecret
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LifecycleWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LifecycleWebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.getLifecycleWebhook(ctx, data.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "read lifecycle webhook", err)
		return
	}

	existingSecret := data.SignatureSecret
	resp.Diagnostics.Append(data.applyLifecycleWebhook(ctx, webhook)...)
	data.SignatureSecret = existingSecret
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LifecycleWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LifecycleWebhookResourceModel
	var state LifecycleWebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	events, diags := stringSetFromTerraform(ctx, plan.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.updateLifecycleWebhook(ctx, state.ID.ValueString(), lifecycleWebhookRequest{
		URL:     plan.URL.ValueString(),
		Enabled: optionalBoolPointer(plan.Enabled),
		Events:  events,
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "update lifecycle webhook", err)
		return
	}

	secret := state.SignatureSecret
	resp.Diagnostics.Append(plan.applyLifecycleWebhook(ctx, updated)...)
	plan.SignatureSecret = secret
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LifecycleWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LifecycleWebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteLifecycleWebhook(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete lifecycle webhook", err)
	}
}

func (d *LifecycleWebhookDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_webhook"
}

func (d *LifecycleWebhookDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Reads an E2B sandbox lifecycle webhook by ID.",
		Attributes:          lifecycleWebhookDataSourceAttributes(true),
	}
}

func (d *LifecycleWebhookDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LifecycleWebhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LifecycleWebhookDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := d.client.getLifecycleWebhook(ctx, data.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "read lifecycle webhook", err)
		return
	}

	resp.Diagnostics.Append(data.applyLifecycleWebhook(ctx, webhook)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *LifecycleWebhooksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_webhooks"
}

func (d *LifecycleWebhooksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists E2B sandbox lifecycle webhooks.",
		Attributes: map[string]datasourceschema.Attribute{
			"webhooks": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "Lifecycle webhooks.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: lifecycleWebhookDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *LifecycleWebhooksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LifecycleWebhooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LifecycleWebhooksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhooks, err := d.client.listLifecycleWebhooks(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list lifecycle webhooks", err)
		return
	}

	data.Webhooks = make([]LifecycleWebhookDataSourceModel, 0, len(webhooks))
	for _, webhook := range webhooks {
		model := LifecycleWebhookDataSourceModel{}
		resp.Diagnostics.Append(model.applyLifecycleWebhook(ctx, &webhook)...)
		data.Webhooks = append(data.Webhooks, model)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *LifecycleWebhookResourceModel) applyLifecycleWebhook(ctx context.Context, webhook *lifecycleWebhookResponse) diag.Diagnostics {
	events, diags := setStringValue(ctx, webhook.Events)

	m.ID = types.StringValue(webhook.ID)
	m.TeamID = types.StringValue(webhook.TeamID)
	m.Name = types.StringValue(webhook.Name)
	m.URL = types.StringValue(webhook.URL)
	m.Enabled = types.BoolValue(webhook.Enabled)
	m.Events = events
	m.CreatedAt = types.StringValue(webhook.CreatedAt)

	return diags
}

func (m *LifecycleWebhookDataSourceModel) applyLifecycleWebhook(ctx context.Context, webhook *lifecycleWebhookResponse) diag.Diagnostics {
	events, diags := setStringValue(ctx, webhook.Events)

	m.ID = types.StringValue(webhook.ID)
	m.TeamID = types.StringValue(webhook.TeamID)
	m.Name = types.StringValue(webhook.Name)
	m.URL = types.StringValue(webhook.URL)
	m.Enabled = types.BoolValue(webhook.Enabled)
	m.Events = events
	m.CreatedAt = types.StringValue(webhook.CreatedAt)

	return diags
}

func lifecycleWebhookDataSourceAttributes(lookup bool) map[string]datasourceschema.Attribute {
	idAttribute := datasourceschema.StringAttribute{
		MarkdownDescription: "Lifecycle webhook ID.",
		Computed:            true,
	}
	if lookup {
		idAttribute = datasourceschema.StringAttribute{
			MarkdownDescription: "Lifecycle webhook ID.",
			Required:            true,
		}
	}

	return map[string]datasourceschema.Attribute{
		"id": idAttribute,
		"team_id": datasourceschema.StringAttribute{
			MarkdownDescription: "E2B team ID that owns the webhook.",
			Computed:            true,
		},
		"name": datasourceschema.StringAttribute{
			MarkdownDescription: "Lifecycle webhook name.",
			Computed:            true,
		},
		"url": datasourceschema.StringAttribute{
			MarkdownDescription: "Webhook endpoint URL.",
			Computed:            true,
		},
		"enabled": datasourceschema.BoolAttribute{
			MarkdownDescription: "Whether the webhook is enabled.",
			Computed:            true,
		},
		"events": datasourceschema.SetAttribute{
			MarkdownDescription: "Lifecycle event names delivered to the webhook.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"created_at": datasourceschema.StringAttribute{
			MarkdownDescription: "Timestamp when the webhook was created.",
			Computed:            true,
		},
	}
}
