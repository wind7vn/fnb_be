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

        stage('Package, Build & Deploy on Host') {
            steps {
                echo "Đóng gói mã nguồn và gửi trực tiếp sang host để build..."
                script {
                    withCredentials([
                        file(credentialsId: 'dev-env-file', variable: 'ENV_FILE'),
                        file(credentialsId: 'dev-firebase-service-account', variable: 'FIREBASE_FILE')
                    ]) {
                        sh """
                            # Đóng gói toàn bộ code workspace (đẩy file nén ra chỗ khác để tránh lỗi changed as we read it)
                            tar --exclude='./.git' -czf /tmp/source_fnb.tar.gz .
                            
                            # Đảm bảo các thư mục tồn tại trên host
                            ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} "mkdir -p ${DEPLOY_PATH}/src"
                            
                            # SCP mã nguồn nén sang Host
                            scp -o StrictHostKeyChecking=no /tmp/source_fnb.tar.gz ${REMOTE_USER}@${REMOTE_IP}:${DEPLOY_PATH}/source.tar.gz
                            rm -f /tmp/source_fnb.tar.gz
                            
                            # SCP file cấu hình sang Host
                            scp -o StrictHostKeyChecking=no \$ENV_FILE ${REMOTE_USER}@${REMOTE_IP}:${DEPLOY_PATH}/.env
                            scp -o StrictHostKeyChecking=no \$FIREBASE_FILE ${REMOTE_USER}@${REMOTE_IP}:${DEPLOY_PATH}/firebase-service-account.json
                            
                            # SSH sang Host để giải nén mã nguồn, build và restart
                            # Lưu ý: nếu máy đích không nhận lệnh 'go', bạn có thể cần thay 'go' bằng '/usr/local/go/bin/go'
                            ssh -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_IP} "\
                                cd ${DEPLOY_PATH} && \
                                tar -xzf source.tar.gz -C ./src && \
                                cd ./src && \
                                export PATH=\\\$PATH:/usr/local/go/bin:/usr/bin && \
                                CGO_ENABLED=0 GOOS=linux go build -a -o ../fnb_be ./cmd/server && \
                                cd .. && \
                                sudo systemctl restart ${SERVICE_NAME} && \
                                rm -f source.tar.gz && \
                                rm -rf ./src \
                            "
                        """
                    }
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
