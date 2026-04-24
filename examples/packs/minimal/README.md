# Minimal Example Pack

This directory is a minimal extra playbook pack for local experimentation and authoring guidance.

It is not part of the bundled catalog and is not auto-loaded by default.

## Contents

This pack includes several example playbooks demonstrating different Faultline features:

- **example-cache-prime-missing.yaml** - Basic playbook structure with match patterns
- **example-setup-error.yaml** - Demonstrates `match.use` for reusable matcher composition
- **example-hypothesis-disambiguation.yaml** - Shows how to use hypothesis fields (supports, contradicts, discriminators, excludes) to disambiguate between confusable failure patterns
- **faultline-matchers.yaml** - Defines shared matcher groups referenced by playbooks

## Features Demonstrated

### Match Composition (`match.use`)
See `example-setup-error.yaml` and `faultline-matchers.yaml`:
- Inline patterns in `match.any`
- Reusable matcher references in `match.use`
- Reduces duplication across multiple playbooks

### Hypothesis Fields
See `example-hypothesis-disambiguation.yaml`:
- `supports`: Signals that increase confidence in this playbook
- `contradicts`: Signals that decrease confidence (suggest other playbooks are more likely)
- `discriminators`: Strong evidence for this playbook over rivals
- `excludes`: Signals whose presence means this playbook should not match

## Try It

List all playbooks in this pack:
```bash
./bin/faultline list --playbook-pack examples/packs/minimal
```

Explain a specific example:
```bash
./bin/faultline explain example-cache-prime-missing --playbook-pack examples/packs/minimal
./bin/faultline explain example-setup-error --playbook-pack examples/packs/minimal
./bin/faultline explain example-hypothesis-disambiguation --playbook-pack examples/packs/minimal
```

Install it persistently if you want to test the installed-pack flow:

```bash
./bin/faultline packs install ./examples/packs/minimal --name example-pack
./bin/faultline packs list
```

## Learning Path

1. Start with **example-cache-prime-missing.yaml** to understand basic playbook structure
2. Explore **faultline-matchers.yaml** and **example-setup-error.yaml** to see how `match.use` reduces duplication
3. Study **example-hypothesis-disambiguation.yaml** to learn how hypothesis fields improve differential diagnosis
4. Use these patterns in your own org example packs
