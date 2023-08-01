def PROJECT_NAME = 'logdna-logspout'
def TRIGGER_PATTERN = ".*@logdnabot.*"
def DEFAULT_BRANCH = 'master'
def CURRENT_BRANCH = [env.CHANGE_BRANCH, env.BRANCH_NAME]?.find{branch -> branch != null}

pipeline {
  agent {
    node {
      label 'ec2-fleet'
      customWorkspace "${PROJECT_NAME}-${BUILD_NUMBER}"
    }
  }

  options {
    timestamps()
    ansiColor 'xterm'
  }

  triggers {
    issueCommentTrigger(TRIGGER_PATTERN)
  }

  stages {
    stage('Validate PR Source') {
      when {
        expression { env.CHANGE_FORK }
        not {
          triggeredBy 'issueCommentCause'
        }
      }
      steps {
        error("A maintainer needs to approve this PR for CI by commenting")
      }
    }
    stage('Lint') {
      steps {
        sh 'make lint'
      }
    }

    stage('Release') {
      when {
        branch "${DEFAULT_BRANCH}"
        not {
          changelog '\\[skip ci\\]'
        }
      }

      environment {
        DEFAULT_BRANCH = "${DEFAULT_BRANCH}"
        GIT_BRANCH = "${CURRENT_BRANCH}"
        GITHUB_TOKEN = credentials('github-api-token')
        USERNAME = 'logdna'
      }

      stages {
        stage('Create Release') {
          steps {
            configFileProvider([configFile(fileId: 'git-askpass', variable: 'GIT_ASKPASS')]) {
              sh 'chmod +x \$GIT_ASKPASS'
              sh 'make version'
            }
          }
        }

        stage('Build Image') {
          steps {
            sh 'make build'
          }
        }

        stage('Publish Image') {
          steps {
            script {
              docker.withRegistry(
                'https://index.docker.io/v1/', 'dockerhub-username-password'
              ) {
                sh 'make publish'
              }
            }
          }
        }
      }
    }
  }
}
