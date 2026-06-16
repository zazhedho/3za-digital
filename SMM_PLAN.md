# SMM, Wallet, And Deposit Architecture

> Current implementation notes for 3ZA Digital SMM service catalog, order flow, wallet ledger, deposit, and payment provider integration.

## Product Scope

Initial product type:

```text
smm
```

The data model and service layer stay generic enough for future products:

```text
pulsa
ppob
game
ewallet
```

Frontend must never call external providers directly. Provider credentials stay in backend environment variables.

## Balance Model

3ZA Digital uses two balance concepts:

| Balance | Source | Visible To | Purpose |
| --- | --- | --- | --- |
| Provider balance | External provider | Admin only | Operational capital used to fulfill provider orders |
| User balance | 3ZA database | User/admin by permission | User funds used to create orders |

Rules:

- User orders debit user wallet first.
- Provider order is created after internal validation and wallet debit.
- Failed provider order refunds user wallet automatically.
- Admin approve for deposits must still check provider main balance when required by business logic.
- Wallet changes must always create `wallet_transactions` ledger rows.
- Wallet mutation must be idempotent.

## Main Tables

| Table | Purpose |
| --- | --- |
| `provider_services` | Cached provider service catalog |
| `orders` | Internal order records for all product types |
| `order_status_logs` | Order status history |
| `provider_balance_snapshots` | Provider balance snapshots |
| `provider_api_logs` | Provider request/response diagnostics without secrets |
| `wallets` | User wallet balances |
| `wallet_transactions` | Wallet ledger |
| `deposit_requests` | Manual, static QRIS, and dynamic QRIS deposit requests |
| `payment_gateway_logs` | Payment provider request/status/webhook logs |

## Provider Services

`provider_services` stores provider catalog data:

- `provider`
- `product_type`
- `provider_service_id`
- `name`
- `category`
- `brand`
- `platform`
- `min_quantity`
- `max_quantity`
- `price`
- `metadata`
- `raw_response`
- `is_active`
- `synced_at`

Frontend behavior:

- SMM Services must use pagination.
- Service ID and service name must be shown separately.
- Create SMM Order service dropdown must support search by `provider_service_id`.
- Service dropdown should show readable name, provider service ID, platform, min/max quantity, and price per 1k.
- Platform filter should reduce service search noise.

## Orders

`orders` stores all internal orders:

- `provider`
- `product_type`
- `ref_id`
- `order_number`
- `service_id`
- `provider_service_id`
- `target`
- `quantity`
- `status`
- `amount`
- `provider_charge`
- `profit`
- `metadata`
- `provider_response`
- `created_by`
- timestamps

SMM Create Order rules:

- `target` must be a URL.
- `quantity` defaults to selected service minimum.
- `quantity` cannot be lower than service minimum.
- `quantity` cannot be higher than service maximum.
- Amount is calculated from selling price, not raw provider price.
- Order creation must protect against race conditions around wallet debit and provider submission.
- If provider submission fails after debit, refund must happen once.

Status:

```text
pending
processing
completed
partial
failed
cancelled
```

Final statuses:

```text
completed
partial
failed
cancelled
```

## Pricing

Pricing uses provider cost plus runtime markup config.

Important fields:

| Field | Meaning |
| --- | --- |
| `provider_services.price` | Raw provider cost |
| `orders.provider_charge` | Provider charge for order |
| `orders.amount` | User selling amount |
| `orders.profit` | Selling amount minus provider charge |

UI rules:

- Show user-facing selling price.
- Do not expose provider brand names such as H2H to normal users.
- Use "Provider" only where admin diagnostics need it.
- Service price labels should say price per 1k when the value is per 1,000 quantity.

## Deposits

Deposit statuses:

```text
pending
paid
expired
failed
cancelled
```

Deposit methods:

```text
manual_admin
qris_static
qris_dynamic
```

Rules:

- Minimum deposit amount is `10000`.
- Deposit fee comes from runtime config.
- UI should show calculated Topup Fee amount, not hardcoded percentage text.
- Manual review deposits must use the same fee calculation as QRIS deposits.
- Paid deposit credits wallet exactly once.
- Expired, failed, or cancelled deposit must not credit wallet.
- Admin cancel/reject requires reason and displays reason in list/detail.

## QRIS

Supported QRIS modes:

| Mode | Config | Behavior |
| --- | --- | --- |
| Static QRIS | `payment.qris.image_url` filled | Uses uploaded QRIS image URL |
| Dynamic QRIS | `payment.qris.image_url` empty and dynamic provider enabled | Creates provider transaction and returns QR image/data |

Dynamic providers:

| Provider | Create | Status |
| --- | --- | --- |
| QRISLY | Provider invoice/QRIS endpoint | Provider status endpoint |
| BOQRIS | `POST /api/v1/transactions` | `GET /api/v1/transactions/:transaction_id` |

Status sync:

- Deposit status should follow dynamic provider status when provider has final state.
- `paid` status triggers wallet credit.
- `expired` status should update local deposit as expired.
- Provider logs must not include API keys or bearer tokens.

## Backend Flow

```text
route -> middleware -> handler -> service -> repository -> database
```

Layer rules:

- Interfaces live in `internal/interfaces`.
- Business services depend on interfaces.
- Provider clients live in `internal/integrations/<provider>`.
- Shared provider HTTP/JSON helper code belongs in reusable packages, not duplicated provider `flexible.go` files.
- Provider-specific translation stays near provider integration.

## Frontend Flow

Important frontend expectations:

- Permission-first route/menu rendering.
- Form/list/detail file split for modules with full CRUD/detail flow.
- Explicit search submit and reset buttons.
- Pagination for large datasets.
- Searchable remote dropdowns for services.
- Confirmation modals for important write actions.
- Mobile-first tables, filters, dropdowns, and modals.

## API Surface

Main endpoints:

```text
GET  /api/smm/balance
GET  /api/smm/services
POST /api/smm/services/sync
GET  /api/smm/orders
POST /api/smm/orders
GET  /api/smm/orders/:id
POST /api/smm/orders/:id/refresh-status

GET  /api/wallet/me
GET  /api/wallet/transactions
GET  /api/admin/wallets
POST /api/admin/wallets/:user_id/adjust

POST /api/deposits
GET  /api/deposits
GET  /api/deposits/:id
POST /api/deposits/:id/status

POST /api/webhooks/payments/:provider
```

## Permissions

Representative permissions:

```text
smm_balance:view
smm_services:list
smm_services:sync
smm_orders:list
smm_orders:view
smm_orders:create
smm_orders:refresh_status
wallet:view
wallet_transactions:list
wallets:list
wallets:adjust
deposits:create
deposits:list
deposits:view
deposits:update_status
payment_webhooks:receive
```

Permission assignment is managed through role permission setup. Role names are labels; permissions control access.
