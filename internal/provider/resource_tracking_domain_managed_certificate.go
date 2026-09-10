package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type trackingDomainManagedCertificateResource struct {
	client *SparkPostClient
}

func NewTrackingDomainManagedCertificateResource() resource.Resource {
	return &trackingDomainManagedCertificateResource{}
}

type trackingDomainManagedCertificateResourceModel struct {
	Domain     types.String `tfsdk:"domain"`
	Subaccount types.Int64  `tfsdk:"subaccount"`
	Id         types.String `tfsdk:"id"`
}

func (r *trackingDomainManagedCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tracking_domain_managed_certificate"
}

func (r *trackingDomainManagedCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tracking domain to enable a SparkPost-managed TLS certificate for",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subaccount": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional subaccount ID that contains the tracking domain",
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

func (r *trackingDomainManagedCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trackingDomainManagedCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trackingDomainManagedCertificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()

	eligible, err := r.client.CheckTrackingDomainCertificateEligibility(domain, subaccount)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if !eligible {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("tracking domain %q is not eligible for a managed certificate", domain))
		return
	}

	if err := r.client.EnableTrackingDomainManagedCertificate(domain, subaccount); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.Id = plan.Domain

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainManagedCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trackingDomainManagedCertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	t, err := r.client.GetTrackingDomain(domain, subaccount)
	if err != nil {
		if errors.Is(err, ErrTrackingDomainNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	// The managed certificate can only be removed by deleting the tracking
	// domain outright, but check the flag anyway so drift (however it
	// happened) is detected rather than silently assumed still enabled.
	if !t.UsesManagedCertificate {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainManagedCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op: all changes require replacement
}

func (r *trackingDomainManagedCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trackingDomainManagedCertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
