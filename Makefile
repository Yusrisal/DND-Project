.PHONY: api web migrate-up migrate-down migrate-status build tidy check
.PHONY: deploy deploy-setup deploy-status deploy-logs deploy-rollback

# Dev: run `make api` and `make web` in two shells.
api:
	cd backend && go run ./cmd/server

web:
	cd frontend && npm run dev

migrate-up:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down 1

migrate-status:
	cd backend && go run ./cmd/migrate status

# Release build: one static binary plus the frontend bundle.
build:
	cd backend && go build -o bin/dnd-api ./cmd/server
	cd backend && go build -o bin/dnd-migrate ./cmd/migrate
	cd frontend && npm ci && npm run build

tidy:
	cd backend && go mod tidy
	cd backend && gofmt -w .

check:
	cd backend && go vet ./...
	cd backend && go test ./...
	cd frontend && npm run typecheck
	cd frontend && npm test

# Deployment to the homeserver. Configure deploy/deploy.env first; see
# docs/deployment.md.
deploy:
	./deploy/dnd-deploy release

deploy-setup:
	./deploy/dnd-deploy setup

deploy-status:
	./deploy/dnd-deploy status

deploy-logs:
	./deploy/dnd-deploy logs 100

deploy-rollback:
	./deploy/dnd-deploy rollback
