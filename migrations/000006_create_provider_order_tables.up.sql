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
    price NUMERIC(18, 4) NOT NULL DEFAULT 0,
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
    amount NUMERIC(18, 4) NOT NULL DEFAULT 0,
    provider_charge NUMERIC(18, 4) NOT NULL DEFAULT 0,
    profit NUMERIC(18, 4) NOT NULL DEFAULT 0,
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
    balance NUMERIC(18, 4) NOT NULL DEFAULT 0,
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
