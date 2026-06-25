// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SandboxResource{}
var _ resource.ResourceWithImportState = &SandboxResource{}

func NewSandboxResource() resource.Resource {
	return &SandboxResource{}
}

type SandboxResource struct {
	client *e2bClient
}

type SandboxResourceModel struct {
	ID                        types.String              `tfsdk:"id"`
	TemplateID                types.String              `tfsdk:"template_id"`
	Timeout                   types.Int64               `tfsdk:"timeout"`
	AutoPause                 types.Bool                `tfsdk:"auto_pause"`
	AutoResume                types.Bool                `tfsdk:"auto_resume"`
	Secure                    types.Bool                `tfsdk:"secure"`
	AllowInternetAccess       types.Bool                `tfsdk:"allow_internet_access"`
	NetworkAllowPublicTraffic types.Bool                `tfsdk:"network_allow_public_traffic"`
	NetworkAllowOut           types.Set                 `tfsdk:"network_allow_out"`
	NetworkDenyOut            types.Set                 `tfsdk:"network_deny_out"`
	NetworkEgressProxy        *SandboxEgressProxyModel  `tfsdk:"network_egress_proxy"`
	NetworkMaskRequestHost    types.String              `tfsdk:"network_mask_request_host"`
	NetworkRules              types.Map                 `tfsdk:"network_rules"`
	VolumeMounts              []SandboxVolumeMountModel `tfsdk:"volume_mounts"`
	Metadata                  types.Map                 `tfsdk:"metadata"`
	EnvVars                   types.Map                 `tfsdk:"env_vars"`
	Alias                     types.String              `tfsdk:"alias"`
	ClientID                  types.String              `tfsdk:"client_id"`
	EnvdVersion               types.String              `tfsdk:"envd_version"`
	EnvdAccessToken           types.String              `tfsdk:"envd_access_token"`
	TrafficAccessToken        types.String              `tfsdk:"traffic_access_token"`
	State                     types.String              `tfsdk:"state"`
	StartedAt                 types.String              `tfsdk:"started_at"`
	EndAt                     types.String              `tfsdk:"end_at"`
	LifecycleAutoResume       types.Bool                `tfsdk:"lifecycle_auto_resume"`
	LifecycleOnTimeout        types.String              `tfsdk:"lifecycle_on_timeout"`
	CPUCount                  types.Int64               `tfsdk:"cpu_count"`
	MemoryMB                  types.Int64               `tfsdk:"memory_mb"`
	DiskSizeMB                types.Int64               `tfsdk:"disk_size_mb"`
}

type SandboxVolumeMountModel struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

