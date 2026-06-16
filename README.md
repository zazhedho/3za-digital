# 3ZA Digital

> Permission-first social commerce panel for SMM services, wallet balance, deposits, provider integrations, and operational audit trails.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react&logoColor=111111)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Permission First](https://img.shields.io/badge/Auth-permission--first-222222?style=for-the-badge)

## Overview

3ZA Digital is a full-stack dashboard for selling SMM services through provider integration while keeping user balance, deposit, pricing, menu visibility, and admin operations under a permission-first backend.

The system is built to stay extendable for future product surfaces such as pulsa, PPOB, game, and e-wallet.

| Area | Stack |
| --- | --- |
| Backend | Go, Gin, GORM |
| Frontend | React, Vite |
| Database | PostgreSQL |
| Cache/session | Redis, optional |
| Auth | JWT, refresh token, optional Redis sessions |
| Provider | H2H.id for SMM catalog/order flow |
| QRIS | Static QRIS, QRISLY, BOQRIS |
| CI/CD | GitHub Actions, Docker, GHCR |

## Contents

- [Features](#features)
- [Architecture](#architecture)
- [Permission Model](#permission-model)
- [Quick Start](#quick-start)
- [Environment](#environment)
- [Payment And Deposit](#payment-and-deposit)
- [SMM Order Flow](#smm-order-flow)
- [Main Routes](#main-routes)
- [Frontend](#frontend)
- [Testing And Quality](#testing-and-quality)
- [Adding Modules](#adding-modules)
- [Documentation](#documentation)

## Features

- Email/password login with email or phone identifier.
- Google login and auto-register controlled by backend config.
- Register OTP controlled by runtime config.
- Permission-first RBAC with role as a grouping label.
- Sidebar menu visibility derived from permissions.
- Runtime application config from `app_configs`.
- SMM service sync, service search, service pagination, and order creation.
- Wallet ledger with idempotent debits/refunds/credits.
- Manual deposit, Static QRIS, Dynamic QRIS through QRISLY or BOQRIS.
- Admin deposit approve/cancel flow with reason.
- Admin wallet adjustment modal flow.
- Audit trails for important write operations.
- Provider balance snapshots and provider API logs.
- Redis-backed sessions, permission cache, and route rate limits when Redis is enabled.

## Architecture

```text
3za-digital/
├── cmd/                         # helper commands
├── frontend/                    # Vite React app
├── infrastructure/              # database, cache, external infra setup
├── internal/
│   ├── domain/                  # domain entities and constants
│   ├── dto/                     # request/response DTOs
│   ├── handlers/http/           # HTTP handlers
│   ├── integrations/            # external provider clients
│   ├── interfaces/              # service/repository contracts
│   ├── repositories/            # database access
│   ├── router/                  # route wiring
│   └── services/                # business logic
├── middlewares/
├── migrations/
├── pkg/
├── postman/
├── utils/
└── main.go
```

Backend flow:

```text
route -> middleware -> handler -> service -> repository -> database
```

Repository convention:

- Use `internal/repositories/generic` for common CRUD/list behavior.
- Keep module repositories focused on custom joins, aggregates, and transactions.
- Define search/filter/sort behavior through generic query options.

Integration convention:

- Provider-specific clients live in `internal/integrations/<provider>`.
- Shared HTTP JSON client helpers live in `internal/integrations/httpjson`.
- Business services depend on interfaces, not concrete provider response structs.

## Permission Model

RBAC rules:

- `permission` is the runtime source of truth.
- `role` is a label and permission grouping mechanism.
- `superadmin` is the only bypass exception.
- Menus are derived from permissions, not manual menu assignment.

Practical behavior:

- Endpoint access uses `PermissionMiddleware(resource, action)`.
- `/api/menus/me` is built from current user permissions.
- A module menu can appear when user has at least one permission for that resource.
- Parent menus are included automatically when a child menu is permitted.

Menu order convention:

```text
100-199  Product and member operations
900-999  System administration
```

Current intended order:

```text
100 SMM Services
101 SMM Orders
102 Wallet
103 Deposits
104 Admin Wallets
105 Admin Deposits
900 Users
901 Roles
902 Menus
903 Configurations
```

## Quick Start

1. Copy env:

```bash
cp .env.example .env
```

2. Run migration and backend:

```bash
go run . -migrate
go run .
```

3. Run frontend:

```bash
cd frontend
npm install
npm run dev
```

4. Health check:

```text
GET /healthcheck
```

## Environment

Minimum backend env:

| Key | Notes |
| --- | --- |
| `APP_NAME` | Application name |
| `APP_ENV` | `development`, `staging`, `production` |
| `PORT` | Backend port |
| `DATABASE_URL` | Preferred database URL |
| `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASS`, `DB_NAME`, `DB_SSLMODE` | Used when `DATABASE_URL` is empty |
| `JWT_KEY` | Minimum 32 chars in production |
| `JWT_EXP` | Access token expiration |
| `PATH_MIGRATE` | Migration source path |

Optional but recommended:

| Area | Keys |
| --- | --- |
| Redis | `REDIS_URL`, `REDIS_PORT`, `REDIS_DB`, `REDIS_PASSWORD` |
| Permission cache | `PERMISSION_CACHE_TTL`, `PERMISSION_CACHE_TTL_SECONDS` |
| Google login | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_IDS` |
| SMTP | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` |
| Storage | storage provider keys when upload is enabled |
| Location sync | `LOCATION_SERVICE_BASE_URL`, `LOCATION_SERVICE_TIMEOUT_SECONDS` |
| Rate limits | `SMM_ORDER_RATE_LIMIT`, `DEPOSIT_CREATE_RATE_LIMIT`, `PAYMENT_WEBHOOK_RATE_LIMIT` |

Provider env:

| Provider | Keys |
| --- | --- |
| H2H | `H2H_BASE_URL`, `H2H_MEMBER_ID`, `H2H_PIN`, `H2H_PASSWORD`, `H2H_TIMEOUT_SECONDS` |
| QRISLY | `QRISLY_BASE_URL`, `QRISLY_API_KEY`, `QRISLY_QRIS_ID`, `QRISLY_OUTPUT_TYPE`, `QRISLY_TIMEOUT_SECONDS` |
| BOQRIS | `BOQRIS_BASE_URL`, `BOQRIS_API_KEY`, `BOQRIS_MERCHANT_ID`, `BOQRIS_TIMEOUT_SECONDS` |

Security rule:

> Provider credentials must stay in backend environment variables only. Do not put H2H, QRISLY, BOQRIS, PIN, password, API key, bearer token, or webhook secrets in frontend code, Postman shared variables, request logs, or provider API logs.

## Runtime Config

Runtime config is stored in `app_configs` and can be changed from the admin Configurations page.

Auth flags:

| Key | Meaning |
| --- | --- |
| `auth.public_registration_enabled` | Enable public register |
| `auth.register_otp_enabled` | Require OTP for email register |
| `auth.password_reset_email_enabled` | Send reset token by email |

Pricing and SMM:

| Key | Meaning |
| --- | --- |
| `pricing.default_markup_percent` | Default provider price markup |
| `pricing.product_markup_percent.smm` | SMM-specific markup |
| `pricing.smm_service_price_max_age` | Max cached SMM price age before lazy sync |

Payment:

| Key | Meaning |
| --- | --- |
| `payment.qris.fee_percent` | Topup fee percent |
| `payment.qris.image_url` | Static QRIS image URL |
| `payment.qris.merchant_name` | Static QRIS merchant name |
| `payment.qris.dynamic_provider` | `qrisly` or `boqris` |

## Payment And Deposit

Deposit methods:

| Method | Provider | Behavior |
| --- | --- | --- |
| Manual review | `manual` | User creates request, admin approves/cancels |
| Static QRIS | `qris` | Uses configured image URL, unique code, fee |
| Dynamic QRIS | `qrisly` or `boqris` | Provider generates QR and payable amount |

Dynamic QRIS selection:

1. Frontend sends provider `qrisly` for “Dynamic QRIS”.
2. Backend resolves active provider from `payment.qris.dynamic_provider`.
3. Backend calls QRISLY or BOQRIS based on config.
4. UI still displays user-facing label `Dynamic QRIS`.

BOQRIS flow:

- Create transaction: `POST /api/v1/transactions`
- Status check: `GET /api/v1/transactions/{transaction_id}`
- `paid` status completes the deposit and credits wallet.
- `expired`, `failed`, or `cancelled` update the deposit status without crediting wallet.

Wallet rules:

- Every balance mutation must create a `wallet_transactions` row.
- Wallet balance cannot go negative.
- Money calculations use integer cents internally.
- Member can access only their own wallet, deposits, and orders.
- Admin and superadmin access is permission-based.
- Admin topup and credit adjustment are blocked when active wallet liability would exceed live H2H main balance.

## SMM Order Flow

1. Admin syncs SMM services from H2H.
2. Backend upserts services into `provider_services`.
3. User creates order through backend.
4. Backend validates service, target URL, and quantity range.
5. Backend calculates price from provider price per 1,000 quantity plus markup.
6. Backend debits wallet and creates internal order in one transaction.
7. Backend calls H2H provider.
8. If provider create fails, backend creates an idempotent wallet refund.
9. If final provider status is failed, backend creates an idempotent refund.

Price formula:

```text
provider_charge = ceil_to_whole(provider_price_per_1k * quantity / 1000)
markup_amount   = provider_charge * markup_percent / 100
amount          = ceil_to_whole(provider_charge + markup_amount)
profit          = amount - provider_charge
```

Catalog freshness:

- Redis stores last successful product sync timestamp when available.
- Database `provider_services.synced_at` is fallback.
- Lazy sync runs when cached price age exceeds `pricing.smm_service_price_max_age`.
- Redis lock prevents concurrent sync for the same product type.

## Main Routes

Auth and users:

```text
POST /api/user/register
POST /api/user/register/otp/send
POST /api/user/login
POST /api/user/google/login
POST /api/user/refresh-token
POST /api/user/forgot-password
POST /api/user/reset-password
POST /api/user/logout
GET  /api/user
GET  /api/users
```

RBAC and configuration:

```text
GET    /api/roles
POST   /api/role
GET    /api/role/:id
PUT    /api/role/:id
DELETE /api/role/:id
POST   /api/role/:id/permissions
GET    /api/permissions
GET    /api/permissions/me
GET    /api/menus/me
GET    /api/menus
GET    /api/menu/:id
PUT    /api/menu/:id
GET    /api/configs
GET    /api/config/:id
PUT    /api/config/:id
```

SMM:

```text
GET  /api/smm/services
POST /api/smm/services/sync
GET  /api/smm/orders
POST /api/smm/orders
GET  /api/smm/orders/:id
GET  /api/smm/orders/:id/status-logs
POST /api/smm/orders/:id/refresh-status
```

Wallet and deposits:

```text
GET  /api/wallet/me
GET  /api/wallet/transactions
GET  /api/admin/wallets
POST /api/admin/wallets/:user_id/topup
POST /api/admin/wallets/:user_id/adjust
GET  /api/deposits
POST /api/deposits
GET  /api/deposits/:id
GET  /api/deposits/settings
GET  /api/admin/deposits
GET  /api/admin/deposits/:id
POST /api/admin/deposits/:id/status
```

Provider, dashboard, audit, sessions:

```text
GET /api/provider/h2h/balance
GET /api/provider/api-logs
GET /api/dashboard/summary
GET /api/audits
GET /api/audit/:id
GET /api/user/sessions
DELETE /api/user/session/:session_id
POST /api/user/sessions/revoke-others
```

Location:

```text
GET  /api/location/province
GET  /api/location/city?province_code=11
GET  /api/location/district?city_code=1101
GET  /api/location/village?district_code=110101
POST /api/location/sync
GET  /api/location/sync/:id
```

## Frontend

Frontend lives in `frontend/`.

```bash
cd frontend
npm install
npm run dev
npm run lint
npm run build
```

Important behavior:

- Login supports local auth and Google Identity Services.
- Register follows backend config.
- Sidebar balance uses cached `/wallet/me` to avoid API spam.
- `/menus/me` is cached for the session and cleared on logout.
- SMM service dropdown uses remote search plus paginated “Load more”.
- Mobile layout is first-class because most users are expected to use mobile.

See [frontend/README.md](frontend/README.md).

## Testing And Quality

Backend:

```bash
go test ./...
golangci-lint run --config .golangci.yml --timeout=5m ./...
go build ./...
```

Frontend:

```bash
cd frontend
npm run lint
npm run build
```

Postman collection:

```text
postman/3za-digital.postman_collection.json
```

CI:

- `.github/workflows/lint.yml`
- `.github/workflows/docker.yml`

GitHub Actions use Node 24-compatible action major versions:

- `actions/checkout@v6`
- `docker/login-action@v4`
- `docker/build-push-action@v7`

## Adding Modules

Keep new modules aligned with permission-first design.

Backend layers:

```text
internal/domain/<module>
internal/dto
internal/interfaces/<module>
internal/repositories/<module>
internal/services/<module>
internal/handlers/http/<module>
internal/router/router.go
```

Required migration pieces:

- business table(s)
- `menu_items` row
- `permissions` rows
- optional `role_permissions` seed

Resource naming must match:

```text
menu_items.name      = projects
permissions.resource = projects
```

Protect routes with permissions:

```go
mdw.PermissionMiddleware("projects", "list")
mdw.PermissionMiddleware("projects", "view")
mdw.PermissionMiddleware("projects", "create")
mdw.PermissionMiddleware("projects", "update")
mdw.PermissionMiddleware("projects", "delete")
```

Seed helper:

```bash
go run ./cmd/module-seed \
  --name projects \
  --display-name "Projects" \
  --path /projects \
  --icon bi-folder \
  --order-index 106
```

## Documentation

| File | Purpose |
| --- | --- |
| [DESIGN.md](DESIGN.md) | Visual design reference |
| [SMM_PLAN.md](SMM_PLAN.md) | SMM, wallet, deposit, and QRIS architecture |
| [frontend/README.md](frontend/README.md) | Frontend setup and behavior |
| [frontend/FRONTEND_STRUCTURE.md](frontend/FRONTEND_STRUCTURE.md) | Frontend module structure |
| [frontend/RESPONSIVE_DESIGN.md](frontend/RESPONSIVE_DESIGN.md) | Responsive UI notes |
| [postman/3za-digital.postman_collection.json](postman/3za-digital.postman_collection.json) | API collection |

## Operational Notes

- Do not use role name checks for normal module access.
- Do not expose provider details like H2H to end users.
- Do not store provider credentials in database config rows.
- Keep payment credit, refund, and provider callback handling idempotent.
- Use confirmation modals for destructive or financially important actions.
- Use audit trails for important admin actions and financial state changes.
