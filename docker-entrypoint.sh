#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."
until goose -dir /app/migrations postgres "${DATABASE_DSN}" status > /dev/null 2>&1; do
    echo "PostgreSQL is unavailable — sleeping"
    sleep 2
done
echo "PostgreSQL is up"

echo "Running migrations..."
goose -dir /app/migrations postgres "${DATABASE_DSN}" up
echo "Migrations done"

echo "Starting server..."
exec "$@"
