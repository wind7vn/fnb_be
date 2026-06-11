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
cp .env .env.temp
sed -i '' 's/^PORT=.*/PORT=4000/' .env.temp || sed -i 's/^PORT=.*/PORT=4000/' .env.temp
sed -i '' 's|^APP_DOMAIN=.*|APP_DOMAIN=https://dev-fnb.windai.online|' .env.temp || sed -i 's|^APP_DOMAIN=.*|APP_DOMAIN=https://dev-fnb.windai.online|' .env.temp

# 4. Create local package directory for single-connection upload
echo "--- 📦 Creating local deployment package ---"
rm -rf deploy_package && mkdir -p deploy_package/configs deploy_package/data deploy_package/web/bank_icons

# Copy binaries
cp bin/fnb_be deploy_package/fnb_be
cp bin/fnb_migrator deploy_package/fnb_migrator

# Copy env
cp .env.temp deploy_package/.env
rm -f .env.temp

# Copy banking JSON & icons
cp data/momo_banks.json deploy_package/data/momo_banks.json
cp -r web/bank_icons/. deploy_package/web/bank_icons/

# Copy keys if present
if [ -f configs/firebase-service-account.json ]; then
    cp configs/firebase-service-account.json deploy_package/configs/firebase-service-account.json
fi
if [ -f configs/AuthKey_S77A4NC4RH.p8 ]; then
    cp configs/AuthKey_S77A4NC4RH.p8 deploy_package/configs/AuthKey_S77A4NC4RH.p8
fi

# Copy service and nginx configs
cp deploy/fnb_be.service deploy_package/fnb_be.service
cp deploy/nginx_dev_fnb.conf deploy_package/nginx_dev_fnb.conf

# Archive
tar -czf deploy.tar.gz -C deploy_package .
rm -rf deploy_package
echo "✔ Package deploy.tar.gz created."

# 5. Upload package (Exactly 1 SCP Connection)
echo "--- 📤 Uploading package to remote host ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "mkdir -p ${DEPLOY_PATH}"
$SCP_CMD deploy.tar.gz ${REMOTE_USER}@${REMOTE_HOST}:${DEPLOY_PATH}/deploy.tar.gz
rm -f deploy.tar.gz

# 6. Stop service, Extract, Migrate and Restart (Exactly 1 SSH Connection)
echo "--- ⚙ Extracting and configuring remote host ---"
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "
    # Stop service
    systemctl --user stop ${SERVICE_NAME} 2>/dev/null || true

    # Create target directories
    mkdir -p ${DEPLOY_PATH}/configs ${DEPLOY_PATH}/data ${DEPLOY_PATH}/web/bank_icons

    # Extract
    tar -xzf ${DEPLOY_PATH}/deploy.tar.gz -C ${DEPLOY_PATH}
    rm -f ${DEPLOY_PATH}/deploy.tar.gz

    # Executable permissions
    chmod +x ${DEPLOY_PATH}/fnb_be ${DEPLOY_PATH}/fnb_migrator

    # Run migrations
    cd ${DEPLOY_PATH} && ./fnb_migrator

    # Setup Systemd Service
    mkdir -p ~/.config/systemd/user && \
    cp -f ${DEPLOY_PATH}/fnb_be.service ~/.config/systemd/user/${SERVICE_NAME} && \
    systemctl --user daemon-reload && \
    systemctl --user enable ${SERVICE_NAME} && \
    systemctl --user restart ${SERVICE_NAME}
"

# 7. Configure Nginx Reverse Proxy (requires sudo)
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

# 8. Verification
echo "--- 🔍 Verifying deployment status ---"
sleep 2
$SSH_CMD ${REMOTE_USER}@${REMOTE_HOST} "systemctl --user is-active ${SERVICE_NAME}"

echo "========================================="
echo "🎉 DEPLOYMENT SUCCESSFUL!"
echo "Server running on port 4000 and proxy-mapped to $DOMAIN"
echo "========================================="
