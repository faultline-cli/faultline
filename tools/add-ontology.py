#!/usr/bin/env python3
"""
Add domain/class/mode ontology fields to playbooks that lack them.
Fields are inserted after the stage_hints: line.
Playbooks that already have domain: are skipped.
"""

import os
import re
import sys

# Complete ontology mapping: playbook id -> (domain, class, mode)
ONTOLOGY = {
    # auth
    "aws-credentials":                  ("auth",        "registry-auth",             "aws-credentials-invalid"),
    "docker-auth":                      ("auth",        "registry-auth",             "docker-registry-auth-failed"),
    "git-auth":                         ("auth",        "credential-expired",        "git-auth-failed"),
    "kubectl-auth":                     ("auth",        "credential-expired",        "kubectl-auth-failed"),
    "missing-env":                      ("auth",        "missing-credential",        "missing-env-credential"),
    "oidc-token-failure":               ("auth",        "token-expired",             "oidc-token-invalid"),
    "ssh-key-auth":                     ("auth",        "credential-expired",        "ssh-key-auth-failed"),
    "ssh-permission-denied":            ("auth",        "credential-expired",        "ssh-permission-denied"),

    # build / dependency
    "npm-ci-lockfile":                  ("dependency",  "lockfile-drift",            "npm-ci-requires-package-lock"),
    "pnpm-lockfile":                    ("dependency",  "lockfile-drift",            "pnpm-frozen-lockfile-outdated"),
    "pnpm-lockfile-missing":            ("dependency",  "lockfile-drift",            "pnpm-frozen-lockfile-missing"),
    "poetry-lockfile-drift":            ("dependency",  "lockfile-drift",            "poetry-lockfile-drift"),
    "yarn-lockfile":                    ("dependency",  "lockfile-drift",            "yarn-lockfile-outdated"),
    "go-sum-missing":                   ("dependency",  "lockfile-drift",            "go-sum-missing"),
    "dependency-drift":                 ("dependency",  "version-conflict",          "dependency-version-drift"),
    "npm-eresolve-conflict":            ("dependency",  "version-conflict",          "npm-eresolve-conflict"),
    "npm-peer-dependency-conflict":     ("dependency",  "peer-dependency-mismatch",  "npm-peer-dependency-conflict"),
    "package-manager-mismatch":         ("dependency",  "version-conflict",          "package-manager-lockfile-mismatch"),
    "workspace-dependency-mismatch":    ("dependency",  "version-conflict",          "workspace-dependency-mismatch"),
    "dotnet-restore":                   ("dependency",  "missing-package",           "dotnet-restore-failed"),
    "install-failure":                  ("dependency",  "missing-package",           "package-install-failed"),
    "maven-dependency-resolution":      ("dependency",  "missing-package",           "maven-dependency-not-found"),
    "pip-install-failure":              ("dependency",  "missing-package",           "pip-install-failed"),
    "python-module-missing":            ("dependency",  "missing-package",           "python-module-not-found"),
    "cache-corruption":                 ("dependency",  "cache-poisoning",           "dependency-cache-corrupted"),
    "pip-hash-mismatch":                ("dependency",  "cache-poisoning",           "pip-hash-mismatch"),

    # build / source
    "eslint-failure":                   ("source",      "lint-failure",              "eslint-lint-failure"),
    "quality-gate-failure":             ("source",      "lint-failure",              "quality-gate-failed"),
    "go-compile-error":                 ("source",      "compile-error",             "go-compile-error"),
    "gradle-build":                     ("source",      "compile-error",             "gradle-build-failure"),
    "syntax-error":                     ("source",      "compile-error",             "generic-syntax-error"),
    "typescript-compile":               ("source",      "compile-error",             "typescript-type-error"),
    "merge-conflict":                   ("source",      "merge-conflict",            "unresolved-merge-conflict"),
    "git-refspec-mismatch":             ("source",      "version-mismatch",          "git-refspec-not-found"),
    "topology-boundary-crossed":        ("source",      "ownership-violation",       "topology-boundary-crossed"),
    "topology-ownership-mismatch":      ("source",      "ownership-violation",       "topology-ownership-mismatch"),
    "topology-failure-clustered":       ("source",      "ownership-violation",       "topology-failure-clustered"),

    # build / runtime (missing executables, interpreter issues)
    "ffmpeg-avconv-missing":            ("runtime",     "missing-executable",        "ffmpeg-avconv-missing"),
    "missing-executable":               ("runtime",     "missing-executable",        "binary-not-found"),
    "node-gyp-missing-build-tools":     ("runtime",     "missing-executable",        "node-gyp-build-tools-missing"),
    "python-command-not-found":         ("runtime",     "missing-executable",        "python-command-not-found"),
    "node-version-mismatch":            ("runtime",     "interpreter-mismatch",      "node-version-mismatch"),
    "python-externally-managed":        ("runtime",     "interpreter-mismatch",      "python-externally-managed"),
    "python-virtualenv-not-activated":  ("runtime",     "interpreter-mismatch",      "python-venv-not-activated"),
    "runtime-mismatch":                 ("runtime",     "interpreter-mismatch",      "runtime-version-mismatch"),
    "node-out-of-memory":               ("runtime",     "oom-killed",                "node-heap-oom"),
    "gradle-daemon-timeout":            ("runtime",     "resource-exhaustion",       "gradle-daemon-timeout"),

    # build / filesystem
    "build-input-file-missing":         ("filesystem",  "missing-file",              "build-input-file-missing"),
    "dockerfile-copy-source-missing":   ("filesystem",  "missing-file",              "docker-copy-source-missing"),
    "npm-enoent-package-json":          ("filesystem",  "missing-file",              "npm-enoent-package-json"),
    "npm-eacces-permission-denied":     ("filesystem",  "permission-denied",         "npm-eacces-permission-denied"),
    "path-case-mismatch":               ("filesystem",  "wrong-working-directory",   "path-case-mismatch"),
    "working-directory":                ("filesystem",  "wrong-working-directory",   "wrong-working-directory"),

    # build / container
    "buildkit-session-lost":            ("container",   "build-session-error",       "buildkit-session-lost"),
    "docker-build-context":             ("container",   "build-context-error",       "docker-build-context-invalid"),
    "docker-manifest-not-found":        ("container",   "image-not-found",           "docker-manifest-not-found"),

    # build / test-runner
    "tox-invocation-error":             ("test-runner", "test-runner-error",         "tox-invocation-error"),

    # ci
    "artifact-upload-failure":          ("ci-config",   "artifact-failure",          "artifact-upload-failed"),
    "azure-pipelines-service-connection": ("auth",      "credential-expired",        "azure-service-connection-failed"),
    "azure-pipelines-task-not-found":   ("ci-config",   "task-not-found",            "azure-task-missing"),
    "circleci-config-validation":       ("ci-config",   "workflow-syntax",           "circleci-config-invalid"),
    "circleci-resource-class-invalid":  ("ci-config",   "resource-overallocation",   "circleci-resource-class-invalid"),
    "circleci-resource-class-oom":      ("ci-config",   "resource-overallocation",   "circleci-oom-resource-class"),
    "github-actions-matrix-axis-invalid": ("ci-config", "workflow-syntax",           "github-actions-matrix-axis-undefined"),
    "github-actions-missing-checkout":  ("ci-config",   "workflow-syntax",           "github-actions-missing-checkout"),
    "github-actions-permission":        ("auth",        "insufficient-scope",        "github-actions-token-permission"),
    "github-actions-runner-capacity":   ("ci-config",   "resource-overallocation",   "github-actions-runner-queue-full"),
    "github-actions-syntax":            ("ci-config",   "workflow-syntax",           "github-actions-workflow-syntax"),
    "gitlab-ci-artifact-expired":       ("ci-config",   "artifact-failure",          "gitlab-artifact-expired"),
    "gitlab-ci-yaml-invalid":           ("ci-config",   "workflow-syntax",           "gitlab-ci-config-invalid"),
    "gitlab-job-log-limit":             ("ci-config",   "resource-overallocation",   "gitlab-log-size-exceeded"),
    "gitlab-no-runner":                 ("ci-config",   "resource-overallocation",   "gitlab-no-matching-runner"),
    "git-shallow-checkout":             ("ci-config",   "workflow-syntax",           "git-shallow-clone-missing-history"),
    "jenkins-agent-offline":            ("platform",    "scheduler-error",           "jenkins-agent-offline"),
    "jenkins-plugin-missing":           ("ci-config",   "task-not-found",            "jenkins-plugin-missing"),
    "pipeline-timeout":                 ("ci-config",   "resource-overallocation",   "pipeline-time-limit-exceeded"),
    "runner-disk-full":                 ("runtime",     "disk-full",                 "runner-disk-full"),
    "runner-update-permission-denied":  ("ci-config",   "permission-denied",         "runner-update-permission-denied"),
    "secrets-not-available":            ("auth",        "missing-credential",        "ci-secret-not-available"),

    # deploy
    "config-mismatch":                  ("platform",    "config-error",              "deployment-config-mismatch"),
    "container-crash":                  ("container",   "container-crash",           "container-exit-unexpected"),
    "health-check-failure":             ("platform",    "health-check-failure",      "service-health-check-failed"),
    "image-pull-backoff":               ("container",   "image-not-found",           "image-pull-backoff"),
    "k8s-crashloopbackoff":             ("platform",    "crash-loop",                "k8s-crashloopbackoff"),
    "port-conflict":                    ("runtime",     "port-in-use",               "deploy-port-already-in-use"),
    "postgres-connection-refused":      ("database",    "service-not-ready",         "postgres-connection-refused"),
    "terraform-init":                   ("ci-config",   "config-error",              "terraform-init-failed"),
    "terraform-state-lock":             ("platform",    "concurrency-conflict",      "terraform-state-lock"),

    # network
    "connection-refused":               ("network",     "connection-refused",        "generic-connection-refused"),
    "dns-enotfound":                    ("network",     "dns-failure",               "dns-enotfound"),
    "dns-resolution":                   ("network",     "dns-failure",               "dns-resolution-failed"),
    "firewall-egress-blocked":          ("network",     "egress-blocked",            "firewall-egress-blocked"),
    "network-timeout":                  ("network",     "timeout",                   "network-request-timeout"),
    "rate-limited":                     ("network",     "rate-limiting",             "request-rate-limited"),
    "ssl-cert-error":                   ("network",     "tls-validation",            "ssl-certificate-error"),

    # runtime
    "arch-mismatch":                    ("runtime",     "arch-mismatch",             "cpu-architecture-mismatch"),
    "disk-full":                        ("runtime",     "disk-full",                 "disk-space-exhausted"),
    "docker-daemon-config-conflict":    ("container",   "daemon-error",              "docker-daemon-config-conflict"),
    "docker-daemon-unavailable":        ("container",   "daemon-error",              "docker-daemon-not-running"),
    "docker-permission-denied-nonroot": ("container",   "permission-denied",         "docker-socket-permission-denied"),
    "env-var-missing":                  ("runtime",     "env-not-persisted",         "env-var-missing"),
    "oom-killed":                       ("runtime",     "oom-killed",                "process-oom-killed"),
    "permission-denied":                ("runtime",     "permission-denied",         "filesystem-permission-denied"),
    "port-in-use":                      ("runtime",     "port-in-use",               "port-already-bound"),
    "resource-limits":                  ("runtime",     "resource-exhaustion",       "resource-limit-exceeded"),
    "segfault":                         ("runtime",     "crash",                     "process-segfault"),

    # test
    "asciidoctor-jbehave-test-failure": ("test-runner", "test-framework-failure",    "jbehave-step-failure"),
    "coverage-gate-failure":            ("test-runner", "coverage-gate",             "coverage-gate-failed"),
    "cucumber-step-failure":            ("test-runner", "scenario-failure",          "cucumber-step-failed"),
    "database-migration-timeout":       ("database",    "migration-timeout",         "database-migration-timeout"),
    "database-test-isolation":          ("database",    "isolation-failure",         "database-state-pollution"),
    "flaky-test":                       ("test-runner", "flaky-test",                "non-deterministic-test-failure"),
    "go-data-race":                     ("test-runner", "concurrency-error",         "go-data-race-detected"),
    "go-test-failure":                  ("test-runner", "test-failure",              "go-test-reported-failures"),
    "jest-command-not-found":           ("runtime",     "missing-executable",        "jest-command-not-found"),
    "jest-worker-crash":                ("test-runner", "test-runner-crash",         "jest-worker-crashed"),
    "mdanalysis-test-failure":          ("test-runner", "test-framework-failure",    "mdanalysis-test-failure"),
    "missing-test-fixture":             ("test-runner", "test-setup-failure",        "missing-test-fixture"),
    "nupic-test-failure":               ("test-runner", "test-framework-failure",    "nupic-test-failure"),
    "openai-gym-test-failure":          ("test-runner", "test-framework-failure",    "openai-gym-test-failure"),
    "order-dependency":                 ("test-runner", "flaky-test",                "test-order-dependency"),
    "parallelism-conflict":             ("test-runner", "concurrency-error",         "test-parallelism-conflict"),
    "pytest-fixture-error":             ("test-runner", "test-setup-failure",        "pytest-fixture-error"),
    "pytest-no-tests":                  ("test-runner", "test-setup-failure",        "pytest-no-tests-collected"),
    "python-parser-test-failure":       ("test-runner", "test-framework-failure",    "python-parser-test-failure"),
    "readthedocs-build-test-failure":   ("test-runner", "test-framework-failure",    "readthedocs-build-test-failure"),
    "snapshot-mismatch":                ("test-runner", "test-failure",              "snapshot-golden-file-mismatch"),
    "testcontainer-startup":            ("container",   "container-startup-failure", "testcontainer-failed-to-start"),
    "test-timeout":                     ("test-runner", "timeout",                   "test-suite-timeout"),
    "youtube-dl-test-failure":          ("test-runner", "test-framework-failure",    "youtube-dl-test-failure"),
}

