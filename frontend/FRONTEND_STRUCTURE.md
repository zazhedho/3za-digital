# Frontend Structure

3ZA Digital frontend follows the same high-level structure as `safety-riding/frontend`, with module pages split into list, form, and detail files when the module has those flows.

## Directory Map

```text
src/
├── components/
│   └── common/                  # reusable UI primitives
├── contexts/                    # app-wide state
├── hooks/                       # shared React hooks
├── pages/
│   ├── auth/
│   ├── dashboard/
│   ├── deposits/
│   ├── menus/
│   ├── roles/
│   ├── sessions/
│   ├── smm/
│   ├── users/
│   └── wallets/
├── services/                    # Axios API clients
├── styles/                      # design system and responsive rules
└── utils/                       # formatters and shared helpers
```

## Page Pattern

Use this pattern for modules with list/form/detail workflow:

```text
UserList.jsx
UserForm.jsx
UserDetail.jsx

DepositList.jsx
DepositForm.jsx
DepositDetail.jsx
```

Rules:

- `List` owns table, pagination, search/filter state, and row actions.
- `Form` owns create/edit modal or page form behavior.
- `Detail` owns read-only record view and contextual actions.
- Shared handlers and reusable UI logic must move to `components`, `hooks`, or `utils`, not stay trapped inside one page file.

## State And Data

Auth:

- `AuthContext` stores session user, token state, and auth actions.
- Auth pages redirect when user is already logged in.
- Google login/register uses frontend client ID only; backend validates token.

Menu:

- Menu data comes from `/menus/me`.
- Frontend must not invent menu visibility from role names.
- Superadmin bypass remains backend-owned.

Wallet:

- Sidebar balance comes from cached `/wallet/me`.
- Mutations that affect wallet balance must trigger wallet refresh.
- Admin wallet pages use admin endpoints and permission checks.

## Lists

Expected list behavior:

- Pagination where backend supports or data volume can grow.
- Explicit Search/Apply button for text search.
- Reset button to clear filters and return to default fetch.
- Search field should be omitted when backend cannot search that dataset reliably.
- Filters should be visually grouped and close to the table.
- Mobile layout should stack controls without leaving empty white frames.

## Dropdowns

Use `SearchableSelect` for large datasets:

- Searchable by visible name and important backend identifiers.
- Remote mode for datasets such as SMM services.
- Load More support when backend returns paginated results.
- Mobile dropdown must keep Load More visible and not clipped.
- Selected option label should be readable, with IDs and prices separated from names.

## Tables

Use consistent table rules:

- Compact spacing.
- Left-aligned values, including money/amount fields.
- Status badges with enough padding and readable colors.
- Row action menu through reusable table action component.
- Action dropdown must overlay above table overflow and close on outside click.
- Detail-like tables must not have cells stuck to the frame edge.

## Actions

Important actions need confirmation:

- Approve
- Cancel/reject
- Delete
- Adjust balance
- Sync provider catalog
- Any action that changes money, permission, user access, or final status

Refresh-status action on SMM Orders does not need confirmation because it is a sync/read-side update and should be low friction.
