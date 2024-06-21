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