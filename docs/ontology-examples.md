# CI Failure Ontology: Example Entries

This document provides detailed, ready-to-use ontology examples. Each example shows the complete structure for a specific failure mode, including evidence patterns, remediation strategies, and fixture guidance.

---

## Example 1: NPM Lockfile Mismatch

**Failure Domain:** `dependency`  
**Failure Class:** `lockfile-drift`  
**Failure Mode:** `npm-ci-requires-package-lock`

### Complete Ontology Record

```yaml
id: npm-ci-lockfile

# === Ontology Classification ===
domain: dependency
class: lockfile-drift
mode: npm-ci-requires-package-lock
aliases:
  - "npm-ci-lockfile-outdated"
  - "npm-lockfile-not-in-sync"

severity: medium
confidence_baseline: 0.95

# === Root Cause ===
root_cause: |
  npm ci enforces deterministic, lockfile-based installation. It refuses to
  proceed if package.json and package-lock.json disagree, because the
  disagreement indicates either:
  
  - The lockfile was not regenerated after package.json changed
  - The lockfile is corrupted or partially updated
  - Different versions of npm generated incompatible lockfile formats
  - Workspaces were modified without regenerating the root lockfile

# === Evidence Pattern ===
evidence:
  required:
    - log.contains: "npm ci can only install packages when"
    - log.regex: "package\\.json and package-lock\\.json.*(?:not in sync|out of sync|are in sync)"
  
  optional:
    - log.contains: "run `npm install` to update the lock"
    - log.contains: "added"
    - log.contains: "changed"
    - delta.signal: dependency.npm.lockfile.changed
    - delta.signal: dependency.npm.package.changed
  
  exclusions:
    - log.contains: "ENOENT"
    - log.contains: "404"
    - log.contains: "ERR! code E"
    - log.contains: "ERR! registry error"
  
  confidence: 0.95
  false_positive_risks:
    - |
      npm may emit this message during automatic recovery if npm install is
      re-run after ci fails. The second run's output might contain cached
      references to the sync failure message.
    - |
      Partial npm ci success (partial install) followed by cleanup may produce
      ambiguous intermediate messages that resemble a sync error.
    - |
      Very old npm versions may have different error message wording; grep
      should account for variations in "sync" language.

# === Remediation ===
remediation:
  strategy: align-lockfile
  
  summary: |
    Regenerate package-lock.json to match package.json, then commit.
  
  steps:
    - |
      Run npm install locally to regenerate package-lock.json:
      ```bash
      npm install
      ```
    
    - |
      Review the diff to ensure it looks sensible (version changes, new
      packages in package.json, etc.):
      ```bash
      git diff package-lock.json | head -100
      ```
    
    - |
      Commit the regenerated lockfile:
      ```bash
      git add package-lock.json
      git commit -m "fix: regenerate lockfile"
      ```
    
    - |
      Ensure package-lock.json is NOT listed in .gitignore.
    
    - |
      If the repo uses npm workspaces, regenerate from the workspace root with
      the same npm major version as CI:
      ```bash
      npm install --workspaces
      ```
  
  validation:
    - |
      Run npm ci locally and confirm it succeeds:
      ```bash
      npm ci
      ```
    
    - |
      Re-run the CI job and confirm the install step succeeds.
    
    - |
      Verify the lockfile checksum is stable across re-runs:
      ```bash
      npm ci
      git diff package-lock.json  # should be empty
      ```
  
  docs_link: "https://docs.npmjs.com/cli/v10/commands/npm-ci"

# === Fixtures ===
fixtures:
  positive:
    - id: npm-ci-lockfile-out-of-sync-simple
      source: |
        Generated from: Node 18.12.x + npm 10.2.x, packge.json modified
        but package-lock.json not updated.
      confidence: 0.95
      fixture_path: "fixtures/real/npm-ci-lockfile-simple.log"
    
    - id: npm-ci-lockfile-workspace-mismatch
      source: |
        npm workspace with root lockfile regenerated using npm 9.x but
        CI uses npm 10.x. Major version mismatch causes format differences.
      confidence: 0.90
      fixture_path: "fixtures/real/npm-workspace-mismatch.log"
    
    - id: npm-ci-lockfile-after-merge
      source: |
        Two branches both modified package.json independently. After
        merge, lockfile is stale. Reproduces common rebase/merge conflict.
      confidence: 0.93
      fixture_path: "fixtures/real/npm-ci-lockfile-merge-conflict.log"
  
  negative:
    - id: npm-ci-enoent-not-lockfile
      description: |
        ENOENT: no such file or directory pkgname. This is a missing
        package error or missing file reference, not a lockfile sync issue.
      confuses_with: npm-enoent-package-json
      fixture_path: "fixtures/real/npm-enoent-file-missing.log"
    
    - id: npm-ci-registry-auth-timeout
      description: |
        npm ERR! code ETIMEDOUT when reaching registry. Looks like a
        timeout, not a lockfile problem. The sync error message should not
        appear in this case.
      confuses_with: registry-auth
      fixture_path: "fixtures/real/npm-registry-timeout.log"
    
    - id: npm-install-works-but-ci-fails
      description: |
        Edge case: npm install succeeds but the next ci invocation in same
        log shows sync message. Indicates partial state or test artifact.
      confuses_with: flaky-install
      fixture_path: "fixtures/real/npm-partial-install-then-ci.log"

# === Related Modes ===
related_modes:
  - id: pnpm-lockfile-missing
    reason: |
      Same root cause (lockfile mismatch) but different package manager.
      Signals differ; pnpm uses .pnpm-lock.yaml and --frozen-lockfile flag.
  
  - id: yarn-lockfile-mismatch
    reason: |
      Same pattern for Yarn classic and Yarn v3+. Different remediation;
      Yarn uses yarn install vs npm install.
  
  - id: dependency-drift
    reason: |
      Broader class. npm-ci-lockfile is one specific manifestation of
      dependency misalignment. Dependency-drift catches other scenarios
      like version conflicts that don't involve lockfile sync.

# === Coverage ===
coverage:
  domains: [dependency]
  classes: [lockfile-drift]
  modes: [npm-ci-requires-package-lock]
  depth: deep
  stage_hints: [build]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

### Integration with Existing Playbook

This ontology would be added to `playbooks/bundled/log/build/npm-ci-lockfile.yaml`. The playbook file remains unchanged except for the new fields:

```yaml
id: npm-ci-lockfile
title: npm ci lockfile mismatch
category: build

