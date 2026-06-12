CREATE TABLE IF NOT EXISTS app_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(150) NOT NULL UNIQUE,
    display_name VARCHAR(150) NOT NULL,
    category VARCHAR(100) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_configs_category ON app_configs(category);
CREATE INDEX IF NOT EXISTS idx_app_configs_is_active ON app_configs(is_active);
CREATE INDEX IF NOT EXISTS idx_app_configs_deleted_at ON app_configs(deleted_at);

INSERT INTO menu_items (id, name, display_name, path, icon, order_index, is_active)
VALUES
    (gen_random_uuid(), 'configs', 'Configurations', '/configs', 'bi-sliders', 903, TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (id, name, display_name, resource, action) VALUES
    (gen_random_uuid(), 'list_configs', 'List Configurations', 'configs', 'list'),
    (gen_random_uuid(), 'view_configs', 'View Configuration Detail', 'configs', 'view'),
    (gen_random_uuid(), 'update_configs', 'Update Configurations', 'configs', 'update')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.resource = 'configs'
WHERE r.name IN ('admin', 'superadmin')
ON CONFLICT DO NOTHING;

INSERT INTO app_configs (id, config_key, display_name, category, value, description, is_active)
VALUES
    (
        gen_random_uuid(),
        'auth.public_registration_enabled',
        'Public Registration Enabled',
        'auth',
        'true',
        'Enable or disable public self-registration. When disabled, public register-related endpoints reject requests.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'auth.register_otp_enabled',
        'Register OTP Enabled',
        'auth',
        'false',
        'Enable OTP verification for public user registration.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'auth.password_reset_email_enabled',
        'Password Reset Email Enabled',
        'auth',
        'false',
        'Enable password reset tokens to be sent through the email sender service.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'pricing.default_markup_percent',
        'Default Markup Percent',
        'pricing',
        '0',
        'Default markup percentage applied when no product-specific pricing markup is configured.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'pricing.product_markup_percent.smm',
        'SMM Markup Percent',
        'pricing',
        '5',
        'Markup percentage applied to SMM provider charge before charging reseller or enduser wallet.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'pricing.smm_service_price_max_age',
        'SMM Service Price Max Age',
        'pricing',
        '24h',
        'Maximum age for cached SMM service price before lazy sync is triggered.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'payment.qris.fee_percent',
        'QRIS Topup Fee Percent',
        'payment',
        '5',
        'QRIS topup fee percentage added to the user requested deposit amount.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'payment.qris.dynamic_provider',
        'Dynamic QRIS Provider',
        'payment',
        'boqris',
        'Dynamic QRIS payment provider. Supported values: qrisly, boqris.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'payment.qris.image_url',
        'QRIS Image URL',
        'payment',
        '',
        'Public QRIS image URL shown to users when creating a QRIS deposit.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'payment.qris.merchant_name',
        'QRIS Merchant Name',
        'payment',
        '',
        'Merchant name shown to users when creating a QRIS deposit.',
        TRUE
    ),
    (
        gen_random_uuid(),
        'support.whatsapp_number',
        'Support WhatsApp Number',
        'support',
        '6281234567890',
        'WhatsApp number for customer support. Format: 628xxx (no plus sign).',
        TRUE
    ),
    (
        gen_random_uuid(),
        'support.telegram_username',
        'Support Telegram Username',
        'support',
        '',
        'Telegram username for customer support (without @).',
        TRUE
    ),
    (
        gen_random_uuid(),
        'support.email_address',
        'Support Email Address',
        'support',
        'support@3zadigital.com',
        'Email address for customer support inquiries.',
        TRUE
    )
ON CONFLICT (config_key) DO NOTHING;
