# 3ZA Digital Product Order Plan

## Scope Awal

Project dimulai sebagai dashboard Social Media Marketing (SMM) only, tetapi pondasi backend harus siap untuk product lain:

- SMM
- pulsa
- PPOB
- game
- e-wallet

Integrasi external API memakai H2H.id: https://h2h.id/docs/api

Frontend tidak boleh memanggil H2H langsung karena credential H2H harus aman di backend.

Prinsip utama: **API publik boleh SMM-specific, tetapi domain inti, tabel inti, wallet, order, pricing, dan provider client harus generic multi-product.**

## Model Bisnis Balance

3ZA Digital memakai dua level balance:

```text
Provider balance = saldo utama di H2H
User balance     = saldo reseller/enduser di sistem 3ZA Digital
```

Provider balance:

- Source of truth ada di H2H.
- Dipakai backend untuk membayar order ke provider.
- Hanya internal/admin yang boleh melihat.
- Di sistem kita disimpan sebagai snapshot histori, bukan saldo operasional user.

User balance:

- Source of truth ada di database 3ZA Digital.
- Dipakai reseller/enduser untuk membuat order.
- Harus cukup sebelum order dibuat.
- Bisa bertambah dari deposit/topup/adjustment.
- Berkurang saat order dibuat.
- Refund saat order gagal.

Implikasi:

- User tidak langsung memakai saldo H2H.
- Semua order user harus melewati wallet 3ZA Digital.
- H2H balance adalah saldo master untuk memenuhi order ke provider.
- Pricing ke user bisa lebih tinggi dari provider charge agar ada margin.

Keputusan bisnis awal:

- Wallet dibuat otomatis saat user dibuat.
- Topup fase awal dilakukan manual: admin/superadmin topup saldo utama di H2H.id sebagai modal, reseller/user membuat request deposit, lalu admin approve/topup wallet user.
- Struktur deposit tidak boleh hardcoded hanya admin topup, karena fase berikutnya user harus bisa topup sendiri lewat payment gateway.
- Payment gateway optional untuk fase berikutnya; tabel dan flow tetap disiapkan agar callback gateway bisa credit wallet tanpa ubah model utama.
- Admin topup dan credit adjustment tidak boleh membuat total liability wallet aktif (`balance + locked_balance`) lebih besar dari live H2H main balance.
- Jika provider gagal membuat order, refund wallet otomatis penuh.
- Jika order final `partial`, sistem harus mendukung refund sisa berdasarkan `remains`, tetapi implementasi bisa masuk fase setelah order dasar stabil.
- Harga awal dihitung dari service price estimate.
- Final charge/profit bisa disesuaikan dari provider charge jika H2H mengembalikan charge final.
- Nominal IDR disimpan sebagai `NUMERIC(18,2)`, tetapi tampilan bisa dibulatkan rupiah tanpa desimal.

Deposit/topup direction:

- Manual topup admin dan payment gateway self-topup future masuk lewat konsep `deposit_request`.
- Fase awal: user membuat `deposit_request` pending method `manual_admin`, admin approve/topup setelah validasi pembayaran manual.
- Wallet hanya bertambah setelah deposit berstatus final `paid` atau admin action yang valid.
- Callback payment gateway harus idempotent agar retry webhook tidak double credit.
- Setiap credit wallet dari deposit wajib punya ledger `wallet_transactions`.

## Kondisi Backend Saat Ini

- Stack: Go, Gin, GORM, PostgreSQL, Redis opsional.
- Modul dasar sudah tersedia: auth, user, role, permission, menu, app config, audit, location.
- `go.mod` sudah memakai module `3za-digital`.
- Import path sudah memakai `3za-digital/...`.
- Dokumentasi dasar sudah memakai nama 3ZA Digital.

## Product Type

Product type awal:

```text
smm
```

Product type berikutnya:

```text
pulsa
ppob
game
ewallet
```

Semua product type disimpan sebagai data, bukan dipisah menjadi tabel besar berbeda.

## Environment Baru

Tambahkan ke `.env.example`:

```text
APP_NAME=3ZA Digital
DB_NAME=3za_digital

H2H_BASE_URL=https://api.h2h.id/api/trx
H2H_MEMBER_ID=
H2H_PIN=
H2H_PASSWORD=
H2H_TIMEOUT_SECONDS=20
```

