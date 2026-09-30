DELETE FROM "user"
WHERE login IN (
    'seed.buyer.one@example.test',
    'seed.buyer.two@example.test',
    'seed.seller.one@example.test',
    'seed.seller.two@example.test',
    'seed.pickup.one@example.test',
    'seed.pickup.two@example.test'
);