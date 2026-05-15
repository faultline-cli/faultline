# Database transaction opened without a deferred rollback

**Playbook ID:** `missing-transaction-rollback`
**Category:** runtime
**Severity:** high
**Tags:** `source`, `go`, `database`, `transaction`, `sql`, `reliability`

## What this failure means

A database transaction is opened and committed in the same function with no
`defer tx.Rollback()` guard, so if an error path returns early the transaction
is silently left open, holding connections and locks.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A `db.Begin()` or `db.BeginTx()` call is followed by `tx.Commit()` in the
same function, but there is no `defer tx.Rollback()` to release the
transaction on error paths.

The idiomatic Go pattern defers a rollback immediately after a successful
`Begin`:

```go
tx, err := db.Begin()
if err != nil { return err }
defer tx.Rollback() // safe no-op if Commit already succeeded
```

Without the deferred rollback:
- Any early return on an error before `Commit` leaves the transaction open
  until the database connection times out.
- Under concurrent load, leaked transactions exhaust the connection pool.
- Table locks held by uncommitted transactions block other writers.

## Fix steps

1. Add `defer tx.Rollback()` immediately after a successful `Begin`:
   ```go
   tx, err := db.BeginTx(ctx, nil)
   if err != nil {
       return fmt.Errorf("begin tx: %w", err)
   }
   defer tx.Rollback() // always deferred; no-op after a successful Commit

   // ... mutations ...

   return tx.Commit()
   ```
2. Return the error from `Commit` so the caller knows if it failed.
3. Avoid capturing `tx` in a closure that may outlive the deferred rollback.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Test that an error on any path before `Commit` releases the connection
  cleanly without exhausting the pool.

## Likely files to inspect

- `**/*.go`


## Run Faultline

```bash
faultline analyze build.log
faultline explain missing-transaction-rollback
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Database transaction opened without a deferred rollback
- Runtime: database transaction opened without a deferred rollback
- faultline explain missing-transaction-rollback
- Go database transaction opened without a deferred rollback


---

*Generated from [playbooks/bundled/source/missing-transaction-rollback.yaml](../../../playbooks/bundled/source/missing-transaction-rollback.yaml). Do not edit directly — run `make docs-generate`.*
