data "bitbucketdc_group" "developers" {
  name = "platform-developers"
}

# Use the group name in a permission resource
resource "bitbucketdc_project_permissions" "platform" {
  project_key = "PLAT"

  group {
    name       = data.bitbucketdc_group.developers.name
    permission = "PROJECT_WRITE"
  }
}
