package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "API URL for SparkPost",
				Default:     "https://api.eu.sparkpost.com/api/v1/",
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "API Key for SparkPost",
				Sensitive:   true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"sparkpost_tracking_domain": resourceTrackingdomain(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiUrl := d.Get("api_url").(string)
	apiKey := d.Get("api_key").(string)

	client := NewSparkPostClient(apiUrl, apiKey)

	return client, diags
}
