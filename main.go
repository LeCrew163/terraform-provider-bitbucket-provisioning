package main

import (
	"context"
	"flag"
	"log"

	"github.com/alpina/terraform-provider-bitbucket-dc/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version will be set by the goreleaser configuration
var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/alpina/bitbucket-dc",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
