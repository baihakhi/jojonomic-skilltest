CREATE TABLE IF NOT EXISTS tbl_harga (
    reff_id VARCHAR (15) PRIMARY KEY NOT NULL,
    admin_id VARCHAR (15) NOT NULL,
    harga_topup INT NOT NULL,
    harga_buyback INT NOT NULL,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tbl_rekening (
    reff_id VARCHAR (15)  PRIMARY KEY NOT NULL,
    norek VARCHAR (15) UNIQUE NOT NULL,
    customer_name VARCHAR (20) NOT NULL,
    gold_balance DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP
);

CREATE TABLE if NOT EXISTS tbl_transaksi (
    reff_id VARCHAR (15)  PRIMARY KEY NOT NULL,
    norek VARCHAR (15),
    type VARCHAR(15),
    gold_weight DECIMAL(12, 2),
    harga_topup INT NOT NULL,
    harga_buyback INT NOT NULL,
    gold_balance DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP
);

INSERT INTO tbl_rekening ( reff_id, norek, customer_name, gold_balance, created_at) VALUES ('xkdfd', 'rek001', 'john doe', 1.34, NOW());