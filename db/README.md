# ДЗ 2. Внедрение СУБД

Backend проекта VibeMarket использует PostgreSQL для пользователей, профилей покупателей, каталога, корзин и заказов. Refresh-токены хранятся в Redis.


## Схема и подключение

- Описание отношений и ограничений: `db/normalized/relations.md`; диаграмма: `db/normalized/er_diagram.md`.
- Миграции на SQL: `db/migrations/*.up.sql` и `*.down.sql`. Миграции `000007` и `000008` содержат данные категорий и товаров. Сервис `migrate` в `docker-compose.yaml` применяет миграции через golang-migrate.
- Подключение к PostgreSQL: `backend/cmd/server/main.go`, `NewPostgresClient`. 
- Refresh-токены хранятся в Redis, реализация: `backend/internal/auth/repository/redis/tokens.go`. Access-токены создаются в приложении.

Сокращённые пути в таблице начинаются от `backend/internal/`. Маршруты регистрируются в `backend/cmd/server/main.go` и в `main.go` соответствующих модулей.

## Ручки и запросы

| Ручка | SQL-запросы и другие операции | Код |
| --- | --- | --- |
| `POST /api/v1/auth/register` | `INSERT INTO "user" ... RETURNING ...`; `INSERT INTO buyer ...`; запись refresh-токена в Redis. При части ошибок вызывается `DELETE FROM "user" WHERE id = $1 RETURNING id`. | `auth/delivery/register.go` → `auth/usecase/register.go` → `auth/repository/postgres/user.go`, `buyers/repository/postgres/buyer.go` |
| `POST /api/v1/auth/login` | `SELECT ... FROM "user" WHERE email = $1`; после проверки пароля запись refresh-токена в Redis. При действующем access-токене SQL-запрос не выполняется. | `auth/delivery/login.go` → `auth/usecase/login.go` → `auth/repository/postgres/user.go` |
| `POST /api/v1/auth/refresh` | Чтение refresh-токена из Redis; `SELECT ... FROM "user" WHERE id = $1`; запись нового refresh-токена в Redis. | `auth/delivery/refresh.go` → `auth/usecase/refresh.go` → `auth/repository/postgres/user.go`, `auth/repository/redis/tokens.go` |
| `POST /api/v1/auth/logout` | Отзыв токенов устройства в Redis; SQL-запросов нет. | `auth/delivery/logout.go` → `auth/usecase/logout.go` → `auth/repository/redis/tokens.go` |
| `GET /api/v1/buyers/{id}/profile` | Один `SELECT` из `buyer` с `LEFT JOIN "user"` по идентификатору пользователя. | `buyers/delivery/profile.go` → `buyers/usecase/profile.go` → `buyers/repository/postgres/buyer.go` |
| `GET /api/v1/products/` | Один `SELECT` из `product` с `JOIN product_category`, `JOIN seller`, `LEFT JOIN LATERAL` для рейтинга и числа отзывов, подзапросом для изображений; фильтры по категории, продавцу, названию и цене; сортировка и курсорная пагинация. | `products/delivery/products.go` → `products/usecase/products.go` → `products/repository/postgres/product.go` |
| `GET /api/v1/buyers/{id}/basket` | `SELECT 1 FROM buyer`; `INSERT INTO basket ... ON CONFLICT DO NOTHING`; `SELECT` содержимого корзины с соединением `basket_product`, `product`, `product_category` и `seller`. | `buyers/delivery/basket.go` → `buyers/usecase/basket.go` → `buyers/repository/postgres/basket.go` |
| `POST /api/v1/buyers/{id}/basket/products` | Проверка покупателя и доступного остатка товара через `SELECT`; создание корзины при отсутствии; чтение текущего количества; `INSERT INTO basket_product ... ON CONFLICT ... DO UPDATE`. | `buyers/delivery/basket.go` → `buyers/usecase/basket.go` → `buyers/repository/postgres/basket.go` |
| `PATCH /api/v1/buyers/{buyer_id}/basket/products/{product_id}` | Проверка покупателя и доступного остатка через `SELECT`; при количестве 0 выполняется `DELETE FROM basket_product`, иначе `UPDATE basket_product SET count = $3`. | `buyers/delivery/basket.go` → `buyers/usecase/basket.go` → `buyers/repository/postgres/basket.go` |
| `POST /api/v1/buyers/{id}/basket/preview` | При наличии промокода — `SELECT` из `promocode`; чтение корзины теми же запросами, что у `GET .../basket`; `SELECT` ближайшего `pickup_point` с `earth_distance`. Данные не меняются, кроме возможного создания пустой корзины методом `GetBasket`. | `buyers/delivery/basket.go` → `buyers/usecase/basket.go` → `buyers/repository/postgres/basket.go` |
| `POST /api/v1/orders` | В транзакции: `SELECT` товаров корзины с `FOR UPDATE OF p`; `SELECT` ближайшего пункта выдачи; при наличии кода — `SELECT` промокода; `INSERT INTO "order" ... RETURNING id`; для каждого товара `INSERT INTO order_product` и `UPDATE product SET available_count = available_count - $2`; затем `INSERT INTO order_notification` и `DELETE FROM basket_product`. | `orders/delivery/create_order.go` → `orders/usecase/order.go` → `orders/repository/postgres/order.go` |

