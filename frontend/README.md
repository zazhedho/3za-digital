# 3ZA Digital Frontend

> Mobile-first React dashboard for 3ZA Digital operations: SMM services, SMM orders, deposits, wallets, permissions, audit trails, sessions, and system configuration.

![React](https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react&logoColor=111111)
![Vite](https://img.shields.io/badge/Vite-7-646CFF?style=for-the-badge&logo=vite&logoColor=white)
![Bootstrap](https://img.shields.io/badge/Bootstrap-5-7952B3?style=for-the-badge&logo=bootstrap&logoColor=white)
![Permission First](https://img.shields.io/badge/UI-permission--first-222222?style=for-the-badge)

## Overview

This frontend is intentionally operational, not marketing-heavy. It is designed for frequent mobile usage while still supporting dense admin workflows on desktop.

Key behavior:

- Menu visibility comes from backend permissions.
- Role is treated as a permission group label, not as the main authorization source.
- Superadmin keeps full bypass behavior.
- Sidebar wallet balance is cached and refreshed safely after wallet/deposit mutations.
- Search and filters are explicit-submit to reduce unnecessary API calls.
- Important actions use confirmation modals.
- Error messages are centered and readable for long backend messages.
- Success messages stay compact and non-blocking.

## Scripts

```bash
npm install
npm run dev
npm run build
npm run lint
```

## Environment

Create `frontend/.env`:

```env
VITE_API_URL=http://localhost:8080/api
VITE_GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
```

Notes:

- `VITE_API_URL` must point to backend API base URL.
- `VITE_GOOGLE_CLIENT_ID` enables Google buttons on Login and Register.
- Backend must also contain matching `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_IDS`, because backend validates submitted Google `id_token`.

## App Behavior

Authentication:

- Login supports email/password, phone/password, and Google login.
- Register supports email/password and Google signup when backend config allows public registration.
- Email register follows backend config:
  - `auth.public_registration_enabled`
  - `auth.register_otp_enabled`
- OTP form appears only after Create Account succeeds and backend requires OTP.
- Password validation requires minimum 8 characters, lowercase, uppercase, number, and symbol before submit.
- Authenticated users visiting `/login` or `/register` are redirected away from auth pages.

Data access:

- `/menus/me` is fetched after login/session bootstrap and reused by layout.
- `/wallet/me` is cached for sidebar balance and refreshed after relevant wallet/deposit actions.
- Pages must rely on backend permissions and backend responses, not role-name shortcuts.

## Structure

```text
frontend/
├── public/
├── src/
│   ├── components/common/       # shared controls, modals, table actions, selects
│   ├── contexts/                # auth, menu, wallet, toast, theme context
│   ├── hooks/                   # reusable UI/data hooks
│   ├── pages/                   # route pages grouped by module
│   ├── services/                # API clients
│   ├── styles/                  # theme and responsive CSS
│   └── utils/                   # formatters and shared helpers
└── vite.config.js
```

Page convention:

```text
ModuleList.jsx
ModuleForm.jsx
ModuleDetail.jsx
```

Use the convention for modules that need create/edit/detail flows, for example Users, Roles, Menus, Configurations, Deposits, Wallets, and SMM Orders.

## UI Guidelines

- Use English for all visible UI text.
- Keep backend/provider implementation names hidden from end users when not needed.
- Do not hardcode status, fee percentage, provider choice, or permission behavior if backend/config already provides it.
- Keep forms compact and clearly separate search/filter controls from cards and tables.
- On mobile, stack filter controls but keep the table close to the filter area.
- Dropdowns with large datasets must be searchable and support remote loading.
- Table action menus must overlay above table overflow and close when clicking outside.
- Amount columns are left-aligned like other table values.
- Detail pages must include a Back button and avoid putting record IDs in page headers.

## Shared Components

Common components should stay reusable:

| Component | Use |
| --- | --- |
| `ConfirmationModal` | Approve, cancel, delete, adjust, and other important actions |
| `SearchableSelect` | Large selectable datasets with search and remote load-more support |
| `TableActionMenu` | Compact row actions inside list tables |
| `BackButton` | Detail and form navigation |
| `PaginationBar` | Paginated list pages |
| `StatusBadge` | Consistent status color and spacing |

## Related Docs

- [`../README.md`](../README.md)
- [`../DESIGN.md`](../DESIGN.md)
- [`FRONTEND_STRUCTURE.md`](FRONTEND_STRUCTURE.md)
- [`RESPONSIVE_DESIGN.md`](RESPONSIVE_DESIGN.md)
