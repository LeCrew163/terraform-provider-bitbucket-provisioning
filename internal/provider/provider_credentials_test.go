package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	bitbucketProvider "github.com/LeCrew163/bitbucket-provisioning/internal/provider"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ── Schema sensitivity unit tests ─────────────────────────────────────────────
// These tests do NOT require a running Bitbucket instance.

// TestProviderSchema_tokenSensitive verifies that the 'token' provider attribute
// is marked Sensitive so Terraform redacts its value in plan/apply output.
func TestProviderSchema_tokenSensitive(t *testing.T) {
	checkProviderAttrSensitive(t, "token", true)
}

// TestProviderSchema_passwordSensitive verifies that the 'password' provider
// attribute is marked Sensitive.
func TestProviderSchema_passwordSensitive(t *testing.T) {
	checkProviderAttrSensitive(t, "password", true)
}

// TestProviderSchema_usernameNotSensitive verifies that 'username' is
// intentionally NOT marked Sensitive — it is a public login identifier,
// not a secret value.
func TestProviderSchema_usernameNotSensitive(t *testing.T) {
	checkProviderAttrSensitive(t, "username", false)
}

// checkProviderAttrSensitive instantiates the provider, calls Schema(), and
// asserts that the named attribute has the expected Sensitive value.
func checkProviderAttrSensitive(t *testing.T, attrName string, wantSensitive bool) {
	t.Helper()

	p := bitbucketProvider.New("test")()

	var req fwprovider.SchemaRequest
	resp := &fwprovider.SchemaResponse{}
	p.Schema(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned diagnostics errors: %s", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes[attrName]
	if !ok {
		t.Fatalf("provider schema does not define attribute %q", attrName)
	}

	sa, ok := attr.(fwschema.StringAttribute)
	if !ok {
		t.Fatalf("attribute %q is not a StringAttribute (got %T)", attrName, attr)
	}

	if sa.Sensitive != wantSensitive {
		t.Errorf("attribute %q: Sensitive = %v, want %v", attrName, sa.Sensitive, wantSensitive)
	}
}

// ── Acceptance test: credentials absent from resource state ──────────────────

// TestAccProvider_credentialsNotInResourceState verifies that the credential
// values supplied via environment variables (BITBUCKET_TOKEN / BITBUCKET_USERNAME
// / BITBUCKET_PASSWORD) do not appear as attribute values in any resource's
// Terraform state after a successful apply.
func TestAccProvider_credentialsNotInResourceState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialLeakConfig(),
				Check:  testAccCheckNoCredentialInState(),
			},
		},
	})
}

// testAccCheckNoCredentialInState returns a TestCheckFunc that fails if any
// resource attribute value in the state equals one of the credential env vars.
func testAccCheckNoCredentialInState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		creds := map[string]string{}
		if v := os.Getenv("BITBUCKET_TOKEN"); v != "" {
			creds["BITBUCKET_TOKEN"] = v
		}
		if v := os.Getenv("BITBUCKET_USERNAME"); v != "" {
			creds["BITBUCKET_USERNAME"] = v
		}
		if v := os.Getenv("BITBUCKET_PASSWORD"); v != "" {
			creds["BITBUCKET_PASSWORD"] = v
		}

		for resourceName, ms := range s.RootModule().Resources {
			if ms.Primary == nil {
				continue
			}
			for attrKey, attrVal := range ms.Primary.Attributes {
				for envName, secret := range creds {
					if attrVal == secret {
						return fmt.Errorf(
							"resource %q attribute %q contains the value of %s — credentials must not be stored in resource state",
							resourceName, attrKey, envName,
						)
					}
				}
			}
		}
		return nil
	}
}

// testAccCredentialLeakConfig returns a minimal Terraform configuration that
// creates one resource so there is state to inspect.
func testAccCredentialLeakConfig() string {
	return `
resource "bitbucketdc_project" "credential_check" {
  key             = "TFCREDCHK"
  name            = "Credential Leak Check"
  prevent_destroy = false
}
`
}
