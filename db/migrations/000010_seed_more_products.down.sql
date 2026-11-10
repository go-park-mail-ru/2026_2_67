DELETE FROM product AS p
USING seller AS s, "user" AS u,
    (VALUES
        ('seed.digital-workshop@example.test', 'Беспроводная мышь Silent Click'),
        ('seed.digital-workshop@example.test', 'Механическая клавиатура KeyCraft TKL'),
        ('seed.digital-workshop@example.test', 'Наушники AirBeat Lite'),
        ('seed.digital-workshop@example.test', 'Внешний аккумулятор PowerGo 20000'),
        ('seed.digital-workshop@example.test', 'USB-C хаб Connect 6 в 1'),
        ('seed.digital-workshop@example.test', 'Веб-камера ClearView Full HD'),
        ('seed.northern-thread@example.test', 'Хлопковая футболка «Белое море»'),
        ('seed.northern-thread@example.test', 'Худи «Северный ветер»'),
        ('seed.northern-thread@example.test', 'Шапка из шерсти мериноса «Тайга»'),
        ('seed.northern-thread@example.test', 'Льняной шоппер «Городской маршрут»'),
        ('seed.clay-and-wood@example.test', 'Керамическая кружка «Туман» 350 мл'),
        ('seed.clay-and-wood@example.test', 'Дубовая сервировочная доска «Лес»'),
        ('seed.clay-and-wood@example.test', 'Керамическая ваза «Песчаный берег»'),
        ('seed.clay-and-wood@example.test', 'Набор подставок «Годовые кольца»')
    ) AS products(seller_email, name)
WHERE p.seller_id = s.user_id
    AND s.user_id = u.id
    AND u.email = products.seller_email
    AND p.name = products.name;

DELETE FROM "user"
WHERE email IN (
    'seed.digital-workshop@example.test',
    'seed.northern-thread@example.test',
    'seed.clay-and-wood@example.test'
);