## Транзакции

Создание заказа выполняется методом `OrderRepoPostgres.CreateOrderTx` в транзакции PostgreSQL с уровнем изоляции `READ COMMITTED`. Запрос `SELECT ... FOR UPDATE OF p` блокирует строки товаров на время проверки остатков и оформления заказа. В этой же транзакции создаются заказ, его позиции и уведомление, уменьшаются остатки и очищается корзина. После успешного выполнения вызывается `Commit`; при ошибке выполняется откат.

Регистрация и изменение корзины пока не объединены в транзакции. При части ошибок регистрации выполняется компенсирующее удаление пользователя. Запись пользователя и профиля в PostgreSQL и сохранение токена в Redis выполняются последовательно.

## Типы данных и ограничения

В миграциях идентификаторы создаются как `BIGINT ... AS IDENTITY`; связи задаются внешними ключами. Для обязательных полей используются `NOT NULL`, для непустых строк — `CHECK (LENGTH(TRIM(...)) > 0)`. Уникальность email обеспечивается индексом по `LOWER(email)`. У корзины уникален `buyer_id`, а у позиции заказа — пара `(order_id, product_id)`.

Количество товара ограничено проверками на положительность или неотрицательность в зависимости от таблицы. Рейтинг отзыва допускает значения от 1 до 5. Даты создания и изменения используют `TIMESTAMPTZ`; дата рождения покупателя в DDL имеет тип `DATE`. Цена товара в DDL — `DOUBLE PRECISION`, цена позиции заказа — `INTEGER`: при оформлении заказа код округляет цену единицы товара после скидок.

## Ссылки на реализацию

- [Подключение СУБД и маршрутов](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/cmd/server/main.go).
- [SQL пользователей](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/auth/repository/postgres/user.go) и [SQL профиля покупателя](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/buyers/repository/postgres/buyer.go).
- [SQL каталога](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/products/repository/postgres/product.go).
- [SQL корзины](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/buyers/repository/postgres/basket.go).
- [Транзакция создания заказа](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/orders/repository/postgres/order.go).
- [Хранение токенов в Redis](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/backend/internal/auth/repository/redis/tokens.go).
- [SQL-миграции](https://github.com/go-park-mail-ru/2026_2_67/tree/c109f890182aa5af148d179305247b9b261b4385/db/migrations) и [описание отношений](https://github.com/go-park-mail-ru/2026_2_67/blob/c109f890182aa5af148d179305247b9b261b4385/db/normalized/relations.md).


