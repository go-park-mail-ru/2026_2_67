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
**3 НФ и НФБК**: `{login}` и `{id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## buyer

### Описание
Профиль покупателя.

### Ограничения целостности
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
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

`{email}` и `{telephone}` являются потенциальными ключами, так как два разных покупателя не могут зарегистрировать одинаковые почты или телефоны.

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}`, `{email}` и `{telephone}` не составные ключи\
**3 НФ и НФБК**: `{user_id}`, `{email}` и `{telephone}` — потенциальные ключи, других функциональных зависимостей нет.

---

## seller

### Описание
Профиль продавца.

### Ограничения целостности
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
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

`{email}` и `{name}` являются потенциальными ключами, так как два разных продавца не могут зарегистрировать одинаковые почты или названия (не может быть две "Пятерочка").

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{user_id}`, `{email}` и `{name}` не составные ключи\
**3 НФ и НФБК**: `{user_id}`, `{email}` и `{name}` — потенциальные ключи, других функциональных зависимостей нет.

---

## order

### Описание
Общая информация о заказе покупателя.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **buyer_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `buyer.user_id`)
- **order_status_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order_status.id`)
- **promocode_id**: `INTEGER`, `NULL`, `FOREIGN KEY` (ссылается на `promocode.id`). Может быть `NULL`, так как заказ может быть без промокода.
- **pickup_point_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `pickup_point.user_id`)
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

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
- **id**: `PRIMARY KEY`
- **order_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order.id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **count**: `INTEGER`, `NOT NULL`, `CHECK (count > 0)`
- **price**: `INTEGER`, `NOT NULL`, `CHECK (price >= 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`

### Функциональная зависимость
`{id} -> order_id, product_id, count, price, created_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` — не составной ключ\
**3 НФ и НФБК**: `{id}` — потенциальный ключ, других функциональных зависимостей нет.

---

## order_status

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
**3 НФ и НФБК**: `{id}` и `{status}` — потенциальные ключи, других функциональных зависимостей нет.

---

## basket

