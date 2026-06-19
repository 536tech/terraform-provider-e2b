// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithImportState = &APIKeyResource{}
var _ datasource.DataSource = &APIKeyDataSource{}
var _ datasource.DataSource = &APIKeysDataSource{}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{}
}

func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{}
}

type APIKeyResource struct {
	client *e2bClient
}

type APIKeyResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Key               types.String `tfsdk:"key"`
	MaskPrefix        types.String `tfsdk:"mask_prefix"`
	MaskValueLength   types.Int64  `tfsdk:"mask_value_length"`
	MaskedValuePrefix types.String `tfsdk:"masked_value_prefix"`
	MaskedValueSuffix types.String `tfsdk:"masked_value_suffix"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByID       types.String `tfsdk:"created_by_id"`
	CreatedByEmail    types.String `tfsdk:"created_by_email"`
	LastUsed          types.String `tfsdk:"last_used"`
}

type APIKeyDataSource struct {
	client *e2bClient
}

type APIKeyDataSourceModel APIKeyResourceModel

type APIKeysDataSource struct {
	client *e2bClient
}

type APIKeysDataSourceModel struct {
	APIKeys []APIKeyDataSourceModel `tfsdk:"api_keys"`
}

func (r *APIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "Manages an E2B team API key. This route requires `access_token` and `team_id` provider configuration.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				MarkdownDescription: "API key ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": resourceschema.StringAttribute{
				MarkdownDescription: "API key name.",
				Required:            true,
			},
			"key": resourceschema.StringAttribute{
				MarkdownDescription: "Raw API key value. E2B only returns this value at creation time.",
				Computed:            true,
				Sensitive:           true,
			},
			"mask_prefix": resourceschema.StringAttribute{
				MarkdownDescription: "Token type prefix in E2B's masked key metadata.",
				Computed:            true,
			},
			"mask_value_length": resourceschema.Int64Attribute{
				MarkdownDescription: "Original token length in E2B's masked key metadata.",
				Computed:            true,
			},
			"masked_value_prefix": resourceschema.StringAttribute{
				MarkdownDescription: "Visible prefix in E2B's masked key metadata.",
				Computed:            true,
			},
			"masked_value_suffix": resourceschema.StringAttribute{
				MarkdownDescription: "Visible suffix in E2B's masked key metadata.",
				Computed:            true,
			},
			"created_at": resourceschema.StringAttribute{
				MarkdownDescription: "Timestamp when the API key was created.",
				Computed:            true,
			},
			"created_by_id": resourceschema.StringAttribute{
				MarkdownDescription: "User ID that created the API key, when returned by E2B.",
				Computed:            true,
			},
			"created_by_email": resourceschema.StringAttribute{
				MarkdownDescription: "Email address that created the API key, when returned by E2B.",
				Computed:            true,
			},
			"last_used": resourceschema.StringAttribute{
				MarkdownDescription: "Timestamp when the API key was last used, when returned by E2B.",
				Computed:            true,
			},
		},
	}
}

func (r *APIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.createAPIKey(ctx, data.Name.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "create API key", err)
		return
	}

	data.applyAPIKey(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.readAPIKeyByID(ctx, data.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "read API key", err)
		return
	}
	if apiKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	existingKey := data.Key
	data.applyAPIKey(apiKey)
	if apiKey.Key == "" {
		data.Key = existingKey
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan APIKeyResourceModel
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		if err := r.client.updateAPIKey(ctx, state.ID.ValueString(), plan.Name.ValueString()); err != nil {
			addClientError(&resp.Diagnostics, "update API key", err)
			return
		}
	}

	plan.ID = state.ID
	plan.Key = state.Key
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteAPIKey(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete API key", err)
	}
}

func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *APIKeyResource) readAPIKeyByID(ctx context.Context, id string) (*apiKeyResponse, error) {
	apiKeys, err := r.client.listAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, apiKey := range apiKeys {
		if apiKey.ID == id {
			return &apiKey, nil
		}
	}

	return nil, nil
}

func (d *APIKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *APIKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Reads an E2B team API key by ID or name. This route requires `access_token` and `team_id` provider configuration.",
		Attributes:          apiKeyDataSourceAttributes(true),
	}
}

func (d *APIKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Missing API key lookup", "Set either id or name.")
		return
	}

	apiKeys, err := d.client.listAPIKeys(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list API keys", err)
		return
	}

	var match *apiKeyResponse
	for _, apiKey := range apiKeys {
		if (!data.ID.IsNull() && apiKey.ID == data.ID.ValueString()) || (!data.Name.IsNull() && apiKey.Name == data.Name.ValueString()) {
			match = &apiKey
			break
		}
	}
	if match == nil {
		resp.Diagnostics.AddError("API key not found", "No E2B API key matched the configured id or name.")
		return
	}

	model := APIKeyResourceModel(data)
	model.applyAPIKey(match)
	state := APIKeyDataSourceModel(model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *APIKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *APIKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists E2B team API keys. This route requires `access_token` and `team_id` provider configuration.",
		Attributes: map[string]datasourceschema.Attribute{
			"api_keys": datasourceschema.ListNestedAttribute{
				MarkdownDescription: "E2B API keys.",
				Computed:            true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: apiKeyDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *APIKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKeys, err := d.client.listAPIKeys(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list API keys", err)
		return
	}

	data.APIKeys = make([]APIKeyDataSourceModel, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		model := APIKeyResourceModel{}
		model.applyAPIKey(&apiKey)
		data.APIKeys = append(data.APIKeys, APIKeyDataSourceModel(model))
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (m *APIKeyResourceModel) applyAPIKey(apiKey *apiKeyResponse) {
	m.ID = types.StringValue(apiKey.ID)
	m.Name = types.StringValue(apiKey.Name)
	if apiKey.Key != "" {
		m.Key = types.StringValue(apiKey.Key)
	}
	applyMask(m, apiKey.Mask)
	m.CreatedAt = types.StringValue(apiKey.CreatedAt)
	if apiKey.CreatedBy != nil {
		m.CreatedByID = types.StringValue(apiKey.CreatedBy.ID)
		m.CreatedByEmail = types.StringValue(apiKey.CreatedBy.Email)
	} else {
		m.CreatedByID = types.StringNull()
		m.CreatedByEmail = types.StringNull()
	}
	m.LastUsed = stringPointerValue(apiKey.LastUsed)
}

func applyMask(m *APIKeyResourceModel, mask *maskingDetailsResponse) {
	if mask == nil {
		m.MaskPrefix = types.StringNull()
		m.MaskValueLength = types.Int64Null()
		m.MaskedValuePrefix = types.StringNull()
		m.MaskedValueSuffix = types.StringNull()
		return
	}

	m.MaskPrefix = types.StringValue(mask.Prefix)
	m.MaskValueLength = types.Int64Value(mask.ValueLength)
	m.MaskedValuePrefix = types.StringValue(mask.MaskedValuePrefix)
	m.MaskedValueSuffix = types.StringValue(mask.MaskedValueSuffix)
}

func apiKeyDataSourceAttributes(lookup bool) map[string]datasourceschema.Attribute {
	idAttribute := datasourceschema.StringAttribute{
		MarkdownDescription: "API key ID.",
		Computed:            true,
	}
	nameAttribute := datasourceschema.StringAttribute{
		MarkdownDescription: "API key name.",
		Computed:            true,
	}
	if lookup {
		idAttribute = datasourceschema.StringAttribute{
			MarkdownDescription: "API key ID. Set either id or name.",
			Optional:            true,
			Computed:            true,
		}
		nameAttribute = datasourceschema.StringAttribute{
			MarkdownDescription: "API key name. Set either id or name.",
			Optional:            true,
			Computed:            true,
		}
	}

	return map[string]datasourceschema.Attribute{
		"id":                  idAttribute,
		"name":                nameAttribute,
		"key":                 datasourceschema.StringAttribute{MarkdownDescription: "Raw API key value. E2B only returns this value at creation time, so data sources usually return null.", Computed: true, Sensitive: true},
		"mask_prefix":         datasourceschema.StringAttribute{MarkdownDescription: "Token type prefix in E2B's masked key metadata.", Computed: true},
		"mask_value_length":   datasourceschema.Int64Attribute{MarkdownDescription: "Original token length in E2B's masked key metadata.", Computed: true},
		"masked_value_prefix": datasourceschema.StringAttribute{MarkdownDescription: "Visible prefix in E2B's masked key metadata.", Computed: true},
		"masked_value_suffix": datasourceschema.StringAttribute{MarkdownDescription: "Visible suffix in E2B's masked key metadata.", Computed: true},
		"created_at":          datasourceschema.StringAttribute{MarkdownDescription: "Timestamp when the API key was created.", Computed: true},
		"created_by_id":       datasourceschema.StringAttribute{MarkdownDescription: "User ID that created the API key, when returned by E2B.", Computed: true},
		"created_by_email":    datasourceschema.StringAttribute{MarkdownDescription: "Email address that created the API key, when returned by E2B.", Computed: true},
		"last_used":           datasourceschema.StringAttribute{MarkdownDescription: "Timestamp when the API key was last used, when returned by E2B.", Computed: true},
	}
}
