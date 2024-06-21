CREATE TABLE IF NOT EXISTS orders (
                        order_uid VARCHAR(255) PRIMARY KEY,
                        track_number VARCHAR(255),
                        entry VARCHAR(50),
                        delivery_id BIGINT REFERENCES deliveries(delivery_id),
                        payment_id BIGINT REFERENCES payments(payment_id),
                        items_ids BIGINT[] NOT NULL,
                        locale VARCHAR(50),
                        internal_signature VARCHAR(255),
                        customer_id VARCHAR(255),
                        delivery_service VARCHAR(100),
                        shardkey VARCHAR(100),
                        sm_id INT,
                        date_created TIMESTAMP NOT NULL,
                        oof_shard VARCHAR(50)
);