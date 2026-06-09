# F&B Management System - Backend Service

This is the high-performance, multi-tenant SaaS backend for the Food & Beverage Management System (POS + Kitchen Display System). It is built with Go, PostgreSQL, and Redis.

For detailed architecture diagrams, Row-Level Security details, ER models, and API routing schemas, please see [docs/architecture.md](file:///Users/wind/projects/fnb_be/docs/architecture.md).

---

## 🛠 Prerequisites

Make sure you have the following installed on your machine:
- **Go 1.26+**
- **PostgreSQL 15+**
- **Redis 6+**

---

## 🚀 Getting Started

### 1. Environment Configuration
Copy `.env.example` to `.env` and fill in your database, Redis, and Firebase credentials:
```bash
cp .env.example .env
```
Ensure your database parameters (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) and Redis URL (`REDIS_URL`) are configured correctly.

### 1.5. Push Notifications & APNs (.p8 Key)
The system uses Firebase Cloud Messaging (FCM) to relay notifications to clients. For iOS notifications:
- Store your Apple APNs private key file (e.g., `AuthKey_S77A4NC4RH.p8`) inside the `configs/` directory.
- Configure your `.env` variables to reference it:
  ```env
  FIREBASE_SERVICE_ACCOUNT_PATH=configs/firebase-service-account.json
  APNS_KEY_PATH=configs/AuthKey_S77A4NC4RH.p8
  APNS_KEY_ID=S77A4NC4RH
  ```
- **Firebase Console Setup**: The `.p8` key must also be uploaded to the **Firebase Console** (under Project Settings -> Cloud Messaging -> Apple app sharing credentials) along with its Key ID (`S77A4NC4RH`) and your Apple Developer Team ID to authorize FCM to push notifications to iOS devices.

### 2. Database Migrations
Run GORM AutoMigrate to create/upgrade the database tables, indices, and constraints:
```bash
go run cmd/migrator/main.go
```

### 3. Database Seeding
Seed the database with initial tenants, default users, product tables, and default items:
```bash
go run cmd/seeder/main.go
```
*Note: This creates a default user with phone number `0982651922` and password `123`.*

### 4. Running the Server Locally
To start the Fiber HTTP and WebSocket API server:
```bash
go run cmd/server/main.go
```
The server will start on port `8080` (or the port defined in your `.env` file).

---

## 🧪 Testing

To run all unit tests in the project:
```bash
go test ./...
```

---

## 🔧 Maintenance Scripts

- **Database Structure Cleanup**: If you have removed/renamed fields in GORM models, run this script to safely drop columns that are no longer mapped in Go models:
  ```bash
  go run cmd/cleanup/main.go
  ```
- **Bank Data Downloader**: Downloads MoMo banks lists and saves to `data/momo_banks.json` (used for QR code billing generation):
  ```bash
  go run cmd/download_banks/main.go
  ```

---

## 📦 Deployment (CI/CD)

The project includes continuous deployment automation:

### Jenkins Pipeline
The build and deployment flow is configured in the root [Jenkinsfile](file:///Users/wind/projects/fnb_be/Jenkinsfile):
- Automatically cleans workspace, checks out code.
- Pulls secure credentials (`.env` and `firebase-service-account.json`).
- Deploys source code to target IP `172.17.0.1` at `/home/wind/fnb/src`.
- Builds Go binaries for the server and migrator.
- Runs database migrations.
- Installs and restarts the user-level systemd service `fnb_be.service`.

### Manual Service Management on Server
On the hosting server, the service is managed via **systemd user services** (bypassing the need for sudo/root permissions):
```bash
# Restart the service
systemctl --user restart fnb_be.service

# View real-time logs
journalctl --user -u fnb_be.service -f -n 100

# Check service status
systemctl --user status fnb_be.service
```
The systemd service configuration file is located at [deploy/fnb_be.service](file:///Users/wind/projects/fnb_be/deploy/fnb_be.service).
