package main

import (
	"context"
	"flag"
	"log"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version will be set by the goreleaser configuration
var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "art01.sldnet.de:8081/swisslife/bitbucket-provisioning",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
