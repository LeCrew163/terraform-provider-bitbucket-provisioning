package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGroupDataSource_basic looks up the built-in "stash-users" group that
// exists in every fresh Bitbucket DC instance used for acceptance tests.
func TestAccGroupDataSource_basic(t *testing.T) {
	dsName := "data.bitbucketdc_group.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig("stash-users"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "name", "stash-users"),
					resource.TestCheckResourceAttr(dsName, "id", "stash-users"),
				),
			},
		},
	})
}

// TestAccGroupDataSource_notFound verifies that a missing group name returns an error.
func TestAccGroupDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupDataSourceConfig("nonexistent-group-xyz"),
				ExpectError: errorRegexp(`Group Not Found`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccGroupDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

data "bitbucketdc_group" "test" {
  name = %q
}
`, name)
}
