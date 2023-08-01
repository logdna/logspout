def BRANCH_ACTUAL = env.CHANGE_BRANCH ? env.CHANGE_BRANCH : env.BRANCH_NAME
def CREDS = [
    string(credentialsId: 'github-api-token',
           variable: 'GITHUB_TOKEN')
]

pipeline {
    agent {
        node {
            label 'ec2-fleet'
            customWorkspace("/tmp/workspace/${env.BUILD_TAG}")
        }
    }

    environment {
        GIT_BRANCH = "${BRANCH_ACTUAL}"
        GOPRIVATE = "github.com/answerbook"
    }

    options {
        timeout time: 1, unit: 'HOURS'
        timestamps()
        ansiColor 'xterm'
        withCredentials(CREDS)
    }

    stages {
        stage('Lint') {
            steps {
                sh 'make lint'
            }
        }

        stage('Test') {
            steps {
                configFileProvider([configFile(fileId: 'git-askpass', variable: 'GIT_ASKPASS')]) {
                    sh 'chmod +x \$GIT_ASKPASS'
                    sh 'make test'
                }
            }
            post {
                always {
                    publishCoverage adapters: [
                        coberturaAdapter(path: 'reports/coverage.xml')
                    ]
                    publishHTML target: [
                        allowMissing: true,
                        alwaysLinkToLastBuild: false,
                        keepAll: false,
                        reportDir: "reports",
                        reportFiles: "coverage.html",
                        reportName: "${env.BUILD_TAG}"
                    ]
                }
            }
        }

        stage('Publish') {
            steps {
                configFileProvider([configFile(fileId: 'git-askpass', variable: 'GIT_ASKPASS')]) {
                    sh 'chmod +x \$GIT_ASKPASS'
                    sh 'make publish'
                }
            }
        }
    }
}
