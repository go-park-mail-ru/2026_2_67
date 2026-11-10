INSERT INTO "user" (email, password_hash)
VALUES
    ('seed.digital-workshop@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.northern-thread@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('seed.clay-and-wood@example.test', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

INSERT INTO seller (user_id, name, description)
SELECT users.id, profiles.name, profiles.description
FROM (VALUES
    ('seed.digital-workshop@example.test', 'Цифровая мастерская', 'Магазин компьютерной периферии, аудиотехники и полезных аксессуаров для работы и дома.'),
    ('seed.northern-thread@example.test', 'Северная нить', 'Повседневная одежда и аксессуары из натуральных материалов для города и путешествий.'),
    ('seed.clay-and-wood@example.test', 'Глина и дерево', 'Мастерская керамики и деревянных предметов интерьера ручной работы.')
) AS profiles(email, name, description)
JOIN "user" AS users ON users.email = profiles.email;

INSERT INTO product (category_id, seller_id, price, discount, name, description, available_count)
SELECT c.id, s.user_id, products.price, products.discount,
       products.name, products.description, products.available_count
FROM (VALUES
    ('Техника', 'seed.digital-workshop@example.test', 3490.00, 10, 'Беспроводная мышь Silent Click', 'Бесшумная мышь с регулируемой чувствительностью до 2400 DPI. Подключается по Bluetooth или через USB-приёмник, подходит для работы за ноутбуком.', 35),
    ('Техника', 'seed.digital-workshop@example.test', 7290.00, 5, 'Механическая клавиатура KeyCraft TKL', 'Компактная клавиатура без цифрового блока с тактильными переключателями, русской раскладкой и белой подсветкой. Съёмный кабель USB-C входит в комплект.', 18),
    ('Техника', 'seed.digital-workshop@example.test', 5490.00, 0, 'Наушники AirBeat Lite', 'Внутриканальные Bluetooth-наушники с микрофоном и сенсорным управлением. Работают до 6 часов без подзарядки и до 24 часов с зарядным футляром.', 24),
    ('Техника', 'seed.digital-workshop@example.test', 3190.00, 15, 'Внешний аккумулятор PowerGo 20000', 'Портативный аккумулятор ёмкостью 20000 мА·ч с разъёмами USB-A и USB-C. Поддерживает быструю зарядку мощностью до 20 Вт и показывает уровень заряда на дисплее.', 30),
    ('Техника', 'seed.digital-workshop@example.test', 4290.00, 0, 'USB-C хаб Connect 6 в 1', 'Алюминиевый адаптер с HDMI, двумя портами USB-A, USB-C Power Delivery и слотами для SD и microSD. Позволяет подключить монитор и периферию к ноутбуку одним кабелем.', 16),
    ('Техника', 'seed.digital-workshop@example.test', 8990.00, 10, 'Веб-камера ClearView Full HD', 'Камера для видеозвонков с разрешением 1920×1080 при 30 кадрах в секунду, автофокусом и встроенным микрофоном. Оснащена защитной шторкой и креплением на монитор.', 12),
    ('Одежда', 'seed.northern-thread@example.test', 2490.00, 0, 'Хлопковая футболка «Белое море»', 'Белая футболка прямого кроя из плотного хлопка 200 г/м² с небольшим принтом волны. Размер M, усиленный ворот и мягкие плоские швы.', 40),
    ('Одежда', 'seed.northern-thread@example.test', 5990.00, 10, 'Худи «Северный ветер»', 'Тёмно-синее худи свободного кроя из хлопкового футера с начёсом. Размер L, объёмный капюшон и карман-кенгуру для прохладной погоды.', 22),
    ('Одежда', 'seed.northern-thread@example.test', 1890.00, 5, 'Шапка из шерсти мериноса «Тайга»', 'Тёплая вязаная шапка оливкового цвета с широким отворотом. Состав: 80% шерсть мериноса и 20% полиамид, подходит на обхват головы 54–58 см.', 28),
    ('Одежда', 'seed.northern-thread@example.test', 1590.00, 0, 'Льняной шоппер «Городской маршрут»', 'Многоразовая сумка из плотной смеси льна и хлопка с внутренним карманом для ключей. Размер 38×42 см, длинные ручки позволяют носить её на плече.', 32),
    ('Уникальное', 'seed.clay-and-wood@example.test', 2190.00, 0, 'Керамическая кружка «Туман» 350 мл', 'Кружка ручной работы с удобной ручкой и матовой серо-голубой глазурью. Подходит для горячих напитков; оттенок и рисунок глазури могут немного отличаться у каждого изделия.', 10),
    ('Уникальное', 'seed.clay-and-wood@example.test', 3790.00, 5, 'Дубовая сервировочная доска «Лес»', 'Доска из массива дуба размером 35×20 см для подачи сыра, фруктов и закусок. Поверхность обработана пищевым минеральным маслом, предусмотрено отверстие для подвешивания.', 8),
    ('Уникальное', 'seed.clay-and-wood@example.test', 4590.00, 0, 'Керамическая ваза «Песчаный берег»', 'Настольная ваза высотой 22 см с фактурной поверхностью и песочной глазурью. Внутреннее покрытие удерживает воду, поэтому ваза подходит и для свежих цветов, и для сухоцветов.', 7),
    ('Уникальное', 'seed.clay-and-wood@example.test', 1290.00, 10, 'Набор подставок «Годовые кольца»', 'Четыре круглые подставки под чашки из берёзовой фанеры с гравировкой древесных колец. Диаметр 10 см, пробковая основа защищает стол от царапин.', 20)
) AS products(category, seller_email, price, discount, name, description, available_count)
JOIN category AS c ON c.category = products.category
JOIN "user" AS u ON u.email = products.seller_email
JOIN seller AS s ON s.user_id = u.id;
