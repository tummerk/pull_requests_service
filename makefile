.PHONY: generate
generate:
	oapi-codegen --config codegen.yaml openapi.yaml


migrate-up:
	migrate -path db/migrations -database ${DB_DSN} up

migrate-down:
	migrate -path db/migrations -database ${DB_DSN} down