type SandboxEgressProxyModel struct {
	Address  types.String `tfsdk:"address"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type SandboxNetworkRuleModel struct {
	Headers types.Map `tfsdk:"headers"`
}

func (r *SandboxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox"
}

func (r *SandboxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a running E2B sandbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Sandbox ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Template ID, alias, or name used to create the sandbox. For example, `base`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Sandbox time to live in seconds.",
				Optional:            true,
			},
			"auto_pause": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox should pause instead of being killed when the timeout is reached.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"auto_resume": schema.BoolAttribute{
				MarkdownDescription: "Whether E2B should automatically resume a paused sandbox on request.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"secure": schema.BoolAttribute{
				MarkdownDescription: "Whether E2B should secure system communication with the sandbox and return access tokens.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"allow_internet_access": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox can access the internet.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"network_allow_public_traffic": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox may receive public traffic.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"network_allow_out": schema.SetAttribute{
				MarkdownDescription: "Destinations that sandbox egress traffic is allowed to reach. Entries can be CIDR blocks, IP addresses, or domain names. When allowing domains, E2B requires `network_deny_out` to include `ALL_TRAFFIC`.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"network_deny_out": schema.SetAttribute{
				MarkdownDescription: "CIDR blocks or IP addresses that sandbox egress traffic is denied from reaching. Use `ALL_TRAFFIC` when pairing domain allow rules with a default-deny policy.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"network_egress_proxy": schema.SingleNestedAttribute{
				MarkdownDescription: "SOCKS5 proxy for sandbox egress traffic, applied after allow and deny filtering.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "SOCKS5 proxy address in `host:port` format.",
						Required:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "Optional SOCKS5 username.",
						Optional:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "Optional SOCKS5 password.",
						Optional:            true,
						Sensitive:           true,
					},
				},
			},
			"network_mask_request_host": schema.StringAttribute{
				MarkdownDescription: "Host value E2B should mask on incoming sandbox requests.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_rules": schema.MapAttribute{
				MarkdownDescription: "Per-domain egress HTTP/HTTPS transform rules. Map keys are domains, and each value is a list of rules that may inject or override request headers. Domains listed here still need to be allowed by `network_allow_out`.",
				ElementType:         sandboxNetworkRuleListType(),
				Optional:            true,
			},
			"volume_mounts": schema.ListNestedAttribute{
				MarkdownDescription: "Volumes to mount into the sandbox at creation time.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "E2B volume name.",
							Required:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path where the volume is mounted in the sandbox.",
							Required:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Metadata assigned to the sandbox.",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"env_vars": schema.MapAttribute{
				MarkdownDescription: "Environment variables injected into the sandbox.",
				ElementType:         types.StringType,
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias of the template used to create the sandbox.",
				Computed:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "E2B client ID for the sandbox.",
				Computed:            true,
			},
			"envd_version": schema.StringAttribute{
				MarkdownDescription: "Version of envd running in the sandbox.",
				Computed:            true,
			},
			"envd_access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authenticated envd requests when secure sandbox mode is enabled.",
				Computed:            true,
				Sensitive:           true,
			},
			"traffic_access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authenticated sandbox traffic when secure sandbox mode is enabled.",
				Computed:            true,
				Sensitive:           true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current sandbox state.",
				Computed:            true,
			},
			"started_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the sandbox started.",
				Computed:            true,
			},
			"end_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the sandbox will expire.",
				Computed:            true,
			},
			"lifecycle_auto_resume": schema.BoolAttribute{
				MarkdownDescription: "Whether the sandbox is configured to auto-resume, as reported by E2B.",
				Computed:            true,
			},
			"lifecycle_on_timeout": schema.StringAttribute{
				MarkdownDescription: "Lifecycle action E2B applies when the sandbox timeout is reached.",
				Computed:            true,
			},
			"cpu_count": schema.Int64Attribute{
				MarkdownDescription: "Number of CPU cores allocated to the sandbox.",
				Computed:            true,
			},
			"memory_mb": schema.Int64Attribute{
				MarkdownDescription: "Memory allocated to the sandbox, in MB.",
				Computed:            true,
			},
			"disk_size_mb": schema.Int64Attribute{
				MarkdownDescription: "Disk allocated to the sandbox, in MB.",
				Computed:            true,
			},
		},
	}
}

func (r *SandboxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SandboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, diags := mapStringFromTerraform(ctx, data.Metadata)
	resp.Diagnostics.Append(diags...)
	envVars, diags := mapStringFromTerraform(ctx, data.EnvVars)
	resp.Diagnostics.Append(diags...)
	network, diags := data.sandboxNetworkConfig(ctx, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.createSandbox(ctx, sandboxCreateRequest{
		TemplateID:           data.TemplateID.ValueString(),
		Timeout:              optionalInt64Pointer(data.Timeout),
		AutoPause:            optionalBoolPointer(data.AutoPause),
		AutoResume:           sandboxAutoResumeFromTerraform(data.AutoResume),
		Secure:               optionalBoolPointer(data.Secure),
		AllowInternetAccess:  optionalBoolPointer(data.AllowInternetAccess),
		Network:              network,
		Metadata:             metadata,
		EnvironmentVariables: envVars,
		VolumeMounts:         sandboxVolumeMountsFromTerraform(data.VolumeMounts),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "create sandbox", err)
		return
	}

	data.ID = types.StringValue(created.SandboxID)
	data.Alias = types.StringValue(created.Alias)
	data.ClientID = types.StringValue(created.ClientID)
	data.EnvdVersion = types.StringValue(created.EnvdVersion)
	data.EnvdAccessToken = stringPointerValue(created.EnvdAccessToken)
	data.TrafficAccessToken = stringPointerValue(created.TrafficAccessToken)

	if detail, err := r.client.getSandbox(ctx, created.SandboxID); err == nil {
		resp.Diagnostics.Append(data.applySandboxDetail(ctx, detail)...)
	} else if !isNotFound(err) {
		addClientError(&resp.Diagnostics, "read created sandbox", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SandboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.getSandbox(ctx, data.ID.ValueString())
	if isNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "read sandbox", err)
		return
	}

	resp.Diagnostics.Append(data.applySandboxDetail(ctx, detail)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SandboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SandboxResourceModel
	var state SandboxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sandboxID := state.ID.ValueString()
	plan.ID = state.ID

	if !plan.Timeout.Equal(state.Timeout) && !plan.Timeout.IsNull() && !plan.Timeout.IsUnknown() {
		if err := r.client.setSandboxTimeout(ctx, sandboxID, plan.Timeout.ValueInt64()); err != nil {
			addClientError(&resp.Diagnostics, "update sandbox timeout", err)
			return
		}
	}

	if sandboxNetworkNeedsUpdate(plan, state) {
		network, diags := plan.sandboxNetworkConfig(ctx, true)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if network == nil {
			network = &sandboxNetworkConfig{}
		}
		if !plan.AllowInternetAccess.IsNull() && !plan.AllowInternetAccess.IsUnknown() {
			network.AllowInternetAccess = optionalBoolPointer(plan.AllowInternetAccess)
		}

		if err := r.client.updateSandboxNetwork(ctx, sandboxID, *network); err != nil {
			addClientError(&resp.Diagnostics, "update sandbox network", err)
			return
		}
	}

	detail, err := r.client.getSandbox(ctx, sandboxID)
	if err != nil {
		addClientError(&resp.Diagnostics, "read updated sandbox", err)
		return
	}

	resp.Diagnostics.Append(plan.applySandboxDetail(ctx, detail)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SandboxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SandboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.deleteSandbox(ctx, data.ID.ValueString()); err != nil {
		addClientError(&resp.Diagnostics, "delete sandbox", err)
	}
}

func (r *SandboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *SandboxResourceModel) applySandboxDetail(ctx context.Context, detail *sandboxDetailResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(detail.SandboxID)
	if m.TemplateID.IsNull() || m.TemplateID.IsUnknown() || m.TemplateID.ValueString() == "" {
		m.TemplateID = types.StringValue(detail.TemplateID)
	}
	m.Alias = types.StringValue(detail.Alias)
	m.ClientID = types.StringValue(detail.ClientID)
	m.StartedAt = types.StringValue(detail.StartedAt)
	m.EndAt = types.StringValue(detail.EndAt)
	m.EnvdVersion = types.StringValue(detail.EnvdVersion)
	m.EnvdAccessToken = stringPointerValue(detail.EnvdAccessToken)
	m.TrafficAccessToken = stringPointerValue(detail.TrafficAccessToken)
	m.AllowInternetAccess = boolPointerValue(detail.AllowInternetAccess)
	m.CPUCount = int64PointerValue(detail.CPUCount)
	m.MemoryMB = int64PointerValue(detail.MemoryMB)
	m.DiskSizeMB = int64PointerValue(detail.DiskSizeMB)
	m.State = types.StringValue(detail.State)
	if detail.Lifecycle != nil {
		m.LifecycleAutoResume = types.BoolValue(detail.Lifecycle.AutoResume)
		m.LifecycleOnTimeout = types.StringValue(detail.Lifecycle.OnTimeout)
	} else {
		m.LifecycleAutoResume = types.BoolNull()
		m.LifecycleOnTimeout = types.StringNull()
	}
	if len(m.VolumeMounts) > 0 || len(detail.VolumeMounts) > 0 {
		m.VolumeMounts = flattenSandboxVolumeMounts(detail.VolumeMounts)
	}

	metadata, mapDiags := mapStringValue(ctx, detail.Metadata)
	diags.Append(mapDiags...)
	if !m.Metadata.IsNull() || len(detail.Metadata) > 0 {
		m.Metadata = metadata
	}

	if detail.Network == nil {
		if m.NetworkAllowPublicTraffic.IsUnknown() {
			m.NetworkAllowPublicTraffic = types.BoolNull()
		}
		return diags
	}

	if !m.NetworkAllowPublicTraffic.IsNull() || detail.Network.AllowPublicTraffic != nil {
		m.NetworkAllowPublicTraffic = boolPointerValue(detail.Network.AllowPublicTraffic)
	}
	if !m.NetworkMaskRequestHost.IsNull() || detail.Network.MaskRequestHost != "" {
		m.NetworkMaskRequestHost = types.StringValue(detail.Network.MaskRequestHost)
	}
	allowOut, setDiags := setStringValue(ctx, detail.Network.AllowOut)
	diags.Append(setDiags...)
	denyOut, setDiags := setStringValue(ctx, detail.Network.DenyOut)
	diags.Append(setDiags...)
	if !m.NetworkAllowOut.IsNull() || len(detail.Network.AllowOut) > 0 {
		m.NetworkAllowOut = allowOut
	}
	if !m.NetworkDenyOut.IsNull() || len(detail.Network.DenyOut) > 0 {
		m.NetworkDenyOut = denyOut
	}
	if m.NetworkEgressProxy != nil || detail.Network.EgressProxy != nil {
		m.NetworkEgressProxy = flattenSandboxEgressProxy(detail.Network.EgressProxy, m.NetworkEgressProxy)
	}
	if detail.Network.Rules != nil {
		rules, rulesDiags := sandboxNetworkRulesValue(ctx, detail.Network.Rules)
		diags.Append(rulesDiags...)
		m.NetworkRules = rules
	} else if m.NetworkRules.IsUnknown() {
		m.NetworkRules = types.MapNull(sandboxNetworkRuleListType())
	}

	return diags
}

func (m SandboxResourceModel) sandboxNetworkConfig(ctx context.Context, force bool) (*sandboxNetworkConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	config := &sandboxNetworkConfig{}
	hasNetwork := force

	if !m.NetworkAllowOut.IsNull() && !m.NetworkAllowOut.IsUnknown() {
		allowOut, setDiags := stringSetFromTerraform(ctx, m.NetworkAllowOut)
		diags.Append(setDiags...)
		config.AllowOut = allowOut
		hasNetwork = true
	}

	if !m.NetworkDenyOut.IsNull() && !m.NetworkDenyOut.IsUnknown() {
		denyOut, setDiags := stringSetFromTerraform(ctx, m.NetworkDenyOut)
		diags.Append(setDiags...)
		config.DenyOut = denyOut
		hasNetwork = true
	}

	if m.NetworkEgressProxy != nil {
		config.EgressProxy = sandboxEgressProxyFromTerraform(m.NetworkEgressProxy)
		hasNetwork = true
	}

	if !m.NetworkAllowPublicTraffic.IsNull() && !m.NetworkAllowPublicTraffic.IsUnknown() {
		config.AllowPublicTraffic = optionalBoolPointer(m.NetworkAllowPublicTraffic)
		hasNetwork = true
	}

	if !m.NetworkMaskRequestHost.IsNull() && !m.NetworkMaskRequestHost.IsUnknown() {
		config.MaskRequestHost = m.NetworkMaskRequestHost.ValueString()
		hasNetwork = true
	}

	if !m.NetworkRules.IsNull() && !m.NetworkRules.IsUnknown() {
		rules, ruleDiags := sandboxNetworkRulesFromTerraform(ctx, m.NetworkRules)
		diags.Append(ruleDiags...)
		config.Rules = rules
		hasNetwork = true
	}

	if !hasNetwork {
		return nil, diags
	}

	return config, diags
}

func sandboxNetworkNeedsUpdate(plan SandboxResourceModel, state SandboxResourceModel) bool {
	return !plan.AllowInternetAccess.Equal(state.AllowInternetAccess) ||
		!plan.NetworkAllowOut.Equal(state.NetworkAllowOut) ||
		!plan.NetworkDenyOut.Equal(state.NetworkDenyOut) ||
		!sandboxEgressProxyModelsEqual(plan.NetworkEgressProxy, state.NetworkEgressProxy) ||
		!plan.NetworkRules.Equal(state.NetworkRules)
}

func sandboxEgressProxyFromTerraform(model *SandboxEgressProxyModel) *sandboxEgressProxyConfig {
	if model == nil {
		return nil
	}

	config := &sandboxEgressProxyConfig{}
	if !model.Address.IsNull() && !model.Address.IsUnknown() {
		config.Address = model.Address.ValueString()
	}
	if !model.Username.IsNull() && !model.Username.IsUnknown() {
		config.Username = model.Username.ValueString()
	}
	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		config.Password = model.Password.ValueString()
	}

	return config
}

func flattenSandboxEgressProxy(config *sandboxEgressProxyConfig, prior *SandboxEgressProxyModel) *SandboxEgressProxyModel {
	if config == nil {
		return prior
	}

	result := &SandboxEgressProxyModel{
		Address:  types.StringValue(config.Address),
		Username: types.StringNull(),
		Password: types.StringNull(),
	}
	if config.Username != "" {
		result.Username = types.StringValue(config.Username)
	} else if prior != nil {
		result.Username = prior.Username
	}
	if config.Password != "" {
		result.Password = types.StringValue(config.Password)
	} else if prior != nil {
		result.Password = prior.Password
	}

	return result
}

func sandboxEgressProxyModelsEqual(a *SandboxEgressProxyModel, b *SandboxEgressProxyModel) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return a.Address.Equal(b.Address) && a.Username.Equal(b.Username) && a.Password.Equal(b.Password)
}

func sandboxNetworkRulesFromTerraform(ctx context.Context, value types.Map) (map[string][]sandboxNetworkRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	result := make(map[string][]sandboxNetworkRule, len(value.Elements()))
	for domain, element := range value.Elements() {
		list, ok := element.(types.List)
		if !ok {
			diags.AddError(
				"Unexpected Network Rule Type",
				fmt.Sprintf("Expected list value for network_rules[%q], got %T.", domain, element),
			)
			continue
		}

		var ruleModels []SandboxNetworkRuleModel
		diags.Append(list.ElementsAs(ctx, &ruleModels, false)...)
		if diags.HasError() {
			continue
		}

		rules := make([]sandboxNetworkRule, 0, len(ruleModels))
		for _, ruleModel := range ruleModels {
			headers, headerDiags := mapStringFromTerraform(ctx, ruleModel.Headers)
			diags.Append(headerDiags...)
			if headerDiags.HasError() {
				continue
			}
			rules = append(rules, sandboxNetworkRule{
				Transform: &sandboxNetworkTransform{
					Headers: headers,
				},
			})
		}
		result[domain] = rules
	}

	return result, diags
}

func sandboxNetworkRulesValue(ctx context.Context, rules map[string][]sandboxNetworkRule) (types.Map, diag.Diagnostics) {
	if len(rules) == 0 {
		return types.MapNull(sandboxNetworkRuleListType()), nil
	}

	model := make(map[string][]SandboxNetworkRuleModel, len(rules))
	var diags diag.Diagnostics
	for domain, domainRules := range rules {
		modelRules := make([]SandboxNetworkRuleModel, 0, len(domainRules))
		for _, rule := range domainRules {
			headers := map[string]string(nil)
			if rule.Transform != nil {
				headers = rule.Transform.Headers
			}
			headerValue, headerDiags := mapStringValue(ctx, headers)
			diags.Append(headerDiags...)
			modelRules = append(modelRules, SandboxNetworkRuleModel{
				Headers: headerValue,
			})
		}
		model[domain] = modelRules
	}
	if diags.HasError() {
		return types.MapUnknown(sandboxNetworkRuleListType()), diags
	}

	value, mapDiags := types.MapValueFrom(ctx, sandboxNetworkRuleListType(), model)
	diags.Append(mapDiags...)
	return value, diags
}

func sandboxNetworkRuleListType() attr.Type {
	return types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"headers": types.MapType{ElemType: types.StringType}}}}
}

func sandboxAutoResumeFromTerraform(value types.Bool) *sandboxAutoResume {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	return &sandboxAutoResume{
		Enabled: value.ValueBool(),
	}
}

func sandboxVolumeMountsFromTerraform(mounts []SandboxVolumeMountModel) []sandboxVolumeMount {
	if len(mounts) == 0 {
		return nil
	}

	result := make([]sandboxVolumeMount, 0, len(mounts))
	for _, mount := range mounts {
		result = append(result, sandboxVolumeMount{
			Name: mount.Name.ValueString(),
			Path: mount.Path.ValueString(),
		})
	}

	return result
}

func flattenSandboxVolumeMounts(mounts []sandboxVolumeMount) []SandboxVolumeMountModel {
	if len(mounts) == 0 {
		return nil
	}

	result := make([]SandboxVolumeMountModel, 0, len(mounts))
	for _, mount := range mounts {
		result = append(result, SandboxVolumeMountModel{
			Name: types.StringValue(mount.Name),
			Path: types.StringValue(mount.Path),
		})
	}

	return result
}
