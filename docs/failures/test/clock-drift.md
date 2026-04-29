# CI clock drift or time synchronization failure

**Playbook ID:** `clock-drift`
**Category:** test
**Severity:** low
**Tags:** `clock`, `time`, `ntp`, `skew`, `drift`, `container`, `ci`, `timezone`, `tls`

## What this failure means

CI container or runner has a clock sufficiently out of sync to cause
TLS handshake failures, token expiry validation errors, or time-sensitive
test assertions to fail non-deterministically.

## Common log signals

```text
clock skew
time skew
NTP
clock drift
time drift
certificate expired
certificate is not yet valid
the system time is off
```

## Diagnosis

Container runtimes inherit the host kernel clock. If the host NTP daemon
falls behind (common on ephemeral spot/preemptible VMs after sleep resume or
heavy load), the container clock drifts. Effects:

- TLS certificate validation fails with "not yet valid" or "has expired"
- `iat`/`exp` claims in JWTs are rejected
- `Date` headers or timestamps in tests fail exact comparisons
- Signed URLs or pre-signed S3 requests reject with 403 + "Request time too skewed"

Check the runner clock:

```bash
date -u
# Compare against: curl -sI https://worldtimeapi.org/api/ip | grep -i date
```

For Docker-in-Docker or nested containers, also check the inner clock:

```bash
docker run --rm alpine date -u
```

## Fix steps

1. **Sync the host clock** (if you manage the runner):

   ```bash
   # Ubuntu/Debian
   sudo timedatectl set-ntp true
   sudo systemctl restart systemd-timesyncd
   timedatectl status

   # Or force immediate sync
   sudo chronyc makestep          # chrony
   sudo ntpdate -u pool.ntp.org   # ntp
   ```

2. **Restart the worker / runner**. Spot-instance runners often resume from
   a saved state with a stale clock; a restart forces NTP re-sync.

3. **Add a clock-sync step** to the CI job:

   ```yaml
   # GitHub Actions — run before tests
   - name: Sync clock
     run: |
       sudo timedatectl set-ntp true
       sleep 2
   ```

4. **Reduce time sensitivity in tests**: avoid `time.Now()` in assertions;
   use relative comparisons with a 5–10 s tolerance.

5. **Set `AWS_RETRY_MODE=adaptive`** or pass `--max-clock-skew` to avoid
   AWS SDK S3 signature failures on minor drift.

## Validation

- Run `date -u` on the runner before and after the sync step.
- Re-run the failing test/job.
- Check TLS or token validation errors are resolved.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain clock-drift
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI clock drift or time synchronization failure
- Test: ci clock drift or time synchronization failure
- certificate is not yet valid
- GitHub Actions ci clock drift or time synchronization failure
- faultline explain clock-drift


---

*Generated from [playbooks/bundled/log/test/clock-drift.yaml](../../playbooks/bundled/log/test/clock-drift.yaml). Do not edit directly — run `make docs-generate`.*
