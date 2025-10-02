# Currency Converter API

## Features

- Authentication
  - Register with email and password (hashed using Argon2)
    - Login returns a JWT access token
    - Logout and new logins invalidate previous sessions via token versioning
- Rates
  - Background job refreshes rates every 6 hours (configurable)
  - Uses exchangerate.host (free) as the source
  - Rate cache in memory + persisted in Postgres for resilience
- Security and Performance
  - JWT auth with token version check against DB
  - Rate limiting (per-IP)
  - Secure HTTP headers
  - Request logging with trace IDs
  - Panic recovery middleware
- Error Handling
  - Consistent JSON error responses with code, message, and trace_id
  - Input validation with clear error messages

## Tech

- Golang
- Gin (HTTP framework)
- GORM (ORM) with Postgres
- Zap (structured logging)
- JWT (github.com/golang-jwt/jwt/v5)
- Docker & docker-compose

## External API

This service uses [exchangerate-api](https://www.exchangerate-api.com) as the source for currency exchange rates. Rates are fetched via their private API and refreshed in the background at a configurable interval (default: every 6 hours). API key is required for exchangerate.host.

- API endpoint: `https://v6.exchangerate-api.com/v6/{api_key}/latest/{base_currency}`
- Base currency and symbols are configurable via query parameters.
- Example request:
    ```
    https://v6.exchangerate-api.com/v6/f1f7a18d707dad8dbd854c9d/latest/USD
    ```
- The fetched rates are cached in memory and persisted in Postgres for resilience and performance.

For more details, see the [exchangerate-api documentation](https://www.exchangerate-api.com/docs).


## Configuration

| Variable                | Description                              | Example Value           |
|-------------------------|------------------------------------------|------------------------|
| `APP_ENV`               | Application environment                  | `development`          |
| `PORT`                  | API server port                          | `8080`                 |
| `DB_HOST`               | Database host                            | `localhost`            |
| `DB_PORT`               | Database port                            | `5432`                 |
| `DB_USER`               | Database username                        | `postgres`             |
| `DB_PASSWORD`           | Database password                        | `S3cret`               |
| `DB_NAME`               | Database name                            | `currencydb`           |
| `DB_SSLMODE`            | Postgres SSL mode                        | `disable`              |
| `DB_TIMEZONE`           | Database timezone                        | `Asia/Bangkok`         |
| `JWT_SECRET`            | Secret key for JWT authentication        | `Wd15JdPhGkwaHx4RCWNxu0thiexfbI3O` |
| `JWT_EXPIRY`            | JWT token expiry duration                | `24h`                  |
| `RATE_BASE_CURRENCY`    | Base currency for exchange rates         | `USD`                  |
| `RATE_REFRESH_INTERVAL` | Interval for refreshing exchange rates   | `6h`                   |
| `HTTP_CLIENT_TIMEOUT`   | HTTP client timeout for API requests     | `10s`                  |
| `RATE_LIMIT_REQUESTS`   | Max requests per rate limit window       | `100`                  |
| `RATE_LIMIT_WINDOW`     | Rate limit window duration               | `1m`                   |
| `EXCHANGE_API_URL`      | URL for the exchange rate API            | `https://v6.exchangerate-api.com/v6/` |
| `EXCHANGE_API_KEY`      | API key for the exchange rate API        | `f1f7a18d707dad8dbd854c9d` |

## Quick Start (Docker)

1. Start services:
   ```bash
   docker-compose up -d
   ```
2. The Open API Document will be available at `http://localhost:8080/docs`.

Postgres will run at `localhost:5432` with a default `currencydb` database.

## Local Development (without Docker)

1. Create a `.env` file in the project root and fill in the configuration variables as shown in the table above.
2. Start the API server:
  ```bash
  go run ./cmd/app
  ```
3. The OpenAPI documentation will be available at `http://localhost:8080/docs`. You can also import the Postman collection (JSON) provided in the project for API testing.

## API

All endpoints return structured error responses on failure:
```json
{
  "code": "validation_error",
  "message": "validation error",
  "details": { ... },
  "trace_id": "..."
}
```

- Health Check
  - GET /healthcheck

- Auth
  - POST /api/v1/auth/register
    - Body: { "email": "user@example.com", "password": "password123" }
    - 201 Created on success
  - POST /api/v1/auth/login
    - Body: { "email": "user@example.com", "password": "password123" }
    - 200 OK: { "access_token": "...", "token_type": "bearer" }
  - POST /api/v1/auth/logout
    - Requires Authorization: Bearer <token>
    - 200 OK

- Rates (Auth required)
  - GET /api/v1/rates?base=USD
    - Returns all rates relative to requested base (derived if different from stored base)
    - 200 OK: { "base": "USD", "rates": { "THB": 36.7, ... }, "updated_at": "..." }
  - GET /api/v1/convert?from=USD&to=THB&amount=123.45
    - 200 OK: { "from": "USD", "to": "THB", "amount": 123.45, "rate": 36.7, "result": 4526.415, "updated_at": "..." }

### cURL Examples

- Register
  ```bash
  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"SuperSecret123"}'
  ```

- Login
  ```bash
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"SuperSecret123"}'
  ```

- Get Rates
  ```bash
  curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/v1/rates?base=USD"
  ```

- Convert
  ```bash
  curl -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8080/api/v1/convert?from=USD&to=THB&amount=10"
  ```

- Logout
  ```bash
  curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/auth/logout
  ```

## Security Notes

- JWT secret must be strong and kept secure (use a secret manager in production).
- Session invalidation via token versioning: new logins or logout increment user token version, revoking prior tokens.
- Rate limiting is IP-based and in-memory; for distributed deployments, use a shared store (e.g., Redis).
- Security headers are set for API safety. CORS is not enabled by default; add CORS middleware if needed.

## Error Handling Strategy

- Centralized error helpers in `pkg/response` ensure consistent JSON structure.
- Validation errors return HTTP 400 with details from Gin binding or custom checks.
- Unauthorized responses return HTTP 401 with clear message.
- Internal errors return HTTP 500 with a generic message and a trace_id for correlation.

## Logging

- Zap-based structured logging.
- Request logs include method, path, status, latency, IP, user-agent, and trace_id.
- A request ID is generated if not supplied via `X-Request-ID`.

## Performance

- In-memory cache for rates with background refresh reduces latency and upstream calls.
- Pooled HTTP client with timeouts.
- Gin in Release mode in production (set APP_ENV=production).