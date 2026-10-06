<!--
Домены:
1) Пользователь (user).
* Пользователь как покупатель - это buyer.
* Пользователь как продавец - это seller.
* Пользователь как пункт выдачи заказов (ПВЗ) - pickup_point.

2) Корзина (basket).
3) Заказ (order).
4) Уведомление покупателя (order_notification).
5) Промокод (promocode).

buyer, seller, pickup_point - "профили пользователя".
Это позволяет сделать так, чтобы под одним (login, password) пользователь
мог быть и buyer и seller.
-->

# Описание таблиц и ограничений

## user

### Описание
Таблица пользователя для аутентификации.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **email**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(email)) > 0)`
- **password_hash**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(password_hash)) > 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{id} -> email, password_hash, created_at, updated_at`\
`{email} -> id, password_hash, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{email}` и `{id}` не составные ключи\
**3 НФ и НФБК**: `{email}` и `{id}` — потенциальные ключи, других функциональных зависимостей нет.
email

---

## refresh_token

### Описание
Хранит `refresh_token` каждого `user` для разных устройств. Само отношение хранится в `Redis`.

### Ограничения целостности
- **user_id**: `BIGINT`, `FOREIGN KEY`, `NOT NULL` (ссылается на `user.id`)
- **token_hash**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(token_hash)) > 0)`
- **device_name**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(device_name)) > 0)`: чтобы поддерживать разные устройства пользователей
- **is_revoked**: `BOOLEAN`, `NOT NULL`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **expires_at**: `TIMESTAMPTZ`, `NOT NULL`

`{user_id, device_name}` является потенциальным ключом, так как один `user` может быть на разных устройств.

### Функциональные зависимости
`(user_id, device_name) -> token_hash, is_revoked, expires_at`

### НФ

**1 НФ**: все атрибуты атомарны\
**2 НФ**: отсутсвуют частичные зависимости от детерминанта `{user_id, device_name}`\
**3 НФ** и **НФБК**: `{user_id, device_name}` - потенциальный ключ, других функциональных зависимостей нет.

---

## buyer

### Описание
Профиль покупателя.
`
### Ограничения целостности
- **user_id**: `BIGINT`, `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
- **name**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(name)) > 0)`
- **surname**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(surname)) > 0)`
- **birth_date**: `TIMESTAMPTZ`, `NOT NULL`
- **description**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(description)) > 0)`
- **avatar_url**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(avatar_url)) > 0)`
- **telephone**: `TEXT`, `NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(telephone)) > 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{user_id} -> name, surname, birth_date, description, avatar_url, telephone, created_at, updated_at`\
`{telephone} -> user_id, name, surname, birth_date, description, avatar_url, created_at, updated_at`

`{telephone}` является потенциальным ключом, так как два разных покупателя не могут зарегистрировать одинаковые телефоны.

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}`и `{telephone}` не составные ключи\
**3 НФ и НФБК**: `{user_id}` и `{telephone}` — потенциальные ключи, других функциональных зависимостей нет.

---

## seller

### Описание
Профиль продавца.

### Ограничения целостности
- **user_id**: `BIGINT`, `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
- **name**: `TEXT`, `NOT NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(name)) > 0)`
- **description**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(description)) > 0)`
- **avatar_url**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(avatar_url)) > 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{user_id} -> name, description, avatar_url, created_at, updated_at`\
`{name} -> user_id, description, avatar_url, created_at, updated_at`

`{name}` являются потенциальном ключом, так как два разных продавца не могут зарегистрировать одинаковые названия (не может быть две "Пятерочка").

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}` и `{name}` не составные ключи\
**3 НФ и НФБК**: `{user_id}` и `{name}` — потенциальные ключи, других функциональных зависимостей нет.

---

## order

### Описание
Общая информация о заказе покупателя.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **buyer_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `buyer.user_id`)
- **order_status_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order_status.id`)
- **promocode_id**: `BIGINT`, `NULL`, `FOREIGN KEY` (ссылается на `promocode.id`). Может быть `NULL`, так как заказ может быть без промокода.
- **pickup_point_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `pickup_point.user_id`)
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{id} -> buyer_id, order_status_id, promocode_id, pickup_point_id, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` — не составной ключ\
**3 НФ и НФБК**: `{id}` — потенциальный ключ, других функциональных зависимостей нет.

---

## order_product

### Описание
Продукт, который добавлен в заказ покупателя, с фиксированием его количества и цены на момент покупки.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **order_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order.id`)
- **product_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **count**: `BIGINT`, `NOT NULL`, `CHECK (count > 0)`
- **price**: `BIGINT`, `NOT NULL`, `CHECK (price >= 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`

