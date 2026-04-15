# <Exact error string or problem-first title>

**Target search query:** `<exact query a developer would paste into Google>`

## Error snippet

```text
<smallest real CI log snippet that reproduces the issue>
```

## What this error means

<1-2 concise sentences explaining the failure and the usual CI cause.>

## Fix steps

1. <first fix>
2. <second fix>
3. <third fix>
4. <verification step>

## How Faultline detects it

Faultline maps this failure to `<playbook-id>`.

Primary signals:

- `<exact match string>`
- `<exact match string>`
- `<supporting context>`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [<related page>](<relative-link>)
- [<related page>](<relative-link>)