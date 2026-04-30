# .NET NuGet package restore failure

**Playbook ID:** `dotnet-restore`
**Category:** build
**Severity:** high
**Tags:** `dotnet`, `nuget`, `restore`, `csharp`, `build`

## What this failure means

`dotnet restore` or a build that implicitly runs restore could not download or resolve one or more NuGet packages. The build cannot proceed until all dependencies are available.

## Common log signals

```text
error NU
Package restore failed
NuGet restore
unable to find package
MSBuild error
Could not resolve SDK
NETStandard
```

## Diagnosis

`dotnet restore` or a build that implicitly runs restore could not download or resolve one or more NuGet packages. The build cannot proceed until all dependencies are available.

## Fix steps

1. Add feed credentials to `NuGet.Config` or as CI secrets for private feeds.
2. Run `dotnet restore --verbosity detailed` to identify which package is failing.
3. Update the package version to one that exists with `dotnet list package --outdated`.
4. If you need an offline restore path, restore from a known packages cache such as `dotnet restore --packages .nuget/packages`.

## Validation

- `dotnet restore` completes successfully.
- The CI build proceeds past the restore phase without the original NuGet error.

## Likely files to inspect

- `*.csproj`
- `*.sln`
- `NuGet.Config`
- `global.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain dotnet-restore
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- .NET NuGet package restore failure
- Build: .net nuget package restore failure
- Package restore failed
- GitHub Actions .net nuget package restore failure
- faultline explain dotnet-restore


---

*Generated from [playbooks/bundled/log/build/dotnet-restore.yaml](../../../playbooks/bundled/log/build/dotnet-restore.yaml). Do not edit directly — run `make docs-generate`.*