# NEW: Ontology metadata
domain: dependency
class: lockfile-drift
mode: npm-ci-requires-package-lock
aliases: []
severity: medium
confidence_baseline: 0.95

# EXISTING: Match logic unchanged
match:
  any:
    - npm ci can only install packages when your package.json and package-lock.json
    - npm error `npm ci` can only install packages when your
    - package.json and package-lock.json are in sync
    - missing package-lock.json
    - run `npm install` to generate a lockfile
# ... rest of existing playbook
```

---

## Example 2: Missing Local Binary (Executable)

**Failure Domain:** `runtime`  
**Failure Class:** `missing-executable`  
**Failure Mode:** `node-missing-from-path`

### Complete Ontology Record

```yaml
id: missing-executable

# === Ontology Classification ===
domain: runtime
class: missing-executable
mode: node-missing-from-path
aliases:
  - "executable-not-found"
  - "missing-binary"

severity: high
confidence_baseline: 0.92

# === Root Cause ===
root_cause: |
  A required executable (node, python, docker, go, etc.) is not available
  in the CI environment's PATH. The CI step invokes the tool but the OS
  cannot locate it because:
  
  - The tool was never installed in the CI image
  - The tool is installed in a non-standard location not on PATH
  - The CI container or runtime environment doesn't include the tool by default
  - The tool installation step was skipped or failed silently
  - The PATH environment variable is not set correctly

# === Evidence Pattern ===
evidence:
  required:
    - log.regex: "command not found|^(sh:.*)?(.*/)?node:|[Nn]ode.*not found|Cannot find|not in PATH"
    - log.regex: "node:|node: command not found"
  
  optional:
    - log.contains: "node --version"
    - log.contains: "exec"
    - delta.signal: runtime.binary.missing
    - context.stage: build
  
  exclusions:
    - log.contains: "SyntaxError"
    - log.contains: "error TS"
    - log.contains: "ERR!"
    - log.contains: "exit code"
  
  confidence: 0.92
  false_positive_risks:
    - |
      Some framework error messages include "node:" prefix (e.g., "node:fs")
      when reporting import errors. These are not "command not found" errors.
      Filter by checking for shell error patterns.
    - |
      Some build tools have internal "node" references that can be confused
      with the executable binary. Prefer regex on shell output.

