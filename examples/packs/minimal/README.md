# Minimal Example Pack

This directory is a minimal extra playbook pack for local experimentation and authoring guidance.

It is not part of the bundled catalog and is not auto-loaded by default.

## Try It

```bash
./bin/faultline list --playbook-pack examples/packs/minimal
./bin/faultline explain example-cache-prime-missing --playbook-pack examples/packs/minimal
```

Install it persistently if you want to test the installed-pack flow:

```bash
./bin/faultline packs install ./examples/packs/minimal --name example-pack
./bin/faultline packs list
```
