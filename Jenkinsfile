pipeline {
    agent any

    environment {
        REMOTE_IP = '172.17.0.1'
        
        REMOTE_USER = 'wind' 
        
        DEPLOY_PATH = '/home/wind/fnb'
        SERVICE_NAME = 'fnb_be.service' 
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Binary') {
            steps {
                echo "Đang build Golang CGO_ENABLED=0..."
                sh 'CGO_ENABLED=0 GOOS=linux go build -a -o fnb_be ./cmd/server'
            }
        }

        stage('Deploy To Host via SSH') {
            steps {
                echo "Đang đẩy qua Docker Gateway IP SSH ra máy Host..."
                script {
                    sh """
                        # SCP đẩy file qua luồng SSH theo IP 172.17.0.1
                        scp -o StrictHostKeyChecking=no fnb_be ${REMOTE_USER}@${REMOTE_IP}:${DEPLOY_PATH}/fnb_be_new
                        
                        # Xuyên thủng gọi SSH để thay đổi file và Restart systemctl
                        ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} "\
                            mkdir -p ${DEPLOY_PATH} && \
                            chmod +x ${DEPLOY_PATH}/fnb_be_new && \
                            mv -f ${DEPLOY_PATH}/fnb_be_new ${DEPLOY_PATH}/fnb_be && \
                            sudo systemctl restart ${SERVICE_NAME} \
                        "
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