# === Remediation ===
remediation:
  strategy: install-missing-tool
  
  summary: |
    Install the missing executable (node, python, docker, etc.) in CI or
    ensure it's available in the PATH.
  
  steps:
    - |
      Identify which tool is missing. The error message should name it:
      ```bash
      # Example: "command not found: node"
      # Action: Install node
      ```
    
    - |
      Add the installation step to your CI configuration before the step
      that needs the tool. For Node:
      
      **GitHub Actions:**
      ```yaml
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      ```
      
      **CircleCI:**
      ```yaml
      - image: cimg/node:20
      ```
      
      **Jenkins:**
      ```bash
      curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
      apt-get install -y nodejs
      ```
    
    - |
      Verify the tool is installed and in PATH:
      ```bash
      which node
      node --version
      ```
  
  validation:
    - |
      Re-run the CI step that invoked the missing tool:
      ```bash
      node --version  # should succeed
      ```
    
    - |
      Confirm the subsequent build steps complete without "command not found".
  
  docs_link: "https://nodejs.org/en/download/package-manager/"

# === Fixtures ===
fixtures:
  positive:
    - id: missing-executable-node-ubuntu
      source: |
        CI container running Ubuntu 22.04 without Node pre-installed.
        Build step tries to run 'node' but gets "node: command not found".
      confidence: 0.95
      fixture_path: "fixtures/real/missing-executable-node-ubuntu.log"
    
    - id: missing-executable-python3-alpine
      source: |
        Alpine Linux container with only python 2.x. Build step invokes
        python3 but only python 2 is available. "python3: not found"
      confidence: 0.93
      fixture_path: "fixtures/real/missing-executable-python3-alpine.log"
    
    - id: missing-executable-docker-nonroot
      source: |
        Docker-in-Docker setup with unprivileged user. User hasn't been
        added to docker group; docker command exists but user has no permission.
      confidence: 0.88
      fixture_path: "fixtures/real/missing-executable-docker-permission.log"
  
  negative:
    - id: executable-exists-but-different-error
      description: |
        node command exists but exits with SyntaxError. Error mentions
        "node:" but it's a code error, not a missing-executable error.
      confuses_with: syntax-error
      fixture_path: "fixtures/real/node-syntax-error.log"
    
    - id: executable-in-different-location
      description: |
        Tool is installed to /opt/tools/node instead of standard location.
        Script uses "which node" but found via full path. Not a missing-executable.
      confuses_with: wrong-working-directory
      fixture_path: "fixtures/real/executable-alternate-location.log"

# === Related Modes ===
related_modes:
  - id: runtime-mismatch
    reason: |
      If the executable exists but is the wrong version (e.g., Node 14
      when code requires Node 18), that's runtime-mismatch, not missing-executable.
  
  - id: docker-permission-denied-nonroot
    reason: |
      Docker command exists but user lacks permission. Different remediation;
      add user to docker group vs install the tool.

# === Coverage ===
coverage:
  domains: [runtime]
  classes: [missing-executable]
  modes: [node-missing-from-path]
  depth: deep
  stage_hints: [build, test]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

---

## Example 3: Python Interpreter Mismatch

**Failure Domain:** `runtime`  
**Failure Class:** `interpreter-mismatch`  
**Failure Mode:** `python-module-installed-to-wrong-interpreter`

### Complete Ontology Record

