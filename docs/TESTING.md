# Backend PostgreSQL integration tests

Integration tests use a dedicated PostgreSQL container, database, port, and named volume. They do not use the development database (`db` on port `5432`, volume `postgres_data`). Keep the test password separate from development credentials. The test service binds only to `127.0.0.1:5433`.

From the repository root, set a URL-safe test password (hex characters are URL-safe):

```sh
export POSTGRES_TEST_PASSWORD="$(openssl rand -hex 24)"
```

Keep using the same `POSTGRES_TEST_PASSWORD` for an existing `postgres_test_data` volume. Changing the container environment does not change the password already initialized inside PostgreSQL.

The existing development service requires `POSTGRES_PASSWORD` during Compose interpolation. Set a throwaway value for Compose commands that explicitly select `test-db`; this value is not passed to or used by the test container:

```sh
POSTGRES_PASSWORD=unused-for-test-db docker compose --profile integration up -d test-db
```

Wait for the isolated test database to become healthy:

```sh
POSTGRES_PASSWORD=unused-for-test-db docker compose --profile integration ps test-db
```

Run the existing migrations against only the test service. The migration command reads the standard `POSTGRES_*` variables, so map the test settings for this command:

```sh
POSTGRES_HOST=127.0.0.1 \
POSTGRES_PORT=5433 \
POSTGRES_DB="${POSTGRES_TEST_DB:-expense_app_test}" \
POSTGRES_USER="${POSTGRES_TEST_USER:-expense_app_test}" \
POSTGRES_PASSWORD="$POSTGRES_TEST_PASSWORD" \
  go -C backend run ./cmd/migrate up
```

Run all backend tests with the PostgreSQL test DSN:

```sh
AUTH_TEST_DATABASE_URL="postgres://${POSTGRES_TEST_USER:-expense_app_test}:${POSTGRES_TEST_PASSWORD}@127.0.0.1:5433/${POSTGRES_TEST_DB:-expense_app_test}?sslmode=disable&timezone=UTC" \
  go -C backend test ./...
```

If you override `POSTGRES_TEST_USER` or `POSTGRES_TEST_DB`, use the same values in `AUTH_TEST_DATABASE_URL`. Keep the password URL-safe (for example, hexadecimal) or URL-encode it in the DSN.

The PostgreSQL integration tests require an already migrated, disposable database. Do not point `AUTH_TEST_DATABASE_URL` at the development database. To stop only the isolated test service without touching the development service or either database volume:

```sh
POSTGRES_PASSWORD=unused-for-test-db docker compose --profile integration stop test-db
```
