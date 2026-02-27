package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccProjectResource_basic tests the full project lifecycle:
// create → read → import → update name → update description → make public.
func TestAccProjectResource_basic(t *testing.T) {
	resourceName := "bitbucketdc_project.test"
	key := "TFACCBASIC"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProjectDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with all fields ───────────────────────────
			{
				Config: testAccProjectAllFieldsConfig(key, "Acceptance Test Project", "Created by acceptance test", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "name", "Acceptance Test Project"),
					resource.TestCheckResourceAttr(resourceName, "description", "Created by acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "public", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ── Step 2: Import by key ─────────────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Update name and description ──────────────────────
			{
				Config: testAccProjectAllFieldsConfig(key, "Acceptance Test Project (Updated)", "Updated by acceptance test", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "name", "Acceptance Test Project (Updated)"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated by acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "public", "false"),
				),
			},
			// ── Step 4: Make project public ───────────────────────────────
			{
				Config: testAccProjectAllFieldsConfig(key, "Acceptance Test Project (Updated)", "Updated by acceptance test", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "public", "true"),
				),
			},
			// ── Step 5: Import after update ───────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccProjectResource_minimal tests a project created with only required fields
// (key and name). Verifies that computed fields (id, public) are set by the provider.
func TestAccProjectResource_minimal(t *testing.T) {
	resourceName := "bitbucketdc_project.minimal"
	key := "TFACCMIN"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectMinimalConfig(key, "Minimal Test Project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "name", "Minimal Test Project"),
					resource.TestCheckResourceAttr(resourceName, "public", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

// TestAccProjectResource_keyReplaces tests that changing the project key
// forces a destroy-then-create (RequiresReplace plan modifier).
func TestAccProjectResource_keyReplaces(t *testing.T) {
	resourceName := "bitbucketdc_project.minimal"
	key1 := "TFACCKEY1"
	key2 := "TFACCKEY2"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectMinimalConfig(key1, "Key Replace Test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key1),
				),
			},
			{
				// Changing the key must replace the resource (destroy old, create new).
				Config: testAccProjectMinimalConfig(key2, "Key Replace Test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key2),
				),
			},
		},
	})
}

// TestAccProjectResource_disappears verifies that when a project is deleted
// outside Terraform, the next plan detects drift and proposes to re-create it.
func TestAccProjectResource_disappears(t *testing.T) {
	resourceName := "bitbucketdc_project.test"
	key := "TFACCDIS"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectAllFieldsConfig(key, "Disappears Test", "Will be deleted out-of-band", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					// Simulate out-of-band deletion immediately after apply.
					testAccDeleteProjectOutOfBand(key),
				),
				// After the check deletes the project, a plan will show it needs
				// to be re-created — this is the expected drift.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccProjectResource_duplicateKey verifies that attempting to create a project
// with a key that already exists produces a clear "Project Already Exists" error.
func TestAccProjectResource_duplicateKey(t *testing.T) {
	key := "TFACCDUPKEY"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProjectDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create the original project.
			{
				Config: testAccProjectAllFieldsConfig(key, "Duplicate Key Test", "For testing duplicate key errors", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_project.test", "key", key),
				),
			},
			// Step 2: add a second resource that tries to use the same key.
			// The apply must fail with the "Project Already Exists" error.
			{
				Config:      testAccProjectDuplicateKeyConfig(key),
				ExpectError: regexp.MustCompile(`Project Already Exists`),
			},
		},
	})
}

// TestAccProjectResource_invalidKeyPlan verifies that the plan-time key validator
// rejects keys that do not match the required format before any API call is made.
func TestAccProjectResource_invalidKeyPlan(t *testing.T) {
	cases := []struct {
		name        string
		key         string
		errorRegexp *regexp.Regexp
	}{
		{
			name:        "lowercase",
			key:         "lowercase",
			errorRegexp: regexp.MustCompile(`Invalid Project Key`),
		},
		{
			name:        "starts_with_underscore",
			key:         "_STARTS",
			errorRegexp: regexp.MustCompile(`Invalid Project Key`),
		},
		{
			name:        "starts_with_digit",
			key:         "1STARTS",
			errorRegexp: regexp.MustCompile(`Invalid Project Key`),
		},
		{
			name:        "contains_hyphen",
			key:         "MY-PROJECT",
			errorRegexp: regexp.MustCompile(`Invalid Project Key`),
		},
		{
			name:        "single_char",
			key:         "A",
			errorRegexp: regexp.MustCompile(`Invalid Project Key`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				Steps: []resource.TestStep{
					{
						Config:      testAccProjectMinimalConfig(tc.key, "Invalid Key Test"),
						PlanOnly:    true,
						ExpectError: tc.errorRegexp,
					},
				},
			})
		})
	}
}

// ── Destroy check ────────────────────────────────────────────────────────────

// testAccCheckProjectDestroyed verifies that all projects managed by the test
// no longer exist in Bitbucket after the test resources have been destroyed.
func testAccCheckProjectDestroyed(s *terraform.State) error {
	httpClient := newTestHTTPClient()

	for name, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucketdc_project" {
			continue
		}

		key := rs.Primary.Attributes["key"]
		url := fmt.Sprintf("%s/rest/api/latest/projects/%s", testBitbucketBaseURL(), key)

		resp, err := httpClient.Do(newTestRequest("GET", url))
		if err != nil {
			return fmt.Errorf("error checking project %s (%s): %w", key, name, err)
		}
		resp.Body.Close()

		if resp.StatusCode != 404 {
			return fmt.Errorf("project %s (%s) still exists (HTTP %d)", key, name, resp.StatusCode)
		}
	}

	return nil
}

// testAccDeleteProjectOutOfBand returns a TestCheckFunc that deletes the project
// directly via the Bitbucket REST API, simulating an out-of-band deletion.
func testAccDeleteProjectOutOfBand(key string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		url := fmt.Sprintf("%s/rest/api/latest/projects/%s", testBitbucketBaseURL(), key)

		resp, err := newTestHTTPClient().Do(newTestRequest("DELETE", url))
		if err != nil {
			return fmt.Errorf("error deleting project %s out-of-band: %w", key, err)
		}
		resp.Body.Close()

		if resp.StatusCode != 204 && resp.StatusCode != 404 {
			return fmt.Errorf("unexpected status %d when deleting project %s out-of-band", resp.StatusCode, key)
		}

		return nil
	}
}

// ── Config helpers ───────────────────────────────────────────────────────────

// testAccProjectAllFieldsConfig returns a configuration for bitbucketdc_project.test
// with all attributes set.
func testAccProjectAllFieldsConfig(key, name, description string, public bool) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key         = %q
  name        = %q
  description = %q
  public      = %t
}
`, key, name, description, public)
}

// testAccProjectMinimalConfig returns a configuration for bitbucketdc_project.minimal
// with only the required attributes (key and name).
func testAccProjectMinimalConfig(key, name string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "minimal" {
  key  = %q
  name = %q
}
`, key, name)
}

// testAccProjectDuplicateKeyConfig returns a configuration that contains the
// original test project AND a second resource that tries to claim the same key,
// which should trigger the "Project Already Exists" error on apply.
func testAccProjectDuplicateKeyConfig(key string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Duplicate Key Test"
}

resource "bitbucketdc_project" "dup" {
  key  = %q
  name = "Duplicate Key Test (copy)"
}
`, key, key)
}
