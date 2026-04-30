# Port already in use

**Playbook ID:** `port-in-use`
**Category:** runtime
**Severity:** medium
**Tags:** `port`, `network`, `bind`, `socket`

## What this failure means

A process attempted to bind to a TCP or UDP port that is already occupied by another process on the same CI runner.

## Common log signals

```text
address already in use
bind: address already in use
listen tcp
EADDRINUSE
lsof
```

## Diagnosis

A process attempted to bind to a TCP or UDP port that is already occupied by another process on the same CI runner.

## Fix steps

1. Identify what is holding the port:

   ```bash
   lsof -i :<port>
   ss -tlnp | grep :<port>
   ```

2. Kill the occupying process:

   ```bash
   fuser -k <port>/tcp
   # or:
   lsof -ti :<port> | xargs kill -9
   ```

3. On shared or persistent CI runners, previous jobs may leave lingering
   processes. Add a cleanup step at the start of the job to free known ports
   before starting services.
4. Parameterise the port with an environment variable so parallel CI jobs
   can bind to different ports without conflict.
5. Prefer ephemeral port 0 where supported: let the OS assign a free port
   and retrieve the assigned value from the listening socket handle.

## Validation

- `lsof -i :<port>` — confirm the port is free before starting the service.
- Re-run the failing step and confirm it binds without `EADDRINUSE`.

## Likely files to inspect

- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.env.example`


## Run Faultline

```bash
faultline analyze build.log
faultline explain port-in-use
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Port already in use
- Runtime: port already in use
- bind: address already in use
- faultline explain port-in-use


---

*Generated from [playbooks/bundled/log/runtime/port-in-use.yaml](../../../playbooks/bundled/log/runtime/port-in-use.yaml). Do not edit directly — run `make docs-generate`.*
