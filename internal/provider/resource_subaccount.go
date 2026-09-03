package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithImportState = &subaccountResource{}

type subaccountResource struct {
	client *SparkPostClient
}

func NewSubaccountResource() resource.Resource {
	return &subaccountResource{}
}

type subaccountResourceModel struct {
	Id     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	IPPool types.String `tfsdk:"ip_pool"`
}

func (r *subaccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (r *subaccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The subaccount ID assigned by SparkPost",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the subaccount",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "One of `active`, `suspended` or `terminated`. Defaults to `active` on creation. Terminating is permanent: SparkPost has no way to reactivate a terminated subaccount.",
			},
			"ip_pool": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the IP pool assigned to the subaccount",
			},
		},
	}
}

func (r *subaccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subaccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	ipPool := plan.IPPool.ValueString()

	id, err := r.client.CreateSubaccount(name, ipPool)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	// The create response doesn't carry status/ip_pool, so read the
	// subaccount back to populate the rest of the state.
	sa, err := r.client.GetSubaccount(id)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("subaccount %d was created but could not be read back: %s", id, err))
		return
	}

	// Status can't be set on create, so apply it as a follow-up update if requested.
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() && plan.Status.ValueString() != sa.Status {
		if err := r.client.UpdateSubaccount(id, name, plan.Status.ValueString(), ipPool); err != nil {
			resp.Diagnostics.AddError("Create Error", fmt.Sprintf("subaccount %d was created but failed to set initial status: %s", id, err))
			return
		}
		sa.Status = plan.Status.ValueString()
	}

	plan.Id = types.Int64Value(int64(sa.ID))
	plan.Status = types.StringValue(sa.Status)
	plan.IPPool = types.StringValue(sa.IPPool)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *subaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subaccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sa, err := r.client.GetSubaccount(int(state.Id.ValueInt64()))
	if err != nil {
		if errors.Is(err, ErrSubaccountNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	state.Name = types.StringValue(sa.Name)
	state.Status = types.StringValue(sa.Status)
	state.IPPool = types.StringValue(sa.IPPool)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *subaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subaccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state subaccountResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.Id.ValueInt64())
	err := r.client.UpdateSubaccount(id, plan.Name.ValueString(), plan.Status.ValueString(), plan.IPPool.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	plan.Id = state.Id
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *subaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected a numeric subaccount ID, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// Delete terminates the subaccount. SparkPost has no delete endpoint;
// "terminated" is a permanent, one-way status.
func (r *subaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subaccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.Id.ValueInt64())
	err := r.client.UpdateSubaccount(id, state.Name.ValueString(), "terminated", state.IPPool.ValueString())
	if err != nil {
		if errors.Is(err, ErrSubaccountNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
