package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccRepositoryResource_basic tests the full repository lifecycle:
// create → read → import → update name → update description → destroy.
func TestAccRepositoryResource_basic(t *testing.T) {
	projectKey := "TFACCREPOBASIC"
	resourceName := "bitbucketdc_repository.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRepositoryDestroyed,
		Steps: []resource.TestStep{
			// ── Step 1: Create with all fields ───────────────────────────
			{
				Config: testAccRepositoryConfig(projectKey, "Test Repository", "Initial description", false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttrSet(resourceName, "slug"),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Repository"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "public", "false"),
					resource.TestCheckResourceAttr(resourceName, "forkable", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_http"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_ssh"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
				),
			},
			// ── Step 2: Import by project_key/slug ───────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Update name and description ──────────────────────
			{
				Config: testAccRepositoryConfig(projectKey, "Test Repository Updated", "Updated description", false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test Repository Updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
			// ── Step 4: Make repository non-forkable ─────────────────────
			{
				Config: testAccRepositoryConfig(projectKey, "Test Repository Updated", "Updated description", false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "forkable", "false"),
				),
			},
		},
	})
}

// TestAccRepositoryResource_minimal tests a repository created with only required fields.
func TestAccRepositoryResource_minimal(t *testing.T) {
	projectKey := "TFACCREPOMIN"
	resourceName := "bitbucketdc_repository.minimal"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckRepositoryDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryMinimalConfig(projectKey, "Minimal Repository"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttrSet(resourceName, "slug"),
					resource.TestCheckResourceAttr(resourceName, "name", "Minimal Repository"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_http"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_ssh"),
				),
			},
		},
	})
}

// ── Destroy check ────────────────────────────────────────────────────────────

func testAccCheckRepositoryDestroyed(s *terraform.State) error {
	httpClient := newTestHTTPClient()

	for name, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucketdc_repository" {
			continue
		}

		projectKey := rs.Primary.Attributes["project_key"]
		slug := rs.Primary.Attributes["slug"]
		url := fmt.Sprintf("%s/rest/api/latest/projects/%s/repos/%s",
			testBitbucketBaseURL(), projectKey, slug)

		req := newTestRequest("GET", url)
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error checking repository %s (%s): %w", name, slug, err)
		}
		resp.Body.Close()

		if resp.StatusCode != 404 {
			return fmt.Errorf("repository %s (%s/%s) still exists (HTTP %d)", name, projectKey, slug, resp.StatusCode)
		}
	}

	return nil
}

// ── Config helpers ───────────────────────────────────────────────────────────

// testAccRepositoryConfig returns a Terraform configuration that creates a project and
// a repository with all optional fields set. The slug is derived by Bitbucket from the name.
func testAccRepositoryConfig(projectKey, name, description string, public, forkable bool) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repo Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = %q
  description = %q
  public      = %t
  forkable    = %t
}
`, projectKey, name, description, public, forkable)
}

// testAccRepositoryMinimalConfig returns a Terraform configuration with a project and
// a repository with only the required fields (project_key, name).
func testAccRepositoryMinimalConfig(projectKey, name string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repo Test Project"
}

resource "bitbucketdc_repository" "minimal" {
  project_key = bitbucketdc_project.test.key
  name        = %q
}
`, projectKey, name)
}
