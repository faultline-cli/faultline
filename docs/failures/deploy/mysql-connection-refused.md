# MySQL connection refused or service unavailable

**Playbook ID:** `mysql-connection-refused`
**Category:** deploy
**Severity:** high
**Tags:** `mysql`, `database`, `connection`, `refused`, `service`, `HY000`

## What this failure means

The application or test suite could not connect to MySQL. The service is not
running, has not finished starting, or the connection string points to the
wrong host or port.

## Common log signals

```text
ERROR 2003 (HY000)
ERROR 2002 (HY000)
Can't connect to MySQL server on
Can't connect to local MySQL server through socket
SQLSTATE[HY000] [2002]
_mysql_exceptions.OperationalError: (2003,
CommunicationsException: Communications link failure
MySQLNonTransientConnectionException
```

## Diagnosis

MySQL connection failures in CI occur when:

- The MySQL service container has not finished initializing before the
  application or test runner attempts its first connection.
- The hostname is wrong — for example, `localhost` is used instead of the
  service name (`mysql`, `db`) in a Docker Compose or GitHub Actions
  services block.
- The port is misconfigured or blocked. MySQL listens on `3306` by default.
- Credentials, database name, or user permissions are incorrect.
- The service definition is missing from the CI workflow entirely.

The MySQL client emits error code `2003` (TCP connection refused) or `2002`
(socket file not found) along with the HY000 SQLSTATE, or a driver-level
equivalent such as `SQLSTATE[HY000] [2002] Connection refused` (PHP/PDO) or
`CommunicationsException: Communications link failure` (Java).

## Fix steps

1. Verify MySQL is accepting connections:

   ```bash
   mysqladmin ping -h 127.0.0.1 -P 3306 -u root -ppassword
   ```

2. Check the service hostname. In Docker Compose and GitHub Actions
   services blocks, connect using the **service name** — not `localhost`:

   **GitHub Actions example:**
   ```yaml
   services:
     mysql:
       image: mysql:8
       env:
         MYSQL_ROOT_PASSWORD: secret
         MYSQL_DATABASE: testdb
       options: >-
         --health-cmd="mysqladmin ping -h 127.0.0.1"
         --health-interval=10s
         --health-retries=5
   ```

   Then connect to `127.0.0.1` (not `localhost`) because the service is
   port-mapped, not on a shared Docker network.

3. For Docker Compose, add a healthcheck dependency:

   ```yaml
   services:
     mysql:
       image: mysql:8
       environment:
         MYSQL_ROOT_PASSWORD: secret
       healthcheck:
         test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
         interval: 10s
         retries: 5
     app:
       depends_on:
         mysql:
           condition: service_healthy
   ```

4. Add a startup wait in CI if healthchecks are unavailable:

   ```bash
   for i in $(seq 1 30); do
     mysqladmin ping -h 127.0.0.1 -u root -ppassword 2>/dev/null && break
     sleep 2
   done
   ```

5. Verify the environment variables for the connection string:

   ```bash
   echo "HOST=$DB_HOST PORT=$DB_PORT USER=$DB_USER"
   ```

## Validation

- `mysqladmin ping -h <host> -P 3306 -u root -p<password>` exits zero.
- `mysql -h <host> -u root -p<password> -e "SELECT 1"` succeeds.
- Re-run the failing test step.

## Likely files to inspect

- `.env`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `config/database.yml`
- `database.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain mysql-connection-refused
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- MySQL connection refused or service unavailable
- Deploy: mysql connection refused or service unavailable
- CommunicationsException: Communications link failure
- faultline explain mysql-connection-refused


---

*Generated from [playbooks/bundled/log/deploy/mysql-connection-refused.yaml](../../playbooks/bundled/log/deploy/mysql-connection-refused.yaml). Do not edit directly — run `make docs-generate`.*
