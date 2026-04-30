# Docker daemon configuration conflict prevents startup

**Playbook ID:** `docker-daemon-config-conflict`
**Category:** runtime
**Severity:** high
**Tags:** `docker`, `daemon`, `config`, `daemon.json`, `startup`

## What this failure means

The Docker daemon failed to start because conflicting options were supplied — the same setting was configured both in `daemon.json` and on the daemon command line. The daemon exits instead of starting.

## Common log signals

```text
unable to configure the Docker daemon with file
conflicting options: hosts
merged json
daemon.json
failed to start daemon
could not change group /var/run/docker.sock
You cannot specify a host flag in daemon.json together with the -H
Flag --hosts cannot be used together with 'hosts' key in daemon configuration file
```

## Diagnosis

Docker does not allow the same option to be specified in both `daemon.json` and via CLI flags. When it detects a conflict during startup, the daemon aborts with a configuration error. This is distinct from the daemon being unreachable after a successful start.

Common conflicts:

- `hosts` set in `daemon.json` and via `-H` or `--host` flags at the same time.
- `data-root` or `log-driver` specified both in the config file and as command-line arguments.
- Remote-access settings (`tcp://`) added to `daemon.json` on a system that already passes `--host` via systemd unit override.
- Stale or malformed `daemon.json` left from a prior Docker install.

## Fix steps

1. Check the daemon error message to identify which option is conflicting:

   ```bash
   journalctl -u docker --no-pager -n 30
   # or on systems without journald:
   cat /var/log/docker.log | tail -30
   ```

2. Open `/etc/docker/daemon.json` and compare its contents against the active dockerd command or systemd unit:

   ```bash
   cat /etc/docker/daemon.json
   systemctl cat docker | grep ExecStart
   ```

3. Remove the conflicting key from `daemon.json` if the same option is already on the command line, or move all configuration into `daemon.json` and remove it from the command line.

4. Validate the JSON syntax before restarting:

   ```bash
   python3 -c "import json, sys; json.load(open('/etc/docker/daemon.json'))"
   ```

5. Restart the daemon:

   ```bash
   sudo systemctl restart docker
   docker info
   ```

## Validation

- `sudo systemctl status docker` shows `active (running)`.
- `docker info` succeeds without error.

## Likely files to inspect

- `/etc/docker/daemon.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-daemon-config-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker daemon configuration conflict prevents startup
- Runtime: docker daemon configuration conflict prevents startup
- You cannot specify a host flag in daemon.json together with the -H
- faultline explain docker-daemon-config-conflict
- Docker docker daemon configuration conflict prevents startup


---

*Generated from [playbooks/bundled/log/runtime/docker-daemon-config-conflict.yaml](../../../playbooks/bundled/log/runtime/docker-daemon-config-conflict.yaml). Do not edit directly — run `make docs-generate`.*
