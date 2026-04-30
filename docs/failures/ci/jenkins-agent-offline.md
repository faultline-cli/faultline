# Jenkins build agent went offline during job

**Playbook ID:** `jenkins-agent-offline`
**Category:** ci
**Severity:** critical
**Tags:** `jenkins`, `agent`, `offline`, `ci`, `connection`

## What this failure means

The Jenkins build agent disconnected or went offline while running a job. The build was lost mid-execution and Jenkins cannot continue until the agent reconnects or the job is rescheduled on another agent.

## Common log signals

```text
Agent.*is offline
Connection was broken
Remoting connection closed
Channel is already closed
java.io.IOException: remoting
No such slave/agent
Node .* was taken offline
SSH Connection timeout
```

## Diagnosis

Jenkins distributes work to agents (nodes). When an agent loses its connection to the controller during a build:

- The build log shows `Agent [name] is offline` or `Connection was broken`.
- The build status shows as aborted or with a non-zero exit.
- `SSH Connection timeout`, `Remoting connection closed`, or `java.io.IOException: remoting` appear in the log.
- Docker-based agents may exit because the container or sidecar was killed by the container runtime.
- Network interruptions, JVM crashes, OOM kills on the agent, or resource limit changes on the agent host cause disconnects.

## Fix steps

1. Check Jenkins Controller → Manage Jenkins → Nodes → [agent name] for the last disconnect reason.

2. Review the agent log for the disconnect reason:

   ```bash
   journalctl -u jenkins-agent -n 100
   # or check the remoting.log on the agent host
   ```

3. If the agent was OOM-killed, increase the JVM heap on the agent or reduce concurrent job parallelism.

4. For SSH agents, verify the SSH key is still valid and the controller can reach the agent host:

   ```bash
   ssh -i ~/.ssh/jenkins_key user@agent-host "java -version"
   ```

5. For Docker-based agents (Kubernetes or Docker plugin), confirm the pod/container has adequate resources and a stable network path to the controller.

6. Retry the build after the agent reconnects. If it fails repeatedly, investigate the agent host's resource and network health.

## Validation

- The agent shows as online in Jenkins → Manage Jenkins → Nodes.
- A re-triggered build completes without an agent disconnect error.

## Likely files to inspect

- `Jenkinsfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain jenkins-agent-offline
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Jenkins build agent went offline during job
- Ci: jenkins build agent went offline during job
- java.io.IOException: remoting
- GitHub Actions jenkins build agent went offline during job
- faultline explain jenkins-agent-offline


---

*Generated from [playbooks/bundled/log/ci/jenkins-agent-offline.yaml](../../../playbooks/bundled/log/ci/jenkins-agent-offline.yaml). Do not edit directly — run `make docs-generate`.*
