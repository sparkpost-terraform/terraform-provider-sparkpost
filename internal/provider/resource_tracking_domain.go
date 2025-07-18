package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type trackingDomainResource struct {
	client *SparkPostClient
}

func NewTrackingDomainResource() resource.Resource {
	return &trackingDomainResource{}
}

type trackingDomainResourceModel struct {
	Domain     types.String `tfsdk:"domain"`
	HTTPS      types.Bool   `tfsdk:"https"`
	Subaccount types.Int64  `tfsdk:"subaccount"`
	Id         types.String `tfsdk:"id"`
}

func (r *trackingDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tracking_domain"
}

func (r *trackingDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain to be used for tracking links",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"https": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Specifies if the domain should use HTTPS",
				Computed:            false,
			},
			"subaccount": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional subnet account ID for creating the tracking domain in",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain name used as the resource ID",
			},
		},
	}
}

func (r *trackingDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*SparkPostClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *SparkPostClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *trackingDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trackingDomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()
	https := plan.HTTPS.ValueBool()

	err := r.client.CreateTrackingDomain(domain, https, subaccount)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.Id = plan.Domain

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trackingDomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	_, err := r.client.GetTrackingDomain(domain, subaccount)
	if err != nil {
		if err == TrackingDomainNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan trackingDomainResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    subaccount := int(plan.Subaccount.ValueInt64())
    domain := plan.Domain.ValueString()
    https := plan.HTTPS.ValueBool()

    err := r.client.UpdateTrackingDomain(domain, https, subaccount)
    if err != nil {
        resp.Diagnostics.AddError("Update Error", err.Error())
        return
    }

    plan.Id = types.StringValue(domain)
    diags = resp.State.Set(ctx, &plan)
    resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trackingDomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	err := r.client.DeleteTrackingDomain(domain, subaccount)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