Credential H2H wajib disimpan di `.env`, bukan di frontend dan bukan di log.

## Endpoint H2H Yang Dipakai Di Fase SMM

Base URL:

```text
https://api.h2h.id/api/trx
```

Endpoint awal:

```text
GET /balance
GET /pricelist?type=smm
GET /pricelist?type=smm&platform=instagram
GET /?type=smm&service={service}&target={target}&quantity={quantity}&refID={refID}
GET /status?refID={refID}
```

Endpoint product lain tetap lewat H2H client yang sama, dengan `type` berbeda sesuai docs H2H.

## Tahap 1: Project Cleanup

Tujuan: backend bisa build stabil sebagai project 3ZA Digital.

Pekerjaan:

- Pastikan import path memakai `3za-digital/...`.
- Update `.env.example` untuk brand dan database default.
- Update README minimal agar memakai 3ZA Digital.
- Pastikan `go test ./...` berjalan.

## Tahap 2: Provider Client

Tujuan: semua akses provider terpusat, dan H2H menjadi provider pertama.

Lokasi:

```text
internal/integrations/h2h
```

Interface generic:

```go
type ProviderClient interface {
    GetBalance(ctx context.Context) (*BalanceResponse, error)
    GetPriceList(ctx context.Context, req PriceListRequest) (*PriceListResponse, error)
    CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error)
    GetOrderStatus(ctx context.Context, refID string) (*OrderStatusResponse, error)
}
```

Request generic:

```text
PriceListRequest:
- Type
- Platform
- Category

CreateOrderRequest:
- Type
- ServiceCode
- Target
- Quantity
- RefID
- Metadata
```

Provider client tidak punya helper product-specific. Caller mengisi `Type`, misalnya `Type=smm`, lalu memanggil method generic.

Aturan:

- Semua request memakai timeout.
- Semua response error H2H dinormalisasi.
- Query credential ditambahkan hanya di client.
- Log tidak boleh menyimpan `pin`, `password`, atau credential penuh.
- Unit test memakai mock HTTP server.

## Tahap 3: Database Generic

Tambahkan migration baru.

Tabel:

```text
provider_services
orders
order_status_logs
provider_balance_snapshots
provider_api_logs
wallets
wallet_transactions
deposit_requests
payment_gateway_logs
```

`provider_services` menyimpan cache pricelist dari provider:

- provider
- product_type
- provider_service_id
- name
- category
- brand
- platform
- min_quantity
- max_quantity
- price
- metadata
- raw_response
- is_active
- synced_at

`orders` menyimpan order internal semua product:

- provider
- product_type
- ref_id unique
- service_id
- provider_service_id
- target
- quantity
- customer_no
- customer_name
- status
- amount
- provider_charge
- profit
- start_count
- remains
- metadata
- provider_response
- created_by
- created_at
- updated_at

Catatan field:

- `target` dipakai SMM untuk link/username.
- `customer_no` dipakai pulsa/PPOB/game/e-wallet untuk nomor pelanggan/player/account.
- `metadata` JSONB dipakai untuk atribut khusus product, agar tabel tidak berubah tiap product baru.

`order_status_logs` menyimpan histori status:

- order_id
- old_status
- new_status
- provider_status
- provider_response
- created_at

`provider_balance_snapshots` menyimpan histori balance:

- provider
- balance
- raw_response
- created_at

`provider_api_logs` opsional untuk debugging:

- provider
- endpoint
- request_ref
- product_type
- response_status
- response_body
- duration_ms
- error_message
- created_at

Credential tidak boleh masuk ke `provider_api_logs`.

`wallets` menyimpan saldo user 3ZA Digital:

- user_id
- balance
- locked_balance
- currency
- is_active
- created_at
- updated_at

`wallet_transactions` menyimpan ledger mutasi saldo:

- wallet_id
- user_id
- order_id
- type
- direction
- amount
- balance_before
- balance_after
- reference
- description
- metadata
- created_by
- created_at

Tipe transaksi awal:

```text
deposit
debit_order
refund_order
adjustment
```

