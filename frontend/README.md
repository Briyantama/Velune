# Velune web frontend

Next.js (App Router) + TypeScript + Tailwind + shadcn-style primitives, with TanStack Query for server state and httpOnly cookie auth (via Next route handlers).

## Setup

1. Install deps:

```bash
cd frontend
npm install
```

2. Configure environment:

Create `frontend/.env.local`:

```bash
NEXT_PUBLIC_GATEWAY_BASE_URL=http://127.0.0.1:8080
ADMIN_SERVICE_URL=http://127.0.0.1:8099
ADMIN_API_KEY=change-me-admin-api-key
```

3. Run dev server:

```bash
npm run dev
```

## Auth model

- Browser never sees tokens directly.
- Next route handlers store `velune_access_token` and `velune_refresh_token` in httpOnly cookies.
- Client pages call `/api/gateway/*` routes (same origin) which attach `Authorization: Bearer ...` to the backend gateway.

