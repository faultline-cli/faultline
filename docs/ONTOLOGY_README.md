# Ontology Docs

## Current status

Faultline uses lightweight ontology metadata today through playbook fields such
as `domain`, `class`, and `mode`. These fields are catalog metadata: they help
humans, docs, and coverage reports group playbooks, but they do not change
matching, ranking, or remediation output.

The current shipped `faultline coverage` implementation is fixture-evidence
oriented. It reports:

- total resolved playbooks
- playbooks with positive fixture evidence
- positive fixture assertions
- negative fixture assertions
- strict top-1 fixture count
- category and domain grouping
- playbooks missing positive fixtures
- duplicate playbook IDs

Supported command surface:

```bash
faultline coverage
faultline coverage --json
faultline coverage --fixture-dir fixtures
faultline coverage --playbooks playbooks/bundled
faultline coverage --playbook-pack examples/packs/minimal
```

The command does not currently support ontology filtering flags such as
`--domain`, `--depth`, `--gaps`, CSV output, or ontology validation gates.

## Documents

- [ontology.md](ontology.md): historical taxonomy design and vocabulary. Treat
  it as background design, not a shipped CLI contract.
- [ontology-examples.md](ontology-examples.md): historical examples for richer
  ontology records. Useful as inspiration when refining metadata, but not the
  current required playbook schema.
- [ontology-quick-reference.md](ontology-quick-reference.md): current concise
  reference for the shipped metadata and `coverage` command.

The old implementation roadmap was removed because it described a planned
coverage implementation that no longer matches the current code.
