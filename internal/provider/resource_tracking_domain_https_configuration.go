package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// trackingDomainHTTPSConfigurationResource controls whether a tracking
// domain serves over HTTPS, optionally via a SparkPost-managed TLS
// certificate. These are combined into one resource because enabling HTTPS
// on a domain with no certificate fails, and there's a strict order that has
// to happen inside a single apply when using a managed certificate: verify
// the domain (over plain HTTP, see CreateTrackingDomain) -> enable the
// managed certificate -> then, only once the certificate exists, turn HTTPS
// on. Set managed_certificate = false to skip that and just set https
// directly, e.g. for a domain that already has its own certificate.
type trackingDomainHTTPSConfigurationResource struct {
	client *SparkPostClient
}

func NewTrackingDomainHTTPSConfigurationResource() resource.Resource {
	return &trackingDomainHTTPSConfigurationResource{}
}

type trackingDomainHTTPSConfigurationResourceModel struct {
	Domain             types.String `tfsdk:"domain"`
	Subaccount         types.Int64  `tfsdk:"subaccount"`
	ManagedCertificate types.Bool   `tfsdk:"managed_certificate"`
	HTTPS              types.Bool   `tfsdk:"https"`
	Id                 types.String `tfsdk:"id"`
}

func (r *trackingDomainHTTPSConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tracking_domain_https_configuration"
}

func (r *trackingDomainHTTPSConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tracking domain to configure. Must already be verified (see sparkpost_tracking_domain_verification).",
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
			"managed_certificate": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to enable a SparkPost-managed TLS certificate for this domain before setting https. Defaults to false. SparkPost has no API to disable a managed certificate once enabled, so this can't be changed back to false afterwards.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"https": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the tracking domain should serve over HTTPS. Defaults to false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain name used as the resource ID",
			},
		},
	}
}

func (r *trackingDomainHTTPSConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trackingDomainHTTPSConfigurationResource) enableManagedCertificate(domain string, subaccount int) error {
	eligible, err := r.client.CheckTrackingDomainCertificateEligibility(domain, subaccount)
	if err != nil {
		return err
	}
	if !eligible {
		return fmt.Errorf("tracking domain %q is not eligible for a managed certificate", domain)
	}

	return r.client.EnableTrackingDomainManagedCertificate(domain, subaccount)
}

func (r *trackingDomainHTTPSConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trackingDomainHTTPSConfigurationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()

	managedCert := plan.ManagedCertificate.ValueBool()

	if managedCert {
		if err := r.enableManagedCertificate(domain, subaccount); err != nil {
			resp.Diagnostics.AddError("Create Error", err.Error())
			return
		}
	}

	https := plan.HTTPS.ValueBool()

	if err := r.client.UpdateTrackingDomain(domain, https, subaccount); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to set https: %s", err))
		return
	}

	plan.Id = plan.Domain
	plan.ManagedCertificate = types.BoolValue(managedCert)
	plan.HTTPS = types.BoolValue(https)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainHTTPSConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trackingDomainHTTPSConfigurationResourceModel
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

	state.HTTPS = types.BoolValue(t.HTTPS)
	state.ManagedCertificate = types.BoolValue(t.UsesManagedCertificate)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainHTTPSConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan trackingDomainHTTPSConfigurationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state trackingDomainHTTPSConfigurationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(plan.Subaccount.ValueInt64())
	domain := plan.Domain.ValueString()
	managedCert := plan.ManagedCertificate.ValueBool()

	if managedCert && !state.ManagedCertificate.ValueBool() {
		if err := r.enableManagedCertificate(domain, subaccount); err != nil {
			resp.Diagnostics.AddError("Update Error", err.Error())
			return
		}
	} else if !managedCert && state.ManagedCertificate.ValueBool() {
		resp.Diagnostics.AddError(
			"Update Error",
			"SparkPost has no API to disable a managed certificate once enabled; managed_certificate cannot be changed back to false",
		)
		return
	}

	https := plan.HTTPS.ValueBool()
	if err := r.client.UpdateTrackingDomain(domain, https, subaccount); err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	plan.Id = state.Id
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *trackingDomainHTTPSConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trackingDomainHTTPSConfigurationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := int(state.Subaccount.ValueInt64())
	domain := state.Id.ValueString()

	// There's no API to un-enable a managed certificate, only to delete the
	// tracking domain outright. Turn HTTPS back off - the one part of this
	// resource that can actually be reverted - and leave the certificate in
	// place. A 404 means the tracking domain is already gone, which isn't
	// this resource's problem to report.
	if err := r.client.UpdateTrackingDomain(domain, false, subaccount); err != nil && !errors.Is(err, ErrTrackingDomainNotFound) {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
