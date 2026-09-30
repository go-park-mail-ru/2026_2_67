DROP INDEX idx_order_notification_unread;
DROP INDEX idx_order_notification_order_status_id;
DROP INDEX idx_order_notification_order_id;
DROP TABLE order_notification;

DROP INDEX idx_order_product_product_id;
DROP INDEX idx_order_product_order_id;
DROP TABLE order_product;

DROP INDEX idx_order_pickup_point_id;
DROP INDEX idx_order_order_status_id;
DROP INDEX idx_order_buyer_id;
DROP TABLE "order";

DROP TABLE order_status;