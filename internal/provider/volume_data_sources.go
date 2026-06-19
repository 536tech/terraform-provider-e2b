// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VolumeDataSource{}
var _ datasource.DataSource = &VolumesDataSource{}

func NewVolumeDataSource() datasource.DataSource {
	return &VolumeDataSource{}
}

func NewVolumesDataSource() datasource.DataSource {
	return &VolumesDataSource{}
}

type VolumeDataSource struct {
	client *e2bClient
}

type VolumeDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type VolumesDataSource struct {
	client *e2bClient
}

type VolumesDataSourceModel struct {
	Volumes []VolumeDataSourceModel `tfsdk:"volumes"`
}

func (d *VolumeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (d *VolumeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an E2B volume by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Volume ID.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Volume name.",
				Computed:            true,
			},
		},
	}
}

func (d *VolumeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VolumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VolumeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := d.client.getVolume(ctx, data.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "read volume", err)
		return
	}

	data.ID = types.StringValue(volume.VolumeID)
	data.Name = types.StringValue(volume.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *VolumesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *VolumesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists E2B volumes.",
		Attributes: map[string]schema.Attribute{
			"volumes": schema.ListNestedAttribute{
				MarkdownDescription: "E2B volumes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Volume ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Volume name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *VolumesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VolumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VolumesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumes, err := d.client.listVolumes(ctx)
	if err != nil {
		addClientError(&resp.Diagnostics, "list volumes", err)
		return
	}

	data.Volumes = make([]VolumeDataSourceModel, 0, len(volumes))
	for _, volume := range volumes {
		data.Volumes = append(data.Volumes, VolumeDataSourceModel{
			ID:   types.StringValue(volume.VolumeID),
			Name: types.StringValue(volume.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
