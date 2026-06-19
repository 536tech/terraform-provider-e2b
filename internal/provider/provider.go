// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultAPIURL = "https://api.e2b.app"

var _ provider.Provider = &E2BProvider{}

// E2BProvider defines the provider implementation.
type E2BProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// E2BProviderModel describes the provider configuration.
type E2BProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}

func (p *E2BProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "e2b"
	resp.Version = p.version
}

func (p *E2BProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing E2B sandboxes, volumes, and template metadata.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "E2B API key. May also be set with the `E2B_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("E2B Platform API base URL. May also be set with `E2B_API_URL`. Defaults to `%s`.", defaultAPIURL),
				Optional:            true,
			},
		},
	}
}

func (p *E2BProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data E2BProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := strings.TrimSpace(os.Getenv("E2B_API_KEY"))
	if !data.APIKey.IsNull() {
		apiKey = strings.TrimSpace(data.APIKey.ValueString())
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing E2B API key",
			"Set api_key in the provider configuration or set the E2B_API_KEY environment variable.",
		)
		return
	}

	apiURL := strings.TrimSpace(os.Getenv("E2B_API_URL"))
	if !data.APIURL.IsNull() {
		apiURL = strings.TrimSpace(data.APIURL.ValueString())
	}

	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	client, err := newE2BClient(apiURL, apiKey, p.version)
	if err != nil {
		resp.Diagnostics.AddError("Invalid E2B client configuration", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *E2BProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSandboxResource,
		NewTemplateResource,
		NewVolumeResource,
	}
}

func (p *E2BProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSandboxDataSource,
		NewSandboxesDataSource,
		NewTemplateDataSource,
		NewTemplatesDataSource,
		NewVolumeDataSource,
		NewVolumesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &E2BProvider{
			version: version,
		}
	}
}
