HOST_PORT ?= 8080
export HOST_PORT

.PHONY: up down reset logs dev test test-backend test-frontend test-docker lint api-types

up:
	docker compose up --build -d --wait
	@echo "Shop is running at http://localhost:$(HOST_PORT) (API docs at http://localhost:$(HOST_PORT)/docs)"

down:
	docker compose down

reset:
	docker compose down -v

logs:
	docker compose logs -f shop

dev:
	@cd frontend && npm install --no-audit --no-fund
	@trap 'kill 0' INT TERM; \
	(cd backend && SEED_CSV=../testdata/products.csv go run ./cmd/api) & \
	(cd frontend && npm run dev) & \
	wait

test: test-backend test-frontend

test-backend:
	cd backend && go vet ./... && go test -race ./...

test-frontend:
	cd frontend && npm install --no-audit --no-fund && npm run typecheck && npm run lint && npm test

test-docker:
	docker run --rm -v "$(CURDIR):/repo" -v shop-go-cache:/go/pkg/mod -w /repo/backend golang:1.26 \
		sh -c "go vet ./... && go test -race ./..."
	docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -w /app node:24-alpine \
		sh -c "npm ci --no-audit --no-fund && npm run typecheck && npm run lint && npm test"

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

api-types:
	cd frontend && npm run gen:api
