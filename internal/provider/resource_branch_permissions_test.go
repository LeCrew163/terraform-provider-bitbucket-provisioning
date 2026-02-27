package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccBranchPermissionsResource_basic tests the full lifecycle:
// create with two rules → import → update (add rule) → update (remove rule) → destroy.
func TestAccBranchPermissionsResource_basic(t *testing.T) {
	projectKey := "TFACCBRPERM"
	resourceName := "bitbucketdc_branch_permissions.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// ── Step 1: Create with one restriction ───────────────────────
			{
				Config: testAccBranchPermissionsOneRuleConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "restriction.*", map[string]string{
						"type":         "no-deletes",
						"matcher_type": "BRANCH",
						"matcher_id":   "main",
					}),
				),
			},
			// ── Step 2: Import by project key ─────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Add a second restriction ──────────────────────────
			{
				Config: testAccBranchPermissionsTwoRulesConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "restriction.*", map[string]string{
						"type":         "no-deletes",
						"matcher_type": "BRANCH",
						"matcher_id":   "main",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "restriction.*", map[string]string{
						"type":         "fast-forward-only",
						"matcher_type": "PATTERN",
						"matcher_id":   "release/*",
					}),
				),
			},
			// ── Step 4: Remove first restriction (only second remains) ─────
			{
				Config: testAccBranchPermissionsSecondRuleConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "restriction.*", map[string]string{
						"type":         "fast-forward-only",
						"matcher_type": "PATTERN",
						"matcher_id":   "release/*",
					}),
				),
			},
			// ── Step 5: Remove all restrictions ───────────────────────────
			{
				Config: testAccBranchPermissionsEmptyConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
				),
			},
		},
	})
}

// TestAccBranchPermissionsResource_invalidType verifies that an invalid restriction
// type is caught at plan time.
func TestAccBranchPermissionsResource_invalidType(t *testing.T) {
	projectKey := "TFACCBRPERMVAL"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccBranchPermissionsInvalidTypeConfig(projectKey),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Branch Restriction Type`),
			},
		},
	})
}

// TestAccBranchPermissionsResource_invalidMatcherType verifies that an invalid
// matcher type is caught at plan time.
func TestAccBranchPermissionsResource_invalidMatcherType(t *testing.T) {
	projectKey := "TFACCBRPERMVAL2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccBranchPermissionsInvalidMatcherConfig(projectKey),
				PlanOnly:    true,
				ExpectError: errorRegexp(`Invalid Branch Matcher Type`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccBranchPermissionsOneRuleConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  restriction {
    type         = "no-deletes"
    matcher_type = "BRANCH"
    matcher_id   = "main"
  }
}
`, projectKey)
}

func testAccBranchPermissionsTwoRulesConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  restriction {
    type         = "no-deletes"
    matcher_type = "BRANCH"
    matcher_id   = "main"
  }

  restriction {
    type         = "fast-forward-only"
    matcher_type = "PATTERN"
    matcher_id   = "release/*"
  }
}
`, projectKey)
}

func testAccBranchPermissionsSecondRuleConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  restriction {
    type         = "fast-forward-only"
    matcher_type = "PATTERN"
    matcher_id   = "release/*"
  }
}
`, projectKey)
}

func testAccBranchPermissionsEmptyConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key
}
`, projectKey)
}

func testAccBranchPermissionsInvalidTypeConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  restriction {
    type         = "INVALID_TYPE"
    matcher_type = "BRANCH"
    matcher_id   = "main"
  }
}
`, projectKey)
}

func testAccBranchPermissionsInvalidMatcherConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Branch Permissions Test Project"
}

resource "bitbucketdc_branch_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  restriction {
    type         = "no-deletes"
    matcher_type = "INVALID_MATCHER"
    matcher_id   = "main"
  }
}
`, projectKey)
}
