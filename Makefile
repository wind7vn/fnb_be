# F&B Management System Backend Makefile

.PHONY: build test clean deploy log status migrate seed

# Local Development Commands
build:
	go build -o bin/fnb_be ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin/

# Database Actions
migrate:
	go run cmd/migrator/main.go

seed:
	go run cmd/seeder/main.go

# Remote Deployment & Operations
deploy:
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh

log:
	@if command -v sshpass >/dev/null 2>&1; then \
		sshpass -p hunter ssh -o StrictHostKeyChecking=no wind@192.168.1.2 "journalctl --user -u fnb_be.service -f -n 100"; \
	else \
		ssh -o StrictHostKeyChecking=no wind@192.168.1.2 "journalctl --user -u fnb_be.service -f -n 100"; \
	fi

status:
	@if command -v sshpass >/dev/null 2>&1; then \
		sshpass -p hunter ssh -o StrictHostKeyChecking=no wind@192.168.1.2 "systemctl --user status fnb_be.service"; \
	else \
		ssh -o StrictHostKeyChecking=no wind@192.168.1.2 "systemctl --user status fnb_be.service"; \
	fi
