package provider_test

import (
	"os"
	"testing"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used by acceptance tests to instantiate
// the provider under test via Protocol V6.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bitbucketdc": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates that the environment variables required by
// acceptance tests are set. Call it from each TestAcc* function via PreCheck.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("BITBUCKET_BASE_URL"); v == "" {
		t.Fatal("BITBUCKET_BASE_URL must be set for acceptance tests.\n" +
			"  Example: export BITBUCKET_BASE_URL=http://localhost:7990")
	}

	hasToken := os.Getenv("BITBUCKET_TOKEN") != ""
	hasBasicAuth := os.Getenv("BITBUCKET_USERNAME") != "" && os.Getenv("BITBUCKET_PASSWORD") != ""

	if !hasToken && !hasBasicAuth {
		t.Fatal("Authentication credentials must be set for acceptance tests.\n" +
			"  Either: export BITBUCKET_TOKEN=<personal-access-token>\n" +
			"  Or:     export BITBUCKET_USERNAME=admin BITBUCKET_PASSWORD=admin")
	}

	if hasToken && hasBasicAuth {
		t.Fatal("Both BITBUCKET_TOKEN and BITBUCKET_USERNAME/PASSWORD are set.\n" +
			"  The provider only accepts one authentication method at a time.\n" +
			"  Unset the token before running acceptance tests:\n" +
			"    unset BITBUCKET_TOKEN")
	}
}