### Описание
Общая информация о корзине покупателя. Каждый покупатель имеет корзину, но в корзине могут НЕ находиться продукты ([basket_product](#basket_product))
 
### Ограничения целостности и бизнес-логика
- **id**: `PRIMARY KEY`
- **buyer_id**: `INTEGER`, `NOT NULL`, `UNIQUE`, `FOREIGN KEY` (ссылается на `buyer.user_id`)

**Обязательность корзины**: Наличие корзины для покупателя **обязательно** (связь 1:1). Поле **buyer_id** имеет ограничения `NOT NULL` и `UNIQUE`. Корзина автоматически создаётся при регистрации профиля покупателя.

### Функциональная зависимость
`{id} -> buyer_id`\
`{buyer_id} -> id`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: `{id}` и `{buyer_id}` не составные ключи\
**3 НФ и НФБК**: `{id}` и `{buyer_id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## basket_product

### Описание
Продукт, содержащийся в корзине покупателя с учётом количества.

### Ограничения целостности
- **basket_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `basket.id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **count**: `INTEGER`, `NOT NULL`, `CHECK (count > 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

`{buyer_id, product_id}` является потенциальным ключом, так как покупатель не может иметь несколько одинаковых товаров в корзине (для этого есть `count`).

### Функциональные зависимости
`{basket_id, product_id} -> count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{basket_id, product_id}`\
**3 НФ и НФБК**: `{basket_id, product_id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## order_notification

### Описание
Уведомление о изменении состояния заказа покупателя.

### Ограничения целостности
- **order_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `order.id`)
- **order_status_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `order_status.id`)
- **is_read**: `BOOLEAN`, `NOT NULL`
- **created_at**: `TIMESTAMP`, `NOT NULL`

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
- **user_id**: `PRIMARY KEY`, `FOREIGN KEY` (ссылается на `user.id`)
- **longitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (longitude BETWEEN -180 AND 180)`
- **latitude**: `DOUBLE PRECISION`, `NOT NULL`, `CHECK (latitude BETWEEN -90 AND 90)`
- **start_time**: `TIME`, `NOT NULL`
- **end_time**: `TIME`, `NOT NULL`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **Уникальность координат**: `UNIQUE (longitude, latitude)`

### Функциональные зависимости
`{user_id} -> longitude, latitude, start_time, end_time, created_at`\
`{longitude, latitude} -> user_id, start_time, end_time, created_at`

`{longtitude, latitude}` является потенциальным ключом, так как долгота и широта однозначно определяют положение ПВЗ.

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{longitude, latitude}`\
**3 НФ и НФБК**: `{user_id}` и `{longitude, latitude}` — потенциальные ключи, других функциональных зависимостей нет.

---

## promocode

### Описание
Промокоды на скидку по категориям товаров.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **promocode**: `TEXT`, `NOT NULL`, `UNIQUE`
- **available_category_id**: `INTEGER`, `NULL`, `FOREIGN KEY` (ссылается на `product_category.id`): `NULL`, если распространяется на весь заказ
- **discount**: `INTEGER`, `NOT NULL`, `CHECK (discount BETWEEN 1 AND 100)`
- **start_datetime**: `TIMESTAMP`, `NOT NULL`
- **end_datetime**: `TIMESTAMP`, `NOT NULL`, `CHECK (start_datetime < end_datetime)`

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
- **id**: `PRIMARY KEY`
- **category**: `TEXT`, `NOT NULL`, `UNIQUE`

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
- **id**: `PRIMARY KEY`
- **category_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product_category.id`)
- **seller_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `seller.user_id`)
- **price**: `INTEGER`, `NOT NULL`, `CHECK (price >= 0)`
- **name**: `TEXT`, `NOT NULL`
- **description**: `TEXT`, `NULL`
- **available_count**: `INTEGER`, `NOT NULL`, `CHECK (available_count >= 0)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

`{name, seller_id, category_id}` является потенциальным ключом, так как у продуктов могут быть одинаковые `name` от разных продавцов. У одного продавца может быть несколько продуктов (несколько `name`). У одного продавца могут быть одинаковые товары по названию, но разной категории: к примеру, клей категории: "Для дома" и "Строительство".

### Функциональные зависимости
`{id} -> category_id, seller_id, price, name, description, available_count, created_at, updated_at`\
`{name, seller_id, category_id} -> id, price, description, available_count, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{name, seller_id, category_id}`\
**3 НФ и НФБК**: `{id}` и `{name, seller_id, category_id}` — потенциальные ключи, других функциональных зависимостей нет.

---

## product_picture

### Описание
Картинка/изображение, приложенное к товару.

### Ограничения целостности
- **id**: `PRIMARY KEY`
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **picture_url**: `TEXT`, `NOT NULL`, `UNIQUE`
- **created_at**: `TIMESTAMP`, `NOT NULL`

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
- **buyer_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `buyer.user_id`)
- **product_id**: `INTEGER`, `NOT NULL`, `FOREIGN KEY` (ссылается на `product.id`)
- **review**: `TEXT`, `NULL`
- **rating**: `INTEGER`, `NOT NULL`, `CHECK (rating BETWEEN 1 AND 5)`
- **created_at**: `TIMESTAMP`, `NOT NULL`
- **updated_at**: `TIMESTAMP`, `NOT NULL`

`{buyer_id, product_id}`является потенциальным ключом, так как у покупателя может быть несколько отзывов. Также на один продукт могут быть отзывы от нескольких покупателей.

### Функциональные зависимости
`{buyer_id, product_id} -> review, rating, created_at, updated_at`

### НФ
**1 НФ**: все атрибуты атомарны\
**2 НФ**: Отсутствуют частичные зависимости от детерминанта `{buyer_id, product_id}`\
**3 НФ и НФБК**: `{buyer_id, product_id}` — потенциальные ключи, других функциональных зависимостей нет.
