## Context

The provider has 11 resources and 4 singular data sources. All are fully tested (52 passing acceptance tests). The remaining work is additive: two list data sources that follow the exact same patterns as existing singular ones, documentation generation, and a Jenkins pipeline.

Existing patterns to follow:
- Data sources use `datasource.DataSource` interface with `Read` only
- Pagination: Bitbucket DC uses `start`/`limit`/`isLastPage` cursor model — must page through all results
- List data sources store results as a list of objects in a computed attribute
- Tests: `TF_ACC=1 go test` against a live Bitbucket instance

## Goals / Non-Goals

**Goals:**
- `data.bitbucketdc_projects` — list all projects the caller can see, optional name filter
- `data.bitbucketdc_repositories` — list all repos in a project, optional name filter
- Auto-generated docs via `tfplugindocs` (schema descriptions are already written)
- Jenkinsfile with lint, unit-test, acceptance-test (parameterised), and release stages

**Non-Goals:**
- Pagination state (all pages are fetched and returned as a single list)
- Branch workflow resource (API not exposed in Bitbucket DC 9.x REST spec)
- Cross-project repository search

## Decisions

### 1. Pagination: fetch all pages internally

**Decision:** Fetch all pages inside the Read method and return a flat list.

**Rationale:** Terraform data sources are not cursored — callers expect a complete list. The Bitbucket `isLastPage` + `nextPageStart` pattern is well-understood from the existing client.

**Alternative:** Expose `limit`/`start` attributes to let callers page manually — rejected because it breaks the data source contract (partial data would cause inconsistent plans).

### 2. Filter attribute: optional `filter` string, client-side contains match

**Decision:** Accept an optional `filter` attribute and apply a case-insensitive `strings.Contains` match on the name after fetching.

**Rationale:** The Bitbucket API accepts a `name` query parameter for projects and a `filter` parameter for repositories — we pass it through to reduce data transfer, but also apply client-side matching to ensure exact semantics.

### 3. Documentation: tfplugindocs with `templates/` scaffolding

**Decision:** Use `tfplugindocs generate` which reads schema `Description` fields and writes to `docs/`.

**Rationale:** Standard Terraform provider tooling; required for Terraform Registry submission.

### 4. Jenkinsfile: declarative pipeline, acceptance tests parameterised

**Decision:** Acceptance tests run in a separate parameterised stage, not on every commit, because they require a live Bitbucket instance.

## Risks / Trade-offs

- [Large projects with many repos] Fetching all pages may be slow → acceptable, data sources are read-once at plan time
- [tfplugindocs] Requires Go 1.21+ toolchain in Jenkins → already satisfied by go.mod
