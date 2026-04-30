# MSBuild task or target failure

**Playbook ID:** `msbuild-error`
**Category:** build
**Severity:** high
**Tags:** `dotnet`, `msbuild`, `csharp`, `fsharp`, `vbnet`, `build`

## What this failure means

MSBuild exited with a fatal `error MSBxxxx` diagnostic. The task or target
could not complete and the build cannot proceed until the root cause is
resolved.

## Common log signals

```text
error MSB
MSB4006
MSB3644
MSB4018
MSB6006
MSB3027
MSB3491
MSB1003
```

## Diagnosis

MSBuild emitted a structured error (`error MSBxxxx`) from a task or target.
Common causes include:

- **MSB4006** — circular dependency in the target dependency graph (e.g.
  a project referencing itself or two projects referencing each other through
  the `_GenerateRestoreProjectPathWalk` target)
- **MSB3644 / MSB3270** — reference assemblies for the requested target
  framework are missing or the processor architecture of a reference does not
  match the build output
- **MSB4018** — an MSBuild task threw an unhandled exception; the stack trace
  appears immediately below the error line
- **MSB6006** — a spawned tool (compiler, linker, custom task) returned a
  non-zero exit code
- **MSB3027 / MSB3491** — the build could not copy or write an output file,
  usually because it is locked by another process or the output path does not
  exist
- **MSB1003** — no project or solution file found; the working directory or
  `DOTNET_PROJECT_DIR` is wrong

## Fix steps

1. Read the full `error MSBxxxx` line in the log — the four-digit code
   pinpoints the category. Look the code up in the
   [MSBuild error reference](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-errors).
2. **For MSB4006 (circular dependency):**
   - Open the `.csproj` / `.vbproj` / `.fsproj` file named in the error.
   - Check `<ProjectReference>` items for cycles (A → B → A).
   - Check custom `<Target>` elements whose `DependsOnTargets` or
     `BeforeTargets` / `AfterTargets` attributes create a cycle.
   - Remove or break the cycle; use a shared library project to hold
     shared code instead of mutual references.
3. **For MSB3644 (missing reference assemblies):**
   - Verify `<TargetFramework>` in the `.csproj` matches an SDK installed
     in the CI image.
   - Pin the SDK version in `global.json` and update the CI image to match.
4. **For MSB4018 (task threw exception):**
   - Read the stack trace that follows the error line.
   - If the failing task is a NuGet task, see the `dotnet-restore` playbook.
   - If the failing task is a custom task, rebuild or update the task assembly.
5. **For MSB6006 (tool returned non-zero):**
   - The real compile or link error will appear above the MSB6006 line —
     scroll up to find it.
6. **For MSB3027 / MSB3491 (file copy / write failure):**
   - Check whether another process is holding the output file.
   - Ensure the output directory exists and the runner has write permission.
7. **For MSB1003 (project not found):**
   - Confirm the working directory in the CI step matches where the
     `.sln` or `.csproj` lives.
   - Check `DOTNET_PROJECT_DIR` if the pipeline sets it.
8. Run `dotnet build --verbosity detailed` locally to reproduce with full
   task-level output.

## Validation

- `dotnet build` (or `dotnet restore && dotnet build`) completes with exit
  code 0 and no `error MSB` lines in the output.
- The CI step that previously failed now exits zero.

## Likely files to inspect

- `*.csproj`
- `*.vbproj`
- `*.fsproj`
- `*.sln`
- `global.json`
- `Directory.Build.props`
- `Directory.Build.targets`


## Run Faultline

```bash
faultline analyze build.log
faultline explain msbuild-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- MSBuild task or target failure
- Build: msbuild task or target failure
- circular dependency in the target dependency graph
- GitHub Actions msbuild task or target failure
- faultline explain msbuild-error


---

*Generated from [playbooks/bundled/log/build/msbuild-error.yaml](../../../playbooks/bundled/log/build/msbuild-error.yaml). Do not edit directly — run `make docs-generate`.*
