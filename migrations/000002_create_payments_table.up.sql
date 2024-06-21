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
