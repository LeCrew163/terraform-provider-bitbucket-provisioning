variable "bitbucket_url" {
  description = "Bitbucket Data Center base URL"
  type        = string
  default     = "http://localhost:7990"
}

variable "bitbucket_username" {
  description = "Bitbucket admin username"
  type        = string
  default     = "admin"
}

variable "bitbucket_password" {
  description = "Bitbucket admin password"
  type        = string
  sensitive   = true
  default     = "admin"
}

variable "insecure_skip_verify" {
  description = "Skip TLS certificate verification (for local/self-signed certs)"
  type        = bool
  default     = false
}
