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

        stage('Lint') {
            steps {
                sh '''
                    if ! command -v golangci-lint &>/dev/null; then
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
                            | sh -s -- -b "$(go env GOPATH)/bin" v1.61.0
                    fi
                    make lint
                '''
            }
        }

        stage('Unit Tests') {
            steps {
                sh 'make test'
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
