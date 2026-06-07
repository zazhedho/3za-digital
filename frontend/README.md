# 3ZA Digital Frontend

Vite React frontend for 3ZA Digital.

## Scripts

```bash
npm run dev
npm run build
npm run lint
```

## Environment

Set `VITE_API_URL` to backend API base URL, for example:

```env
VITE_API_URL=http://localhost:8080/api
```

Set `VITE_GOOGLE_CLIENT_ID` to enable Google Identity Services buttons on Login and Register:

```env
VITE_GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
```

The backend must also be configured with the matching `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_IDS` value because the backend validates the submitted Google `id_token`.

## Auth UI Behavior

- Login supports email/password and Google login.
- Register supports Google signup only when public registration is enabled by backend config.
- Email register follows backend config:
  - `auth.public_registration_enabled`
  - `auth.register_otp_enabled`
- Register password validation requires minimum 8 characters, lowercase, uppercase, number, and symbol before submit.
