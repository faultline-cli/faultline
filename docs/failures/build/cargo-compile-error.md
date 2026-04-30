# Rust cargo build compilation failure

**Playbook ID:** `cargo-compile-error`
**Category:** build
**Severity:** high
**Tags:** `rust`, `cargo`, `compile`, `build`, `rustc`, `borrow-checker`

## What this failure means

The Rust compiler rejected the crate with one or more hard errors — type
mismatches, borrow checker violations, unresolved imports, or missing trait
implementations. The binary cannot be produced until all errors are resolved.

## Common log signals

```text
error[E
error: could not compile
aborting due to previous error
aborting due to
undefined reference to
multiple definition of
```

## Diagnosis

`rustc` emits structured errors in the form `error[EXXXX]: <description>`
followed by source locations. Every error is deterministic and must be
resolved before the build can succeed.

Common error classes:

- **E0308 — mismatched types:** a value of one type is used where a
  different type is expected. Check function signatures and generic bounds.
- **E0382 — borrow of moved value / E0502 — conflicting borrows:** the
  borrow checker rejected unsafe aliasing or use-after-move. Review ownership
  and lifetime annotations.
- **E0412 / E0433 — unresolved name or module:** a type, function, or crate
  is referenced that is not in scope. Add the correct `use` statement or
  add the crate to `Cargo.toml`.
- **E0277 — trait not implemented:** a concrete type is used where a trait
  bound is required but the type does not implement the trait. Implement the
  trait or use a type adapter.
- **E0599 — method not found:** the method does not exist for this type, or
  a required feature flag is not enabled.

The compiler always names the exact file, line, and column. Start there.

## Fix steps

1. Run `cargo build` locally to reproduce the exact error output:

   ```bash
   cargo build 2>&1 | head -60
   ```

2. Read the first error and its suggestion. `rustc` often suggests the
   exact fix (e.g., `help: consider borrowing here: &val`).

3. For **E0308 type mismatch**: check the types on both sides of the
   assignment or call and add an explicit conversion or change the type.

4. For **E0382/E0502 borrow errors**: restructure to avoid overlapping
   borrows, use `.clone()` where ownership transfer is undesirable, or
   annotate lifetimes explicitly.

5. For **E0412/E0433 unresolved path**: verify the crate is listed in
   `Cargo.toml` and the `use` path matches the module tree:

   ```bash
   cargo tree | grep <crate-name>
   ```

6. For **E0277 trait not implemented**: implement the missing trait for
   your type, or add a `#[derive(...)]` macro if the trait supports it.

7. Run `cargo check` after each fix for fast feedback without producing
   a binary:

   ```bash
   cargo check
   ```

## Validation

- `cargo build` exits zero with no `error[E` lines in output.
- `cargo check` exits zero.
- Confirm the CI build step exits zero.

## Likely files to inspect

- `**/*.rs`
- `Cargo.toml`
- `Cargo.lock`
- `rust-toolchain.toml`
- `rust-toolchain`


## Run Faultline

```bash
faultline analyze build.log
faultline explain cargo-compile-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Rust cargo build compilation failure
- Build: rust cargo build compilation failure
- aborting due to previous error
- GitHub Actions rust cargo build compilation failure
- faultline explain cargo-compile-error
- Rust rust cargo build compilation failure


---

*Generated from [playbooks/bundled/log/build/cargo-compile-error.yaml](../../../playbooks/bundled/log/build/cargo-compile-error.yaml). Do not edit directly — run `make docs-generate`.*
