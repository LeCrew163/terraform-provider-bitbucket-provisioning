# Jenkins Pipeline for Terraform Provider

## Overview

Since we're using Jenkins instead of GitHub Actions, we need to create custom Jenkinsfiles for CI/CD. The Cloud provider's GitHub Actions workflows are not applicable.

## Example Jenkinsfile

### Main CI Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent {
        label 'golang'
    }

    environment {
        GO_VERSION = '1.21'
        CGO_ENABLED = '0'
        GOPATH = "${WORKSPACE}/go"
        PATH = "${GOPATH}/bin:${env.PATH}"
    }

    options {
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 1, unit: 'HOURS')
        timestamps()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Setup') {
            steps {
                sh '''
                    go version
                    go mod download
                    go mod verify
                '''
            }
        }

        stage('Format Check') {
            steps {
                sh 'make fmtcheck'
            }
        }

        stage('Lint') {
            steps {
                sh '''
                    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
                    golangci-lint run --timeout 5m
                '''
            }
        }

        stage('Unit Tests') {
            steps {
                sh 'make test'
            }
            post {
                always {
                    junit '**/test-results/*.xml'
                }
            }
        }

        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Acceptance Tests') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                    expression { params.RUN_ACCEPTANCE_TESTS }
                }
            }
            environment {
                TF_ACC = '1'
                BITBUCKET_BASE_URL = credentials('bitbucket-test-url')
                BITBUCKET_TOKEN = credentials('bitbucket-test-token')
            }
            steps {
                sh 'make testacc'
            }
            post {
                always {
                    junit '**/acc-test-results/*.xml'
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
        failure {
            emailext(
                subject: "Build Failed: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                body: "Build failed. Check console output at ${env.BUILD_URL}",
                to: "${env.TEAM_EMAIL}"
            )
        }
    }
}
```

---

### Release Pipeline

```groovy
// Jenkinsfile.release
pipeline {
    agent {
        label 'golang'
    }

    parameters {
        string(name: 'VERSION', description: 'Release version (e.g., v1.0.0)', defaultValue: '')
        booleanParam(name: 'DRY_RUN', description: 'Dry run (skip actual release)', defaultValue: false)
    }

    environment {
        GO_VERSION = '1.21'
        CGO_ENABLED = '0'
        GOPATH = "${WORKSPACE}/go"
        PATH = "${GOPATH}/bin:${env.PATH}"
        GPG_FINGERPRINT = credentials('gpg-fingerprint')
    }

    stages {
        stage('Validate Version') {
            steps {
                script {
                    if (!params.VERSION) {
                        error("VERSION parameter is required")
                    }
                    if (!params.VERSION.matches(/^v\d+\.\d+\.\d+$/)) {
                        error("VERSION must match semantic versioning (e.g., v1.0.0)")
                    }
                }
            }
        }

        stage('Checkout') {
            steps {
                checkout scm
                sh """
                    git fetch --tags
                    git tag ${params.VERSION}
                """
            }
        }

        stage('Setup GoReleaser') {
            steps {
                sh '''
                    curl -sfL https://goreleaser.com/static/run | VERSION=latest sh -s -- check
                '''
            }
        }

        stage('Import GPG Key') {
            steps {
                withCredentials([
                    file(credentialsId: 'gpg-private-key', variable: 'GPG_KEY_FILE'),
                    string(credentialsId: 'gpg-passphrase', variable: 'GPG_PASSPHRASE')
                ]) {
                    sh '''
                        gpg --batch --import ${GPG_KEY_FILE}
                        echo "${GPG_PASSPHRASE}" | gpg --batch --yes --passphrase-fd 0 --pinentry-mode loopback --sign-key ${GPG_FINGERPRINT}
                    '''
                }
            }
        }

        stage('Build and Release') {
            when {
                expression { !params.DRY_RUN }
            }
            environment {
                GITHUB_TOKEN = credentials('github-token')
            }
            steps {
                sh 'goreleaser release --clean'
            }
        }

        stage('Dry Run') {
            when {
                expression { params.DRY_RUN }
            }
            steps {
                sh 'goreleaser release --snapshot --skip=publish --clean'
            }
        }

        stage('Publish to Terraform Registry') {
            when {
                expression { !params.DRY_RUN }
            }
            environment {
                TFC_TOKEN = credentials('terraform-cloud-token')
            }
            steps {
                sh '''
                    # Upload to Terraform Cloud private registry
                    # This step depends on your Terraform Cloud setup
                    # You may need to use terraform-registry-manifest tool
                    # or Terraform Cloud API
                    echo "Publishing to Terraform Cloud private registry..."
                    # TODO: Add Terraform Cloud upload logic
                '''
            }
        }

        stage('Push Tag') {
            when {
                expression { !params.DRY_RUN }
            }
            steps {
                sh """
                    git push origin ${params.VERSION}
                """
            }
        }

        stage('Update Changelog') {
            when {
                expression { !params.DRY_RUN }
            }
            steps {
                sh '''
                    # Generate changelog entry
                    echo "## ${VERSION}" >> CHANGELOG.md
                    git log --pretty=format:"- %s" $(git describe --tags --abbrev=0 HEAD^)..HEAD >> CHANGELOG.md
                    git add CHANGELOG.md
                    git commit -m "Update CHANGELOG for ${VERSION}"
                    git push origin main
                '''
            }
        }
    }

    post {
        success {
            script {
                if (!params.DRY_RUN) {
                    emailext(
                        subject: "Release ${params.VERSION} Published Successfully",
                        body: "Provider version ${params.VERSION} has been released and published.",
                        to: "${env.TEAM_EMAIL}"
                    )
                }
            }
        }
        failure {
            emailext(
                subject: "Release ${params.VERSION} Failed",
                body: "Release pipeline failed. Check console output at ${env.BUILD_URL}",
                to: "${env.TEAM_EMAIL}"
            )
        }
        always {
            cleanWs()
        }
    }
}
```

---

### Scheduled Dependency Check

```groovy
// Jenkinsfile.dependency-check
pipeline {
    agent {
        label 'golang'
    }

    triggers {
        // Run every Monday at 9 AM
        cron('0 9 * * 1')
    }

    environment {
        GO_VERSION = '1.21'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Check Outdated Dependencies') {
            steps {
                sh '''
                    go list -u -m -json all | jq -r 'select(.Update != null) | "\\(.Path) \\(.Version) -> \\(.Update.Version)"' > outdated.txt
                    if [ -s outdated.txt ]; then
                        echo "Outdated dependencies found:"
                        cat outdated.txt
                    else
                        echo "All dependencies are up to date"
                    fi
                '''
            }
        }

        stage('Security Audit') {
            steps {
                sh '''
                    go install golang.org/x/vuln/cmd/govulncheck@latest
                    govulncheck ./... || true
                '''
            }
        }

        stage('Mod Tidy Check') {
            steps {
                sh '''
                    go mod tidy
                    git diff --exit-code go.mod go.sum
                '''
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'outdated.txt', allowEmptyArchive: true
        }
        failure {
            emailext(
                subject: "Dependency Check Issues Found",
                body: "Dependency check found outdated or vulnerable dependencies. Check ${env.BUILD_URL}",
                to: "${env.TEAM_EMAIL}"
            )
        }
    }
}
```

---

## Jenkins Setup Requirements

### Required Plugins

1. **Pipeline Plugin** - For Jenkinsfile support
2. **Git Plugin** - For repository checkout
3. **Credentials Plugin** - For secrets management
4. **Email Extension Plugin** - For notifications
5. **JUnit Plugin** - For test result reporting
6. **Blue Ocean** (optional) - For better UI

### Required Credentials

Configure these in Jenkins Credentials:

| ID | Type | Description |
|----|------|-------------|
| `bitbucket-test-url` | Secret text | Test Bitbucket DC URL |
| `bitbucket-test-token` | Secret text | Test Bitbucket PAT |
| `gpg-private-key` | Secret file | GPG private key for signing |
| `gpg-passphrase` | Secret text | GPG key passphrase |
| `gpg-fingerprint` | Secret text | GPG key fingerprint |
| `github-token` | Secret text | GitHub token for releases |
| `terraform-cloud-token` | Secret text | Terraform Cloud API token |

### Jenkins Agent Requirements

**Labels:** `golang`

**Required Tools:**
- Go 1.21+
- Git
- GPG
- jq (for dependency checks)
- curl
- Make

### Environment Variables

Set these in Jenkins global configuration:

```properties
TEAM_EMAIL=platform-team@company.com
```

---

## Makefile Targets (Reused from Cloud Provider)

```makefile
# These targets work with Jenkins
default: build

