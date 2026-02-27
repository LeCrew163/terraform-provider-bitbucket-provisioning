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
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
