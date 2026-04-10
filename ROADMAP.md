# Faultline Roadmap

## Current Position

- V1 is shipped as a deterministic CI failure analysis CLI
- the repository now contains only the CLI-focused code and bundled playbooks
- future work should extend the CLI directly rather than reintroducing service-side components

## Next Up

### 1. Core CLI

- keep `analyze`, `list`, and `explain` stable as the public interface
- preserve deterministic file and stdin behavior
- keep startup and analysis latency low as playbook coverage grows

### 2. Playbook Coverage

- expand the bundled playbook set based on high-frequency CI failures
- prioritize common auth, dependency, test, and environment failures
- keep every playbook explicit and easy to audit

### 3. Output Quality

- keep text output short and fix-oriented
- support stable JSON for agent usage
- keep confidence and evidence grounded in exact matches

### 4. Docker Delivery

- keep the image small and fast-starting
- keep mounted-log analysis the default CI workflow
- document GitHub Actions and generic CI usage

### 5. Validation

- add fixture-driven tests for all bundled playbooks
- verify stable ranking and deterministic output
- add container smoke tests

## Not On The Roadmap

- hosted webhook ingestion
- provider-specific PR comment flows
- frontend UI
- probabilistic or AI-based diagnosis
