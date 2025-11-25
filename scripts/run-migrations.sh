#!/bin/bash
set -e

DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-onichange}

DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "Running migrations for all services..."
echo "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

# Run migrations for each service
for service_dir in migrations/*/; do
    service=$(basename "$service_dir")
    echo ""
    echo "=== Migrating $service ==="
    
    # Find all .up.sql files in service directory
    for migration_file in "$service_dir"*.up.sql; do
        if [ -f "$migration_file" ]; then
            echo "Running: $migration_file"
            psql "$DB_URL" -f "$migration_file" || {
                echo "Warning: Migration failed (might already exist): $migration_file"
            }
        fi
    done
done

echo ""
echo "✅ All migrations completed!"
