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
	APIKey      types.String `tfsdk:"api_key"`
	AccessToken types.String `tfsdk:"access_token"`
	TeamID      types.String `tfsdk:"team_id"`
	APIURL      types.String `tfsdk:"api_url"`
}

func (p *E2BProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "e2b"
	resp.Version = p.version
}

func (p *E2BProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for managing E2B sandboxes, volumes, templates, API keys, lifecycle webhooks, and governance evidence data sources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "E2B API key. May also be set with the `E2B_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "E2B access token for bearer-authenticated team and API key routes. May also be set with the `E2B_ACCESS_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "E2B team ID to send as `X-Team-ID` for bearer-authenticated team routes. May also be set with the `E2B_TEAM_ID` environment variable.",
				Optional:            true,
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

	accessToken := strings.TrimSpace(os.Getenv("E2B_ACCESS_TOKEN"))
	if !data.AccessToken.IsNull() {
		accessToken = strings.TrimSpace(data.AccessToken.ValueString())
	}

	teamID := strings.TrimSpace(os.Getenv("E2B_TEAM_ID"))
	if !data.TeamID.IsNull() {
		teamID = strings.TrimSpace(data.TeamID.ValueString())
	}

	if apiKey == "" && accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing E2B authentication token",
			"Set api_key/E2B_API_KEY for normal E2B API routes or access_token/E2B_ACCESS_TOKEN for bearer-authenticated team routes.",
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

	client, err := newE2BClientWithConfig(apiURL, e2bClientConfig{
		APIKey:      apiKey,
		AccessToken: accessToken,
		TeamID:      teamID,
		Version:     p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Invalid E2B client configuration", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *E2BProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessTokenResource,
		NewAPIKeyResource,
		NewLifecycleWebhookResource,
		NewSandboxResource,
		NewSnapshotResource,
		NewTemplateResource,
		NewTemplateTagsResource,
		NewVolumeResource,
	}
}

func (p *E2BProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIKeyDataSource,
		NewAPIKeysDataSource,
		NewLifecycleEventsDataSource,
		NewLifecycleWebhookDataSource,
		NewLifecycleWebhooksDataSource,
		NewSandboxDataSource,
		NewSandboxesDataSource,
		NewSnapshotsDataSource,
		NewTemplateDataSource,
		NewTemplateTagsDataSource,
		NewTemplatesDataSource,
		NewTeamMetricsDataSource,
		NewTeamMetricMaxDataSource,
		NewTeamsDataSource,
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
