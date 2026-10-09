CREATE TABLE products (
    id           INTEGER PRIMARY KEY,
    sku          TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    category     TEXT    NOT NULL,
    price_cents  INTEGER NOT NULL CHECK (price_cents >= 0),
    stock        INTEGER NOT NULL CHECK (stock >= 0),
    weight_grams INTEGER CHECK (weight_grams IS NULL OR weight_grams >= 0),
    created_at   TEXT    NOT NULL,
    updated_at   TEXT    NOT NULL,
    deleted_at   TEXT
);

CREATE UNIQUE INDEX products_sku_active ON products (sku) WHERE deleted_at IS NULL;
CREATE INDEX products_category ON products (category) WHERE deleted_at IS NULL;
CREATE INDEX products_name ON products (name) WHERE deleted_at IS NULL;

CREATE VIRTUAL TABLE products_fts USING fts5 (
    name,
    sku,
    description,
    content = 'products',
    content_rowid = 'id',
    tokenize = 'unicode61 remove_diacritics 2'
);

CREATE TRIGGER products_fts_insert AFTER INSERT ON products BEGIN
    INSERT INTO products_fts (rowid, name, sku, description)
    VALUES (new.id, new.name, new.sku, new.description);
END;

CREATE TRIGGER products_fts_delete AFTER DELETE ON products BEGIN
    INSERT INTO products_fts (products_fts, rowid, name, sku, description)
    VALUES ('delete', old.id, old.name, old.sku, old.description);
END;

CREATE TRIGGER products_fts_update AFTER UPDATE OF name, sku, description ON products BEGIN
    INSERT INTO products_fts (products_fts, rowid, name, sku, description)
    VALUES ('delete', old.id, old.name, old.sku, old.description);
    INSERT INTO products_fts (rowid, name, sku, description)
    VALUES (new.id, new.name, new.sku, new.description);
END;

CREATE TABLE orders (
    id              INTEGER PRIMARY KEY,
    idempotency_key TEXT    NOT NULL UNIQUE,
    request_hash    TEXT    NOT NULL,
    status          TEXT    NOT NULL CHECK (status IN ('pending', 'paid', 'failed')),
    total_cents     INTEGER NOT NULL CHECK (total_cents >= 0),
    card_last4      TEXT    NOT NULL,
    failure_reason  TEXT    NOT NULL DEFAULT '',
    created_at      TEXT    NOT NULL,
    updated_at      TEXT    NOT NULL
);

CREATE INDEX orders_status ON orders (status, created_at);

CREATE TABLE order_items (
    order_id         INTEGER NOT NULL REFERENCES orders (id),
    product_id       INTEGER NOT NULL REFERENCES products (id),
    sku              TEXT    NOT NULL,
    name             TEXT    NOT NULL,
    unit_price_cents INTEGER NOT NULL CHECK (unit_price_cents >= 0),
    quantity         INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, product_id)
);
