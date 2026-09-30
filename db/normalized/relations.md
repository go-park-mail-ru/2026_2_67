<!--
Домены:
1) Пользователь (User).
* Пользователь как покупатель - это Buyer.
* Пользователь как продавец - это Seller.
* Пользователь как пункт выдачи заказов (ПВЗ) - PickupPoint.

2) Корзина (Basket).
3) Заказ (Order).
4) Уведомление покупателя (OrderNotification).
5) Промокод (Promocode).

Buyer, Seller, PickupPoint - "профили пользователя".
Это позволяет сделать так, чтобы под одним (login, password) пользователь
мог быть и Buyer и Seller.
-->

# Обоснование выбора типов данных PostgreSQL

1. **Даты и временя**: Используется стандартный тип `TIMESTAMP`. Для времени работы ПВЗ используется тип `TIME`.
2. **Строковые данные**: тип `TEXT` был выбран из-за преимуществами производительности над такими типами как `CHAR`.
3. **Числа с плавающей точкой**: для координат ПВЗ (`latitude`, `longitude`) используется `double precision` (`float8`), так как обеспечивает большую точность, чем `float4`.
4. **ID**: выбран `BIGINT`, так как значение шире стандартного `INTEGER`.

---

# Описание таблиц и ограничений

## User

### Описание
Таблица пользователя для аутентификации.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **login**: `TEXT`, `NOT NULL`, `UNIQUE`
- **password_hash**: `TEXT`, `NOT NULL`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{id} -> login, password_hash, created_at, updated_at`\
`{login} -> id, password_hash, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{login}` и `{id}` не составные ключи\
**3 НФ и НФБК**: `{login}` и `{id}` — потенциальные ключи, других ФЗ нет.

---

## Buyer

### Описание
Профиль покупателя.

### Ограничения целостности
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `User.id`)
- **name**: `TEXT`, `NOT NULL`
- **surname**: `TEXT`, `NULL`
- **birth_date**: `TIMESTAMP`, `NOT NULL`
- **description**: `TEXT`, `NULL`
- **avatar_url**: `TEXT`, `NULL`
- **email**: `TEXT`, `NOT NULL`, `UNIQUE`
- **telephone**: `TEXT`, `NULL`, `UNIQUE`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{user_id} -> name, surname, birth_date, description, avatar_url, email, telephone, created_at, updated_at`\
`{email} -> user_id, name, surname, birth_date, description, avatar_url, telephone, created_at, updated_at`\
`{telephone} -> user_id, name, surname, birth_date, description, avatar_url, email, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}`, `{email}` и `{telephone}` не составные ключи\
**3 НФ и НФБК**: `{user_id}`, `{email}` и `{telephone}` — потенциальные ключи, других ФЗ нет.

---

## Seller

### Описание
Профиль продавца.

### Ограничения целостности
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `User.id`)
- **name**: `TEXT`, `NOT NULL`, `UNIQUE`
- **description**: `TEXT`, `NULL`
- **avatar_url**: `TEXT`, `NULL`
- **email**: `TEXT`, `NULL`, `UNIQUE`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{user_id} -> name, description, avatar_url, email, created_at, updated_at`\
`{email} -> user_id, name, description, avatar_url, created_at, updated_at`\
`{name} -> user_id, description, avatar_url, email, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}`, `{email}` и `{name}` не составные ключи\
**3 НФ и НФБК**: `{user_id}`, `{email}` и `{name}` — потенциальные ключи, других ФЗ нет.

---

## Order

### Описание
Общая информация о заказе покупателя.

### Ограничения целостности и бизнес-логика
- **Оформление без промокода**: Поле **promocode_id** имеет ограничение **`NULL`**. Заказ **можно** оформить без промокода.
- **id**: `PRIMARY KEY`
- **buyer_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Buyer.user_id`)
- **order_status_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `OrderStatus.id`)
- **promocode_id**: `INTEGER`, `NULL`, `FOREIGN KEY` (ссылается на `Promocode.id`)
- **pickup_point_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `PickupPoint.user_id`)
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{id} -> buyer_id, order_status_id, promocode_id, pickup_point_id, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` — не составной ключ\
**3 НФ и НФБК**: `{id}` — потенциальный ключ, других ФЗ нет.

---

## OrderProduct

### Описание
Продукт, который добавлен в заказ покупателя, с фиксированием его количества и цены на момент покупки.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **order_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Order.id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Product.id`)
- **count**: `INTEGER`, `NOT NULL`, `CHECK (count > 0)`
- **price**: `INTEGER`, `NOT NULL`, `CHECK (price >= 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`

### Функциональная зависимость
`{id} -> order_id, product_id, count, price, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` — не составной ключ\
**3 НФ и НФБК**: `{id}` — потенциальный ключ, других ФЗ нет.

---

## OrderStatus

### Описание
Статусы заказа покупателя.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **status**: `TEXT`, `NOT NULL`, `UNIQUE`

### Функциональная зависимость
`{id} -> status`\
`{status} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{status}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{status}` — потенциальные ключи, других ФЗ нет.

---

## Basket

