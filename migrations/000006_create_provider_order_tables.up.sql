CREATE TABLE IF NOT EXISTS provider_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    product_type VARCHAR(50) NOT NULL,
    provider_service_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(150),
    brand VARCHAR(150),
    platform VARCHAR(100),
    min_quantity BIGINT,
    max_quantity BIGINT,
    price NUMERIC(18, 2) NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_response JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(provider, product_type, provider_service_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    product_type VARCHAR(50) NOT NULL,
    ref_id VARCHAR(100) UNIQUE NOT NULL,
    service_id UUID,
    provider_service_id VARCHAR(100) NOT NULL,
    target TEXT,
    quantity BIGINT,
    customer_no VARCHAR(150),
    customer_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    provider_charge NUMERIC(18, 2) NOT NULL DEFAULT 0,
    profit NUMERIC(18, 2) NOT NULL DEFAULT 0,
    start_count BIGINT,
    remains BIGINT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    provider_response JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_orders_service
        FOREIGN KEY (service_id) REFERENCES provider_services(id) ON DELETE SET NULL,
    CONSTRAINT fk_orders_created_by
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_orders_status
        CHECK (status IN ('pending', 'processing', 'completed', 'partial', 'failed', 'cancelled'))
);

CREATE TABLE IF NOT EXISTS order_status_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    provider_status VARCHAR(100),
    provider_response JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_order_status_logs_order
        FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS provider_balance_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    balance NUMERIC(18, 2) NOT NULL DEFAULT 0,
    raw_response JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_api_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    product_type VARCHAR(50),
    endpoint VARCHAR(255) NOT NULL,
    request_ref VARCHAR(100),
    response_status INT,
    response_body JSONB NOT NULL DEFAULT '{}'::jsonb,
    duration_ms INT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    balance NUMERIC(18, 2) NOT NULL DEFAULT 0,
    locked_balance NUMERIC(18, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_wallets_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_wallets_balance_non_negative
        CHECK (balance >= 0),
    CONSTRAINT chk_wallets_locked_balance_non_negative
        CHECK (locked_balance >= 0)
);

CREATE TABLE IF NOT EXISTS deposit_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount NUMERIC(18, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    method VARCHAR(50) NOT NULL,
    provider VARCHAR(50),
    payment_reference VARCHAR(150),
    payment_url TEXT,
    expired_at TIMESTAMP,
    paid_at TIMESTAMP,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_deposit_requests_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_deposit_requests_created_by
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_deposit_requests_amount_positive
        CHECK (amount > 0),
    CONSTRAINT chk_deposit_requests_status
        CHECK (status IN ('pending', 'paid', 'expired', 'failed', 'cancelled')),
    CONSTRAINT chk_deposit_requests_method
        CHECK (method IN ('manual_admin', 'payment_gateway'))
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL,
    user_id UUID NOT NULL,
    order_id UUID,
    deposit_request_id UUID,
    type VARCHAR(50) NOT NULL,
    direction VARCHAR(10) NOT NULL,
    amount NUMERIC(18, 2) NOT NULL,
    balance_before NUMERIC(18, 2) NOT NULL,
    balance_after NUMERIC(18, 2) NOT NULL,
    reference VARCHAR(150),
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_wallet_transactions_wallet
        FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
    CONSTRAINT fk_wallet_transactions_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_wallet_transactions_order
        FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
    CONSTRAINT fk_wallet_transactions_deposit
        FOREIGN KEY (deposit_request_id) REFERENCES deposit_requests(id) ON DELETE SET NULL,
    CONSTRAINT fk_wallet_transactions_created_by
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_wallet_transactions_type
        CHECK (type IN ('deposit', 'debit_order', 'refund_order', 'adjustment')),
    CONSTRAINT chk_wallet_transactions_direction
        CHECK (direction IN ('credit', 'debit')),
    CONSTRAINT chk_wallet_transactions_amount_positive
        CHECK (amount > 0)
);

