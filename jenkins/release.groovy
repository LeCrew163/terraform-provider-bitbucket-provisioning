pipeline {
    agent any

    environment {
        GOPATH = "${WORKSPACE}/.go"
        PATH   = "${WORKSPACE}/.go/bin:${PATH}"
    }

    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Release') {
            environment {
                ARTIFACTORY_USERNAME = credentials('artifactory-username')
                ARTIFACTORY_SECRET   = credentials('artifactory-secret')
            }
            steps {
                sh '''
                    if ! command -v goreleaser &>/dev/null; then
                        go install github.com/goreleaser/goreleaser/v2@latest
                    fi
                    goreleaser release --clean
                '''

                script {
                    def version = sh(script: "git describe --tags --abbrev=0 | sed 's/^v//'", returnStdout: true).trim()
                    def binary  = 'terraform-provider-bitbucket-provisioning'
                    def base    = "http://art01.sldnet.de:8081/artifactory/terraform/alpina-operation/bitbucket-provisioning/${version}"

                    // Terraform requires a manifest.json alongside the zips so that the
                    // Artifactory registry can serve the provider protocol correctly.
                    writeFile file: 'manifest.json', text: '''{
  "version": 1,
  "metadata": {
    "protocol_versions": ["6.0"]
  }
}
'''
                    sh """
                        curl -fSu "\${ARTIFACTORY_USERNAME}:\${ARTIFACTORY_SECRET}" \\
                            -T manifest.json \\
                            "${base}/${binary}_${version}_manifest.json"
                    """
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
