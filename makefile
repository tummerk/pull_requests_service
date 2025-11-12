.PHONY: generate-api
generate-api:
	@echo "Generating API code from OpenAPI spec..."
	powershell -Command "New-Item -ItemType Directory -Force -Path 'internal/server/generated'"
	oapi-codegen -generate types -package generated openapi.yaml > internal/server/generated/models.gen.go
	oapi-codegen -generate chi-server -package generated openapi.yaml > internal/server/generated/server.gen.go

migrate-up:
	migrate -path db/migrations -database ${DB_DSN} up

migrate-down:
	migrate -path db/migrations -database ${DB_DSN} down