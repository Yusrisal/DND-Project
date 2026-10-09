.PHONY: api web migrate-up migrate-down migrate-status build tidy check

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
	cd frontend && npm run typecheck