```yaml
id: python-virtualenv-not-activated

# === Ontology Classification ===
domain: runtime
class: interpreter-mismatch
mode: python-module-installed-to-wrong-interpreter
aliases:
  - "python-venv-mismatch"
  - "python-interpreter-version-conflict"

severity: medium
confidence_baseline: 0.88

# === Root Cause ===
root_cause: |
  A Python package was installed using one Python interpreter (e.g., system
  python 2.7) but the code is executed with a different interpreter (e.g.,
  python 3.9 or a venv). The installed package is not available to the
  running interpreter because:
  
  - pip was not run in a virtual environment (installed to system Python)
  - Different python versions have different site-packages locations
  - CI activated the venv AFTER running pip, or re-activated a different venv
  - Local development created site-packages for Python 2 but CI uses Python 3
  - A workflow step changed PYTHONPATH between install and execution

# === Evidence Pattern ===
evidence:
  required:
    - log.regex: "ModuleNotFoundError|ImportError.*no module named"
  
  optional:
    - log.contains: "virtualenv"
    - log.contains: "venv"
    - log.contains: "PYTHONPATH"
    - log.contains: "site-packages"
    - log.regex: "python.*\\.x"
    - delta.signal: runtime.python.version.changed
  
  exclusions:
    - log.contains: "SyntaxError"
    - log.contains: "pip install"
    - log.contains: "ERR!"
  
  confidence: 0.88
  false_positive_risks:
    - |
      ModuleNotFoundError can occur for many reasons (missing dependency,
      typo in import statement, circular imports). Combine with evidence of
      interpreter mismatch (venv mentions, PYTHONPATH changes).
    - |
      Some CI logs mention "virtualenv" in infrastructure setup without
      indicating an actual mismatch. Require import failure evidence.

# === Remediation ===
remediation:
  strategy: activate-venv
  
  summary: |
    Create and activate a Python virtual environment, then install
    dependencies into it before executing code.
  
  steps:
    - |
      Create a virtual environment:
      ```bash
      python3 -m venv venv
      ```
    
    - |
      Activate it (before any pip installs or code execution):
      ```bash
      source venv/bin/activate  # on Linux/macOS
      # or
      .\\venv\\Scripts\\activate  # on Windows
      ```
    
    - |
      Install dependencies into the venv:
      ```bash
      pip install -r requirements.txt
      ```
    
    - |
      Verify the interpreter is correct:
      ```bash
      which python
      python --version
      python -c "import sys; print(sys.prefix)"
      ```
    
    - |
      Ensure subsequent steps either keep the venv activated or export
      PYTHONPATH if running bare python commands:
      ```bash
      export PYTHONPATH=${PWD}/venv/lib/python3.9/site-packages
      ```
  
  validation:
    - |
      Run the import that failed:
      ```bash
      python -c "import requests"  # or whatever module failed
      ```
    
    - |
      Re-run the full test or application and confirm ModuleNotFoundError
      is gone.
  
  docs_link: "https://docs.python.org/3/tutorial/venv.html"

# === Fixtures ===
fixtures:
  positive:
    - id: python-venv-not-activated
      source: |
        CI script installs pip packages globally, then runs tests in a
        step that should use venv. ModuleNotFoundError: no module named 'pytest'.
      confidence: 0.88
      fixture_path: "fixtures/real/python-venv-not-activated.log"
    
    - id: python-venv-wrong-version
      source: |
        requirements.txt locked to Python 3.9, but CI activated Python 3.8
        venv. Import fails due to incompatible compiled extensions or
        missing version-specific modules.
      confidence: 0.82
      fixture_path: "fixtures/real/python-venv-version-mismatch.log"
    
    - id: python-venv-deactivated-mid-pipeline
      source: |
        Earlier step activates venv correctly. Middle step deactivates
        it (e.g., subprocess call loses environment). Later step fails
        with ModuleNotFoundError.
      confidence: 0.80
      fixture_path: "fixtures/real/python-venv-deactivated-mid-run.log"
  
  negative:
    - id: python-import-syntax-error
      description: |
        Code has SyntaxError in import statement or module file. Not a
        missing module; the module exists but can't be parsed.
      confuses_with: syntax-error
      fixture_path: "fixtures/real/python-import-syntax-error.log"
    
    - id: python-circular-import
      description: |
        ModuleNotFoundError occurs but root cause is circular imports in
        application code, not venv or interpreter mismatch.
      confuses_with: code-structure-error
      fixture_path: "fixtures/real/python-circular-import.log"

# === Related Modes ===
related_modes:
  - id: runtime-mismatch
    reason: |
      If Python is the right version but a compiled extension is
      incompatible (e.g., binary compiled for Python 3.8 used in 3.9),
      that's runtime-mismatch, not interpreter-mismatch.
  
  - id: missing-executable
    reason: |
      If python command itself doesn't exist, that's missing-executable,
      not interpreter-mismatch.

# === Coverage ===
coverage:
  domains: [runtime]
  classes: [interpreter-mismatch]
  modes: [python-module-installed-to-wrong-interpreter]
  depth: deep
  stage_hints: [build, test]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

---

## Example 4: Docker COPY Missing File

**Failure Domain:** `filesystem`  
**Failure Class:** `wrong-working-directory`  
**Failure Mode:** `docker-copy-source-missing`

### Complete Ontology Record

```yaml
id: dockerfile-copy-source-missing

# === Ontology Classification ===
domain: filesystem
class: wrong-working-directory
mode: docker-copy-source-missing
aliases:
  - "dockerfile-copy-not-found"
  - "docker-build-context-missing"

severity: high
confidence_baseline: 0.94

