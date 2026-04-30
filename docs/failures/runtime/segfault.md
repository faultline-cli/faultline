# Process terminated by segmentation fault or fatal signal

**Playbook ID:** `segfault`
**Category:** runtime
**Severity:** critical
**Tags:** `segfault`, `sigsegv`, `signal`, `crash`, `core-dump`

## What this failure means

The process received a fatal signal — most commonly SIGSEGV (segmentation
fault), SIGABRT (assertion failure), or SIGBUS (bus error) — and was
terminated by the operating system. This is a hard crash with no graceful
recovery.

## Common log signals

```text
segmentation fault
Segmentation fault
SIGSEGV
signal: segmentation fault
core dumped
signal: abort trap
signal: aborted
SIGABRT
```

## Diagnosis

The process accessed memory it was not permitted to use (null pointer
dereference, buffer overrun, use-after-free) or explicitly aborted via an
assertion failure or `abort()` call. The OS kills the process immediately and
may produce a core dump.

Common causes:
- **Null pointer dereference**: a pointer is used before being initialised or
  after being freed.
- **Buffer overrun / stack overflow**: writing past the end of a fixed-size
  buffer or allocating an unbounded recursive call stack.
- **Use-after-free**: memory that was freed is accessed again.
- **C extension fault in a managed runtime**: a native library (CGo, Python C
  extension, JNI) crashes and takes the host process with it.
- **Stack overflow in tests**: a deeply recursive function or an infinite
  mutual recursion blows the default stack size.

## Fix steps

1. Identify the crashing binary and reproduce locally with verbose signal
   handling:

   ```bash
   # Linux: enable core dumps and inspect with gdb
   ulimit -c unlimited
   ./failing-binary
   gdb ./failing-binary core

   # Go: set GOTRACEBACK for a full goroutine stack on crash
   GOTRACEBACK=all go test ./...

   # Python C extension
   python -c "import faulthandler; faulthandler.enable(); import your_module"
   ```

2. Attach a memory sanitiser to find the root cause:

   ```bash
   # C/C++: compile with AddressSanitizer
   gcc -fsanitize=address -g -o binary source.c
   ./binary

   # Go: use the race detector for concurrent access faults
   go test -race ./...
   ```

3. For stack overflows: add a base case to the recursion, increase the goroutine
   stack limit, or rewrite the logic iteratively.

4. For C extension crashes: pin the extension to a known-good version, or
   run it inside a subprocess so the host process survives a crash.

5. Check for version mismatches between shared libraries and the binary —
   a `SIGABRT` from `abort()` often signals an ABI incompatibility.

## Validation

- Run the reproducing command and confirm the process exits zero, not with
  a signal.
- Confirm no `Segmentation fault`, `signal: killed`, or core dump appears
  in the output.
- Run under AddressSanitizer or Valgrind at least once to rule out latent
  memory errors.

## Likely files to inspect

- `**/*.c`
- `**/*.cpp`
- `**/*.go`
- `**/*.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain segfault
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Process terminated by segmentation fault or fatal signal
- Runtime: process terminated by segmentation fault or fatal signal
- Access violation in generated code
- faultline explain segfault


---

*Generated from [playbooks/bundled/log/runtime/segfault.yaml](../../../playbooks/bundled/log/runtime/segfault.yaml). Do not edit directly — run `make docs-generate`.*
