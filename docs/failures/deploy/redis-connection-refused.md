# Redis connection refused or service unavailable

**Playbook ID:** `redis-connection-refused`
**Category:** deploy
**Severity:** high
**Tags:** `redis`, `cache`, `queue`, `connection`, `refused`, `service`, `6379`

## What this failure means

The application or test suite could not connect to Redis. The service is not
running, has not finished starting, or the connection string points to the
wrong host or port.

## Common log signals

```text
Redis::CannotConnectError
redis.exceptions.ConnectionError
JedisConnectionException: Could not connect to Redis
Could not connect to Redis at
Error 111 connecting to
ECONNREFUSED 127.0.0.1:6379
ECONNREFUSED ::1:6379
[ioredis] Unhandled error event
```

## Diagnosis

Redis connection failures in CI occur when:

- The Redis service container has not finished starting before the application
  attempts its first connection.
- The hostname is wrong — for example, `localhost` instead of the service
  name (`redis`, `cache`) in a Docker Compose or GitHub Actions services
  block.
- The default Redis port `6379` is misconfigured or blocked.
- The Redis server requires a password (`AUTH`) and the client is not
  providing one, or providing an incorrect one.

Client libraries surface this differently:
- **Python (redis-py):** `redis.exceptions.ConnectionError: Error 111
  connecting to localhost:6379. Connection refused.`
- **Ruby:** `Redis::CannotConnectError: Error connecting to Redis on
  localhost:6379 (Errno::ECONNREFUSED)`
- **Node.js (ioredis / node-redis):**
  `Error: connect ECONNREFUSED 127.0.0.1:6379`
- **Java (Jedis):** `JedisConnectionException: Could not connect to Redis at
  127.0.0.1:6379: Connection refused.`

## Fix steps

1. Verify Redis is running and accepting connections:

   ```bash
   redis-cli -h 127.0.0.1 -p 6379 ping
   # Expected response: PONG
   ```

2. Add a healthcheck to the Redis service in your CI configuration:

   **GitHub Actions example:**
   ```yaml
   services:
     redis:
       image: redis:7
       options: >-
         --health-cmd "redis-cli ping"
         --health-interval=10s
         --health-retries=5
   ```

3. For Docker Compose, depend on the healthcheck:

   ```yaml
   services:
     redis:
       image: redis:7
       healthcheck:
         test: ["CMD", "redis-cli", "ping"]
         interval: 5s
         retries: 5
     app:
       depends_on:
         redis:
           condition: service_healthy
   ```

4. If Redis requires a password, set `REDIS_PASSWORD` and configure the
   client accordingly. A missing `AUTH` produces `NOAUTH Authentication
   required`, not a connection refused — if that is the error, check
   credentials instead.

5. Check the hostname used by the client. In GitHub Actions, the Redis
   service is accessible at `127.0.0.1`, not `redis`, because it is
   port-mapped to the host:

   ```bash
   echo "REDIS_URL=$REDIS_URL"
   redis-cli -h 127.0.0.1 ping
   ```

## Validation

- `redis-cli -h <host> -p 6379 ping` returns `PONG`.
- Re-run the failing test step and confirm no connection error appears.

## Likely files to inspect

- `.env`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `config/redis.yml`
- `config/cable.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain redis-connection-refused
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Redis connection refused or service unavailable
- Deploy: redis connection refused or service unavailable
- JedisConnectionException: Could not connect to Redis
- faultline explain redis-connection-refused


---

*Generated from [playbooks/bundled/log/deploy/redis-connection-refused.yaml](../../playbooks/bundled/log/deploy/redis-connection-refused.yaml). Do not edit directly — run `make docs-generate`.*
