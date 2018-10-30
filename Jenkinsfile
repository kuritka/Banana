def message = ''

pipeline {
    environment {
        PACKAGE = "github.com/Deutsche-Boerse/edt-sftp"
        GOPATH = "${HOME}/executors/${EXECUTOR_NUMBER}"

        EDT_INFRASTRUCTURE_PATH="${env.WORKSPACE}/edt-infrastructure"
        BRANCH_NAME = "${env.BRANCH_NAME}"
    }

    agent { 
        label 'swarm'
    }
    
    options {
        skipDefaultCheckout true
    }

    stages {
        stage('Checkout') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'github-user', usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN')]) {
                    script {
                        message = goCheckout(env.PACKAGE)
                    }
                }
            }
        }

        stage('Linter') {
            steps {
                script {
                    message = "${message}; " + goLinter(env.PACKAGE)
                }
            }
        }

        stage('Test') {
            steps {
	        withEnv(["EDT_SFTP_CONFIG=${env.GOPATH}/src/${env.PACKAGE}/conf/sftp_dev.json"]) {
                    script {
                        message = "${message}; " + goTest(env.PACKAGE)
                    }
                }
            }
        }

        stage('Ansible Build + Deploy') {
            steps {
               withCredentials([
                   usernamePassword(credentialsId: 'github-user', usernameVariable: 'GH_USER', passwordVariable: 'GH_TOKEN'),
                   usernamePassword(credentialsId: 'proxy_address', usernameVariable: 'HTTP_PROXY', passwordVariable: 'HTTPS_PROXY'),
                   usernamePassword(credentialsId: 'container_registry_account', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASSWORD'),
                   file(credentialsId: 'kubeconfig', variable: 'KUBECONFIG'),
                   zip(credentialsId: 'grpc-certs-zip', variable: 'GRPC_CERTIFICATES')
               ]) {
                 script {
                    if (BRANCH_NAME == 'master') {
                        runAnsible(env.PACKAGE)
                    }
                 }
               }
            }
        }
    }

    post {
        failure { 
            script { 
                edtSlackSend 'danger', 'Build Failed', message
                if (env.CHANGE_ID) {
                    withCredentials([string(credentialsId: 'github-oauth-token', variable: 'GITHUB_OAUTH_TOKEN')]) {
                        setPRLabel(env.PACKAGE, 'Build Failed')
                    }
                }
            }
        }

        success { 
            script { 
                edtSlackSend 'good', 'Build Passed', message
                if (env.CHANGE_ID) {
                    withCredentials([string(credentialsId: 'github-oauth-token', variable: 'GITHUB_OAUTH_TOKEN')]) {
                        setPRLabel(env.PACKAGE, 'Build Passed')
                    }
                }
            }
        }

        unstable { 
            script { 
                edtSlackSend 'warning', 'Build Unstable', message
                if (env.CHANGE_ID) {
                    withCredentials([string(credentialsId: 'github-oauth-token', variable: 'GITHUB_OAUTH_TOKEN')]) {
                        setPRLabel(env.PACKAGE, 'Build Unstable')
                    }
                }
            }
        }

        cleanup {
            sh 'rm -rf $GOPATH/src/$PACKAGE'
            sh 'rm -rf $EDT_INFRASTRUCTURE_PATH'
        }
    }
}
