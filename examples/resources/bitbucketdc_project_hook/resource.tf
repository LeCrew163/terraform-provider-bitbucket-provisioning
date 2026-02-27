resource "bitbucketdc_project_hook" "jenkins" {
  project_key = "PLAT"
  hook_key    = "com.nerdwin15.stash-stash-webhook-jenkins:postReceiveHook"
  enabled     = true
  settings_json = jsonencode({
    jenkinsBase = "https://jenkins.example.com"
    cloneType   = "ssh"
    branchOptions = ""
    branchOptionsBranches = ""
  })
}
