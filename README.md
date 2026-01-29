# Terraform Provider for Bitbucket Data Center

A Terraform provider for managing Bitbucket Data Center resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21
- Bitbucket Data Center >= 10.0

## Features

- **Projects**: Create and manage Bitbucket projects
- **Repositories**: Manage repositories within projects
- **Permissions**: Control project and repository access permissions
- **Branch Permissions**: Configure branch protection rules
- **Access Keys**: Manage SSH access keys for projects
- **Data Sources**: Query existing Bitbucket resources

## Building the Provider

Clone the repository and build the provider:

```bash
git clone ssh://git@bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning.git
cd bitbucket-provisioning
make build
```

## Installing the Provider for Local Development

To install the provider locally for testing:

```bash
make install
```

This will build and install the provider to your local Terraform plugins directory.

## Using the Provider

```hcl
terraform {
  required_providers {
    bitbucketdc = {
      source  = "colab.internal.sldo.cloud/alpina/bitbucket-dc"
      version = "~> 0.1"
    }
  }
}

provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token
}

resource "bitbucketdc_project" "example" {
  key         = "MYPROJ"
  name        = "My Project"
  description = "Example project"
  visibility  = "private"
}

resource "bitbucketdc_repository" "example" {
  project_key = bitbucketdc_project.example.key
  slug        = "my-repo"
  name        = "My Repository"
  description = "Example repository"
}
```

## Authentication

The provider supports two authentication methods:

### Personal Access Token (Recommended)

```hcl
provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token
}
```

Environment variable: `BITBUCKET_TOKEN`

### HTTP Basic Authentication

```hcl
provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  username = var.bitbucket_username
  password = var.bitbucket_password
}
```

Environment variables: `BITBUCKET_USERNAME`, `BITBUCKET_PASSWORD`

## Development

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires Bitbucket instance)
export BITBUCKET_BASE_URL="https://bitbucket.example.com"
export BITBUCKET_TOKEN="your-token"
make testacc
```

### Generating Documentation

```bash
make docs
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

See [LICENSE](LICENSE) for license information.

## Support

- Issues: [Bitbucket Issues](https://bitbucket.colab.internal.sldo.cloud/projects/ALPINA-OPERATION/repos/bitbucket-provisioning/issues)
- Documentation: See README.md and IMPLEMENTATION_STATUS.md in this repository