# === Root Cause ===
root_cause: |
  A Dockerfile COPY or ADD command references a source file or directory
  that does not exist in the build context. Docker build context is defined
  by the git repository root, .dockerignore filters, or explicit context
  path. The failure occurs because:
  
  - File path in COPY is wrong (typo, case sensitivity, path separator)
  - File was not checked into git (or is in .gitignore)
  - Directory structure changed (refactoring moved the file)
  - .dockerignore is too aggressive and excludes needed files
  - COPY command runs before the file is generated (wrong step order)

# === Evidence Pattern ===
evidence:
  required:
    - log.regex: "COPY|ADD"
    - log.regex: "no such file or directory|not found in build context|lstat.*no such file"
  
  optional:
    - log.contains: "Dockerfile"
    - log.contains: "step"
    - log.regex: "line \\d+"
    - delta.signal: filesystem.file.missing
  
  exclusions:
    - log.contains: "npm install"
    - log.contains: "ERR!"
    - log.contains: "pip install"
  
  confidence: 0.94
  false_positive_risks:
    - |
      Generic "no such file" errors can occur in many contexts (npm install,
      pip install, git operations). Require both COPY/ADD keyword AND the
      "not found in build context" phrase for high confidence.
    - |
      Some base image operations might emit similar errors unrelated to
      user COPY commands. Require line number or step indicator.

# === Remediation ===
remediation:
  strategy: use-correct-file-path
  
  summary: |
    Verify the file exists in the build context at the path specified in
    the COPY or ADD command.
  
  steps:
    - |
      Identify the file and line in the Dockerfile that failed:
      ```dockerfile
      COPY src/main.py /app/  # Line 5: "src/main.py" not found
      ```
    
    - |
      Check if the file exists locally:
      ```bash
      ls -la src/main.py
      ```
    
    - |
      If the file doesn't exist, either:
      
      a) Create the file (if it's supposed to exist)
      
      b) Update the COPY path to match file's actual location:
         ```dockerfile
         COPY app/src/main.py /app/
         ```
      
      c) Check if the file is gitignored and should be committed:
         ```bash
         git check-ignore -v src/main.py
         git add src/main.py
         ```
    
    - |
      If using .dockerignore, verify it's not filtering needed files:
      ```bash
      # List what docker will include
      docker build --progress=plain . 2>&1 | grep -i ignore
      ```
    
    - |
      Verify case sensitivity (especially on Windows/macOS developers
      pushing to Linux CI):
      ```bash
      # Check actual file name
      ls src/Main.py vs ls src/main.py
      ```
  
  validation:
    - |
      Re-run docker build:
      ```bash
      docker build -t test:latest .
      ```
    
    - |
      Confirm the build succeeds and the COPY step passes.
  
  docs_link: "https://docs.docker.com/engine/reference/builder/#copy"

# === Fixtures ===
fixtures:
  positive:
    - id: dockerfile-copy-file-not-found
      source: |
        Dockerfile has COPY src/app.js /app/ but file is actually at
        app/src/app.js. Docker build fails with "no such file in build context".
      confidence: 0.94
      fixture_path: "fixtures/real/dockerfile-copy-missing.log"
    
    - id: dockerfile-add-directory-missing
      source: |
        Dockerfile ADD ./dist /app/dist assumes dist/ is built, but build
        step failed silently or dist/ is gitignored. Docker can't find
        the directory.
      confidence: 0.92
      fixture_path: "fixtures/real/dockerfile-add-missing-directory.log"
    
    - id: dockerfile-copy-gitignore-conflict
      source: |
        Dockerfile wants to COPY .env.example but .env* is in .gitignore.
        File doesn't exist in build context because it's not committed.
      confidence: 0.90
      fixture_path: "fixtures/real/dockerfile-copy-gitignore-excludes.log"
  
  negative:
    - id: dockerfile-copy-succeeds-with-wrong-image
      description: |
        COPY succeeds but later RUN command references missing tool.
        The failure is not about COPY, but about a missing executable in
        the image.
      confuses_with: missing-executable
      fixture_path: "fixtures/real/dockerfile-run-command-missing.log"
    
    - id: dockerfile-from-not-found
      description: |
        FROM ubuntu:99.99 references non-existent base image tag. Error
        is about image pull, not COPY.
      confuses_with: image-pull-backoff
      fixture_path: "fixtures/real/dockerfile-from-not-found.log"

