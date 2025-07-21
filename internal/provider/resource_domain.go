package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainResource struct {
	client *SparkPostClient
}

func NewDomainResource() resource.Resource {
	return &domainResource{}
}

type domainResourceModel struct {
	Domain         types.String `tfsdk:"domain"`
	Subaccount     types.Int64  `tfsdk:"subaccount"`
	Id             types.String `tfsdk:"id"`
	Shared         types.Bool   `tfsdk:"shared_with_subaccounts"`
	DefaultBounce  types.Bool   `tfsdk:"default_bounce_domain"`
}

func (r *domainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain to be used for sending or bounces",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subaccount": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional subaccount ID for creating the tracking domain in",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"shared_with_subaccounts": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Optional to share the domain with all subaccounts. Cannot be used if a subaccount is set",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},  
			"default_bounce_domain": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Optional to set as default bounce domain for the account. Cannot be used if a subaccount is set",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},      
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain name used as the resource ID",
			},
		},
	}
}

func (r *domainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*SparkPostClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *SparkPostClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *domainResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var config domainResourceModel
    diags := req.Config.Get(ctx, &config)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    subaccountSet := !config.Subaccount.IsNull() && !config.Subaccount.IsUnknown()
    sharedSet := !config.Shared.IsNull() && !config.Shared.IsUnknown()
    defaultSet := !config.DefaultBounce.IsNull() && !config.DefaultBounce.IsUnknown()

    if subaccountSet && sharedSet {
        resp.Diagnostics.AddError(
            "Invalid Configuration",
            "The attributes 'subaccount' and 'shared_with_subaccounts' cannot both be set. Please specify only one.",
        )
    }

    if subaccountSet && defaultSet {
        resp.Diagnostics.AddError(
            "Invalid Configuration",
            "The attributes 'subaccount' and 'default_bounce_domain' cannot both be set. Please specify only one.",
        )
    }
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()
	shared := plan.Shared.ValueBool()
	defaultBounce := plan.DefaultBounce.ValueBool()

	err := r.client.CreateDomain(domain, subaccount, shared, defaultBounce)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.Id = plan.Domain

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	_, err := r.client.GetDomain(domain, subaccount)
	if err != nil {
		if err == DomainNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"This resource does not support update in-place. To change attributes, the resource must be recreated.",
	)
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	err := r.client.DeleteDomain(domain, subaccount)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
