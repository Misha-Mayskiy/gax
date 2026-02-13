# docker-entrypoint.sh
#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."

# Ждем доступности PostgreSQL
until pg_isready -h postgres -p 5432 -U postgres; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is up - running migrations..."

# Применяем миграции
echo "Running database migrations..."
goose -dir ./migrations postgres "host=postgres port=5432 user=postgres password=postgres dbname=users sslmode=disable" up

echo "Migrations completed - starting application..."
exec ./user-service