# === Related Modes ===
related_modes:
  - id: working-directory
    reason: |
      If COPY path itself is correct but the working directory in the
      container is wrong, that's a different issue.
  
  - id: docker-build-context
    reason: |
      Broader class covering other docker build failures. COPY missing file
      is one specific manifestation.

# === Coverage ===
coverage:
  domains: [filesystem]
  classes: [wrong-working-directory]
  modes: [docker-copy-source-missing]
  depth: deep
  stage_hints: [build]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

---

## Example 5: GitHub Actions Env Not Persisted

**Failure Domain:** `ci-config`  
**Failure Class:** `env-not-persisted`  
**Failure Mode:** `github-actions-env-not-carried-to-next-step`

### Complete Ontology Record

```yaml
id: github-actions-environment-persistence

# === Ontology Classification ===
domain: ci-config
class: env-not-persisted
mode: github-actions-env-not-carried-to-next-step
aliases:
  - "github-actions-env-lost"
  - "github-actions-env-isolation"

severity: high
confidence_baseline: 0.91

# === Root Cause ===
root_cause: |
  A GitHub Actions workflow step sets an environment variable, but a
  subsequent step does not see it. This occurs because:
  
  - Step used bare export (not ::set-env or $GITHUB_ENV)
  - Variable was exported to a subshell that exited
  - Each step runs in its own shell session with separate environment
  - Used set-env (deprecated GitHub Actions syntax)
  - The workflow defines variables at job level vs step level
  - Syntax error in GITHUB_ENV export (missing delimiters, newlines)

# === Evidence Pattern ===
evidence:
  required:
    - log.contains: "GITHUB_ENV"
    - log.regex: "env.*not.*found|${.*}.*not.*found|undefined.*variable"
  
  optional:
    - log.contains: "GitHub Actions"
    - log.contains: "::set-env"
    - log.regex: "step \\d+"
    - context.stage: build
  
  exclusions:
    - log.contains: "secret"
    - log.contains: "permission"
    - log.contains: "404"
  
  confidence: 0.91
  false_positive_risks:
    - |
      Some dependency errors include "not found" but aren't about env vars.
      Require explicit GitHub Actions syntax or GITHUB_ENV reference.
    - |
      Step output (::set-output) is different from environment persistence.
      Confirm the error is about env var, not step output variable.

# === Remediation ===
remediation:
  strategy: persist-ci-env-correctly
  
  summary: |
    Use GitHub Actions environment variable syntax (GITHUB_ENV) to persist
    variables across steps instead of bare export.
  
  steps:
    - |
      Identify which step exports the variable. Example:
      
      **WRONG (won't persist to next step):**
      ```yaml
      - name: Build
        run: export MY_VERSION=1.0.0
      
      - name: Deploy
        run: echo $MY_VERSION  # empty!
      ```
    
    - |
      Use $GITHUB_ENV to persist the variable:
      
      **CORRECT:**
      ```yaml
      - name: Build
        run: echo "MY_VERSION=1.0.0" >> $GITHUB_ENV
      
      - name: Deploy
        run: echo $MY_VERSION  # outputs: 1.0.0
      ```
    
    - |
      For multi-line values or special characters, use delimiters:
      ```yaml
      - name: Build
        run: |
          {
            echo "MULTI_LINE_VAR<<EOF"
            echo "line 1"
            echo "line 2"
            echo "EOF"
          } >> $GITHUB_ENV
      
      - name: Use it
        run: echo "${{ env.MULTI_LINE_VAR }}"
      ```
    
    - |
      Verify the syntax:
      ```bash
      echo "KEY=value" >> $GITHUB_ENV
      echo "KEY with spaces=value" >> $GITHUB_ENV
      # Do NOT quote the left side of >>
      ```
  
  validation:
    - |
      Re-run the workflow and verify the variable is available in the next step:
      ```yaml
      - name: Verify
        run: echo "MY_VERSION=${{ env.MY_VERSION }}"
      ```
    
    - |
      Check workflow logs; the Deploy step should show the non-empty value.
  
  docs_link: "https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#environment-files"

# === Fixtures ===
fixtures:
  positive:
    - id: github-actions-env-export-not-persisted
      source: |
        Step runs 'export VERSION=...' but doesn't use GITHUB_ENV.
        Next step tries to use $VERSION and gets "not found" error.
      confidence: 0.91
      fixture_path: "fixtures/real/github-actions-env-export-fails.log"
    
    - id: github-actions-set-env-deprecated
      source: |
        Step uses deprecated ::set-env syntax which GitHub deprecated
        in August 2020. New runner ignores it.
      confidence: 0.89
      fixture_path: "fixtures/real/github-actions-set-env-deprecated.log"
    
    - id: github-actions-env-multiline-delimiter-wrong
      source: |
        Step tries to set multi-line env var but uses wrong delimiter
        syntax. Next step sees incomplete or corrupted value.
      confidence: 0.85
      fixture_path: "fixtures/real/github-actions-env-multiline-wrong.log"
  
  negative:
    - id: github-actions-secrets-not-env
      description: |
        Workflow uses secrets which are handled differently than env vars.
        Secrets are available directly; this is not an env persistence issue.
      confuses_with: missing-secret
      fixture_path: "fixtures/real/github-actions-secret-not-found.log"
    
    - id: github-actions-condition-false
      description: |
        Step is skipped due to 'if:' condition being false. The variable
        isn't "missing"—the step producing it didn't run.
      confuses_with: conditional-skip
      fixture_path: "fixtures/real/github-actions-step-skipped.log"

# === Related Modes ===
related_modes:
  - id: secrets-not-available
    reason: |
      If the missing value is a GitHub Secret (not a custom env var),
      that's a different detection and remediation path.
  
  - id: ci-config-validation
    reason: |
      Broader class for workflow configuration issues. Env persistence is
      one specific manifestation.

# === Coverage ===
coverage:
  domains: [ci-config]
  classes: [env-not-persisted]
  modes: [github-actions-env-not-carried-to-next-step]
  depth: deep
  stage_hints: [build, test, deploy]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

---

## Example 6: Postgres Service Not Ready

**Failure Domain:** `database`  
**Failure Class:** `service-not-ready`  
**Failure Mode:** `postgres-connection-refused-startup-lag`

### Complete Ontology Record

```yaml
id: postgres-connection-refused

