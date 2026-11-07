# Auth: SQL-запросы

Контракт ручек: [POST `/api/v1/auth`](../../../backend/docs/auth/README.md).

## Миграции

[`db/migrations/000001_create_user.up.sql`](../../migrations/000001_create_user.up.sql)

## SQL repository/postgres

### `POST /api/v1/auth/register`

Создаёт пользователя и возвращает его данные.

```sql
INSERT INTO "user" (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at, updated_at;
```

### `POST /api/v1/auth/login`

Находит пользователя по email для входа.

```sql
SELECT id, email, password_hash, created_at, updated_at FROM "user" WHERE email = $1;
```
