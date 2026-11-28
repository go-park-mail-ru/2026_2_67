INSERT INTO oc_config (key, scope, value_type, value, description, created_by, updated_by)
VALUES
    (
        'min_order_amount',
        'common',
        'number',
        '500'::jsonb,
        'Минимальная сумма заказа для оформления (backend + frontend)',
        1,
        1
    )
ON CONFLICT (scope, key) DO NOTHING;

INSERT INTO oc_feature_flag (key, scope, enabled, rollout_percent, whitelist, blacklist, description, created_by, updated_by)
VALUES
    (
        'checkout_enabled',
        'common',
        TRUE,
        100,
        '{}',
        '{}',
        'Разрешение на оформление заказов (backend + frontend)',
        1,
        1
    ),
    (
        'reviews_enabled',
        'common',
        TRUE,
        100,
        '{}',
        '{}',
        'Разрешение на создание отзывов к товарам (backend + frontend)',
        1,
        1
    )
ON CONFLICT (scope, key) DO NOTHING;
