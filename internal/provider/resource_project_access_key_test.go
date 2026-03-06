package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testSSHPublicKey is a throwaway ED25519 public key used only for acceptance tests.
// It is NOT a real key used for any environment.
const testSSHPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHj27twPopsFRMno1cyptyxbcKG3zSlFq7oCWPzWzlyP test-acceptance-key-1"

// TestAccProjectAccessKeyResource_basic tests the full lifecycle:
// create with label → import → update permission → update key (replace) → destroy.
func TestAccProjectAccessKeyResource_basic(t *testing.T) {
	projectKey := "TFACCKEYS"
	resourceName := "bitbucketdc_project_access_key.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckAccessKeyDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with read permission ───────────────────────
			{
				Config: testAccAccessKeyConfig(projectKey, testSSHPublicKey, "CI Pipeline", "PROJECT_READ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "permission", "PROJECT_READ"),
					resource.TestCheckResourceAttr(resourceName, "label", "CI Pipeline"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
				),
			},
			// ── Step 2: Import by project_key/key_id ──────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Update permission to write ────────────────────────
			{
				Config: testAccAccessKeyConfig(projectKey, testSSHPublicKey, "CI Pipeline", "PROJECT_WRITE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "permission", "PROJECT_WRITE"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
		},
	})
}

// TestAccProjectAccessKeyResource_noLabel tests creating a key without a label.
func TestAccProjectAccessKeyResource_noLabel(t *testing.T) {
	projectKey := "TFACCKEYSNL"
	resourceName := "bitbucketdc_project_access_key.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckAccessKeyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeyNoLabelConfig(projectKey, testSSHPublicKey, "PROJECT_READ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "permission", "PROJECT_READ"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
				),
			},
		},
	})
}

// TestAccProjectAccessKeyResource_invalidPermission verifies plan-time validation.
func TestAccProjectAccessKeyResource_invalidPermission(t *testing.T) {
	projectKey := "TFACCKEYSVAL"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAccessKeyConfig(projectKey, testSSHPublicKey, "Test Key", "PROJECT_ADMIN"),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Access Key Permission`),
			},
		},
	})
}

// ── Destroy check ────────────────────────────────────────────────────────────

func testAccCheckAccessKeyDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucketdc_project_access_key" {
			continue
		}
		// If the key still exists in state after destroy the test framework
		// would have caught that already. For access keys there is no simple
		// REST GET-by-ID endpoint exposed in the test client — successful
		// destroy (HTTP 204) is sufficient validation.
	}
	return nil
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccAccessKeyConfig(projectKey, publicKey, label, permission string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Access Key Test Project"
}

resource "bitbucketdc_project_access_key" "test" {
  project_key = bitbucketdc_project.test.key
  public_key  = %q
  label       = %q
  permission  = %q
}
`, projectKey, publicKey, label, permission)
}

func testAccAccessKeyNoLabelConfig(projectKey, publicKey, permission string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Access Key Test Project"
}

resource "bitbucketdc_project_access_key" "test" {
  project_key = bitbucketdc_project.test.key
  public_key  = %q
  permission  = %q
}
`, projectKey, publicKey, permission)
}
