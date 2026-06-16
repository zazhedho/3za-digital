# Responsive Design

3ZA Digital is mobile-first. Most users are expected to use the product from phones, while admins still need dense desktop tools.

## Breakpoints

| Width | Behavior |
| --- | --- |
| `> 1100px` | Full sidebar, multi-column dashboard cards, wider tables |
| `860px - 1100px` | Reduced grids, stacked content panels |
| `< 860px` | Off-canvas sidebar, compact header, stacked filters |
| `< 560px` | Single-column actions, full-width form controls, compact modals |

## Layout Rules

- Sidebar is fixed on desktop and off-canvas on mobile.
- Mobile sidebar must not overlap content after close/open.
- Header title must stay readable and not collide with actions.
- Cards should not be nested inside other cards.
- Search/filter blocks should sit close to tables and cards, with consistent height.
- Form controls should use one visual height across modules.
- Avoid empty frames when search is hidden for modules that only support filters.

## Tables On Mobile

- Tables can scroll horizontally, but action menus must not be clipped.
- Row action hamburger menu should render above table overflow.
- Menu must close when user taps outside the dropdown.
- Amount, price, fee, and balance columns stay left-aligned.
- Detail tables need internal padding so labels and values do not touch the frame.

## Forms On Mobile

- Primary form actions become full-width when space is tight.
- Modal body should scroll internally if content is long.
- Confirmation modals must keep action buttons visible.
- Create Deposit should open as a modal and show QRIS immediately when backend returns QR data.
- QRIS image modal should be scan-friendly and compact around the image.

## Dropdowns On Mobile

- Use searchable dropdowns for large data.
- Dropdown content must fit viewport height.
- Load More button must remain visible and not be cut off.
- Selected text can wrap, but the control height should remain polished.

## QA Checklist

Check these viewports before release:

```text
390 x 844   common Android/iPhone portrait
430 x 932   larger phone portrait
768 x 1024  tablet portrait
1024 x 768  tablet landscape
1440 x 900  desktop
```

For each viewport:

- Login and Register.
- Sidebar open/close.
- Dashboard metrics and sidebar wallet balance.
- SMM Services list and service search.
- Create SMM Order service dropdown and quantity validation.
- Deposits list, create modal, QRIS modal, and detail.
- Admin Deposits action menu.
- Admin Wallets adjust modal.
- Users/Roles/Menus/Configurations list and detail pages.
