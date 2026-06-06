INSERT INTO menu_items (id, name, display_name, path, icon, order_index) VALUES
    (gen_random_uuid(), 'smm_services', 'SMM Services', '/smm/services', 'bi-megaphone', 100),
    (gen_random_uuid(), 'smm_orders', 'SMM Orders', '/smm/orders', 'bi-receipt', 101),
    (gen_random_uuid(), 'wallet', 'Wallet', '/wallet', 'bi-wallet2', 102),
    (gen_random_uuid(), 'deposits', 'Deposits', '/deposits', 'bi-cash-coin', 103),
    (gen_random_uuid(), 'wallets', 'Admin Wallets', '/admin/wallets', 'bi-safe', 905),
    (gen_random_uuid(), 'admin_deposits', 'Admin Deposits', '/admin/deposits', 'bi-bank', 906)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (id, name, display_name, resource, action) VALUES
    (gen_random_uuid(), 'list_smm_services', 'List SMM Services', 'smm_services', 'list'),
    (gen_random_uuid(), 'sync_smm_services', 'Sync SMM Services', 'smm_services', 'sync'),
    (gen_random_uuid(), 'list_smm_orders', 'List SMM Orders', 'smm_orders', 'list'),
    (gen_random_uuid(), 'list_all_smm_orders', 'List All SMM Orders', 'smm_orders', 'list_all'),
    (gen_random_uuid(), 'view_smm_orders', 'View SMM Order Detail', 'smm_orders', 'view'),
    (gen_random_uuid(), 'create_smm_orders', 'Create SMM Orders', 'smm_orders', 'create'),
    (gen_random_uuid(), 'refresh_status_smm_orders', 'Refresh SMM Order Status', 'smm_orders', 'refresh_status'),
    (gen_random_uuid(), 'view_provider_balance', 'View Provider Balance', 'provider_balance', 'view'),
    (gen_random_uuid(), 'list_provider_api_logs', 'List Provider API Logs', 'provider_api_logs', 'list'),
    (gen_random_uuid(), 'view_wallet', 'View Wallet', 'wallet', 'view'),
    (gen_random_uuid(), 'list_wallet_transactions', 'List Wallet Transactions', 'wallet_transactions', 'list'),
    (gen_random_uuid(), 'list_wallets', 'List Wallets', 'wallets', 'list'),
    (gen_random_uuid(), 'topup_wallets', 'Topup Wallets', 'wallets', 'topup'),
    (gen_random_uuid(), 'adjust_wallets', 'Adjust Wallets', 'wallets', 'adjust'),
    (gen_random_uuid(), 'list_admin_deposits', 'List Admin Deposits', 'admin_deposits', 'list'),
    (gen_random_uuid(), 'view_admin_deposits', 'View Admin Deposit Detail', 'admin_deposits', 'view'),
    (gen_random_uuid(), 'approve_admin_deposits', 'Approve Admin Deposits', 'admin_deposits', 'approve'),
    (gen_random_uuid(), 'cancel_admin_deposits', 'Cancel Admin Deposits', 'admin_deposits', 'cancel'),
    (gen_random_uuid(), 'create_deposits', 'Create Deposits', 'deposits', 'create'),
    (gen_random_uuid(), 'list_deposits', 'List Deposits', 'deposits', 'list'),
    (gen_random_uuid(), 'view_deposits', 'View Deposit Detail', 'deposits', 'view'),
    (gen_random_uuid(), 'receive_payment_webhooks', 'Receive Payment Webhooks', 'payment_webhooks', 'receive')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name IN ('superadmin', 'admin')
AND p.resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs', 'wallet', 'wallet_transactions', 'wallets', 'admin_deposits', 'deposits', 'payment_webhooks')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'member'
AND p.name IN (
    'list_smm_services',
    'list_smm_orders',
    'view_smm_orders',
    'create_smm_orders',
    'refresh_status_smm_orders',
    'view_dashboard',
    'view_wallet',
    'list_wallet_transactions',
    'create_deposits',
    'list_deposits',
    'view_deposits'
)
ON CONFLICT DO NOTHING;
