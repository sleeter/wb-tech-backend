CREATE TABLE IF NOT EXISTS deliveries (
    delivery_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);
CREATE TABLE IF NOT EXISTS payments (
    payment_id BIGSERIAL PRIMARY KEY,
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(50),
    provider VARCHAR(100),
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR(100),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);
CREATE TABLE IF NOT EXISTS items (
    item_id BIGSERIAL PRIMARY KEY,
    chrt_id INT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id INT,
    brand VARCHAR(100),
    status INT
);
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