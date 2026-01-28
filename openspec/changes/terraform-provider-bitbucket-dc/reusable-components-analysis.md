# Reusable Components Analysis

## Can We Reuse Code from the Cloud Provider?

**Short Answer:** Yes, but not the API/resource logic. We can reuse **infrastructure, patterns, and tooling** (30-40% of the work), but must implement **all API client and resource logic** from scratch (60-70% of the work).

## What CAN Be Reused ✅

### 1. Project Structure & Build Infrastructure (HIGH REUSE)

**GoReleaser Configuration** - Can reuse ~95%
```yaml
# .goreleaser.yml - Almost identical
version: 2
builds:
  - env: [CGO_ENABLED=0]
    mod_timestamp: '{{ .CommitTimestamp }}'
    flags: [-trimpath]
    ldflags: '-s -w -X main.version={{.Version}}'
    goos: [freebsd, windows, linux, darwin]
    goarch: [amd64, '386', arm, arm64]
    # Same platform matrix, signing, checksums
```

**Benefit:** Ready-to-use multi-platform build and release configuration.

---

**Makefile Targets** - Can adapt ~90%
```makefile
# GNUmakefile - Same targets, same patterns
default: build

build: fmtcheck
	go install

test: fmtcheck
	go test -i $(TEST) || exit 1
	go test $(TEST) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v -parallel=10 -timeout 120m
```

**Benefit:** Standard testing, building, formatting targets work identically.

---

**CI/CD Workflows** - Jenkins (custom implementation needed)
```groovy
// Jenkinsfile - Must create from scratch for Jenkins
pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Test') {
            steps {
                sh 'make test'
            }
        }

        stage('Release') {
            when {
                tag pattern: "v*", comparator: "REGEXP"
            }
            steps {
                sh 'goreleaser release --clean'
            }
        }
    }
}
```

**Note:** GitHub Actions workflows from Cloud provider are NOT applicable - Jenkins requires custom Jenkinsfile.

---

**Additional Build Files:**
- `.gitignore` - Can reuse 100% (same Go project patterns)
- `.golangci.yml` - Can reuse ~90% (linter configuration)
- `LICENSE` - Choose appropriate license (their: MPL-2.0)
- `Jenkinsfile` - Must create from scratch (Cloud provider uses GitHub Actions)

**Time Saved:** 1 day of setup and configuration (Jenkins setup adds ~1 day)

---

### 2. Provider Core Patterns (MODERATE REUSE)

**Provider Configuration Structure** - Can adapt ~60%

```go
// Similar pattern, different auth details
func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "base_url": {
                Type:        schema.TypeString,
                Required:    true,
                DefaultFunc: schema.EnvDefaultFunc("BITBUCKET_BASE_URL", nil),
                Description: "Bitbucket Data Center URL",
            },
            "token": {
                Type:        schema.TypeString,
                Optional:    true,
                Sensitive:   true,
                DefaultFunc: schema.EnvDefaultFunc("BITBUCKET_TOKEN", ""),
                Description: "Personal Access Token",
                ConflictsWith: []string{"username", "password"},
            },
            // ... similar structure, different specifics
        },
        ResourcesMap: map[string]*schema.Resource{
            "bitbucketdc_project":     resourceProject(),
            "bitbucketdc_repository":  resourceRepository(),
            // ...
        },
        DataSourcesMap: map[string]*schema.Resource{
            "bitbucketdc_project":    dataSourceProject(),
            // ...
        },
        ConfigureContextFunc: providerConfigure,
    }
}
```

**What to Adapt:**
- Authentication options (PAT instead of Cloud app passwords)
- Base URL handling (Data Center vs Cloud API)
- Resource and data source names
- Configuration validation

**What Stays Same:**
- Schema structure and patterns
- Environment variable defaults
- Sensitive field marking
- ConflictsWith validation
- ConfigureContextFunc pattern

**Time Saved:** 1 day (vs writing from scratch)

---

**Authentication Pattern** - Can adapt ~50%

```go
// Cloud provider pattern (adapt, not copy)
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
    config := &ProviderConfig{
        BaseURL: d.Get("base_url").(string),
    }

    // Different auth logic, same pattern
    if token := d.Get("token").(string); token != "" {
        // Configure PAT auth (DC-specific)
    } else if username := d.Get("username").(string); username != "" {
        // Configure Basic auth (DC-specific)
    }

    // Similar client initialization pattern
    client, err := initializeClient(config)
    // ... error handling pattern

    return config, nil
}
```

**Time Saved:** 0.5 day (pattern reference)

---

### 3. Resource Implementation Patterns (LOW-MODERATE REUSE)

**IMPORTANT:** SDK v2 vs Plugin Framework difference means we CAN'T copy code directly, but we CAN reuse **patterns and structure**.

**Cloud Provider Uses:** Terraform SDK v2
**Our Provider Uses:** Terraform Plugin Framework (modern, different API)

