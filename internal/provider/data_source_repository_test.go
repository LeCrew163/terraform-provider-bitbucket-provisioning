package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRepositoryDataSource_basic creates a project + repository via resources
// then reads the repository back via data source and verifies computed attributes.
func TestAccRepositoryDataSource_basic(t *testing.T) {
	projectKey := "TFDSREPO"
	repoName := "ds-test-repository"
	dsName := "data.bitbucketdc_repository.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDataSourceConfig(projectKey, repoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "project_key", projectKey),
					resource.TestCheckResourceAttr(dsName, "name", repoName),
					resource.TestCheckResourceAttrSet(dsName, "slug"),
					resource.TestCheckResourceAttrSet(dsName, "id"),
					resource.TestCheckResourceAttrSet(dsName, "state"),
					resource.TestCheckResourceAttrSet(dsName, "clone_url_http"),
					resource.TestCheckResourceAttrSet(dsName, "clone_url_ssh"),
				),
			},
		},
	})
}

// TestAccRepositoryDataSource_notFound verifies that a missing repository returns an error.
func TestAccRepositoryDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoryDataSourceOnlyConfig("TFDSREPO", "nonexistent-repo-slug"),
				ExpectError: errorRegexp(`Repository Not Found`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccRepositoryDataSourceConfig(projectKey, repoName string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repository Data Source Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = %q
}

data "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  slug        = bitbucketdc_repository.test.slug
  depends_on  = [bitbucketdc_repository.test]
}
`, projectKey, repoName)
}

func testAccRepositoryDataSourceOnlyConfig(projectKey, slug string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

data "bitbucketdc_repository" "test" {
  project_key = %q
  slug        = %q
}
`, projectKey, slug)
}
