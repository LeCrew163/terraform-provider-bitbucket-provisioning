package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProjectHookResource_basic tests enabling/disabling a built-in
// Bitbucket project-level hook using the generic project_hook resource.
// Hook used: com.atlassian.bitbucket.server.bitbucket-bundled-hooks:force-push-hook
// (bundled in every Bitbucket DC instance, no extra license required)
func TestAccProjectHookResource_basic(t *testing.T) {
	hookKey := "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:force-push-hook"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Enable the hook at project level
			{
				Config: testAccProjectHookConfig(hookKey, true, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "hook_key", hookKey),
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "enabled", "true"),
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "settings_json", `{}`),
				),
			},
			// Step 2: Import
			{
				ResourceName:      "bitbucketdc_project_hook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Disable the hook
			{
				Config: testAccProjectHookConfig(hookKey, false, `{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "enabled", "false"),
				),
			},
		},
	})
}

// TestAccProjectHookResource_jenkinsWebhook tests the "Webhook to Jenkins for
// Bitbucket Server" plugin hook key and settings at project level.
// Requires a Bitbucket instance with a valid plugin license.
// Set BITBUCKET_JENKINS_HOOK_LICENSED=1 to enable.
func TestAccProjectHookResource_jenkinsWebhook(t *testing.T) {
	if os.Getenv("BITBUCKET_JENKINS_HOOK_LICENSED") == "" {
		t.Skip("Set BITBUCKET_JENKINS_HOOK_LICENSED=1 to run the Webhook-to-Jenkins hook test")
	}

	hookKey := "com.nerdwin15.stash-stash-webhook-jenkins:jenkinsPostReceiveHook"
	settingsJSON := `{"jenkinsBase":"http://jenkins.example.com","cloneType":"http","omitHashCode":false,"omitBranchName":false}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Configure the Jenkins webhook hook at project level
			{
				Config: testAccProjectHookConfig(hookKey, true, settingsJSON),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "hook_key", hookKey),
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "enabled", "true"),
				),
			},
			// Step 2: Import
			{
				ResourceName:      "bitbucketdc_project_hook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update Jenkins base URL
			{
				Config: testAccProjectHookConfig(hookKey, true,
					`{"jenkinsBase":"http://jenkins-new.example.com","cloneType":"http","omitHashCode":false,"omitBranchName":false}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bitbucketdc_project_hook.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccProjectHookConfig(hookKey string, enabled bool, settingsJSON string) string {
	return fmt.Sprintf(`
resource "bitbucketdc_project" "test" {
  key             = "PHKTEST"
  name            = "Project Hook Test"
  prevent_destroy = false
}

resource "bitbucketdc_project_hook" "test" {
  project_key   = bitbucketdc_project.test.key
  hook_key      = %q
  enabled       = %t
  settings_json = %q
}
`, hookKey, enabled, settingsJSON)
}
