## 1. data.bitbucketdc_projects

- [x] 1.1 Implement `data_source_projects.go` with pagination, optional `filter` attribute, and `projects` list attribute
- [x] 1.2 Register `NewProjectsDataSource` in `provider.go`
- [x] 1.3 Write acceptance tests: all projects, filtered, no-match, verify attributes
- [x] 1.4 Add `data.bitbucketdc_projects` example to `tests/terraform/main.tf`

## 2. data.bitbucketdc_repositories

- [x] 2.1 Implement `data_source_repositories.go` with pagination, optional `filter` attribute, and `repositories` list attribute
- [x] 2.2 Register `NewRepositoriesDataSource` in `provider.go`
- [x] 2.3 Write acceptance tests: all repos, filtered, no-match, project-not-found, verify attributes
- [x] 2.4 Add `data.bitbucketdc_repositories` example to `tests/terraform/main.tf`

## 3. E2E Verification

- [x] 3.1 Run full acceptance test suite (52+ tests pass)
- [x] 3.2 Run E2E cycle: `make install` → `terraform apply` → zero-drift plan → `terraform destroy`
- [x] 3.3 Commit data sources

## 4. Documentation

- [x] 4.1 Install `tfplugindocs` tool (`go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest`)
- [x] 4.2 Create `templates/` directory with `index.md.tmpl` provider template
- [x] 4.3 Add `make docs` target to `Makefile`
- [x] 4.4 Run `make docs` and verify `docs/` directory is generated
- [x] 4.5 Commit generated docs

## 5. CI/CD Pipeline

- [x] 5.1 Create `jenkins/` folder with `ci.groovy`, `cd.groovy`, and `release.groovy` pipelines
- [x] 5.2 Commit Jenkinsfile

## 6. Final

- [x] 6.1 Update CHANGELOG with new data sources, docs, and CI/CD
- [ ] 6.2 Commit all remaining changes
