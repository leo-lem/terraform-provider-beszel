package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/leo-lem/terraform-provider-beszel/internal/provider"
)

// version is set via -ldflags at release build time (see .goreleaser.yml).
var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/leo-lem/beszel",
	})
	if err != nil {
		log.Fatal(err)
	}
}
