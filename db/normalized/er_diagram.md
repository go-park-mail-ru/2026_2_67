```mermaid
erDiagram
    User ||--o| Buyer : "1:0..1"
    User ||--o| Seller : "1:0..1"
    User ||--o| PickupPoint : "1:0..1"

    Seller ||--o{ Product : "1:N"
    ProductCategory ||--o{ Product : "1:N"
    Product ||--o{ ProductPicture : "1:N"
    
    Buyer ||--o{ ProductReview : "1:N"
    Product ||--o{ ProductReview : "1:N"

    Buyer ||--o| Basket : "1:1"
    Basket ||--o{ BasketProduct : "1:N"
    Product ||--o{ BasketProduct : "1:N"

    ProductCategory ||--o{ Promocode : "1:N"
    
    Buyer ||--o{ Order : "1:N"
    Order ||--o| OrderStatus : "1:1"
    Promocode ||--o{ Order : "1:N"
    PickupPoint ||--o{ Order : "1:N"
    Order ||--o{ OrderProduct : "1:N"
    Product ||--o{ OrderProduct : "1:N"
    Order ||--o| OrderNotification : "1:N"
    OrderNotification ||--o| OrderStatus : "1:N"

    User {
        int id PK
        char(127) login "UNIQUE, NOT NULL"
        char(255) password_hash "NOT NULL"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    Buyer {
        int user_id PK, FK
        char(127) name "NOT NULL"
        char(127) surname
        datetime birth_date "NOT NULL"
        text description
        char(127) avatar_url "NULL; URL к S3, где хранится картинка"
        char(63) email "UNIQUE, NOT NULL"
        char(31) telephone "UNIQUE"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    Seller {
        int user_id PK, FK
        char(63) name "UNIQUE, NOT NULL"
        text description
        char(127) avatar_url "NULL; URL к S3, где хранится картинка"
        char(63) email "UNIQUE"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    ProductCategory {
        int id PK
        char(63) category "UNIQUE, NOT NULL"
    }

    Product {
        int id PK
        int category_id FK
        int seller_id FK
        int price "NOT NULL"
        char(127) name "NOT NULL"
        text description
        int available_count "NOT NULL"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    ProductPicture {
        int id PK
        int product_id FK
        char(127) picture_url "NOT NULL; URL к S3, где хранится картинка"
        datetime created_at "NOT NULL"
    }

    ProductReview {
        int id PK
        int buyer_id FK
        int product_id FK
        text review
        int rating "NOT NULL"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    Basket {
        int id PK
        int buyer_id FK
    }

    BasketProduct {
        int id PK
        int basket_id FK
        int product_id FK
        int count "NOT NULL"
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    OrderStatus {
        int id PK
        char(63) status "UNIQUE, NOT NULL"
    }

    Promocode {
        int id PK
        char(31) promocode "UNIQUE, NOT NULL"
        int available_category_id FK
        int discount "NOT NULL"
        datetime start_datetime "NOT NULL"
        datetime end_datetime "NOT NULL"
    }

    PickupPoint {
        int user_id PK, FK
        double longitude "NOT NULL"
        double latitude "NOT NULL"
        time start_time "NOT NULL"
        time end_time "NOT NULL"
        datetime created_at "NOT NULL"
    }

    Order {
        int id PK
        int buyer_id FK
        int order_status_id FK
        int promocode_id FK "NULL"
        int pickup_point_id FK
        datetime created_at "NOT NULL"
        datetime updated_at "NOT NULL"
    }

    OrderProduct {
        int id PK
        int order_id FK
        int product_id FK
        int count "NOT NULL"
        int price "NOT NULL"
        datetime created_at "NOT NULL"
    }

    OrderNotification {
        int order_id PK, FK
        int order_status_id FK
        boolean is_read "NOT NULL"
        datetime created_at "NOT NULL"
    }
```