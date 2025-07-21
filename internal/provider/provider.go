package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies the framework interfaces
var _ provider.Provider = &sparkpostProvider{}

func New() provider.Provider {
	return &sparkpostProvider{}
}

type sparkpostProvider struct {
	client *SparkPostClient
}

// Provider-level config model
type providerModel struct {
	APIUrl types.String `tfsdk:"api_url"`
	APIKey types.String `tfsdk:"api_key"`
}

func (p *sparkpostProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sparkpost"
}

func (p *sparkpostProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:            true,
				Description:         "API URL for SparkPost",
				MarkdownDescription: "API URL for SparkPost. Check the sparkpost documentation for possible URLs.",
			},
			"api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "API Key for SparkPost",
				MarkdownDescription: "API Key for SparkPost",
			},
		},
	}
}

func (p *sparkpostProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

    apiUrl := config.APIUrl.ValueString()
    if !strings.HasSuffix(apiUrl, "/") {
        apiUrl = apiUrl + "/"
    }

	client := NewSparkPostClient(apiUrl, config.APIKey.ValueString())

	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *sparkpostProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTrackingDomainResource,
		NewDomainResource,
		NewDomainVerificationResource,
		NewBounceVerificationResource,
		NewTrackingDomainVerificationResource,
		NewTrackingDomainAssociationResource,
	}
}

func (p *sparkpostProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
	    NewSubAccountsDataSource,
	}
}