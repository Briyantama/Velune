# auth-service

Auth identity, password hashing, JWT access tokens, and opaque refresh-token rotation.

All endpoints are under:
`/api/v1/auth`

## Endpoints

1. `POST /api/v1/auth/register`
Request:

```json
{
  "email": "user@example.com",
  "password": "password123",
  "baseCurrency": "USD"
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

2. `POST /api/v1/auth/login`

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

3. `POST /api/v1/auth/refresh`

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

4. `GET /api/v1/auth/me`

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

The refresh token stored in Postgres is a hashed, opaque value; only the plaintext refresh token is returned to the client.

## Local curl examples

Direct to the service (docker-compose split stack exposes `8081`):

```bash
BASE_URL="http://localhost:8081"

ACCESS="$(curl -sS -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","baseCurrency":"USD"}' | jq -r .access_token)"

curl -sS -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS" | jq
```

If you are using the gateway (`docker compose --profile split up`), the same endpoints are available under:
`http://localhost:8080/api/v1/auth/...`
