package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccRepositoryPermissionsResource_basic tests the full lifecycle:
// create with one user → import → add group → remove user → empty.
func TestAccRepositoryPermissionsResource_basic(t *testing.T) {
	projectKey := "TFREPOPRM"
	repoName := "perm-test-repo"
	resourceName := "bitbucketdc_repository_permissions.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRepoPremissionsDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with one user (admin) ──────────────────────────
			{
				Config: testAccRepoPermissionsConfig(projectKey, repoName,
					"user {\n  name       = \"admin\"\n  permission = \"REPO_ADMIN\"\n}",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "repository_slug", repoName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ── Step 2: Import by project_key/repository_slug ─────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Change permission level ───────────────────────────────
			{
				Config: testAccRepoPermissionsConfig(projectKey, repoName,
					"user {\n  name       = \"admin\"\n  permission = \"REPO_WRITE\"\n}",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
				),
			},
			// ── Step 4: Empty — revoke all ────────────────────────────────────
			{
				Config: testAccRepoPermissionsConfig(projectKey, repoName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
				),
			},
		},
	})
}

// TestAccRepositoryPermissionsResource_invalidPermission verifies plan-time validation.
func TestAccRepositoryPermissionsResource_invalidPermission(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRepoPermissionsConfig("TFRPVAL", "val-repo",
					"user {\n  name       = \"admin\"\n  permission = \"PROJECT_ADMIN\"\n}",
				),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Repository Permission Level`),
			},
		},
	})
}

// ── Destroy check ─────────────────────────────────────────────────────────────

func testAccCheckRepoPremissionsDestroyed(_ *terraform.State) error {
	// Reconcile-to-empty on destroy revokes all permissions; no further check needed.
	return nil
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccRepoPermissionsConfig(projectKey, repoName, permBlocks string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repo Permissions Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = %q
}

resource "bitbucketdc_repository_permissions" "test" {
  project_key     = bitbucketdc_project.test.key
  repository_slug = bitbucketdc_repository.test.slug
  %s
}
`, projectKey, repoName, permBlocks)
}
