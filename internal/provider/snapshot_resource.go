// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SnapshotResource{}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

type SnapshotResource struct {
	client *e2bClient
}

type SnapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SandboxID types.String `tfsdk:"sandbox_id"`
	Name      types.String `tfsdk:"name"`
	Names     types.List   `tfsdk:"names"`
}

func (r *SnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a persistent E2B snapshot from a sandbox. E2B does not expose snapshot deletion, so destroying this resource removes it from Terraform state but leaves the remote snapshot intact.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Snapshot ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sandbox_id": schema.StringAttribute{
				MarkdownDescription: "Sandbox ID to snapshot.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Optional snapshot template name. If a snapshot template with this name already exists, E2B assigns a new build to the existing template.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"names": schema.ListAttribute{
				MarkdownDescription: "Full snapshot names returned by E2B.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (r *SnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name = data.Name.ValueString()
	}

	snapshot, err := r.client.createSnapshot(ctx, data.SandboxID.ValueString(), name)
	if err != nil {
		addClientError(&resp.Diagnostics, "create snapshot", err)
		return
	}

	resp.Diagnostics.Append(data.applySnapshot(ctx, snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshots, err := r.client.listSnapshots(ctx, data.SandboxID.ValueString(), 100, "")
	if err != nil {
		addClientError(&resp.Diagnostics, "read snapshot", err)
		return
	}

	for _, snapshot := range snapshots {
		if snapshot.SnapshotID == data.ID.ValueString() {
			resp.Diagnostics.Append(data.applySnapshot(ctx, &snapshot)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Snapshot updates require replacement",
		"E2B snapshots cannot be updated through the public API. Terraform should replace the snapshot resource instead of updating it in place.",
	)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
}

func (m *SnapshotResourceModel) applySnapshot(ctx context.Context, snapshot *snapshotResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(snapshot.SnapshotID)
	names, listDiags := listStringValue(ctx, snapshot.Names)
	diags.Append(listDiags...)
	m.Names = names

	return diags
}
