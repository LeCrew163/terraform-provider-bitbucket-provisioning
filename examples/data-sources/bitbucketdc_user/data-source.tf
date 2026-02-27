data "bitbucketdc_user" "alice" {
  slug = "alice"
}

output "display_name" {
  value = data.bitbucketdc_user.alice.display_name
}
