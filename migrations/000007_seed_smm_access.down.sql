DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs')
);

DELETE FROM permissions WHERE resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs');
DELETE FROM menu_items WHERE name IN ('smm_services', 'smm_orders');
