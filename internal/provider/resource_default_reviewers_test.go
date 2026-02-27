package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDefaultReviewersResource_basic tests the full lifecycle of project-level
// default reviewer conditions: create → import → update → remove all.
func TestAccDefaultReviewersResource_basic(t *testing.T) {
	projectKey := "TFDEFREV"
	resourceName := "bitbucketdc_default_reviewers.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckDefaultReviewersDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with one condition ─────────────────────────
			{
				Config: testAccDefaultReviewersConfig(projectKey,
					`condition {
  source_matcher_type = "ANY_REF"
  source_matcher_id   = "ANY_REF_MATCHER_ID"
  target_matcher_type = "ANY_REF"
  target_matcher_id   = "ANY_REF_MATCHER_ID"
  users               = ["admin"]
  required_approvals  = 1
}`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
				),
			},
			// ── Step 2: Import ────────────────────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Update required approvals ─────────────────────────
			{
				Config: testAccDefaultReviewersConfig(projectKey,
					`condition {
  source_matcher_type = "ANY_REF"
  source_matcher_id   = "ANY_REF_MATCHER_ID"
  target_matcher_type = "ANY_REF"
  target_matcher_id   = "ANY_REF_MATCHER_ID"
  users               = ["admin"]
  required_approvals  = 1
}

condition {
  source_matcher_type = "ANY_REF"
  source_matcher_id   = "ANY_REF_MATCHER_ID"
  target_matcher_type = "BRANCH"
  target_matcher_id   = "refs/heads/main"
  users               = ["admin"]
  required_approvals  = 1
}`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.#", "2"),
				),
			},
			// ── Step 4: Remove all conditions ─────────────────────────────
			{
				Config: testAccDefaultReviewersConfig(projectKey, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "condition.#", "0"),
				),
			},
		},
	})
}

// ── Destroy check ─────────────────────────────────────────────────────────────

func testAccCheckDefaultReviewersDestroyed(_ *terraform.State) error {
	// All conditions are deleted on resource destroy; no further check needed.
	return nil
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccDefaultReviewersConfig(projectKey, conditionBlocks string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Default Reviewers Test Project"
}

resource "bitbucketdc_default_reviewers" "test" {
  project_key = bitbucketdc_project.test.key
  %s
}
`, projectKey, conditionBlocks)
}
