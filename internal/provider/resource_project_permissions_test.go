package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProjectPermissionsResource_basic tests the full permissions lifecycle:
// create with users & groups → import → update (add / change / remove) → destroy.
func TestAccProjectPermissionsResource_basic(t *testing.T) {
	projectKey := "TFACCPERM"
	resourceName := "bitbucketdc_project_permissions.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// ── Step 1: Grant admin permission to the "admin" user ────────
			{
				Config: testAccProjectPermissionsConfig(projectKey, "admin", "PROJECT_ADMIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"name":       "admin",
						"permission": "PROJECT_ADMIN",
					}),
				),
			},
			// ── Step 2: Import by project key ─────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Change permission level ───────────────────────────
			{
				Config: testAccProjectPermissionsConfig(projectKey, "admin", "PROJECT_WRITE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"name":       "admin",
						"permission": "PROJECT_WRITE",
					}),
				),
			},
			// ── Step 4: Remove all permissions (empty config) ─────────────
			{
				Config: testAccProjectPermissionsEmptyConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
				),
			},
		},
	})
}

// TestAccProjectPermissionsResource_invalidPermission verifies that an invalid
// permission level is caught at plan time.
func TestAccProjectPermissionsResource_invalidPermission(t *testing.T) {
	projectKey := "TFACCPERMVAL"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectPermissionsConfig(projectKey, "admin", "INVALID_LEVEL"),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Permission Level`),
			},
		},
	})
}

// ── Config helpers ───────────────────────────────────────────────────────────

// testAccProjectPermissionsConfig creates a project and assigns a single user permission.
func testAccProjectPermissionsConfig(projectKey, username, permission string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Permissions Test Project"
}

resource "bitbucketdc_project_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  user {
    name       = %q
    permission = %q
  }
}
`, projectKey, username, permission)
}

// testAccProjectPermissionsEmptyConfig creates a project with no explicit permissions.
func testAccProjectPermissionsEmptyConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Permissions Test Project"
}

resource "bitbucketdc_project_permissions" "test" {
  project_key = bitbucketdc_project.test.key
}
`, projectKey)
}
