# Auth: SQL-запросы

## Миграции

Таблица создаётся миграцией `db/migrations/000001_create_user.up.sql`:

## POST /api/v1/auth/register

Контракт написан в: `backend/docs/auth/README.md`

```sql
INSERT INTO "user" (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, created_at, updated_at;
```

## POST /api/v1/auth/register

Данный запрос написан на временное использование. В дальнейшем предполагается перевод на транзакцию.

```sql
DELETE FROM "user"
WHERE id = $1
RETURNING id;
```

## GET /api/v1/auth/login

```sql
SELECT id, email, password_hash, created_at, updated_at
FROM "user"
WHERE email = $1;
```

