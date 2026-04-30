# PostgreSQL connection refused or service unavailable

**Playbook ID:** `postgres-connection-refused`
**Category:** deploy
**Severity:** high
**Tags:** `postgres`, `database`, `connection`, `refused`, `service`, `psql`

## What this failure means

The application could not connect to PostgreSQL. The service is not responding, not yet started, or the credentials/hostname are wrong.

## Common log signals

```text
psql: error: could not connect
pg_isready.*rejecting
could not translate host name
```

## Diagnosis

PostgreSQL connection failures occur when:

- The PostgreSQL service is not running or not yet fully started (common in race conditions).
- The hostname is wrong (e.g., `localhost` instead of `postgres` service name in containers).
- The port is wrong or blocked by a firewall.
- Credentials (username, password) are incorrect or not set in environment variables.
- The database does not exist or the user lacks permissions.
- In Kubernetes or Docker Compose, the service has not had time to start.

The error explicitly shows `psql: error: could not connect` with details about the failed hostname and port.

## Fix steps

1. Verify PostgreSQL is running and listening:

   ```bash
   # Local check
   pg_isready -h localhost -p 5432

   # Docker container
   docker ps | grep postgres

   # Kubernetes
   kubectl get pods -l app=postgres
   ```

2. Check the connection string and environment variables:

   ```bash
   echo $DATABASE_URL
   env | grep POSTGRES
   env | grep DB
   ```

3. Verify credentials and hostname:

   ```bash
   psql -h <hostname> -U <username> -d <database> -c "SELECT 1"
   ```

4. If using Docker Compose or Kubernetes, ensure the service starts before the application:

   **Docker Compose example:**
   ```yaml
   services:
     postgres:
       image: postgres:15
       environment:
         POSTGRES_PASSWORD: secret
     app:
       depends_on:
         - postgres  # Ensures postgres starts first
   ```

5. Add a health check or retry logic if the service needs startup time:

   ```bash
   for i in {1..30}; do
     pg_isready -h postgres -p 5432 && break
     sleep 1
   done
   ```

## Validation

- `pg_isready -h <hostname> -p 5432` returns success.
- `psql -h <hostname> -U <username> -d <database> -c "SELECT 1"` executes successfully.
- Re-run the application or test suite.

## Likely files to inspect

- `.env`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `migrations/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain postgres-connection-refused
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- PostgreSQL connection refused or service unavailable
- Deploy: postgresql connection refused or service unavailable
- psql: error: could not connect
- faultline explain postgres-connection-refused


---

*Generated from [playbooks/bundled/log/deploy/postgres-connection-refused.yaml](../../../playbooks/bundled/log/deploy/postgres-connection-refused.yaml). Do not edit directly — run `make docs-generate`.*
