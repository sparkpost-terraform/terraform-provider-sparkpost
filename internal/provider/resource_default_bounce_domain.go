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
)

type defaultBounceDomainResource struct {
	client *SparkPostClient
}

func NewDefaultBounceDomainResource() resource.Resource {
	return &defaultBounceDomainResource{}
}

func (r *defaultBounceDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_bounce_domain"
}

func (r *defaultBounceDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Makes a bounce domain the default bounce domain for the account. SparkPost only accepts this once the bounce CNAME is verified, so order it after sparkpost_domain_bounce_verification. Destroying it unsets the default.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The bounce domain to make the default",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subaccount": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional subaccount ID that contains the domain",
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

func (r *defaultBounceDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *defaultBounceDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainVerificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetDefaultBounceDomain(plan.Domain.ValueString(), int(plan.Subaccount.ValueInt64()), true); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	plan.Id = plan.Domain
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *defaultBounceDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainVerificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.GetDomain(state.Id.ValueString(), int(state.Subaccount.ValueInt64()))
	if err != nil {
		if errors.Is(err, ErrDomainNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	// Another domain took over as the default, or it was unset outside Terraform.
	if !d.DefaultBounceDomain {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *defaultBounceDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op: all attributes require replacement, so Terraform never calls Update.
}

func (r *defaultBounceDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainVerificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetDefaultBounceDomain(state.Id.ValueString(), int(state.Subaccount.ValueInt64()), false)
	if err != nil && !errors.Is(err, ErrDomainNotFound) {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
