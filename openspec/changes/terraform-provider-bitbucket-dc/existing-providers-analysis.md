# Existing Terraform Providers Analysis

## DrFaust92/terraform-provider-bitbucket

**Repository:** https://github.com/DrFaust92/terraform-provider-bitbucket
**Registry:** https://registry.terraform.io/providers/DrFaust92/bitbucket

### Summary

**This provider is for Bitbucket Cloud ONLY, NOT for Bitbucket Data Center.**

### Evidence

1. **Repository Description:** Explicitly states "Terraform Bitbucket Cloud provider"
2. **Documentation:** References Bitbucket Cloud authentication (bitbucket.org/account/settings)
3. **API Target:** Uses Bitbucket Cloud REST API v2.0
4. **Authentication:** Uses Cloud-specific auth methods (App passwords, OAuth)

### Key Differences: Bitbucket Cloud vs Data Center

| Aspect | Bitbucket Cloud | Bitbucket Data Center |
|--------|----------------|----------------------|
| **Hosting** | SaaS (bitbucket.org) | Self-hosted/on-premise |
| **API** | REST API 2.0 (different structure) | REST API 1.0, 2.0, 3.0 (OpenAPI spec) |
| **API Base URL** | api.bitbucket.org | Your own domain (e.g., bitbucket.company.com) |
| **Authentication** | App passwords, OAuth | Personal Access Tokens, HTTP Basic, OAuth |
| **Features** | Cloud-specific features | Data Center-specific features (clustering, etc.) |
| **API Endpoints** | `/2.0/repositories`, `/2.0/workspaces` | `/rest/api/1.0/projects`, `/rest/api/1.0/repos` |

### Why This Provider Doesn't Work for Data Center

1. **Incompatible API Structure:**
   - Cloud uses API 2.0 with different endpoint paths
   - Data Center uses API 1.0/3.0 with different resource structure
   - Example Cloud: `GET /2.0/repositories/{workspace}/{repo_slug}`
   - Example DC: `GET /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}`

2. **Different Authentication:**
   - Cloud uses app passwords and OAuth with workspace access
   - Data Center uses Personal Access Tokens and project-based permissions

3. **Different Resource Model:**
   - Cloud uses "workspaces" as the top-level container
   - Data Center uses "projects" as the top-level container

4. **Different Features:**
   - Cloud has pipelines, deployments, workspace-level settings
   - Data Center has clustering, mesh, project-level hooks, branch permissions

### Provider Maintenance Status

**Active and Well-Maintained:**
- 67 releases (latest v2.50.0 - October 2025)
- 547 commits
- 49 stars, 35 open issues
- Active community support
- Licensed under MPL-2.0

### Resources Provided (Cloud-Specific)

Based on the provider's focus on Bitbucket Cloud:
- Repositories (Cloud workspaces)
- Projects (Cloud team projects)
- Branch restrictions
- Deploy keys
- Webhooks
- Repository variables
- SSH keys
- And more Cloud-specific resources

## Conclusion

**There is NO existing Terraform provider for Bitbucket Data Center.**

The only available provider (DrFaust92/terraform-provider-bitbucket) is specifically designed for Bitbucket Cloud and cannot be used with Bitbucket Data Center due to fundamental API and architecture differences.

**This confirms the need for a new, dedicated Terraform provider for Bitbucket Data Center as proposed in this change.**

## Alternative Providers Checked

### 1. Atlassian Official Provider
- **Status:** Does NOT exist
- Atlassian does not provide an official Terraform provider for Bitbucket Data Center
- They provide providers for Jira Cloud, but not Bitbucket DC

### 2. Community Providers
- **DrFaust92/bitbucket:** Cloud only (analyzed above)
- **aeirola/bitbucket:** Archived/deprecated (superseded by DrFaust92's fork)
- **No other active providers found**

### 3. Search Results
- Terraform Registry search for "bitbucket": Only shows Cloud providers
- GitHub search for "terraform-provider-bitbucket-server" or "terraform-provider-bitbucket-datacenter": No results
- No mentions in Terraform community forums or HashiCorp discussions

## Impact on Our Proposal

**This analysis confirms:**

1. ✅ **No existing solution** - Our provider fills a genuine gap
2. ✅ **Cannot adapt existing provider** - Cloud and DC are fundamentally different
3. ✅ **Clear market need** - Organizations using Bitbucket DC need IaC
4. ✅ **Greenfield development** - No code to fork or extend
5. ✅ **Unique offering** - Will be the ONLY Terraform provider for Bitbucket DC

**Recommendations:**

1. **Naming:** Use clear distinction in naming:
   - Our provider: `terraform-provider-bitbucketdc` (emphasize DC)
   - Existing: `terraform-provider-bitbucket` (Cloud)

2. **Documentation:** Clearly state "for Bitbucket Data Center" in all docs to avoid confusion

3. **Registry:** If publishing publicly later, ensure description explicitly states "Data Center" vs "Cloud"

4. **Community:** Consider reaching out to DrFaust92 or HashiCorp to cross-reference providers and avoid user confusion

## Additional Resources

### Bitbucket API Documentation

**Cloud API:**
- https://developer.atlassian.com/cloud/bitbucket/rest/
- Uses REST API 2.0
- Different endpoint structure

**Data Center API:**
- https://docs.atlassian.com/bitbucket-server/rest/
- Uses REST API 1.0 and 3.0
- OpenAPI specification available
- https://dac-static.atlassian.com/server/bitbucket/10.0.swagger.v3.json

### Key Takeaway

**Bitbucket Cloud and Bitbucket Data Center are SEPARATE products with INCOMPATIBLE APIs. They require completely different Terraform providers.**
