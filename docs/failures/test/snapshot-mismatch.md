# Snapshot or golden-file mismatch

**Playbook ID:** `snapshot-mismatch`
**Category:** test
**Severity:** low
**Tags:** `test`, `snapshot`, `golden`, `fixtures`

## What this failure means

A test compared current output against a committed snapshot or golden file and found a difference.

## Common log signals

```text
snapshot does not match
snapshot mismatch
stored snapshot
received value does not match stored snapshot
golden file mismatch
does not match the expected snapshot
update snapshots
toMatchSnapshot
```

## Diagnosis

A test compared current output against a committed snapshot or golden file and found a difference.

The behavior may have changed intentionally without regenerating the snapshot, or the output may now depend on unstable ordering, timestamps, or environment details.

## Fix steps

1. Inspect the diff between the current output and the stored snapshot to decide whether the behavior change is expected.
2. If the new output is correct, regenerate and commit the updated snapshots or golden files.
3. If the diff is unstable, remove nondeterministic fields such as timestamps, UUIDs, or map iteration order before snapshotting.
4. Re-run the specific test locally to confirm the output is stable across repeated runs.

## Validation

- Re-run the affected test after updating the snapshot or normalizing unstable output.
- Confirm the test passes repeatedly with the same output.

## Likely files to inspect

- `testdata/**`
- `**/*.snap`
- `**/*.golden`
- `**/*snapshot*`


## Run Faultline

```bash
faultline analyze build.log
faultline explain snapshot-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Snapshot or golden-file mismatch
- Test: snapshot or golden-file mismatch
- received value does not match stored snapshot
- faultline explain snapshot-mismatch


---

*Generated from [playbooks/bundled/log/test/snapshot-mismatch.yaml](../../../playbooks/bundled/log/test/snapshot-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