`{order_id, product_id}` является потенциальным ключом, так как в одном заказе может быть несколько продуктов. А один [product](#product) может быть в нескольких заказах.

### Функциональная зависимость
`{id} -> order_id, product_id, count, price, created_at`
`{order_id, product_id} -> id, count, price, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` — не составной ключ и от `{order_id, product_id}` отсутствуют частичные зависимости.
**3 НФ и НФБК**: `{id}` — потенциальный ключ, других функциональных зависимостей нет.

---

## order_status

### Описание
Статусы заказа покупателя. Является перечислением, поэтому множество `status` создаётся при миграции.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **status**: `TEXT`, `NOT NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(status)) > 0)`

### Функциональная зависимость
`{id} -> status`\
`{status} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{status}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{status}` — потенциальные ключи, других функциональных зависимостей нет.

---

## basket

### Описание
Общая информация о корзине покупателя. Каждый покупатель имеет корзину, но в корзине могут НЕ находиться продукты ([basket_product](#basket_product))
 
### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **buyer_id**: `BIGINT`, `NOT NULL`, `UNIQUE`, `FOREIGN KEY` (ссылается на `buyer.user_id`)
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

**Обязательность корзины**: Наличие корзины для покупателя **обязательно** (связь 1:1). Поле **buyer_id** имеет ограничения `NOT NULL` и `UNIQUE`. Корзина автоматически создаётся при регистрации профиля покупателя.

### Функциональная зависимость
`{id} -> buyer_id, updated_at`\
`{buyer_id} -> id, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{buyer_id}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{buyer_id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## basket_product

### Описание
Продукт, содержащийся в корзине покупателя с учётом количества.

### Ограничения целостности
- **basket_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `basket.id`)
- **product_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **count**: `BIGINT`, `NOT NULL`, `CHECK (count > 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

`{buyer_id, product_id}` является потенциальным ключом, так как покупатель не может иметь несколько одинаковых товаров в корзине (для этого есть `count`).

### Функциональные зависимости
`{basket_id, product_id} -> count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: отсутствуют частичные зависимости от детерминанта `{basket_id, product_id}`\
**3 НФ и НФБК**: `{basket_id, product_id}` — потенциальный ключ, других функциональных зависимостей нет.

---

## order_notification

### Описание
Уведомление о изменении состояния заказа покупателя.

### Ограничения целостности
- **order_id**: `BIGINT`, `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `order.id`)
- **order_status_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order_status.id`)
- **is_read**: `BOOLEAN`, `NOT NULL`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{order_id} -> order_status_id, is_read, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{order_id}` — не составной ключ\
**3 НФ и НФБК**: `{order_id}` — потенциальный ключ, других функциональных зависимостей нет.

---

## pickup_point

### Описание
Профиль пункта выдачи заказов (ПВЗ).

<!--
longtitude - долгота,
latitude - широта
-->

### Ограничения целостности
- **user_id**: `BIGINT`, `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
- **longitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (longitude BETWEEN -180 AND 180)`
- **latitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (latitude BETWEEN -90 AND 90)`
- **start_time**: `TIME`, `NOT NULL`
- **end_time**: `TIME`, `NOT NULL`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`

`{longtitude, latitude}` является потенциальным ключом, так как долгота и широта однозначно определяют положение ПВЗ.

### Функциональные зависимости
`{user_id} -> longitude, latitude, start_time, end_time, created_at`\
`{longitude, latitude} -> user_id, start_time, end_time, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: отсутствуют частичные зависимости от детерминанта `{longitude, latitude}`\
**3 НФ и НФБК**: `{user_id}` и `{longitude, latitude}` — потенциальные ключи, других функциональных зависимостей нет.

---

## promocode

### Описание
Промокоды на скидку по категориям товаров.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **promocode**: `TEXT`, `NOT NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(promocode)) > 0)`
- **available_category_id**: `BIGINT`, `NULL`, `FOREIGN KEY` (ссылается на `product_category.id`): `NULL`, если распространяется на все категории.
- **discount**: `BIGINT`, `NOT NULL`, `CHECK (discount BETWEEN 1 AND 100)`
- **start_datetime**: `TIMESTAMPTZ`, `NOT NULL`
- **end_datetime**: `TIMESTAMPTZ`, `NOT NULL`, `CHECK (start_datetime < end_datetime)`

### Функциональные зависимости
`{id} -> promocode, available_category_id, discount, start_datetime, end_datetime`\
`{promocode} -> id, available_category_id, discount, start_datetime, end_datetime`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{promocode}` и `{id}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{promocode}` — потенциальные ключи, других функциональных зависимостей нет.

---

## product_category

### Описание
Категории товаров.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **category**: `TEXT`, `NOT NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(category)) > 0)`

### Функциональные зависимости
`{id} -> category`\
`{category} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{category}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{category}` — потенциальные ключи, других функциональных зависимостей нет.

---

## product

### Описание
Продукт, выставленный на продажу.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **category_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product_category.id`)
- **seller_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `seller.user_id`)
- **price**: `INTEGER`, `NOT NULL`, `CHECK (price >= 0)`
- **name**: `TEXT`, `NOT NULL`, `CHECK (LENGTH(TRIM(name)) > 0)`
- **description**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(description)) > 0)`
- **available_count**: `INTEGER`, `NOT NULL`, `CHECK (available_count >= 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

`{name, seller_id, category_id}` является потенциальным ключом, так как у продуктов могут быть одинаковые `name` от разных продавцов. У одного продавца может быть несколько продуктов (несколько `name`). У одного продавца могут быть одинаковые товары по названию, но разной категории: к примеру, клей категории: "Для дома" и "Строительство".

### Функциональные зависимости
`{id} -> category_id, seller_id, price, name, description, available_count, created_at, updated_at`\
`{name, seller_id, category_id} -> id, price, description, available_count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: отсутствуют частичные зависимости от детерминанта `{name, seller_id, category_id}`\
**3 НФ и НФБК**: `{id}` и `{name, seller_id, category_id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## product_picture

### Описание
Картинка/изображение, приложенное к товару.

### Ограничения целостности
- **id**: `BIGINT`, `PRIMARY KEY`
- **product_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **picture_url**: `TEXT`, `NOT NULL`, `UNIQUE`, `CHECK (LENGTH(TRIM(picture_url)) > 0)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`

### Функциональные зависимости
`{id} -> product_id, picture_url, created_at`\
`{picture_url} -> id, product_id, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{picture_url}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{picture_url}` — потенциальные ключи, других функциональных зависимостей нет.

---

## product_review

### Описание
Отзыв покупателя на товар.

### Ограничения целостности
- **buyer_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `buyer.user_id`)
- **product_id**: `BIGINT`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **review**: `TEXT`, `NULL`, `CHECK (LENGTH(TRIM(review)) > 0)`
- **rating**: `INTEGER`, `NOT NULL`, `CHECK (rating BETWEEN 1 AND 5)`
- **created_at**: `TIMESTAMPTZ`, `NOT NULL`
- **updated_at**: `TIMESTAMPTZ`, `NOT NULL`

`{buyer_id, product_id}`является потенциальным ключом, так как у покупателя может быть несколько отзывов. Также на один продукт могут быть отзывы от нескольких покупателей.

### Функциональные зависимости
`{buyer_id, product_id} -> review, rating, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: отсутствуют частичные зависимости от детерминанта `{buyer_id, product_id}`\
**3 НФ и НФБК**: `{buyer_id, product_id}` — потенциальный ключ, других функциональных зависимостей нет.
