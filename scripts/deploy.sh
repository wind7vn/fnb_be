#!/bin/bash
set -e

# Configuration
REMOTE_HOST="192.168.1.2"
REMOTE_USER="wind"
REMOTE_PASS="hunter"
DEPLOY_PATH="/home/wind/fnb"
SERVICE_NAME="fnb_be.service"
DOMAIN="dev-fnb.windai.online"

echo "========================================="
echo "🚀 Starting deployment to $REMOTE_HOST..."
echo "========================================="

# 1. Check for sshpass
SSH_CMD="ssh -o StrictHostKeyChecking=no"
SCP_CMD="scp -o StrictHostKeyChecking=no"

if command -v sshpass &> /dev/null; then
    echo "✔ sshpass found. Using automated password authentication."
    SSH_CMD="sshpass -p $REMOTE_PASS ssh -o StrictHostKeyChecking=no"
    SCP_CMD="sshpass -p $REMOTE_PASS scp -o StrictHostKeyChecking=no"
else
    echo "⚠️  sshpass not found. If prompted, please enter password: $REMOTE_PASS (or set up SSH keys)."
fi

# 2. Build Go binaries for Linux (amd64)
echo "--- 🛠 Building Golang binaries for Linux (amd64) ---"
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/fnb_be ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/fnb_migrator ./cmd/migrator
echo "✔ Binaries compiled successfully."

# 3. Prepare config files
echo "--- 📝 Preparing remote config files ---"
# Make temporary .env with port 4000
cp .env .env.temp
# Replace PORT and APP_DOMAIN values in temp .env
sed -i '' 's/^PORT=.*/PORT=4000/' .env.temp || sed -i 's/^PORT=.*/PORT=4000/' .env.temp
sed -i '' 's|^APP_DOMAIN=.*|APP_DOMAIN=https://dev-fnb.windai.online|' .env.temp || sed -i 's|^APP_DOMAIN=.*|APP_DOMAIN=https://dev-fnb.windai.online|' .env.temp

# 4. Stop service, create remote directories & clean write-protected config files
echo "--- 📁 Stopping remote service and removing old config files ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "systemctl --user stop ${SERVICE_NAME} 2>/dev/null || true"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "mkdir -p ${DEPLOY_PATH}/configs && rm -f ${DEPLOY_PATH}/.env ${DEPLOY_PATH}/configs/firebase-service-account.json 2>/dev/null || true"

# 5. Upload files to remote host
echo "--- 📤 Uploading binaries and configs to remote host ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "mkdir -p ${DEPLOY_PATH}/data ${DEPLOY_PATH}/web/bank_icons"
$SCP_CMD data/momo_banks.json ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/data/momo_banks.json
$SCP_CMD -r web/bank_icons/. ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/web/bank_icons/
$SCP_CMD bin/fnb_be ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/fnb_be
$SCP_CMD bin/fnb_migrator ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/fnb_migrator
$SCP_CMD .env.temp ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/.env
rm -f .env.temp

# Upload Firebase key and APNs keys
if [ -f configs/firebase-service-account.json ]; then
    $SCP_CMD configs/firebase-service-account.json ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/configs/firebase-service-account.json
fi
if [ -f configs/AuthKey_S77A4NC4RH.p8 ]; then
    $SCP_CMD configs/AuthKey_S77A4NC4RH.p8 ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/configs/AuthKey_S77A4NC4RH.p8
fi

# Upload service files and Nginx configs
$SCP_CMD deploy/fnb_be.service ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/fnb_be.service
$SCP_CMD deploy/nginx_dev_fnb.conf ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/nginx_dev_fnb.conf

# 6. Setup and run migrations on remote host
echo "--- ⚙ Running database migrations on remote host ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "chmod +x ${DEPLOY_PATH}/fnb_be ${DEPLOY_PATH}/fnb_migrator && cd ${DEPLOY_PATH} && ./fnb_migrator"

# 7. Setup Systemd Service
echo "--- ⚙ Configuring Systemd User Service ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "
    mkdir -p ~/.config/systemd/user && \
    cp -f ${DEPLOY_PATH}/fnb_be.service ~/.config/systemd/user/${SERVICE_NAME} && \
    systemctl --user daemon-reload && \
    systemctl --user enable ${SERVICE_NAME} && \
    systemctl --user restart ${SERVICE_NAME}
"

# 8. Configure Nginx Reverse Proxy (requires sudo)
echo "--- 🌐 Configuring Nginx Reverse Proxy for $DOMAIN ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "
    if [ ! -f /usr/sbin/nginx ] && ! command -v nginx &> /dev/null; then
        echo 'Nginx not found. Installing Nginx...' && \
        echo '$REMOTE_PASS' | sudo -S apt-get update && \
        echo '$REMOTE_PASS' | sudo -S apt-get install -y nginx
    fi && \
    echo '$REMOTE_PASS' | sudo -S cp -f ${DEPLOY_PATH}/nginx_dev_fnb.conf /etc/nginx/sites-available/${DOMAIN} && \
    echo '$REMOTE_PASS' | sudo -S ln -sf /etc/nginx/sites-available/${DOMAIN} /etc/nginx/sites-enabled/ && \
    echo '$REMOTE_PASS' | sudo -S nginx -t && \
    echo '$REMOTE_PASS' | sudo -S systemctl reload nginx
"

# 9. Verification
echo "--- 🔍 Verifying deployment status ---"
sleep 2
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "systemctl --user is-active ${SERVICE_NAME}"

echo "========================================="
echo "🎉 DEPLOYMENT SUCCESSFUL!"
echo "Server running on port 4000 and proxy-mapped to $DOMAIN"
echo "========================================="
