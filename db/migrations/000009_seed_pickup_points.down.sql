-- Не допускаем каскадного удаления заказов или других профилей аккаунта.
DO $$
DECLARE
    seed_emails CONSTANT TEXT[] := ARRAY[
        'seed.pickup.moscow.01@example.test',
        'seed.pickup.moscow.02@example.test',
        'seed.pickup.moscow.03@example.test',
        'seed.pickup.moscow.04@example.test',
        'seed.pickup.moscow.05@example.test',
        'seed.pickup.moscow.06@example.test',
        'seed.pickup.moscow.07@example.test',
        'seed.pickup.moscow.08@example.test',
        'seed.pickup.moscow.09@example.test',
        'seed.pickup.moscow.10@example.test',
        'seed.pickup.moscow.11@example.test',
        'seed.pickup.moscow.12@example.test'
    ];
    seed_ids BIGINT[];
BEGIN
    -- Блокировки исключают добавление новых FK-ссылок между проверкой и удалением.
    PERFORM id FROM "user" WHERE email = ANY(seed_emails) ORDER BY id FOR UPDATE;
    SELECT ARRAY_AGG(id) INTO seed_ids FROM "user" WHERE email = ANY(seed_emails);
    PERFORM user_id FROM pickup_point WHERE user_id = ANY(seed_ids) ORDER BY user_id FOR UPDATE;

    IF EXISTS (SELECT 1 FROM "order" WHERE pickup_point_id = ANY(seed_ids)) THEN
        RAISE EXCEPTION 'Cannot remove seeded Moscow pickup points: orders reference them';
    END IF;

    IF EXISTS (SELECT 1 FROM buyer WHERE user_id = ANY(seed_ids))
        OR EXISTS (SELECT 1 FROM seller WHERE user_id = ANY(seed_ids)) THEN
        RAISE EXCEPTION 'Cannot remove seeded Moscow pickup points: accounts have other profiles';
    END IF;

    -- Профили pickup_point удаляются по ON DELETE CASCADE.
    DELETE FROM "user" WHERE id = ANY(seed_ids);
END;
$$;
