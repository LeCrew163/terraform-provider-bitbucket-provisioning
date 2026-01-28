# Quick Start: Reusing Components from Cloud Provider

## TL;DR

✅ **You can reuse ~35% of work from the Cloud provider**
- Build infrastructure, patterns, and templates
- Saves 10-15 days of development time

❌ **You must build ~65% from scratch**
- Complete API client (different APIs)
- All resource logic (different endpoints)

## What to Copy/Adapt

### 1. Direct Copy (with attribution)

These files can be copied with minimal changes:

```bash
# From: https://github.com/DrFaust92/terraform-provider-bitbucket

# Build and Release
cp .goreleaser.yml <your-repo>/        # Change binary name
cp GNUmakefile <your-repo>/            # Update TEST variable
cp .gitignore <your-repo>/             # No changes needed

# Code Quality
cp .golangci.yml <your-repo>/          # No changes needed
cp .markdownlint.yml <your-repo>/      # No changes needed

# CI/CD - NOT APPLICABLE
# Cloud provider uses GitHub Actions, you use Jenkins
# See jenkins-pipeline-example.md for Jenkinsfile examples
```

**Attribution:** Add to README:
```markdown
Build infrastructure adapted from [DrFaust92/terraform-provider-bitbucket](https://github.com/DrFaust92/terraform-provider-bitbucket) (MPL-2.0)
```

---

### 2. Adapt and Reference

Use these as patterns/references, don't copy directly:

#### Provider Configuration Pattern

**Reference:** `bitbucket/provider.go` in Cloud provider

**Adapt for:**
- Plugin Framework (not SDK v2)
- DC authentication (PAT, not app passwords)
- DC base URL handling

```go
// Cloud provider (SDK v2) - REFERENCE ONLY
func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "username": { /* ... */ },
        },
        // ...
    }
}

// Your provider (Plugin Framework) - NEW CODE
func (p *provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "token": { /* ... */ },
        },
    }
}
```

---

#### Resource Implementation Pattern

**Reference:** Any `resource_*.go` file in Cloud provider

**Patterns to use:**
- CRUD operation structure
- Error handling approach (adapt for DC errors)
- State management patterns
- Import ID patterns

**Must rewrite:**
- All API calls (DC endpoints)
- All response parsing (DC response structure)
- Validation (DC rules: uppercase keys vs lowercase slugs)

```go
// Pattern from Cloud provider (REFERENCE)
func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    // 1. Get config values
    // 2. Call API
    // 3. Handle errors
    // 4. Set state
    // 5. Read to populate full state
}

// Your implementation (NEW CODE, same pattern)
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Get plan values
    // 2. Call DC API (different endpoint)
    // 3. Handle errors (different error structure)
    // 4. Set state
    // 5. Read to populate full state
}
```

---

#### Testing Pattern

**Reference:** Any `*_test.go` file in Cloud provider

**Reusable patterns:**
```go
// Test structure (ADAPT)
func TestAccProject_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        // ... your provider factories
        Steps: []resource.TestStep{
            {
                Config: testAccProjectConfig_basic(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("...", "key", "TEST"),
                ),
            },
            // Import test
            {
                ImportState: true,
                ImportStateVerify: true,
            },
        },
    })
}
```

---

#### Documentation Templates

**Reference:** `templates/` directory in Cloud provider

**Adapt:**
- Change `bitbucket_*` → `bitbucketdc_*`
- Update examples for DC (projects not workspaces)
- Update authentication sections (PAT not app passwords)

---

### 3. Build from Scratch

Do NOT try to adapt these - build new:

#### API Client
```bash
# Generate from DC OpenAPI spec
openapi-generator-cli generate \
    -i specs/bitbucket-datacenter-openapi.json \
    -g go \
    -o internal/client/generated
```

❌ Do NOT use Cloud provider's API client
- Completely different endpoints
- Different authentication
- Different response structures

---

#### Resource Implementations

❌ Do NOT copy resource code from Cloud provider
- Reference patterns only
- API calls are 100% different
- Response parsing is 100% different
- Validation rules are different

**Example differences:**

| Cloud | Data Center |
|-------|------------|
| `/2.0/workspaces/{workspace}` | `/rest/api/1.0/projects` |
| Workspace slugs (lowercase) | Project keys (UPPERCASE) |
| OAuth tokens | Personal Access Tokens |

---

## Step-by-Step Setup

### Phase 1: Setup Infrastructure (Day 1)

