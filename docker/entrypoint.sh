#!/bin/sh

set -e


echo "Running database migrations..."
./migrate -path ./db/migrations -database "$DB_DSN" up

echo "Migrations applied successfully."


exec "$@"