# === Ontology Classification ===
domain: database
class: service-not-ready
mode: postgres-connection-refused-startup-lag
aliases:
  - "postgresql-not-listening"
  - "postgres-not-accepting-connections"

severity: medium
confidence_baseline: 0.89

# === Root Cause ===
root_cause: |
  CI tries to connect to PostgreSQL before the service has finished
  starting or before it's accepting connections. This occurs because:
  
  - Container or service was just started; initialization is not complete
  - Database initialization (createdb, migrations) is still running
  - Service is accepting TCP connections but not ready for application traffic
  - No health check or readiness probe before application connects
  - Network delays cause the application to be ready before the DB
  - Port conflict or binding issue prevents the service from listening

# === Evidence Pattern ===
evidence:
  required:
    - log.regex: "could not connect to server|connection refused|psql:.*could not translate"
    - log.regex: "postgres|postgresql|ECONNREFUSED.*5432"
  
  optional:
    - log.contains: "FATAL"
    - log.contains: "starting"
    - log.contains: "port 5432"
    - log.regex: "waiting|sleep|retry"
    - delta.signal: database.postgres.startup_delay
  
  exclusions:
    - log.contains: "password authentication failed"
    - log.contains: "permission denied"
    - log.contains: "does not exist"
  
  confidence: 0.89
  false_positive_risks:
    - |
      "Connection refused" can come from wrong credentials (authentication
      failure), not service readiness. Exclude authentication errors.
    - |
      Service may be running but configured to a different port. Require
      evidence of port 5432 or generic postgres not listening.

# === Remediation ===
remediation:
  strategy: wait-for-service-readiness
  
  summary: |
    Add retry logic or a health check to wait for PostgreSQL to be ready
    before connecting.
  
  steps:
    - |
      Add a wait loop before running the application. For GitHub Actions:
      
      ```yaml
      - name: Wait for PostgreSQL
        run: |
          for i in {1..30}; do
            psql -h localhost -U postgres -d postgres -c "SELECT 1" && break
            echo "PostgreSQL not ready, waiting... ($i/30)"
            sleep 2
          done
      ```
    
    - |
      For Docker Compose, use depends_on with condition:
      
      ```yaml
      services:
        app:
          depends_on:
            postgres:
              condition: service_healthy
        postgres:
          image: postgres:15
          healthcheck:
            test: ["CMD-SHELL", "pg_isready -U postgres"]
            interval: 5s
            timeout: 5s
            retries: 5
      ```
    
    - |
      For Docker directly, use pg_isready:
      
      ```bash
      docker run --rm \
        --network my-net \
        postgres:15 \
        pg_isready -h postgres -U postgres
      ```
    
    - |
      Increase the retry timeout if the database needs to:
      - Copy large data files
      - Run initialization scripts
      - Perform migrations
      
      ```yaml
      for i in {1..60}; do  # Increased from 30 to 60
        psql ... && break
        sleep 2
      done
      ```
  
  validation:
    - |
      Confirm PostgreSQL is listening:
      ```bash
      psql -h localhost -U postgres -d postgres -c "SELECT version();"
      ```
    
    - |
      Re-run the CI job and confirm the connection step succeeds.
    
    - |
      If timing is still flaky, add explicit health check:
      ```bash
      psql ... -c "SELECT 1 FROM information_schema.tables"
      ```
  
  docs_link: "https://www.postgresql.org/docs/current/app-pg-isready.html"

