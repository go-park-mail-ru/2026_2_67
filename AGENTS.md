## Codestyle

### Структура модулей и зависимости

- Разделяйте код предметной области по слоям:
  - `domain/` содержит сущности, DTO, интерфейсы репозиториев и usecase, а также ошибки предметной области. Например, `domain/product.go` объявляет `ProductRepo` и `ProductUsecase`.
  - `usecase/` содержит прикладную логику и реализации интерфейсов usecase. Например, `usecase/product.go` использует `domain.ProductRepo` через `ProductUsecase`.
  - `delivery/` содержит HTTP-обработчики, структуры запросов и ответов, преобразование HTTP-данных в DTO domain. Например, `delivery/products.go` вызывает `h.productUsecase.Products(dto)`.
  - `repository/<driver>/` содержит реализации репозиториев для конкретного хранилища. Например, `repository/postgres/product.go` реализует `domain.ProductRepo` через `*sql.DB`.
- Направляйте зависимости внутрь: `delivery` зависит от `domain` и usecase, `usecase` зависит от `domain`, конкретный `repository/postgres` зависит от `domain` и драйвера базы данных. `domain` не должен импортировать `delivery`, `usecase` или конкретную реализацию репозитория.
- Оставляйте интерфейсы в `domain`, рядом с сущностями, которые они обслуживают: `ProductRepo` и `ProductReviewRepo` описывают необходимое поведение, а `ProductRepoPostgres` и `ProductReviewRepoPostgres` реализуют его в `repository/postgres`.
- Используйте алиасы импортов только при конфликте имён пакетов. Например, `orderDomain` и `sellerDomain` отличают домены заказов и продавцов от локального пакета `products/domain`.
- Файл `main.go` внутри модуля используйте для сборки зависимостей слоя: `products/main.go` создаёт usecase, handler и HTTP-маршруты; `usecase/main.go` создаёт структуры usecase и принимает интерфейсы репозиториев.

### Именование файлов

- Называйте файлы строчными буквами в `snake_case`; имя должно описывать основную сущность или роль файла. Примеры: `product.go`, `product_review.go`, `products_handler.go`.
- `handler`-файлы должны наименоваться с префиксами во множественном числе: `products_handler.go`, а не `product_handler.go`.
- Повторяйте имя сущности в каждом слое, если файл содержит её обработку: `domain/product.go`, `usecase/product.go`, `delivery/product.go`, `repository/postgres/product.go`.
- Не смешивайте в одном файле разные слои или несвязанные сущности. HTTP-типы и преобразования оставляйте в `delivery`, SQL-запросы — в `repository/postgres`, предметные типы — в `domain`.

### Именование переменных и типов

- Выбирайте имя по роли и типу зависимости: `productRepo` для `domain.ProductRepo`, `orderRepo` для `orderDomain.OrderRepo`, `productReviewUsecase` для `domain.ProductReviewUsecase`.
- Наименование импортов доменов должно быть: `entity` + `Domain`, где `entity` - в единственном числе. Пример: `orderDomain`, `sellerDomain`.
- Для receiver используйте короткое имя, связанное с типом: `u` для `ProductUsecase`, `h` для `ProductsHandler`, `r` для `ProductRepoPostgres`. Не используйте безымянные или несвязанные receiver вроде `x`.
- Называйте входные DTO `dto`, когда метод принимает структуру предметной области: `Products(dto domain.ProductsSearchParams)` и `Post(dto domain.PostProductReviewInDTO)`.
- Называйте результаты по смыслу: `productByID` для `domain.ProductByID`, `pricesResult` для рассчитанных цен, `decodedCursor` для декодированного курсора, `accessToken` для payload токена.
- Используйте общепринятые сокращения Go в верхнем регистре: `ID`, `DTO`, `URL`, `HTTP`, `JWT`, `SQL`. Примеры: `ProductByID`, `PostProductReviewInDTO`, `PictureURLs`, `jwtSecret`.
- Ошибки должны быть с префиксом `Err` (`err`). Если ошибка касается конкретного поля в `struct`, то должно быть: `Err` (`err`) + `PropertyName` + `ErrorContent`. Пример: `ErrReviewInvalid`.
- Если `handler` реализует ручку: `GET ...`, то вначале `Get` не пишется. В других случаях (`POST`, `PATCH`, ...) пишется: `PostProductReview`. Если `URL`-ручки содержит глагол, то независимо от метода, метод не пишется в названии.
- Для `delivery/request` & `delivery/response` структур используется имя: `deliveryFunc` + `Response/Request`. Если это структура для ошибочного ответ, то: `error` + `deliveryFunc` + `краткий смысл ошибки` + `Response`.
- Хендлеры, требующие стандартной авторизации, оборачиваются в middleware: `auth/middleware/DefaultAuthMiddleware`.
- Если `JSON`-поле обязательно, то в `delivery/`-структурах всё равно используется указатель, с последующей проверкой, что он не `nil`.
- `DTO`-структуры именуются: `usecaseFuncName` + `In/Out` + `DTO`.

### Содержание ошибок

- Все содержания ошибок должны быть на русском.

### Обязанности delivery/usecase/domain/repository

- `usecase` НИЧЕГО не знает о `JSON` и т.п. Его `OutDTO` не должны быть с JSON-тегами.
- `usecase` НИЧЕГО не должен знать о том, как реализован `delivery`.
