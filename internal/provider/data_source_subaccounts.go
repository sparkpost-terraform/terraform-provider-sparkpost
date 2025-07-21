package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var _ datasource.DataSource = &subaccountsDataSource{}

func NewSubAccountsDataSource() datasource.DataSource {
	return &subaccountsDataSource{}
}

type subaccountsDataSource struct {
	client *SparkPostClient
}

func (d *subaccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "sparkpost_subaccounts"
		resp.TypeName = req.ProviderTypeName + "_subaccounts"
}

func (d *subaccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
	    Attributes: map[string]schema.Attribute{
		    "subaccounts": schema.ListNestedAttribute{
			    Computed: true,
			    NestedObject: schema.NestedAttributeObject{
				    Attributes: map[string]schema.Attribute{
					    "id": schema.Int64Attribute{
						    Computed: true,
					    },
					    "name": schema.StringAttribute{
						    Computed: true,
					    },
				    },
			    },
		    },
	    },
    }
}

func (d *subaccountsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}


func (d *subaccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    subaccounts, err := d.client.ListSubaccounts()
    if err != nil {
        resp.Diagnostics.AddError("Failed to fetch subaccounts", fmt.Sprintf("Error: %s", err))
        return
    }

    var objs []attr.Value
    attrTypes := map[string]attr.Type{
        "id":   types.Int64Type,
        "name": types.StringType,
    }

    for _, sa := range subaccounts {
        obj, diag := types.ObjectValue(attrTypes, map[string]attr.Value{
            "id":   types.Int64Value(int64(sa.ID)),
            "name": types.StringValue(sa.Name),
        })
        if diag.HasError() {
            resp.Diagnostics.Append(diag...)
            return
        }
        objs = append(objs, obj)
    }

    listVal, diag := types.ListValue(types.ObjectType{
        AttrTypes: attrTypes, 
    }, objs)
    if diag.HasError() {
        resp.Diagnostics.Append(diag...)
        return
    }

    resp.State.SetAttribute(ctx, path.Root("subaccounts"), listVal)
}

