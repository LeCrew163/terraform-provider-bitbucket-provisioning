package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProjectsDataSource_all creates a project and verifies it appears in the list.
func TestAccProjectsDataSource_all(t *testing.T) {
	projectKey := "TFDSPROJSALL"
	dsName := "data.bitbucketdc_projects.all"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectsDataSourceAllConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "id", "projects"),
					resource.TestCheckResourceAttrSet(dsName, "projects.#"),
				),
			},
		},
	})
}

// TestAccProjectsDataSource_filtered verifies that filter returns only matching projects.
func TestAccProjectsDataSource_filtered(t *testing.T) {
	projectKey := "TFDSPROJSFLTR"
	dsName := "data.bitbucketdc_projects.filtered"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectsDataSourceFilteredConfig(projectKey, "FilteredProject", "filteredproject"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "projects.#", "1"),
					resource.TestCheckResourceAttr(dsName, "projects.0.key", projectKey),
					resource.TestCheckResourceAttr(dsName, "projects.0.name", "FilteredProject"),
				),
			},
		},
	})
}

// TestAccProjectsDataSource_noMatch verifies that a non-matching filter returns an empty list.
func TestAccProjectsDataSource_noMatch(t *testing.T) {
	projectKey := "TFDSPROJSNONE"
	dsName := "data.bitbucketdc_projects.nomatch"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectsDataSourceNoMatchConfig(projectKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "projects.#", "0"),
				),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccProjectsDataSourceAllConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key             = %q
  name            = "Projects DS All Test"
  prevent_destroy = false
}

data "bitbucketdc_projects" "all" {
  depends_on = [bitbucketdc_project.test]
}
`, projectKey)
}

func testAccProjectsDataSourceFilteredConfig(projectKey, projectName, filter string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key             = %q
  name            = %q
  prevent_destroy = false
}

data "bitbucketdc_projects" "filtered" {
  filter     = %q
  depends_on = [bitbucketdc_project.test]
}
`, projectKey, projectName, filter)
}


func testAccProjectsDataSourceNoMatchConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key             = %q
  name            = "NoMatch Project"
  prevent_destroy = false
}

data "bitbucketdc_projects" "nomatch" {
  filter     = "nonexistent-xyz-filter-abc"
  depends_on = [bitbucketdc_project.test]
}
`, projectKey)
}
