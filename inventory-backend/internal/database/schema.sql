-- Every statement here is idempotent ("IF NOT EXISTS"), so the whole file is
-- applied on every startup without touching existing data.

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL   PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS suppliers (
    id         BIGSERIAL   PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    email      TEXT,
    phone      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id          BIGSERIAL   PRIMARY KEY,
    name        TEXT        NOT NULL,
    sku         TEXT        NOT NULL UNIQUE,
    description TEXT,
    price_cents BIGINT      NOT NULL CHECK (price_cents >= 0),
    category_id BIGINT,
    quantity    INTEGER     NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE products ADD COLUMN IF NOT EXISTS supplier_id BIGINT;

-- Postgres has no "ADD CONSTRAINT IF NOT EXISTS", so check the catalog first.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_category_id_fkey') THEN
        ALTER TABLE products ADD CONSTRAINT products_category_id_fkey
            FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_supplier_id_fkey') THEN
        ALTER TABLE products ADD CONSTRAINT products_supplier_id_fkey
            FOREIGN KEY (supplier_id) REFERENCES suppliers (id) ON DELETE SET NULL;
    END IF;
END $$;

-- One row per stock change. "change" is signed (+5 received, -2 sold) and
-- "balance_after" is the product quantity right after this movement.
CREATE TABLE IF NOT EXISTS stock_movements (
    id            BIGSERIAL   PRIMARY KEY,
    product_id    BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    type          TEXT        NOT NULL CHECK (type IN ('IN', 'OUT', 'ADJUST')),
    change        INTEGER     NOT NULL CHECK (change <> 0),
    balance_after INTEGER     NOT NULL CHECK (balance_after >= 0),
    note          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS stock_movements_product_id_idx ON stock_movements (product_id);