Direction:

```text
credit
debit
```

Aturan wallet:

- Semua perubahan saldo harus lewat `wallet_transactions`.
- Saldo tidak boleh diubah tanpa ledger.
- Debit order harus atomic dengan pembuatan order internal.
- Refund order harus idempotent agar tidak double refund.
- Admin adjustment wajib punya audit trail.
- Balance user tidak boleh negatif kecuali ada kebijakan kredit khusus di masa depan.

`deposit_requests` menyimpan permintaan topup user/admin dan disiapkan untuk payment gateway:

- user_id
- amount
- status
- method
- provider
- payment_reference
- payment_url
- expired_at
- paid_at
- metadata
- created_by
- created_at
- updated_at

Status deposit:

```text
pending
paid
expired
failed
cancelled
```

Method deposit:

```text
manual_admin
payment_gateway
```

Aturan deposit:

- `manual_admin` dibuat oleh user sebagai request deposit atau oleh admin/superadmin sebagai direct topup; status menjadi `paid` setelah validasi admin.
- `payment_gateway` optional untuk fase berikutnya; backend nantinya membuat invoice/payment link ke gateway.
- Deposit `paid` harus credit wallet tepat satu kali.
- Deposit `expired`, `failed`, atau `cancelled` tidak boleh mengubah wallet.
- `payment_reference` harus unik per provider jika tersedia.

`payment_gateway_logs` menyimpan log callback/invoice dari gateway:

- provider
- event_type
- request_id
- payment_reference
- deposit_request_id
- signature
- status
- payload
- error_message
- created_at

Credential, API key, token, dan secret payment gateway tidak boleh masuk ke log.

## Tahap 4: Module Generic

Struktur mengikuti pola existing:

```text
route -> handler -> service -> repository -> database
```

File/module:

```text
internal/domain/catalog
internal/domain/order
internal/domain/provider
internal/dto/catalog.go
internal/dto/order.go
internal/repositories/catalog
internal/repositories/order
internal/services/catalog
internal/services/order
internal/handlers/http/smm
internal/interfaces/catalog
internal/interfaces/order
```

Catalog service responsibility:

- Sync pricelist dari provider berdasarkan product type.
- Upsert ke `provider_services`.
- Filter service by type, platform, category, provider, active status.

Order service responsibility:

- Validasi service aktif.
- Validasi quantity masuk min/max.
- Generate `refID` unik.
- Hitung `amount` user berdasarkan pricing/markup.
- Cek wallet user aktif.
- Cek saldo user cukup.
- Debit wallet user secara atomic dengan order internal.
- Simpan order internal dengan status awal `pending`.
- Hit H2H create order.
- Update status awal dari response provider.
- Simpan status log.
- Refund wallet otomatis jika provider gagal membuat order.
- Refresh status order via H2H.

Handler SMM tetap boleh spesifik untuk pengalaman API yang jelas:

- `internal/handlers/http/smm` memakai `catalog.Service` dan `order.Service`.
- Handler mengisi `product_type = smm`.

## Tahap 5: API Internal Fase SMM

Endpoint backend:

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
POST /api/admin/wallets/:user_id/topup
POST /api/admin/wallets/:user_id/adjust
POST /api/deposits
GET  /api/deposits
GET  /api/deposits/:id
POST /api/webhooks/payments/:provider
```

Semua endpoint authenticated.

Permission:

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
wallets:topup
wallets:adjust
deposits:create
deposits:list
deposits:view
payment_webhooks:receive
```

Walaupun endpoint bernama SMM, data masuk ke tabel generic:

```text
provider_services.product_type = smm
orders.product_type = smm
```

## Tahap 6: API Internal Masa Depan

Jika product lain ditambahkan, pola API bisa memakai salah satu dari dua opsi.

Opsi A, product-specific:

```text
GET  /api/pulsa/services
POST /api/pulsa/orders
GET  /api/ppob/services
POST /api/ppob/orders
GET  /api/game/services
POST /api/game/orders
GET  /api/ewallet/services
POST /api/ewallet/orders
```

Opsi B, generic:

```text
GET  /api/products/:type/services
POST /api/products/:type/orders
GET  /api/products/:type/orders
```

