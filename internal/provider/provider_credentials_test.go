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

// TestProviderSchema_credentialSensitivity verifies that secret credential
// attributes are marked Sensitive (so Terraform redacts them in plan/apply
// output) and that the non-secret username is intentionally not sensitive.
func TestProviderSchema_credentialSensitivity(t *testing.T) {
	tests := []struct {
		attr          string
		wantSensitive bool
	}{
		{"token", true},
		{"password", true},
		{"username", false}, // public login identifier, not a secret
	}

	p := bitbucketProvider.New("test")()
	var req fwprovider.SchemaRequest
	resp := &fwprovider.SchemaResponse{}
	p.Schema(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned diagnostics errors: %s", resp.Diagnostics)
	}

	for _, tc := range tests {
		t.Run(tc.attr, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[tc.attr]
			if !ok {
				t.Fatalf("provider schema does not define attribute %q", tc.attr)
			}
			sa, ok := attr.(fwschema.StringAttribute)
			if !ok {
				t.Fatalf("attribute %q is not a StringAttribute (got %T)", tc.attr, attr)
			}
			if sa.Sensitive != tc.wantSensitive {
				t.Errorf("Sensitive = %v, want %v", sa.Sensitive, tc.wantSensitive)
			}
		})
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
		creds := []struct {
			envName string
			value   string
		}{
			{"BITBUCKET_TOKEN", os.Getenv("BITBUCKET_TOKEN")},
			{"BITBUCKET_USERNAME", os.Getenv("BITBUCKET_USERNAME")},
			{"BITBUCKET_PASSWORD", os.Getenv("BITBUCKET_PASSWORD")},
		}

		for resourceName, ms := range s.RootModule().Resources {
			if ms.Primary == nil {
				continue
			}
			for attrKey, attrVal := range ms.Primary.Attributes {
				for _, cred := range creds {
					if cred.value != "" && attrVal == cred.value {
						return fmt.Errorf(
							"resource %q attribute %q contains the value of %s — credentials must not be stored in resource state",
							resourceName, attrKey, cred.envName,
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
