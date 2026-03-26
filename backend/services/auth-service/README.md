# auth-service

Auth identity with JWT access tokens, refresh-token rotation, and email OTP verification.

All endpoints are under:
`/api/v1/auth`

## Endpoints

1. `POST /api/v1/auth/register`
   Request:

```json
{
  "email": "user@example.com",
  "password": "password123",
  "baseCurrency": "USD",
  "displayName": "Optional Name"
}
```

Response:

```json
{
  "message": "OTP sent"
}
```

2. `POST /api/v1/auth/verify-otp` (activates the account)

Request:

```json
{
  "email": "user@example.com",
  "otp": "123456"
}
```

Response:

```json
{
  "message": "OTP verified"
}
```

3. `POST /api/v1/auth/resend-otp`

Request:

```json
{
  "email": "user@example.com"
}
```

Response:

```json
{
  "message": "OTP sent"
}
```

4. `POST /api/v1/auth/login`

Request:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "access_token": "…",
  "refresh_token": "…",
  "expires_in": 86400
}
```

Note: `POST /api/v1/auth/login` only works for users that have been verified via `verify-otp`.

5. `POST /api/v1/auth/refresh`

Request:

```json
{
  "refresh_token": "…"
}
```

Response:

```json
{
  "access_token": "…",
  "refresh_token": "…",
  "expires_in": 86400
}
```

6. `GET /api/v1/auth/me`

Headers:
`Authorization: Bearer <access_token>`

Response:

```json
{
  "user_id": "…",
  "email": "user@example.com",
  "base_currency": "USD"
}
```

## Environment variables

auth-service uses shared configuration from environment:

- `DATABASE_URL` (required): Postgres DSN for the `velune_auth` database.
- `JWT_SECRET` (required): signing key for JWT access tokens.
- `HTTP_PORT` (optional, default `8080`): listens on this port.
- `JWT_EXPIRY` (optional, default `24h`): access token TTL.
- `MIGRATIONS_PATH` (optional, default `file://migrations`): where SQL migrations live.
- `REFRESH_TOKEN_TTL` (optional, default `30d`): refresh token TTL.
- `OTP_VALIDITY_SECONDS` (optional, default `300`): OTP validity duration in seconds.
- `OTP_RESEND_COOLDOWN_SECONDS` (optional, default `30`): minimum delay between OTP resends.
- `OTP_MAX_RESENDS` (optional, default `3`): max number of resend attempts.
- `OTP_MAX_VERIFY_ATTEMPTS` (optional, default `3`): max number of OTP verification attempts.
- SMTP (optional; if unset or incomplete, auth-service uses a dev stub sender):
  - `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_TLS`

The refresh token stored in Postgres is a hashed, opaque value; only the plaintext refresh token is returned to the client.

## Local curl examples

Direct to the service (docker-compose split stack exposes `8081`):

```bash
BASE_URL="http://localhost:8081"

curl -sS -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","baseCurrency":"USD"}' | jq

curl -sS -X POST "$BASE_URL/api/v1/auth/verify-otp" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"123456"}' | jq

ACCESS="$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' | jq -r .access_token)"

curl -sS -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS" | jq
```

If you are using the gateway (`docker compose --profile split up`), the same endpoints are available under:
`http://localhost:8080/api/v1/auth/...`
