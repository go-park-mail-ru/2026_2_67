DELETE FROM oc_config
WHERE (scope, key) IN (
    ('common', 'min_order_amount')
);

DELETE FROM oc_feature_flag
WHERE (scope, key) IN (
    ('common', 'checkout_enabled'),
    ('common', 'reviews_enabled')
);
