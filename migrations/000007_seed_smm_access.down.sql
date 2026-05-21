DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs', 'wallet', 'wallet_transactions', 'wallets', 'admin_deposits', 'deposits', 'payment_webhooks')
);

DELETE FROM permissions WHERE resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs', 'wallet', 'wallet_transactions', 'wallets', 'admin_deposits', 'deposits', 'payment_webhooks');
DELETE FROM menu_items WHERE name IN ('smm_services', 'smm_orders', 'wallet', 'deposits', 'admin_wallets');