### Описание
Общая информация о корзине покупателя. Каждый покупатель имеет корзину, но в корзине могут НЕ находиться продукты ([BasketProduct](#basketproduct))
 
### Ограничения целостности и бизнес-логика
- **Обязательность корзины**: Наличие корзины для покупателя **обязательно** (связь 1:1). Поле **buyer_id** имеет ограничения `NOT NULL` и `UNIQUE`. Корзина автоматически создаётся при регистрации профиля покупателя.
- **id**: `PRIMARY KEY`
- **buyer_id**: `INTEGER`, `NOT NULL`, `UNIQUE`, `FOREIGN KEY` (ссылается на `Buyer.user_id`)

### Функциональная зависимость
`{id} -> buyer_id`\
`{buyer_id} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{buyer_id}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{buyer_id}` — потенциальные ключи, других ФЗ нет.

---

## BasketProduct

### Описание
Продукт, содержащийся в корзине покупателя с учётом количества.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **basket_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Basket.id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Product.id`)
- **count**: `INTEGER`, `NOT NULL`, `CHECK (count > 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`
- **Уникальность составного ключа**: `UNIQUE (basket_id, product_id)`

### Функциональные зависимости
`{id} -> basket_id, product_id, count, created_at, updated_at`\
`{basket_id, product_id} -> id, count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{basket_id, product_id}`\
**3 НФ и НФБК**: `{id}` и `{basket_id, product_id}` — потенциальные ключи, других ФЗ нет.

---

## OrderNotification

### Описание
Уведомление о изменении состояния заказа покупателя.

### Ограничения целостности
- **order_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `Order.id`)
- **order_status_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `OrderStatus.id`)
- **is_read**: `BOOLEAN`, `NOT NULL`
- **created_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{order_id} -> order_status_id, is_read, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{order_id}` — не составной ключ\
**3 НФ и НФБК**: `{order_id}` — потенциальный ключ, других ФЗ нет.

---

## PickupPoint

### Описание
Профиль пункта выдачи заказов (ПВЗ).

### Ограничения целостности
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `User.id`)
- **longitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (longitude BETWEEN -180 AND 180)`
- **latitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (latitude BETWEEN -90 AND 90)`
- **start_time**: `TIME`, `NOT NULL`
- **end_time**: `TIME`, `NOT NULL`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **Уникальность координат**: `UNIQUE (longitude, latitude)`

### Функциональные зависимости
`{user_id} -> longitude, latitude, start_time, end_time, created_at`\
`{longitude, latitude} -> user_id, start_time, end_time, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{longitude, latitude}`\
**3 НФ и НФБК**: `{user_id}` и `{longitude, latitude}` — потенциальные ключи, других ФЗ нет.

---

## Promocode

### Описание
Промокоды на скидку по категориям товаров.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **promocode**: `TEXT`, `NOT NULL`, `UNIQUE`
- **available_category_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `ProductCategory.id`)
- **discount**: `INTEGER`, `NOT NULL`, `CHECK (discount BETWEEN 1 AND 100)`
- **start_datetime**: `TIMESTAMP`, `NOT NULL`
- **end_datetime**: `TIMESTAMP`, `NOT NULL`, `CHECK (start_datetime < end_datetime)`

### Функциональные зависимости
`{id} -> promocode, available_category_id, discount, start_datetime, end_datetime`\
`{promocode} -> id, available_category_id, discount, start_datetime, end_datetime`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{promocode}` и `{id}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{promocode}` — потенциальные ключи, других ФЗ нет.

---

## ProductCategory

### Описание
Категории товаров.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **category**: `TEXT`, `NOT NULL`, `UNIQUE`

### Функциональные зависимости
`{id} -> category`\
`{category} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{category}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{category}` — потенциальные ключи, других ФЗ нет.

---

## Product

### Описание
Продукт, выставленный на продажу.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **category_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `ProductCategory.id`)
- **seller_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Seller.user_id`)
- **price**: `INTEGER`, `NOT NULL`, `CHECK (price >= 0)`
- **name**: `TEXT`, `NOT NULL`
- **description**: `TEXT`, `NULL`
- **available_count**: `INTEGER`, `NOT NULL`, `CHECK (available_count >= 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`
- **Уникальность позиционирования товара**: `UNIQUE (name, seller_id, category_id)`

### Функциональные зависимости
`{id} -> category_id, seller_id, price, name, description, available_count, created_at, updated_at`\
`{name, seller_id, category_id} -> id, price, description, available_count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{name, seller_id, category_id}`\
**3 НФ и НФБК**: `{id}` и `{name, seller_id, category_id}` — потенциальные ключи, других ФЗ нет.

---

## ProductPicture

### Описание
Картинка/изображение, приложенное к товару.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Product.id`)
- **picture_url**: `TEXT`, `NOT NULL`, `UNIQUE`
- **created_at**: `TIMESTAMP`, `NOT NULL`

### Функциональные зависимости
`{id} -> product_id, picture_url, created_at`\
`{picture_url} -> id, product_id, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{picture_url}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{picture_url}` — потенциальные ключи, других ФЗ нет.

---

## ProductReview

### Описание
Отзыв покупателя на товар.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **buyer_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Buyer.user_id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `Product.id`)
- **review**: `TEXT`, `NULL`
- **rating**: `INTEGER`, `NOT NULL`, `CHECK (rating BETWEEN 1 AND 5)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`
- **Один отзыв от одного покупателя на один товар**: `UNIQUE (buyer_id, product_id)`

### Функциональные зависимости
`{id} -> buyer_id, product_id, review, rating, created_at, updated_at`\
`{buyer_id, product_id} -> id, review, rating, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{buyer_id, product_id}`\
**3 НФ и НФБК**: `{id}` и `{buyer_id, product_id}` — потенциальные ключи, других ФЗ нет.