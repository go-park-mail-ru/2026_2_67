# Buyers: SQL-запросы

## Миграции

Основные таблицы создаются миграцией: `db/migrations/000001_create_user.up.sql`,
`db/migrations/000005_update_updated_at.up.sql`

## GET /api/v1/buyers/{id}/profile

Контракт написан в: `backend/docs/buyers/profile/README.md`

### SQL-запрос

```sql
SELECT
    "user".id,
    "user".email,
    "user".password_hash,
    "user".created_at,
    "user".updated_at,
    buyer.name,
    buyer.surname,
    buyer.birth_date,
    buyer.description,
    buyer.avatar_url,
    buyer.telephone,
    buyer.created_at,
    buyer.updated_at
FROM buyer
LEFT JOIN "user" ON buyer.user_id = "user".id
WHERE "user".id = $1;
```

## POST /api/v1/auth/register

Контракт написан в: `backend/docs/auth/README.md`

Данный запрос написан на временное использование. В дальнейшем предполагается перевод на транзакцию.

```sql
INSERT INTO buyer
    (user_id, name, surname, birth_date, description, avatar_url, telephone)
VALUES
    ($1, $2, $3, $4, $5, $6, $7);
```
