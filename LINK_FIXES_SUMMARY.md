# Link Fixes Summary

## What These Links Are Used For

### 1. **GitHub Issues Link** (README.md:142)
- **Purpose**: Bug tracking and support
- **Old**: `https://github.com/alpina/terraform-provider-bitbucket-dc/issues`
- **New**: `https://bitbucket.colab.internal.sldo.cloud/projects/ALPINA-OPERATION/repos/bitbucket-provisioning/issues`
- **Usage**: Users click this link to report bugs or request features

### 2. **Terraform Registry Documentation** (README.md:143)
- **Purpose**: Provider documentation on public Terraform Registry
- **Old**: `https://registry.terraform.io/providers/alpina/bitbucket-dc/latest/docs`
- **New**: Removed - now references internal documentation (README.md and IMPLEMENTATION_STATUS.md)
- **Usage**: For public providers, this links to auto-generated docs on Terraform Registry. Since you're hosting privately, this isn't applicable.

### 3. **Go Module Path** (go.mod:1)
- **Purpose**: Identifies the Go module for imports
- **Old**: `github.com/alpina/terraform-provider-bitbucket-dc`
- **New**: `bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning`
- **Usage**: Used in all Go import statements throughout the codebase

### 4. **Provider Address** (main.go:22)
- **Purpose**: Tells Terraform where to find the provider
- **Old**: `registry.terraform.io/alpina/bitbucket-dc`
- **New**: `colab.internal.sldo.cloud/alpina/bitbucket-dc`
- **Usage**: This is the namespace Terraform uses when you specify the provider in your `.tf` files

### 5. **Installation Path** (Makefile:9-10)
- **Purpose**: Local installation directory for development
- **Old**: `~/.terraform.d/plugins/registry.terraform.io/alpina/bitbucket-dc/0.1.0/darwin_arm64/`
- **New**: `~/.terraform.d/plugins/colab.internal.sldo.cloud/alpina/bitbucket-dc/0.1.0/darwin_arm64/`
- **Usage**: Where `make install` places the provider for local Terraform testing

### 6. **Git Clone URL** (README.md:25)
- **Purpose**: Instructions for cloning the repository
- **Old**: `https://github.com/alpina/terraform-provider-bitbucket-dc`
- **New**: `ssh://git@bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning.git`
- **Usage**: Command developers use to clone the repository

### 7. **Terraform Provider Source** (README.md:46, QUICKSTART.md:27, IMPLEMENTATION_STATUS.md:305)
- **Purpose**: How users specify the provider in their Terraform configuration
- **Old**: `source = "registry.terraform.io/alpina/bitbucket-dc"`
- **New**: `source = "colab.internal.sldo.cloud/alpina/bitbucket-dc"`
- **Usage**: Used in the `required_providers` block of Terraform configurations

## Files Modified

1. **go.mod** - Module path
2. **main.go** - Provider address and import
3. **internal/provider/provider.go** - Import path
4. **internal/provider/resource_project.go** - Import paths
5. **internal/client/client.go** - Import path
6. **internal/client/generated/test/*.go** - Import paths (21 files)
7. **README.md** - Clone URL, provider source, support links
8. **Makefile** - Installation path
9. **QUICKSTART.md** - Installation path, provider source
10. **IMPLEMENTATION_STATUS.md** - Provider source

## Testing Status

✅ **Compilation**: Successfully builds with `go build`
✅ **Dependencies**: `go mod tidy` completes without errors
✅ **Binary**: Provider binary runs and shows help text
⏳ **Runtime**: No unit tests exist yet (as noted in IMPLEMENTATION_STATUS.md)
⏳ **Integration**: Requires Bitbucket Data Center instance for acceptance tests

## About Terraform Registry

**Public Terraform Registry** (`registry.terraform.io`):
- Hosts publicly available providers
- Auto-generates documentation from code
- Requires specific publishing process
- Not necessary for private/internal providers

**Private Provider Hosting**:
- Use custom namespace (e.g., `colab.internal.sldo.cloud/alpina/bitbucket-dc`)
- Install locally with `make install` for development
- Can set up private Terraform Registry for production use
- Documentation lives in your repository

## How to Use This Provider Now

### Development/Testing

1. Build and install locally:
```bash
make build
make install
```

2. Use in Terraform:
```hcl
terraform {
  required_providers {
    bitbucketdc = {
      source = "colab.internal.sldo.cloud/alpina/bitbucket-dc"
      version = "~> 0.1"
    }
  }
}

provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token
}
```

3. Run Terraform:
```bash
terraform init
terraform plan
terraform apply
```

### Production (Future)

For production use, consider:
1. Setting up a private Terraform Registry
2. Using a CI/CD pipeline to build releases
3. Implementing GPG signing for releases
4. Creating versioned releases with proper tagging

## Next Steps

1. ✅ **Fixed**: All GitHub and public registry references
2. ✅ **Fixed**: Go module path updated
3. ✅ **Fixed**: Provider address updated
4. ⏳ **TODO**: Write unit tests for provider
5. ⏳ **TODO**: Test against real Bitbucket Data Center instance
6. ⏳ **TODO**: Set up CI/CD pipeline in Jenkins
7. ⏳ **TODO**: Consider private Terraform Registry for team use
