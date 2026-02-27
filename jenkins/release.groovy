pipeline {
    agent any

    environment {
        GOPATH       = "${WORKSPACE}/.go"
        PATH         = "${WORKSPACE}/.go/bin:${PATH}"
        GITHUB_TOKEN = credentials('github-token')
    }

    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Release') {
            steps {
                sh '''
                    if ! command -v goreleaser &>/dev/null; then
                        go install github.com/goreleaser/goreleaser/v2@latest
                    fi
                    goreleaser release --clean
                '''
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
