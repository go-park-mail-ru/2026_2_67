DROP TRIGGER buyer_create_basket ON buyer;
DROP FUNCTION create_basket_for_buyer();
DROP INDEX idx_basket_product_product_id;
DROP TABLE basket_product;
DROP TABLE basket;