.PHONY: test race vet web-test web-build run migrate
test:
	go test ./...
race:
	go test -race ./...
vet:
	go vet ./...
web-test:
	cd web && npm test
web-build:
	cd web && npm run build
run:
	go run ./cmd/server
migrate:
	psql "$${DATABASE_URL}" -f migrations/010_review_schema.sql -f migrations/020_demo_review.sql
