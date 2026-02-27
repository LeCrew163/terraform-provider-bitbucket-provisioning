## Requirements

### Requirement: Jenkinsfile declarative pipeline
The system SHALL provide a `Jenkinsfile` at the repository root with a declarative pipeline covering build, lint, test, and release stages.

#### Scenario: Build stage compiles the provider
- **WHEN** a commit is pushed to any branch
- **THEN** Jenkins runs `make build` and fails the build if compilation fails

#### Scenario: Lint stage runs golangci-lint
- **WHEN** a commit is pushed to any branch
- **THEN** Jenkins runs golangci-lint and reports any lint errors

#### Scenario: Unit test stage runs non-acceptance tests
- **WHEN** a commit is pushed to any branch
- **THEN** Jenkins runs `go test ./...` (without TF_ACC) and reports results

#### Scenario: Acceptance tests are parameterised
- **WHEN** user manually triggers the pipeline with `RUN_ACC_TESTS=true`
- **THEN** Jenkins runs acceptance tests with the configured Bitbucket credentials

#### Scenario: Release stage runs on master branch
- **WHEN** a tag matching `v*` is pushed to master
- **THEN** Jenkins runs GoReleaser to build and publish release artifacts