Rekomendasi:

- Mulai dengan endpoint SMM-specific untuk UX awal.
- Siapkan service layer generic agar product lain tinggal menambah handler/route tipis.

## Tahap 7: RBAC Dan Menu

Tambah menu:

```text
SMM Dashboard -> /smm/dashboard
SMM Services  -> /smm/services
SMM Orders    -> /smm/orders
```

Tambah permission sesuai endpoint.

Admin dan superadmin mendapat semua permission SMM.
Operator mendapat akses list, view, create, refresh status.
Member hanya list/view.

Nanti product lain bisa punya resource:

```text
pulsa_services
pulsa_orders
ppob_services
ppob_orders
game_services
game_orders
ewallet_services
ewallet_orders
```

Atau jika ingin lebih generic:

```text
provider_services
orders
```

Rekomendasi RBAC awal: tetap resource per product agar permission mudah dipahami admin.

## Tahap 8: Status Flow

Status internal:

```text
pending
processing
completed
partial
failed
cancelled
```

Mapping awal mengikuti status H2H:

```text
pending -> pending
processing -> processing
completed -> completed
partial -> partial
failed -> failed
```

Final status:

```text
completed
partial
failed
cancelled
```

Final status tidak boleh berubah kecuali ada alasan bisnis eksplisit.

Status ini harus reusable untuk semua product.

## Tahap 9: Pricing Dan Margin

Pondasi pricing harus siap dari awal walaupun fase pertama belum kompleks.

Field minimal:

```text
provider_services.price
orders.provider_charge
orders.amount
orders.profit
```

Fase awal:

- `provider_charge = biaya H2H`.
- `amount = harga yang dibayar user dari wallet`.
- `profit = amount - provider_charge`.
- Jika markup belum dikonfigurasi, boleh fallback `amount = provider_charge`.

Pricing config disimpan di `app_configs` agar bisa diubah tanpa deploy.

Config awal:

```text
pricing.default_markup_percent
pricing.product_markup_percent.smm
pricing.product_markup_percent.pulsa
pricing.product_markup_percent.ppob
pricing.product_markup_percent.game
pricing.product_markup_percent.ewallet
```

Nilai config berupa persentase:

```text
5
7.5
10
```

Formula:

```text
markup_amount = provider_charge * markup_percent / 100
amount        = provider_charge + markup_amount
profit        = amount - provider_charge
```

Contoh:

```text
provider_charge = 10000
markup_percent  = 5
amount          = 10500
profit          = 500
```

Prioritas config:

```text
specific product/category/platform/client markup
product type markup
default markup
0 percent
```

Fase awal cukup implement:

```text
pricing.product_markup_percent.smm
pricing.default_markup_percent
```

Fase berikut:

- markup per product type
- markup per category/platform
- markup per client
- promo/discount

## Tahap 10: Wallet Dan Saldo User

Wallet harus ada sebelum frontend transaksi dibuka untuk reseller/enduser.

Flow create order:

1. User memilih service.
2. Backend menghitung harga user (`amount`).
3. Backend cek wallet user.
4. Jika saldo kurang, order ditolak.
5. Jika saldo cukup, backend membuat order internal dan debit wallet dalam database transaction.
6. Backend call H2H memakai provider balance utama.
7. Jika H2H berhasil, order update mengikuti status provider.
8. Jika H2H gagal sebelum order provider tercipta, wallet user direfund.
9. Jika status akhir `failed` dari provider, wallet user direfund sesuai kebijakan.
10. Semua mutasi saldo masuk `wallet_transactions`.

API wallet fase awal:

```text
GET  /api/wallet/me
GET  /api/wallet/transactions
GET  /api/admin/wallets
POST /api/admin/wallets/:user_id/topup
POST /api/admin/wallets/:user_id/adjust
GET  /api/admin/deposits
GET  /api/admin/deposits/:id
POST /api/admin/deposits/:id/approve
```

Catatan frontend admin: `GET /api/admin/deposits` dan `GET /api/admin/deposits/:id` mengembalikan `user` summary aman (`id`, `name`, `email`, `phone`, `role`, `avatar_url`) agar tabel/detail admin tidak perlu lookup user terpisah.

