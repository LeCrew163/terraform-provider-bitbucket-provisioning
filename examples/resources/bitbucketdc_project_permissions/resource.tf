resource "bitbucketdc_project_permissions" "platform" {
  project_key = "PLAT"

  group {
    name       = "platform-developers"
    permission = "PROJECT_WRITE"
  }

  group {
    name       = "platform-leads"
    permission = "PROJECT_ADMIN"
  }

  user {
    name       = "jsmith"
    permission = "PROJECT_READ"
  }
}
