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