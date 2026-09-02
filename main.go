package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/sparkpost-terraform/terraform-provider-sparkpost/internal/provider"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.4 generate --provider-name sparkpost

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/sparkpost-terraform/sparkpost",
	})
	if err != nil {
		log.Fatal(err)
	}
}
