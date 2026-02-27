pipeline {
    agent any

    parameters {
        booleanParam(
            name: 'RUN_ACC_TESTS',
            defaultValue: true,
            description: 'Run acceptance tests against a live Bitbucket instance'
        )
    }

    environment {
        GOPATH            = "${WORKSPACE}/.go"
        PATH              = "${WORKSPACE}/.go/bin:${PATH}"
        BITBUCKET_BASE_URL = credentials('bitbucket-base-url')
        BITBUCKET_USERNAME = credentials('bitbucket-username')
        BITBUCKET_PASSWORD = credentials('bitbucket-password')
        TF_ACC            = '1'
    }

    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }

        stage('Acceptance Tests') {
            when {
                expression { return params.RUN_ACC_TESTS }
            }
            steps {
                sh 'make testacc'
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