**Resource Structure Pattern** - Can reference ~40%

```go
// Pattern is similar, but SDK v2 vs Framework syntax differs

// SDK v2 (Cloud provider):
func resourceProject() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceProjectCreate,
        ReadContext:   resourceProjectRead,
        UpdateContext: resourceProjectUpdate,
        DeleteContext: resourceProjectDelete,
        Importer: &schema.ResourceImporter{
            StateContext: resourceProjectImport,
        },
        Schema: map[string]*schema.Schema{ /* ... */ },
    }
}

// Plugin Framework (our provider):
type projectResource struct {
    client *client.Client
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{ /* ... */ },
    }
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Similar logic, different API
}
```

**Reusable Patterns:**
- ✅ CRUD operation structure and flow
- ✅ Error handling approach
- ✅ State management patterns
- ✅ Import ID patterns
- ✅ Validation logic structure

**Must Rewrite:**
- ❌ Syntax (SDK v2 → Plugin Framework)
- ❌ Schema definition (different API)
- ❌ All API calls (Cloud → DC)
- ❌ Response parsing (Cloud → DC)

**Time Saved:** 2-3 days (reference patterns, avoid pitfalls)

---

**Error Handling Pattern** - Can adapt ~70%

```go
// Similar pattern, adapt for DC API errors
func handleClientError(resp *http.Response, err error) diag.Diagnostics {
    if err != nil {
        return diag.FromErr(err)
    }

    switch resp.StatusCode {
    case http.StatusNotFound:
        // Remove from state, don't error
        return nil
    case http.StatusForbidden:
        return diag.Errorf("Permission denied: %s", parseErrorMessage(resp))
    case http.StatusConflict:
        return diag.Errorf("Resource already exists: %s", parseErrorMessage(resp))
    default:
        return diag.Errorf("API error %d: %s", resp.StatusCode, parseErrorMessage(resp))
    }
}
```

**Time Saved:** 0.5 day

---

### 4. Testing Patterns (MODERATE REUSE)

**Test Structure** - Can reuse ~60%