```bash
# 1. Create new repo
git init terraform-provider-bitbucketdc
cd terraform-provider-bitbucketdc

# 2. Copy build files from Cloud provider
# (See "Direct Copy" section above)

# 3. Update provider name in copied files
sed -i 's/bitbucket/bitbucketdc/g' .goreleaser.yml
sed -i 's/bitbucket/bitbucketdc/g' GNUmakefile

# 4. Initialize Go module
go mod init github.com/your-org/terraform-provider-bitbucketdc

# 5. Add Plugin Framework dependency
go get github.com/hashicorp/terraform-plugin-framework
```

### Phase 2: Generate API Client (Day 2)

```bash
# 1. Download DC OpenAPI spec
curl -o specs/bitbucket-dc.json \
    https://dac-static.atlassian.com/server/bitbucket/10.0.swagger.v3.json

# 2. Generate Go client
openapi-generator-cli generate \
    -i specs/bitbucket-dc.json \
    -g go \
    -o internal/client/generated

# 3. Create wrapper
# Reference Cloud provider's client.go pattern
# Build new wrapper for DC API
```

### Phase 3: Implement Provider (Day 3-5)

```bash
# Reference Cloud provider's provider.go
# Implement in Plugin Framework syntax
# Adapt authentication for DC
```

### Phase 4: Implement Resources (Day 6-30)

```bash
# Reference Cloud resource patterns
# Implement CRUD for DC API
# Write tests using Cloud provider test patterns
```

---

## Time Savings

| Without Cloud Reference | With Cloud Reference | Savings |
|------------------------|---------------------|---------|
| 60 days | 52 days | 8 days (13%) |

**Where savings come from:**
- ✅ 1 day: Build infrastructure (GoReleaser, Makefile)
- ✅ 1.5 days: Provider patterns (reference)
- ✅ 2.5 days: Resource patterns (reference)
- ✅ 2.5 days: Testing patterns (reference)
- ✅ 0.5 days: Documentation templates (adapt)
- ✅ 1 day: Avoiding common pitfalls

**Additional time required:**
- ❌ +1 day: Jenkins CI/CD setup (vs copying GitHub Actions)

---

## Common Pitfalls to Avoid

### ❌ Don't: Copy API calls
```go
// Cloud provider - DON'T COPY
client.Get(fmt.Sprintf("/2.0/workspaces/%s", workspace))
```

### ✅ Do: Reference patterns
```go
// Pattern: "Get config, call API, handle errors, set state"
// Your DC code:
client.Get(fmt.Sprintf("/rest/api/1.0/projects/%s", projectKey))
```

---

### ❌ Don't: Copy validation rules
```go
// Cloud: lowercase slugs
validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]+$`), "...")
```

### ✅ Do: Write DC validation
```go
// DC: UPPERCASE keys
validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`), "...")
```

---

### ❌ Don't: Copy SDK v2 syntax
```go
// Cloud provider uses SDK v2
func resourceProject() *schema.Resource { /* ... */ }
```

### ✅ Do: Use Plugin Framework
```go
// Use Plugin Framework
type projectResource struct { /* ... */ }
func (r *projectResource) Schema(...) { /* ... */ }
```

---

## License Compliance

**Cloud Provider License:** MPL-2.0 (Mozilla Public License 2.0)

**Requirements:**
- ✅ Can reference and learn from code
- ✅ Can copy build infrastructure with attribution
- ⚠️ If copying significant code, must retain MPL-2.0 for those files
- ⚠️ File-level licensing (not project-level)

**Our Approach:**
- Copy build files (with attribution) ✅
- Reference patterns (no license requirement) ✅
- Write all provider/resource code from scratch ✅
- Choose our own license (MPL-2.0 compatible recommended) ✅

**Attribution in README:**
```markdown
## Acknowledgments

Build infrastructure adapted from [DrFaust92/terraform-provider-bitbucket](https://github.com/DrFaust92/terraform-provider-bitbucket).
Licensed under MPL-2.0.
```

---

## Resources

- **Cloud Provider Repo:** https://github.com/DrFaust92/terraform-provider-bitbucket
- **DC API Docs:** https://docs.atlassian.com/bitbucket-server/rest/
- **DC OpenAPI Spec:** https://dac-static.atlassian.com/server/bitbucket/10.0.swagger.v3.json
- **Plugin Framework Docs:** https://developer.hashicorp.com/terraform/plugin/framework
- **Cloud vs DC Analysis:** [existing-providers-analysis.md](./existing-providers-analysis.md)
- **Detailed Reusability:** [reusable-components-analysis.md](./reusable-components-analysis.md)

---

## Next Steps

1. ✅ Review this guide
2. ✅ Clone/reference Cloud provider repo
3. ✅ Copy build infrastructure files
4. ✅ Generate DC API client
5. ✅ Start provider implementation (reference patterns)
6. ✅ Implement resources (new code, DC API)
7. ✅ Write tests (adapt patterns)
8. ✅ Generate documentation (adapt templates)
