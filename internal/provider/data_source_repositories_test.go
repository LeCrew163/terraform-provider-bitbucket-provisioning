package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRepositoriesDataSource_all creates a project+repo and verifies the repo appears in the list.
func TestAccRepositoriesDataSource_all(t *testing.T) {
	projectKey := "TFDSREPOSALL"
	dsName := "data.bitbucketdc_repositories.all"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceAllConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "project_key", projectKey),
					resource.TestCheckResourceAttr(dsName, "repositories.#", "1"),
					resource.TestCheckResourceAttr(dsName, "repositories.0.name", "Test Repo"),
					resource.TestCheckResourceAttrSet(dsName, "repositories.0.slug"),
					resource.TestCheckResourceAttrSet(dsName, "repositories.0.clone_url_http"),
				),
			},
		},
	})
}

// TestAccRepositoriesDataSource_filtered verifies that filter returns only matching repositories.
func TestAccRepositoriesDataSource_filtered(t *testing.T) {
	projectKey := "TFDSREPOSFLTR"
	dsName := "data.bitbucketdc_repositories.filtered"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceFilteredConfig(projectKey, "Service API", "service"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "repositories.#", "1"),
					resource.TestCheckResourceAttr(dsName, "repositories.0.name", "Service API"),
				),
			},
		},
	})
}

// TestAccRepositoriesDataSource_noMatch verifies a non-matching filter returns empty list.
func TestAccRepositoriesDataSource_noMatch(t *testing.T) {
	projectKey := "TFDSREPOSNONE"
	dsName := "data.bitbucketdc_repositories.nomatch"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoriesDataSourceNoMatchConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "repositories.#", "0"),
				),
			},
		},
	})
}

// TestAccRepositoriesDataSource_projectNotFound verifies error when project does not exist.
func TestAccRepositoriesDataSource_projectNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoriesDataSourceProjectNotFoundConfig("NONEXISTENTPRJ"),
				ExpectError: errorRegexp(`Project Not Found`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccRepositoriesDataSourceAllConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repos DS All Test"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = "Test Repo"
}

data "bitbucketdc_repositories" "all" {
  project_key = bitbucketdc_project.test.key
  depends_on  = [bitbucketdc_repository.test]
}
`, projectKey)
}

func testAccRepositoriesDataSourceFilteredConfig(projectKey, repoName, filter string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repos DS Filter Test"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = %q
}

data "bitbucketdc_repositories" "filtered" {
  project_key = bitbucketdc_project.test.key
  filter      = %q
  depends_on  = [bitbucketdc_repository.test]
}
`, projectKey, repoName, filter)
}

func testAccRepositoriesDataSourceNoMatchConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Repos DS NoMatch Test"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = "Some Repo"
}

data "bitbucketdc_repositories" "nomatch" {
  project_key = bitbucketdc_project.test.key
  filter      = "nonexistent-xyz-filter-abc"
  depends_on  = [bitbucketdc_repository.test]
}
`, projectKey)
}

func testAccRepositoriesDataSourceProjectNotFoundConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

data "bitbucketdc_repositories" "bad" {
  project_key = %q
}
`, projectKey)
}
