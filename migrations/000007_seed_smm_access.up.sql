INSERT INTO menu_items (id, name, display_name, path, icon, order_index) VALUES
    (gen_random_uuid(), 'smm_services', 'SMM Services', '/smm/services', 'bi-megaphone', 100),
    (gen_random_uuid(), 'smm_orders', 'SMM Orders', '/smm/orders', 'bi-receipt', 101)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (id, name, display_name, resource, action) VALUES
    (gen_random_uuid(), 'list_smm_services', 'List SMM Services', 'smm_services', 'list'),
    (gen_random_uuid(), 'sync_smm_services', 'Sync SMM Services', 'smm_services', 'sync'),
    (gen_random_uuid(), 'list_smm_orders', 'List SMM Orders', 'smm_orders', 'list'),
    (gen_random_uuid(), 'view_smm_orders', 'View SMM Order Detail', 'smm_orders', 'view'),
    (gen_random_uuid(), 'create_smm_orders', 'Create SMM Orders', 'smm_orders', 'create'),
    (gen_random_uuid(), 'refresh_status_smm_orders', 'Refresh SMM Order Status', 'smm_orders', 'refresh_status'),
    (gen_random_uuid(), 'view_provider_balance', 'View Provider Balance', 'provider_balance', 'view'),
    (gen_random_uuid(), 'list_provider_api_logs', 'List Provider API Logs', 'provider_api_logs', 'list')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name IN ('superadmin', 'admin')
AND p.resource IN ('smm_services', 'smm_orders', 'provider_balance', 'provider_api_logs')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'member'
AND p.name IN ('list_smm_services', 'list_smm_orders', 'view_smm_orders', 'create_smm_orders', 'refresh_status_smm_orders', 'view_dashboard')
ON CONFLICT DO NOTHING;
