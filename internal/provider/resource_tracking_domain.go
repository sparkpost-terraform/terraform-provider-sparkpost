package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTrackingdomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrackingDomainCreate,
		ReadContext:   resourceTrackingDomainRead,
		DeleteContext: resourceTrackingDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain to be used for tracking links",
			},
			"https": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Specifies if the domain should use HTTPS",
			},
		},
	}
}

func resourceTrackingDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Get("domain").(string)
	https := d.Get("https").(bool)

	err := client.CreateTrackingDomain(domain, https)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain)

	return resourceTrackingDomainRead(ctx, d, m)
}

func resourceTrackingDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Id()

	_, err := client.GetTrackingDomain(domain)
	if err != nil {
		if err == TrackingDomainNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceTrackingDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Id()

	err := client.DeleteTrackingDomain(domain)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("") // Remove resource from state

	return nil
}
