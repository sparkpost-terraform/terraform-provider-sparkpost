package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		DeleteContext: resourceDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain to be used for sending or bounces",
			},
			"subaccount": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional subnet account ID for creating the tracking domain in",
			},        
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Get("domain").(string)

	var subaccount int
	if v, ok := d.GetOk("subaccount"); ok {
		subaccount = v.(int)
	}

	err := client.CreateDomain(domain, subaccount)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domain)

	return resourceDomainRead(ctx, d, m)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Id()
	
	var subaccount int
	if v, ok := d.GetOk("subaccount"); ok {
		subaccount = v.(int)
	}	

	_, err := client.GetDomain(domain, subaccount)
	if err != nil {
		if err == DomainNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*SparkPostClient)
	domain := d.Id()
	
	var subaccount int
	if v, ok := d.GetOk("subaccount"); ok {
		subaccount = v.(int)
	}	

	err := client.DeleteDomain(domain, subaccount)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("") // Remove resource from state

	return nil
}
