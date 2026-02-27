package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProjectDataSource_basic creates a project via resource then reads it
// back via data source and verifies the computed attributes match.
func TestAccProjectDataSource_basic(t *testing.T) {
	projectKey := "TFDSPROJ"
	dsName := "data.bitbucketdc_project.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig(projectKey, "Data Source Test Project", "Created for data source test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "key", projectKey),
					resource.TestCheckResourceAttr(dsName, "name", "Data Source Test Project"),
					resource.TestCheckResourceAttr(dsName, "description", "Created for data source test"),
					resource.TestCheckResourceAttr(dsName, "public", "false"),
					resource.TestCheckResourceAttr(dsName, "id", projectKey),
				),
			},
		},
	})
}

// TestAccProjectDataSource_notFound verifies that a missing project returns an error.
func TestAccProjectDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectDataSourceOnlyConfig("NONEXISTENT99"),
				ExpectError: errorRegexp(`Project Not Found`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccProjectDataSourceConfig(key, name, description string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key         = %q
  name        = %q
  description = %q
}

data "bitbucketdc_project" "test" {
  key        = bitbucketdc_project.test.key
  depends_on = [bitbucketdc_project.test]
}
`, key, name, description)
}

func testAccProjectDataSourceOnlyConfig(key string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

data "bitbucketdc_project" "test" {
  key = %q
}
`, key)
}