```go
// Similar acceptance test structure
func TestAccProject_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, // Our version
        Steps: []resource.TestStep{
            {
                Config: testAccProjectConfig_basic("TEST", "Test Project"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bitbucketdc_project.test", "key", "TEST"),
                    resource.TestCheckResourceAttr("bitbucketdc_project.test", "name", "Test Project"),
                ),
            },
            // Import test
            {
                ResourceName:      "bitbucketdc_project.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

**Reusable Patterns:**
- ✅ Test case structure
- ✅ PreCheck pattern
- ✅ Config generation functions
- ✅ Check functions
- ✅ Import test pattern
- ✅ Update test pattern

**Time Saved:** 2-3 days (avoid test pattern discovery)

---

### 5. Documentation Templates (MODERATE REUSE)

**tfplugindocs Templates** - Can reuse ~50%

```markdown
<!-- templates/resources/project.md.tmpl -->
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
  {{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Type}} ({{.Name}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/resources/bitbucketdc_project/resource.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the following syntax:

{{ codefile "shell" "examples/resources/bitbucketdc_project/import.sh" }}
```

**Time Saved:** 1 day

---

## What CANNOT Be Reused ❌

### 1. API Client - 0% Reusable

**Why:** Completely different APIs

```
Cloud API:
  GET /2.0/repositories/{workspace}/{repo_slug}
  POST /2.0/workspaces/{workspace}/projects

Data Center API:
  GET /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}
  POST /rest/api/1.0/projects
```

**Must Build:** 100% new API client from DC OpenAPI spec

---

### 2. Resource Logic - ~10% Reusable

**Why:** Different API responses, different data structures

```go
// Cloud: Workspace-based
type CloudRepo struct {
    Workspace struct {
        Slug string `json:"slug"`
    } `json:"workspace"`
    Slug string `json:"slug"`
}

// DC: Project-based
type DCRepo struct {
    Project struct {
        Key string `json:"key"`
    } `json:"project"`
    Slug string `json:"slug"`
}
```

**Can Reference:** Patterns and logic flow
**Must Rewrite:** All API calls and response handling

---

### 3. Data Sources - ~10% Reusable

Same as resources - patterns are reusable, but all API calls must be rewritten.

---

### 4. Validation Logic - ~30% Reusable

**Some validators are similar:**
```go
// Similar: Permission levels (though names might differ)
validation.StringInSlice([]string{"PROJECT_READ", "PROJECT_WRITE", "PROJECT_ADMIN"}, false)
```

**Most are different:**
```go
// Cloud: workspace slugs (lowercase)
// DC: project keys (UPPERCASE)
// Completely different validation regex
```

---

## Reusability Summary

| Component | Reusability | Time Saved | Effort Level |
|-----------|------------|------------|--------------|
| **Build Infrastructure** | 80% | 1 day | Copy + adapt |
| GoReleaser, Makefile (no GitHub Actions) | | | |
| **Provider Core** | 50% | 1.5 days | Reference + adapt |
| Configuration, auth patterns | | | |
| **Resource Patterns** | 40% | 2-3 days | Reference patterns |
| CRUD structure, error handling | | | |
| **Testing Patterns** | 60% | 2-3 days | Adapt structure |
| Test cases, helpers | | | |
| **Documentation** | 50% | 1 day | Adapt templates |
| Templates, examples | | | |
| **API Client** | 0% | 0 days | Build from scratch |
| All API calls, client | | | |
| **Resource Logic** | 10% | 0.5 days | Rewrite everything |
| API interactions | | | |
| **CI/CD (Jenkins)** | 0% | 0 days | Build from scratch |
| Jenkins pipelines, not GitHub Actions | | | |
| **TOTAL** | ~32% | **8-10 days** | Mixed |

---

## Recommended Approach

### Phase 1: Setup (Reuse Heavy)

1. **Fork or Reference** the Cloud provider repo
   - Don't fork directly (confusing for users)
   - Copy build infrastructure files
   - Copy and adapt: `.goreleaser.yml`, `GNUmakefile`, workflows

2. **Use as Reference** for patterns
   - Keep Cloud provider open for pattern reference
   - Don't copy code directly (different frameworks)
   - Adapt patterns to Plugin Framework

3. **Copy and Adapt** non-code files
   - CI/CD workflows
   - Documentation templates (adjust for DC)
   - Testing helpers (adjust for DC)

### Phase 2: Implementation (New Code)

1. **Generate API Client** from DC OpenAPI spec
   - 100% new, cannot reuse Cloud client
   - Use openapi-generator for Go

2. **Implement Provider Core**
   - Reference Cloud provider patterns
   - Rewrite in Plugin Framework syntax
   - Adapt auth for DC (PAT vs app passwords)

3. **Implement Resources**
   - Reference Cloud resource structure
   - Rewrite all CRUD operations for DC API
   - Adapt validation for DC (project keys vs workspace slugs)

4. **Implement Tests**
   - Copy test structure patterns
   - Rewrite test configs for DC resources
   - Adapt PreCheck for DC instance

---

## Time Savings Analysis

**Without Reference to Cloud Provider:** ~60 days
**With Cloud Provider as Reference:** ~52 days

**Savings:** ~8 days (13% time reduction)

**Where Savings Come From:**
- Build infrastructure (GoReleaser, Makefile, .golangci.yml): 1 day saved
- Provider patterns: 1.5 days saved
- Resource patterns: 2.5 days saved
- Testing patterns: 2.5 days saved
- Documentation: 0.5 days saved
- Avoiding pitfalls: 1 day saved

**Additional Time Required (vs copying GitHub Actions):**
- Jenkins CI/CD pipeline setup: +1 day

**Where We Still Must Invest:**
- Jenkins Jenkinsfiles (CI, release, scheduled): 1 day
- API client generation: 3 days
- API wrapper: 2 days
- Resource implementations: 25 days
- Testing: 15 days
- Documentation writing: 5 days

---

## SDK v2 vs Plugin Framework Migration

**Important Note:** Cloud provider uses SDK v2, we're using Plugin Framework.

**Cannot directly copy code** due to different APIs, but patterns translate:

```go
// SDK v2 (Cloud)
CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics

// Plugin Framework (Ours)
func (r *resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
```

**Benefit:** Cloud provider shows proven patterns; we implement in modern framework.

---

## Licensing Considerations

**Cloud Provider License:** MPL-2.0 (Mozilla Public License 2.0)

**What This Means:**
- ✅ Can reference and learn from code
- ✅ Can copy build/infra files (with attribution)
- ✅ Can use same patterns and approaches
- ⚠️ If copying significant code portions, must maintain MPL-2.0 license
- ⚠️ File-level licensing (copied files must remain MPL-2.0)

**Recommendation:**
- Reference for patterns, don't copy-paste resource code
- Copy build infrastructure (GoReleaser, Makefile) with attribution
- Write all API and resource logic from scratch
- Choose appropriate license for our provider (MPL-2.0 compatible)

---

## Conclusion

**Yes, we can reuse ~35% of the work:**
- ✅ Build and release infrastructure (high value, low effort)
- ✅ Project structure and patterns (moderate value, low effort)
- ✅ Testing patterns (moderate value, moderate effort)
- ✅ Documentation templates (moderate value, low effort)

**But we must build ~65% from scratch:**
- ❌ Complete API client (100% new)
- ❌ All resource implementations (90% new)
- ❌ DC-specific validation (70% new)
- ❌ DC-specific tests (40% new)

**Net Benefit:** 10-15 days time savings, proven patterns, reduced risk

**Approach:** Use Cloud provider as a **reference and template**, not a codebase to fork.
