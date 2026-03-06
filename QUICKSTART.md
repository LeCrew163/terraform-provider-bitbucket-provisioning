# Quick Start Guide - Terraform Bitbucket DC Provider

## Build the Provider

```bash
# Install dependencies
go mod download

# Build the provider
make build

# Install locally for testing
make install
```

The provider will be installed to:
`~/.terraform.d/plugins/registry.terraform.io/LeCrew163/bitbucket-provisioning/0.10.0/darwin_arm64/`

## Create Your First Project

Create a file named `main.tf`:

```hcl
terraform {
  required_providers {
    bitbucketdc = {
      source = "LeCrew163/bitbucket-provisioning"
      version = "~> 0.10"
    }
  }
}

provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token

  # Alternative: use username/password
  # username = var.bitbucket_username
  # password = var.bitbucket_password
}

variable "bitbucket_token" {
  description = "Bitbucket Personal Access Token"
  type        = string
  sensitive   = true
}

resource "bitbucketdc_project" "example" {
  key         = "MYPROJ"
  name        = "My Project"
  description = "Created with Terraform"
  public      = false
}

output "project_id" {
  value = bitbucketdc_project.example.id
}

output "project_key" {
  value = bitbucketdc_project.example.key
}
```

Create a `terraform.tfvars` file:

```hcl
bitbucket_token = "your-token-here"
```

## Run Terraform

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Create the project
terraform apply

# View outputs
terraform output

# Import existing project
terraform import bitbucketdc_project.example EXISTINGKEY

# Destroy resources
terraform destroy
```

## Using Environment Variables

Instead of using variables, you can use environment variables:

```bash
export BITBUCKET_BASE_URL="https://bitbucket.example.com"
export BITBUCKET_TOKEN="your-token-here"

# Or with username/password
export BITBUCKET_USERNAME="your-username"
export BITBUCKET_PASSWORD="your-password"

terraform plan
terraform apply
```

## Troubleshooting

### Provider Not Found

If Terraform can't find the provider, make sure you've run `make install` and the provider is in the correct directory.

### Connection Errors

Check that:
- The `base_url` is correct and accessible
- Your token/credentials are valid
- The Bitbucket instance is running
- There are no firewall issues

### Project Key Validation

Project keys must:
- Be 2-128 characters long
- Start with an uppercase letter
- Contain only uppercase letters, numbers, and underscores
- Examples: `MYPROJ`, `MY_PROJECT`, `PROJ123`

## Next Steps

Once the basic setup works:
1. Read the full documentation in `docs/`
2. Explore all available resources and data sources
3. Check the [GitHub Issues](https://github.com/LeCrew163/bitbucket-provisioning/issues) for known problems

## Getting Help

- Review `README.md` for comprehensive documentation
- Check [GitHub Issues](https://github.com/LeCrew163/bitbucket-provisioning/issues) for known problems