CREATE TABLE IF NOT EXISTS payment_gateway_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    event_type VARCHAR(100),
    request_id VARCHAR(150),
    payment_reference VARCHAR(150),
    deposit_request_id UUID,
    signature VARCHAR(255),
    status VARCHAR(100),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_payment_gateway_logs_deposit
        FOREIGN KEY (deposit_request_id) REFERENCES deposit_requests(id) ON DELETE SET NULL
);

CREATE OR REPLACE FUNCTION create_wallet_for_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallets (id, user_id, currency, is_active, created_at)
    VALUES (gen_random_uuid(), NEW.id, 'IDR', TRUE, CURRENT_TIMESTAMP)
    ON CONFLICT (user_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_create_wallet_for_new_user ON users;
CREATE TRIGGER trg_create_wallet_for_new_user
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_wallet_for_new_user();

CREATE INDEX IF NOT EXISTS idx_provider_services_provider_type ON provider_services(provider, product_type);
CREATE INDEX IF NOT EXISTS idx_provider_services_platform ON provider_services(platform);
CREATE INDEX IF NOT EXISTS idx_provider_services_category ON provider_services(category);
CREATE INDEX IF NOT EXISTS idx_provider_services_active ON provider_services(is_active);
CREATE INDEX IF NOT EXISTS idx_provider_services_deleted_at ON provider_services(deleted_at);

CREATE INDEX IF NOT EXISTS idx_orders_provider_type ON orders(provider, product_type);
CREATE INDEX IF NOT EXISTS idx_orders_ref_id ON orders(ref_id);
CREATE INDEX IF NOT EXISTS idx_orders_service_id ON orders(service_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_by ON orders(created_by);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

CREATE INDEX IF NOT EXISTS idx_order_status_logs_order_id ON order_status_logs(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_logs_created_at ON order_status_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_provider_balance_snapshots_provider ON provider_balance_snapshots(provider);
CREATE INDEX IF NOT EXISTS idx_provider_balance_snapshots_created_at ON provider_balance_snapshots(created_at);

CREATE INDEX IF NOT EXISTS idx_provider_api_logs_provider ON provider_api_logs(provider);
CREATE INDEX IF NOT EXISTS idx_provider_api_logs_product_type ON provider_api_logs(product_type);
CREATE INDEX IF NOT EXISTS idx_provider_api_logs_request_ref ON provider_api_logs(request_ref);
CREATE INDEX IF NOT EXISTS idx_provider_api_logs_created_at ON provider_api_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_deleted_at ON wallets(deleted_at);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_id ON wallet_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_order_id ON wallet_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_deposit_request_id ON wallet_transactions(deposit_request_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_transactions_reference_unique
    ON wallet_transactions(reference)
    WHERE reference IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);

CREATE INDEX IF NOT EXISTS idx_deposit_requests_user_id ON deposit_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_deposit_requests_status ON deposit_requests(status);
CREATE INDEX IF NOT EXISTS idx_deposit_requests_method ON deposit_requests(method);
CREATE UNIQUE INDEX IF NOT EXISTS idx_deposit_requests_provider_reference_unique
    ON deposit_requests(provider, payment_reference)
    WHERE provider IS NOT NULL AND provider <> ''
      AND payment_reference IS NOT NULL AND payment_reference <> '';
CREATE INDEX IF NOT EXISTS idx_deposit_requests_created_at ON deposit_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_deposit_requests_deleted_at ON deposit_requests(deleted_at);

CREATE INDEX IF NOT EXISTS idx_payment_gateway_logs_provider ON payment_gateway_logs(provider);
CREATE INDEX IF NOT EXISTS idx_payment_gateway_logs_payment_reference ON payment_gateway_logs(payment_reference);
CREATE INDEX IF NOT EXISTS idx_payment_gateway_logs_deposit_request_id ON payment_gateway_logs(deposit_request_id);
CREATE INDEX IF NOT EXISTS idx_payment_gateway_logs_created_at ON payment_gateway_logs(created_at);
