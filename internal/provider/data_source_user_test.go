package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserDataSource_basic looks up the built-in "admin" user that exists in
// every fresh Bitbucket DC instance used for acceptance tests.
func TestAccUserDataSource_basic(t *testing.T) {
	dsName := "data.bitbucketdc_user.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig("admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "slug", "admin"),
					resource.TestCheckResourceAttr(dsName, "id", "admin"),
					resource.TestCheckResourceAttrSet(dsName, "name"),
					resource.TestCheckResourceAttrSet(dsName, "display_name"),
					resource.TestCheckResourceAttrSet(dsName, "active"),
				),
			},
		},
	})
}

// TestAccUserDataSource_notFound verifies that a missing user slug returns an error.
func TestAccUserDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig("nonexistent-user-slug-xyz"),
				ExpectError: errorRegexp(`User Not Found`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccUserDataSourceConfig(slug string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

data "bitbucketdc_user" "test" {
  slug = %q
}
`, slug)
}
