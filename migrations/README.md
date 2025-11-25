# Database Migrations

This directory contains database migration files for all microservices.

## Structure

```
migrations/
├── order/
│   └── 000001_create_orders_table.up.sql
├── user/
│   └── 000001_create_users_table.up.sql
├── store/
├── payment/
├── inventory/
└── notification/
```

## Usage

### Install golang-migrate

```bash
brew install golang-migrate
# or
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Run migrations

```bash
# Up migrations
migrate -path migrations/order -database "postgres://user:password@localhost:5432/onichange?sslmode=disable" up

# Down migrations
migrate -path migrations/order -database "postgres://user:password@localhost:5432/onichange?sslmode=disable" down
```

## Migration Naming Convention

- Format: `{version}_{description}.{direction}.sql`
- Example: `000001_create_users_table.up.sql`
- Version: Sequential number (000001, 000002, etc.)
- Direction: `up` for migration, `down` for rollback