API deposit yang disiapkan untuk payment gateway:

```text
POST /api/deposits
GET  /api/deposits
GET  /api/deposits/:id
POST /api/webhooks/payments/:provider
```

Flow admin manual topup fase awal:

1. Admin topup saldo utama di H2H.id sebagai modal.
2. Reseller/user membuat `deposit_request` pending method `manual_admin`.
3. Admin validasi pembayaran manual dari user.
4. Backend cek live H2H main balance.
5. Backend hitung `available = H2H main balance - total active wallet liability`.
   Liability = total active wallet `balance + locked_balance`.
6. Jika amount request lebih besar dari available, topup ditolak.
7. Jika ditolak karena main balance kurang, deposit tetap `pending`; admin topup H2H main balance dulu lalu approve ulang.
8. Jika valid, sistem set deposit `paid`.
9. Sistem credit wallet dalam database transaction.
10. Sistem tulis `wallet_transactions` type `deposit`.

Flow payment gateway future:

1. User membuat `deposit_request` method `payment_gateway`.
2. Backend membuat invoice/payment link ke gateway.
3. Backend menyimpan `payment_reference`, `payment_url`, `expired_at`, dan metadata gateway.
4. User membayar lewat gateway.
5. Gateway mengirim callback ke `/api/webhooks/payments/:provider`.
6. Backend verifikasi signature dan status callback.
7. Backend update deposit menjadi `paid`.
8. Backend credit wallet dalam database transaction.
9. Backend tulis `wallet_transactions` type `deposit`.
10. Backend tulis `payment_gateway_logs`.

Aturan payment gateway:

- Webhook harus idempotent berdasarkan `payment_reference` dan status deposit.
- Webhook payment gateway default disabled sampai provider gateway benar-benar aktif.
- Saat webhook diaktifkan, `PAYMENT_WEBHOOK_ENABLED=true` dan `PAYMENT_WEBHOOK_SECRET_<PROVIDER>` wajib ada.
- Wallet tidak boleh credit jika signature callback tidak valid.
- Wallet tidak boleh credit jika amount callback berbeda dari amount deposit.
- Deposit yang sudah `paid` tidak boleh diproses ulang.
- Gateway provider harus lewat interface agar bisa menambah Midtrans/Xendit/Duitku/provider lain tanpa ubah wallet core.

Permission:

```text
wallet:view
wallet_transactions:list
wallets:list
wallets:topup
wallets:adjust
deposits:create
deposits:list
deposits:view
payment_webhooks:receive
```

Catatan:

- `member` melihat wallet miliknya sendiri.
- `admin` dan `superadmin` bisa melihat wallet user.
- Topup dan adjustment hanya admin/superadmin.
- `member` boleh membuat deposit self-topup saat payment gateway aktif.
- Payment webhook tidak memakai session user, tetapi wajib validasi signature provider dan secret config.
- H2H balance tidak boleh dianggap wallet user.

## Tahap 11: Fase 2

Setelah flow order stabil:

- Payment gateway adapter untuk self-topup.
- Webhook/callback dari H2H jika tersedia.
- Polling otomatis untuk order non-final.
- Dashboard analytics.
- Client management.
- Margin/markup pricing.
- Invoice dan laporan.
- Frontend React dashboard.

## Urutan Eksekusi Rekomendasi

1. Cleanup import path dan brand.
2. Tambah env H2H.
3. Buat provider client interface dan H2H client plus test.
4. Buat migration generic: `provider_services`, `orders`, `order_status_logs`, `provider_balance_snapshots`, `provider_api_logs`.
5. Buat sync pricelist untuk `product_type=smm`.
6. Buat list services.
7. Buat wallet tables dan wallet service.
8. Buat `deposit_requests` dan `payment_gateway_logs` agar topup siap berkembang ke gateway.
9. Buat API wallet user dan admin topup/adjustment lewat deposit flow.
10. Siapkan interface payment gateway tanpa wajib integrasi provider gateway dulu.
11. Integrasikan debit wallet ke create order.
12. Integrasikan refund wallet untuk provider gagal/status gagal.
13. Buat create order.
14. Buat refresh order status.
15. Tambah RBAC dan menu.
16. Jalankan test penuh.