PLAYBOOK_DIR = "playbooks/bundled/log"


def process_file(path: str, playbook_id: str, domain: str, cls: str, mode: str) -> bool:
    with open(path, "r") as f:
        content = f.read()

    # Skip files that already have domain:
    if re.search(r"^domain:", content, re.MULTILINE):
        print(f"  SKIP (already has domain): {path}")
        return False

    # Find stage_hints line and insert after it
    lines = content.splitlines(keepends=True)
    insert_idx = None
    for i, line in enumerate(lines):
        if re.match(r"^stage_hints:", line):
            insert_idx = i + 1
            break

    if insert_idx is None:
        print(f"  WARN (no stage_hints): {path}", file=sys.stderr)
        return False

    ontology_block = (
        f"domain: {domain}\n"
        f"class: {cls}\n"
        f"mode: {mode}\n"
    )

    lines.insert(insert_idx, ontology_block)
    new_content = "".join(lines)

    with open(path, "w") as f:
        f.write(new_content)

    print(f"  UPDATED: {path}")
    return True


def main():
    updated = 0
    skipped = 0
    missing = 0

    for dirpath, _, filenames in os.walk(PLAYBOOK_DIR):
        for filename in sorted(filenames):
            if not filename.endswith(".yaml"):
                continue
            path = os.path.join(dirpath, filename)
            playbook_id = filename[:-5]  # strip .yaml

            if playbook_id not in ONTOLOGY:
                print(f"  NO MAPPING: {path}", file=sys.stderr)
                missing += 1
                continue

            domain, cls, mode = ONTOLOGY[playbook_id]
            result = process_file(path, playbook_id, domain, cls, mode)
            if result:
                updated += 1
            else:
                skipped += 1

    print(f"\nDone: {updated} updated, {skipped} skipped (already have ontology), {missing} no mapping")


if __name__ == "__main__":
    main()
