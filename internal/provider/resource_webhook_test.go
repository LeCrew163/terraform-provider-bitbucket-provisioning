package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccWebhookResource_repository tests the full lifecycle of a repository-scoped webhook.
func TestAccWebhookResource_repository(t *testing.T) {
	projectKey := "TFWEBHOOK"
	repoName := "webhook-test-repo"
	resourceName := "bitbucketdc_webhook.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// ── Step 1: Create ────────────────────────────────────────────
			{
				Config: testAccWebhookRepoConfig(projectKey, repoName,
					"Test Webhook", "http://example.com/hook",
					`["repo:refs_changed"]`, true, false,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Webhook"),
					resource.TestCheckResourceAttr(resourceName, "url", "http://example.com/hook"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "ssl_verification_required", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "webhook_id"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ── Step 2: Import ────────────────────────────────────────────
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// ── Step 3: Update name, url and events ──────────────────────
			{
				Config: testAccWebhookRepoConfig(projectKey, repoName,
					"Updated Webhook", "http://example.com/hook-v2",
					`["repo:refs_changed", "pr:opened"]`, true, false,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Webhook"),
					resource.TestCheckResourceAttr(resourceName, "url", "http://example.com/hook-v2"),
					resource.TestCheckResourceAttr(resourceName, "events.#", "2"),
				),
			},
			// ── Step 4: Disable ───────────────────────────────────────────
			{
				Config: testAccWebhookRepoConfig(projectKey, repoName,
					"Updated Webhook", "http://example.com/hook-v2",
					`["repo:refs_changed", "pr:opened"]`, false, false,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
				),
			},
		},
	})
}

// TestAccWebhookResource_project tests a project-scoped webhook.
func TestAccWebhookResource_project(t *testing.T) {
	projectKey := "TFWHPROJ"
	resourceName := "bitbucketdc_webhook.proj"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookProjectConfig(projectKey,
					"Project Webhook", "http://example.com/project-hook",
					`["repo:refs_changed"]`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_key", projectKey),
					resource.TestCheckResourceAttr(resourceName, "name", "Project Webhook"),
					resource.TestCheckResourceAttrSet(resourceName, "webhook_id"),
					resource.TestCheckNoResourceAttr(resourceName, "repository_slug"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func testAccWebhookRepoConfig(projectKey, repoName, name, url, eventsJSON string, active, sslVerif bool) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Webhook Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = %q
}

resource "bitbucketdc_webhook" "test" {
  project_key              = bitbucketdc_project.test.key
  repository_slug          = bitbucketdc_repository.test.slug
  name                     = %q
  url                      = %q
  events                   = %s
  active                   = %t
  ssl_verification_required = %t
}
`, projectKey, repoName, name, url, eventsJSON, active, sslVerif)
}

func testAccWebhookProjectConfig(projectKey, name, url, eventsJSON string) string {
	return fmt.Sprintf(`
provider "bitbucketdc" {}

resource "bitbucketdc_project" "test" {
  key  = %q
  name = "Webhook Project Test"
}

resource "bitbucketdc_webhook" "proj" {
  project_key = bitbucketdc_project.test.key
  name        = %q
  url         = %q
  events      = %s
}
`, projectKey, name, url, eventsJSON)
}