# === Fixtures ===
fixtures:
  positive:
    - id: postgres-connection-refused-immediate
      source: |
        Application tries to connect to postgres immediately after
        docker-compose up. Service is starting but not accepting connections yet.
      confidence: 0.89
      fixture_path: "fixtures/real/postgres-connection-refused.log"
    
    - id: postgres-initialization-not-complete
      source: |
        postgres container is running but initdb is still executing.
        Connection attempts fail until initialization completes.
      confidence: 0.87
      fixture_path: "fixtures/real/postgres-initialization-incomplete.log"
    
    - id: postgres-slow-startup-migration
      source: |
        postgres starts normally but database migrations are running
        in a parallel container. Application connects before migrations finish.
      confidence: 0.85
      fixture_path: "fixtures/real/postgres-migration-timeout.log"
  
  negative:
    - id: postgres-auth-failed-wrong-password
      description: |
        postgres is running and accepting connections, but connection
        fails due to wrong password. Not a startup issue.
      confuses_with: docker-auth
      fixture_path: "fixtures/real/postgres-wrong-credentials.log"
    
    - id: postgres-database-not-exist
      description: |
        Connection succeeds but SELECT fails because the target database
        was not created. Not a service-not-ready issue.
      confuses_with: postgres-database-missing
      fixture_path: "fixtures/real/postgres-database-does-not-exist.log"

# === Related Modes ===
related_modes:
  - id: connection-refused
    reason: |
      Broader class for network connection refusals. Postgres is one specific
      database service; MySQL, Redis, etc. follow the same pattern.
  
  - id: network-timeout
    reason: |
      If the connection hangs (slow network), that's different from refused.
      Timeout suggests service exists but is slow; refused suggests not listening.

# === Coverage ===
coverage:
  domains: [database]
  classes: [service-not-ready]
  modes: [postgres-connection-refused-startup-lag]
  depth: deep
  stage_hints: [test, deploy]
  has_negative_fixtures: true
  has_workflow_hooks: false
```

---

## Summary: Cross-Example Patterns

These six examples demonstrate the ontology in action:

| Example | Domain | Class | Mode | Confidence | Remediation Strategy |
|---------|--------|-------|------|------------|----------------------|
| npm-ci-lockfile | dependency | lockfile-drift | npm-ci-requires-package-lock | 0.95 | align-lockfile |
| missing-binary | runtime | missing-executable | node-missing-from-path | 0.92 | install-missing-tool |
| python-venv | runtime | interpreter-mismatch | python-module-installed | 0.88 | activate-venv |
| dockerfile-copy | filesystem | wrong-working-directory | docker-copy-source-missing | 0.94 | use-correct-file-path |
| github-actions-env | ci-config | env-not-persisted | github-actions-env-not-persisted | 0.91 | persist-ci-env-correctly |
| postgres-conn | database | service-not-ready | postgres-connection-refused | 0.89 | wait-for-service-readiness |

**Key Insights:**

1. **Each ontology record is self-contained** - domain, class, mode, evidence, remediation, and fixtures are all defined in one place
2. **Confidence varies by domain** - Some failures (npm lockfiles, docker COPY) are highly deterministic (0.94+); others (python interpreter, postgres startup) require more context (0.88-0.89)
3. **Negative fixtures are critical** - Each example includes false positives (what should NOT match) to prevent over-ambitious matchers
4. **Related modes guide ranking** - When multiple modes might match the same log, the `related_modes` section helps discriminate
5. **Remediation strategies are reusable** - Multiple failure modes may use the same strategy (e.g., both `missing-executable` and `python-venv` use "fix the environment")

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-25
