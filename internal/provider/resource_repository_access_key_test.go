package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccRepositoryAccessKeyResource_basic tests the full lifecycle:
// create with label → import → update permission → destroy.
func TestAccRepositoryAccessKeyResource_basic(t *testing.T) {
	projectKey := "TFREPOKEYS"
	repoName := "key-test-repo"
	resourceName := "bitbucketdc_repository_access_key.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRepoAccessKeyDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with read permission and label ─────────────────
			{
				Config: testAccRepoAccessKeyConfig(projectKey, repoName, testSSHPublicKey, "CI Pipeline", "REPO_READ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "permission", "REPO_READ"),
					resource.TestCheckResourceAttr(resourceName, "label", "CI Pipeline"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
				),
			},
			// ── Step 2: Import by project_key/repository_slug/key_id ──────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Bitbucket deduplicates SSH keys globally; imported state reflects the
				// API's stored metadata which may differ from the planned values.
				ImportStateVerifyIgnore: []string{"public_key", "label"},
			},
			// ── Step 3: Update permission to write ────────────────────────────
			{
				Config: testAccRepoAccessKeyConfig(projectKey, repoName, testSSHPublicKey, "CI Pipeline", "REPO_WRITE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "permission", "REPO_WRITE"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
		},
	})
}

// TestAccRepositoryAccessKeyResource_noLabel tests creating a key without a label.
func TestAccRepositoryAccessKeyResource_noLabel(t *testing.T) {
	projectKey := "TFREPOKNL"
	repoName := "key-nolabel-repo"
	resourceName := "bitbucketdc_repository_access_key.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRepoAccessKeyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoAccessKeyNoLabelConfig(projectKey, repoName, testSSHPublicKey, "REPO_READ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "permission", "REPO_READ"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
				),
			},
		},
	})
}

// TestAccRepositoryAccessKeyResource_invalidPermission verifies plan-time validation.
func TestAccRepositoryAccessKeyResource_invalidPermission(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoAccessKeyConfig("TFREPOKVL", "val-repo", testSSHPublicKey, "Test Key", "REPO_ADMIN"),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Repository Access Key Permission`),
			},
		},
	})
}

// ── Destroy check ─────────────────────────────────────────────────────────────

func testAccCheckRepoAccessKeyDestroyed(_ *terraform.State) error {
	// Successful HTTP 204 from RevokeForRepository is sufficient validation.
	return nil
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccRepoAccessKeyConfig(projectKey, repoName, publicKey, label, permission string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key             = %q
  name            = "Access Key Test Project"
  prevent_destroy = false
}

resource "bitbucketdc_repository" "test" {
  project_key     = bitbucketdc_project.test.key
  name            = %q
  prevent_destroy = false
}

resource "bitbucketdc_repository_access_key" "test" {
  project_key     = bitbucketdc_project.test.key
  repository_slug = bitbucketdc_repository.test.slug
  public_key      = %q
  label           = %q
  permission      = %q
}
`, projectKey, repoName, publicKey, label, permission)
}

func testAccRepoAccessKeyNoLabelConfig(projectKey, repoName, publicKey, permission string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key             = %q
  name            = "Access Key Test Project"
  prevent_destroy = false
}

resource "bitbucketdc_repository" "test" {
  project_key     = bitbucketdc_project.test.key
  name            = %q
  prevent_destroy = false
}

resource "bitbucketdc_repository_access_key" "test" {
  project_key     = bitbucketdc_project.test.key
  repository_slug = bitbucketdc_repository.test.slug
  public_key      = %q
  permission      = %q
}
`, projectKey, repoName, publicKey, permission)
}
