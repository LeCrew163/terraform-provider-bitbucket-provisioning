package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRepositoryHookResource_basic tests enabling/disabling a built-in
// Bitbucket hook using the generic repository_hook resource.
// Hook used: com.atlassian.bitbucket.server.bitbucket-bundled-hooks:force-push-hook
// (bundled in every Bitbucket DC instance, no extra license required)
func TestAccRepositoryHookResource_basic(t *testing.T) {
	hookKey := "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:force-push-hook"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Enable the hook (no plugin-specific settings)
			{
				Config: testAccRepositoryHookConfig(hookKey, true, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "hook_key", hookKey),
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "enabled", "true"),
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "settings_json", `{}`),
				),
			},
			// Step 2: Import
			{
				ResourceName:      "bitbucketdc_repository_hook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Disable the hook
			{
				Config: testAccRepositoryHookConfig(hookKey, false, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "enabled", "false"),
				),
			},
		},
	})
}

// TestAccRepositoryHookResource_jenkinsWebhook tests the "Webhook to Jenkins for
// Bitbucket Server" plugin hook key and settings structure.
//
// This test requires a Bitbucket instance with a valid "Webhook to Jenkins for
// Bitbucket Server" license. Set BITBUCKET_JENKINS_HOOK_LICENSED=1 to enable.
func TestAccRepositoryHookResource_jenkinsWebhook(t *testing.T) {
	if os.Getenv("BITBUCKET_JENKINS_HOOK_LICENSED") == "" {
		t.Skip("Set BITBUCKET_JENKINS_HOOK_LICENSED=1 to run the Webhook-to-Jenkins hook test")
	}

	hookKey := "com.nerdwin15.stash-stash-webhook-jenkins:jenkinsPostReceiveHook"
	settingsJSON := `{"jenkinsBase":"http://jenkins.example.com","cloneType":"http","omitHashCode":false,"omitBranchName":false}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Configure the Jenkins webhook hook
			{
				Config: testAccRepositoryHookConfig(hookKey, true, settingsJSON),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "hook_key", hookKey),
					resource.TestCheckResourceAttr("bitbucketdc_repository_hook.test", "enabled", "true"),
				),
			},
			// Step 2: Import
			{
				ResourceName:      "bitbucketdc_repository_hook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryHookConfig(hookKey string, enabled bool, settingsJSON string) string {
	return fmt.Sprintf(`
resource "bitbucketdc_project" "test" {
  key  = "HOOKTEST"
  name = "Hook Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  name        = "Hook Test Repo"
}

resource "bitbucketdc_repository_hook" "test" {
  project_key     = bitbucketdc_project.test.key
  repository_slug = bitbucketdc_repository.test.slug
  hook_key        = %q
  enabled         = %t
  settings_json   = %q
}
`, hookKey, enabled, settingsJSON)
}
