// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AccessTokenResource{}

func NewAccessTokenResource() resource.Resource {
	return &AccessTokenResource{}
}

type AccessTokenResource struct {
	client *e2bClient
}

type AccessTokenResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Token             types.String `tfsdk:"token"`
	MaskPrefix        types.String `tfsdk:"mask_prefix"`
	MaskValueLength   types.Int64  `tfsdk:"mask_value_length"`
	MaskedValuePrefix types.String `tfsdk:"masked_value_prefix"`
	MaskedValueSuffix types.String `tfsdk:"masked_value_suffix"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (r *AccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_token"
}

func (r *AccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates and deletes an E2B access token. E2B does not expose read/list/update for access tokens, so Terraform preserves state after creation and can delete by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Access token ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Access token name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Raw access token value. E2B only returns this value at creation time.",
				Computed:            true,
				Sensitive:           true,
			},
			"mask_prefix": schema.StringAttribute{
				MarkdownDescription: "Token type prefix in E2B's masked token metadata.",
				Computed:            true,
			},
			"mask_value_length": schema.Int64Attribute{
				MarkdownDescription: "Original token length in E2B's masked token metadata.",
				Computed:            true,
			},
			"masked_value_prefix": schema.StringAttribute{
				MarkdownDescription: "Visible prefix in E2B's masked token metadata.",
				Computed:            true,
			},
			"masked_value_suffix": schema.StringAttribute{
				MarkdownDescription: "Visible suffix in E2B's masked token metadata.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the access token was created.",
				Computed:            true,
			},
		},
	}
}

func (r *AccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.createAccessToken(ctx, data.Name.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "create access token", err)
		return
	}

	data.applyAccessToken(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Access token updates require replacement",
		"E2B access token names cannot be updated through the public API. Terraform should replace the token instead of updating it in place.",
	)
}

func (r *AccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteAccessToken(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete access token", err)
	}
}

func (m *AccessTokenResourceModel) applyAccessToken(accessToken *accessTokenResponse) {
	m.ID = types.StringValue(accessToken.ID)
	m.Name = types.StringValue(accessToken.Name)
	m.Token = types.StringValue(accessToken.Token)
	if accessToken.Mask != nil {
		m.MaskPrefix = types.StringValue(accessToken.Mask.Prefix)
		m.MaskValueLength = types.Int64Value(accessToken.Mask.ValueLength)
		m.MaskedValuePrefix = types.StringValue(accessToken.Mask.MaskedValuePrefix)
		m.MaskedValueSuffix = types.StringValue(accessToken.Mask.MaskedValueSuffix)
	} else {
		m.MaskPrefix = types.StringNull()
		m.MaskValueLength = types.Int64Null()
		m.MaskedValuePrefix = types.StringNull()
		m.MaskedValueSuffix = types.StringNull()
	}
	m.CreatedAt = types.StringValue(accessToken.CreatedAt)
}
