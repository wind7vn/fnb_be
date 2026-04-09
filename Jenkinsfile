pipeline {
    agent none  // Không chiếm dụng agent ngay từ đầu
    
    environment {
        JENKINS_NODE_COOKIE = 'dontKillMe'
        REMOTE_USER = 'wind'
        REMOTE_IP   = '172.17.0.1' 
        DEPLOY_PATH = '/home/wind/fnb'
        SERVICE_NAME = 'fnb_be.service'
    }

    stages {
        stage('Build & Deploy Flow') {
            agent any
            
            stages {
                stage('Clean Workspace') {
                    steps {
                        cleanWs()
                    }
                }

                stage('Checkout') {
                    steps {
                        checkout scm
                    }
                }

                stage('Prepare Secrets') {
                    steps {
                        withCredentials([
                            file(credentialsId: 'dev-env-file', variable: 'ENV_FILE'),
                            file(credentialsId: 'dev-firebase-service-account', variable: 'FIREBASE_FILE')
                        ]) {
                            script {
                                sh '''
                                    cp $ENV_FILE .env
                                    cp $FIREBASE_FILE firebase-service-account.json
                                '''
                            }
                        }
                    }
                }

                stage('Deploy & Build on Host (SSH)') {
                    steps {
                        sshagent(['jenkin-ssh-key']) { 
                            script {
                                echo "Deploying to ${REMOTE_USER}@${REMOTE_IP}..."

                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'mkdir -p ${DEPLOY_PATH}/src'"
                                
                                echo "--- Streaming source code & secrets to target ---"
                                sh """
                                    tar -czf - --exclude='.git' . | \
                                    ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} \
                                    "tar -xzf - -C ${DEPLOY_PATH}/src"
                                """

                                echo "--- Building Golang on target ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'cd ${DEPLOY_PATH}/src && export PATH=\\$PATH:/usr/local/go/bin:/usr/bin && CGO_ENABLED=0 GOOS=linux go build -a -o ../fnb_be ./cmd/server && CGO_ENABLED=0 GOOS=linux go build -a -o ../fnb_migrator ./cmd/migrator'"
                                
                                echo "--- Moving secrets to target ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'mv -f ${DEPLOY_PATH}/src/.env ${DEPLOY_PATH}/.env'"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'mv -f ${DEPLOY_PATH}/src/firebase-service-account.json ${DEPLOY_PATH}/firebase-service-account.json'"
                                
                                echo "--- Running Database Migrations ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'cd ${DEPLOY_PATH} && ./fnb_migrator'"
                                
                                echo "--- Setting up Systemd Service File ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'mkdir -p ~/.config/systemd/user'"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'cp -f ${DEPLOY_PATH}/src/deploy/${SERVICE_NAME} ~/.config/systemd/user/${SERVICE_NAME}'"

                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'systemctl --user daemon-reload'"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'systemctl --user enable ${SERVICE_NAME}'"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'systemctl --user restart ${SERVICE_NAME}'"
                                
                                echo "--- Cleaning up source folder ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'rm -rf ${DEPLOY_PATH}/src'"

                                echo "--- Verify Deployment Status ---"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'sleep 3 && systemctl --user is-active ${SERVICE_NAME} || { echo \"CRITICAL: Dịch vụ đã Crash ngay khi khởi động!\"; journalctl --user -u ${SERVICE_NAME} -n 50 --no-pager; exit 1; }'"
                                sh "ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} 'journalctl --user -u ${SERVICE_NAME} -n 15 --no-pager'"
                            }
                        }
                    }
                }
            }
        }
    }

    post {
        success {
            echo 'Deployment Finished Successfully!'
        }
        failure {
            echo 'Deployment Failed.'
        }
    }
}
