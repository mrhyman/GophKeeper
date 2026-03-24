VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS = -X gophkeeper/pkg/buildinfo.Version=$(VERSION) \
          -X gophkeeper/pkg/buildinfo.BuildDate=$(BUILD_DATE)
DSN ?= postgres://gophkeeper:gophkeeper_secret@localhost:5432/gophkeeper?sslmode=disable

# ============================================
# Сборка
# ============================================
.PHONY: build-server
build-server:
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

.PHONY: build-client
build-client:
	go build -ldflags "$(LDFLAGS)" -o bin/client ./cmd/client

.PHONY: build
build: build-server build-client

# ============================================
# Кросс-компиляция клиента
# ============================================
.PHONY: build-client-all
build-client-all:
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-linux-amd64 ./cmd/client
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/client-linux-arm64 ./cmd/client
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-darwin-amd64 ./cmd/client
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/client-darwin-arm64 ./cmd/client
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/client-windows-amd64.exe ./cmd/client

# ============================================
# TLS-сертификаты
# ============================================
.PHONY: certs
certs:
	./scripts/gen-certs.sh

.PHONY: certs-clean
certs-clean:
	rm -rf ./certs

# ============================================
# Тесты
# ============================================
.PHONY: test
test:
	go test -race -cover -coverprofile=coverage.out ./...

.PHONY: coverage
coverage: test
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: test
	go tool cover -html=coverage.out -o coverage.html

# ============================================
# Линтер
# ============================================
.PHONY: lint
lint:
	golangci-lint run ./...

# ============================================
# Документация
# ============================================

# Godoc - документация кода
.PHONY: docs
docs:
	@echo "Starting godoc server at http://localhost:6060/pkg/gophkeeper"
	godoc -http=:6060

.PHONY: docs-install
docs-install:
	go install golang.org/x/tools/cmd/godoc@latest

# Swagger - API документация
.PHONY: swagger
swagger:
	swag init -g ./cmd/server/main.go -o ./docs

.PHONY: swagger-serve
swagger-serve:
	swag init -g ./cmd/server/main.go -o ./docs
	@echo "Swagger UI available at http://localhost:8080/swagger/"

# ============================================
# Миграции
# ============================================
.PHONY: migrate-up
migrate-up:
	goose -dir internal/server/migrations postgres "$(DSN)" up

.PHONY: migrate-down
migrate-down:
	goose -dir internal/server/migrations postgres "$(DSN)" down

.PHONY: migrate-status
migrate-status:
	goose -dir internal/server/migrations postgres "$(DSN)" status

# ============================================
# Docker
# ============================================
.PHONY: docker-build
docker-build:
	docker compose build --build-arg VERSION=$(VERSION) --build-arg BUILD_DATE=$(BUILD_DATE)

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f server

.PHONY: docker-dev
docker-dev:
	docker compose --profile dev up -d

.PHONY: docker-clean
docker-clean:
	docker compose down -v --remove-orphans

# ============================================
# Всё вместе
# ============================================
.PHONY: run-server
run-server: docker-up
	@echo "Server is running at http://localhost:8080"
	@echo "Swagger UI at http://localhost:8080/swagger/"

.PHONY: run-client
run-client: build-client
	./bin/client

.PHONY: all
all: lint test build swagger