build: fmtcheck
	go install

test: fmtcheck
	go test -i $(TEST) || exit 1
	go test $(TEST) -timeout=30s -parallel=4 -v -json | tee test-results.json
	go-junit-report < test-results.json > test-results/junit.xml

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v -parallel=10 -timeout 120m -json | tee acc-test-results.json
	go-junit-report < acc-test-results.json > acc-test-results/junit.xml

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"
```

---

## Pipeline Triggers

### Main CI Pipeline
- **Trigger:** On every commit to any branch
- **Branches:** All
- **Acceptance Tests:** Only on `main`, `develop`, or manual trigger

### Release Pipeline
- **Trigger:** Manual (parameterized)
- **Parameters:** Version number, dry-run flag

### Dependency Check
- **Trigger:** Scheduled (cron)
- **Schedule:** Weekly (Monday 9 AM)

---

## Comparison: GitHub Actions vs Jenkins

| Feature | GitHub Actions (Cloud) | Jenkins (Our Setup) |
|---------|----------------------|---------------------|
| **Configuration** | YAML in `.github/workflows/` | Groovy in `Jenkinsfile` |
| **Reusability** | Can reuse HashiCorp workflows | Must write custom pipelines |
| **Secrets** | GitHub Secrets | Jenkins Credentials |
| **Artifacts** | GitHub Releases | Jenkins artifacts + external storage |
| **Notifications** | GitHub notifications | Email/Slack via plugins |
| **Test Reports** | GitHub Actions UI | Jenkins JUnit plugin |
| **Setup Time** | ~1 hour (copy workflows) | ~1 day (write Jenkinsfiles) |

---

## Additional Considerations

### GoReleaser Configuration

The `.goreleaser.yml` from the Cloud provider can be reused with minimal changes:

```yaml
# .goreleaser.yml - Works with Jenkins
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: '{{ .CommitTimestamp }}'
    flags:
      - -trimpath
    ldflags:
      - '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}}'
    goos:
      - freebsd
      - windows
      - linux
      - darwin
    goarch:
      - amd64
      - '386'
      - arm
      - arm64
    ignore:
      - goos: darwin
        goarch: '386'
    binary: '{{ .ProjectName }}_v{{ .Version }}'

archives:
  - format: zip
    name_template: '{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}'

checksum:
  name_template: '{{ .ProjectName }}_{{ .Version }}_SHA256SUMS'
  algorithm: sha256

signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"

release:
  github:
    owner: your-org
    name: terraform-provider-bitbucketdc
  draft: false
  prerelease: auto

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

This configuration works the same way whether triggered by GitHub Actions or Jenkins.

---

## Next Steps

1. Create base Jenkinsfile for CI pipeline
2. Configure Jenkins credentials
3. Set up Jenkins agents with required tools
4. Test pipeline with sample commits
5. Create release pipeline
6. Document Jenkins setup for team
7. Set up scheduled dependency checks
