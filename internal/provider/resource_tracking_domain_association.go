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
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type trackingDomainAssociationResource struct {
	client *SparkPostClient
}

func NewTrackingDomainAssociationResource() resource.Resource {
	return &trackingDomainAssociationResource{}
}

type trackingDomainAssociationResourceModel struct {
	Domain         types.String `tfsdk:"domain"`
	TrackingDomain types.String  `tfsdk:"tracking_domain"`
	Subaccount     types.Int64  `tfsdk:"subaccount"`
	Id             types.String `tfsdk:"id"`
}

func (r *trackingDomainAssociationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tracking_domain_association"
}

func (r *trackingDomainAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain to be associated",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tracking_domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tracking domain to be associated",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},        
			"subaccount": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional subnet account ID that contains the domain",
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

func (r *trackingDomainAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trackingDomainAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trackingDomainAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()
	trackingDomain := plan.TrackingDomain.ValueString()

	err := r.client.AssociateTrackingDomain(domain, subaccount, trackingDomain)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.Id = plan.Domain

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trackingDomainAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()
	trackingDomain := state.TrackingDomain.ValueString()

	actualTrackingDomain, err := r.client.GetTrackingDomainAssociation(domain, subaccount, trackingDomain)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

    if actualTrackingDomain != trackingDomain {
    	resp.Diagnostics.AddWarning(
    		"Tracking Domain Mismatch",
    		fmt.Sprintf("The current tracking domain '%s' does not match the configured value '%s'. "+
    			"This may indicate it was edited outside of Terraform.", actualTrackingDomain, trackingDomain),
    	)
    }

    diags = resp.State.SetAttribute(ctx, path.Root("tracking_domain"), types.StringValue(actualTrackingDomain))
    resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op: all changes require replacement
}

func (r *trackingDomainAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trackingDomainAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

    subaccount := int(state.Subaccount.ValueInt64())
    domain := state.Domain.ValueString()
    
    err := r.client.AssociateTrackingDomain(domain, subaccount, "")
    if err != nil {
    	resp.Diagnostics.AddError("Delete Error", err.Error())
    	return
    }    

	resp.State.RemoveResource(ctx)
}
