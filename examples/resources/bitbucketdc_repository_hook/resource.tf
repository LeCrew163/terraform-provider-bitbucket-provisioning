resource "bitbucketdc_repository_hook" "jenkins" {
  project_key     = "PLAT"
  repository_slug = "platform-api"
  hook_key        = "com.nerdwin15.stash-stash-webhook-jenkins:postReceiveHook"
  enabled         = true
  settings_json = jsonencode({
    jenkinsBase = "https://jenkins.example.com"
    cloneType   = "ssh"
  })
}
