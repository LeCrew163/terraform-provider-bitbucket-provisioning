package main

import (
	"context"
	"log"

	"github.com/LeCrew163/bitbucket-provisioning/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version will be set by the goreleaser configuration
var version string = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{})
	if err != nil {
		log.Fatal(err.Error())
	}
}
