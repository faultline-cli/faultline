package engine

import (
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/playbooks"
)

func TestAdditionalBundledPlaybookFixtures(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	tests := []struct {
		name   string
		wantID string
		log    string
	}{
		{
			name:   "aws credentials",
			wantID: "aws-credentials",
			log:    "NoCredentialProviders: no valid providers in chain\nUnable to locate credentials\n",
		},
		{
			name:   "gcp credentials",
			wantID: "gcp-credentials",
			log:    "google: could not find default credentials\nApplication Default Credentials are not available\n",
		},
		{
			name:   "kubectl auth",
			wantID: "kubectl-auth",
			log:    "error: you must be logged in\nno configuration has been provided\n",
		},
		{
			name:   "npm publish auth",
			wantID: "npm-publish-auth",
			log:    "npm ERR! 401 Unauthorized\nInvalid npm token\n",
		},
		{
			name:   "cache corruption",
			wantID: "cache-corruption",
			log:    "cache hit but install failed\ncorrupted tar archive\n",
		},
		{
			name:   "cargo build",
			wantID: "cargo-build",
			log:    "error[E0308]: mismatched types\nerror: could not compile `app`\n",
		},
		{
			name:   "compose build",
			wantID: "docker-compose-build",
			log:    "docker-compose build\nService 'app' failed to build\nCOPY failed: file not found\n",
		},
		{
			name:   "dotnet restore",
			wantID: "dotnet-restore",
			log:    "Package restore failed\nNuGet restore\nCould not resolve SDK\n",
		},
		{
			name:   "eslint failure",
			wantID: "eslint-failure",
			log:    "ESLint\nFormatting check failed\n@typescript-eslint\n",
		},
		{
			name:   "install failure",
			wantID: "install-failure",
			log:    "npm ERR! 404 Not Found\npackage not found\n",
		},
		{
			name:   "makefile error",
			wantID: "makefile-error",
			log:    "make: *** [all] Error 2\nNo rule to make target `deps`\n",
		},
		{
			name:   "maven dependency resolution",
			wantID: "maven-dependency-resolution",
			log:    "Could not resolve dependencies for project demo: could not find artifact com.example:lib:jar:1.2.3\n",
		},
		{
			name:   "path case mismatch",
			wantID: "path-case-mismatch",
			log:    "wrong case in import path\ncapitalization mismatch\n",
		},
		{
			name:   "pip install failure",
			wantID: "pip-install-failure",
			log:    "ERROR: pip's dependency resolver does not currently take into account all the packages\nERROR: Failed building wheel for somepackage\n",
		},
		{
			name:   "python module missing",
			wantID: "python-module-missing",
			log:    "ModuleNotFoundError: No module named 'requests'\n",
		},
		{
			name:   "ruby bundler",
			wantID: "ruby-bundler",
			log:    "gemfile.lock not found\nbundler: command not found\n",
		},
		{
			name:   "syntax error",
			wantID: "syntax-error",
			log:    "SyntaxError: invalid syntax\nunexpected token\n",
		},
		{
			name:   "typescript compile",
			wantID: "typescript-compile",
			log:    "error TS2322: Type 'string' is not assignable to type 'number'\n",
		},
		{
			name:   "azure pipeline task failure",
			wantID: "azure-pipeline-task-failure",
			log:    "##[error]Task failed.\nAzureDevOps\n",
		},
		{
			name:   "circleci resource class",
			wantID: "circleci-resource-class",
			log:    "resource_class\nThe build agent was unable to allocate\n",
		},
		{
			name:   "github actions concurrency",
			wantID: "github-actions-concurrency",
			log:    "This run was cancelled because a newer run started in the concurrency group\ncancel-in-progress\n",
		},
		{
			name:   "github actions permission",
			wantID: "github-actions-permission",
			log:    "HttpError: Resource not accessible by integration\nrequires the 'contents: write' permission\n",
		},
		{
			name:   "gitlab artifact expired",
			wantID: "gitlab-ci-artifact-expired",
			log:    "artifacts have expired\nartifact download failed\n",
		},
		{
			name:   "gitlab no runner",
			wantID: "gitlab-no-runner",
			log:    "This job is stuck because you don't have any active runners\nwaiting for a runner\n",
		},
		{
			name:   "jenkins build failure",
			wantID: "jenkins-build-failure",
			log:    "hudson.AbortException\nFinished: FAILURE\n",
		},
		{
			name:   "pipeline timeout",
			wantID: "pipeline-timeout",
			log:    "Job's timeout exceeded\njob timed out after 1h\n",
		},
		{
			name:   "config mismatch",
			wantID: "config-mismatch",
			log:    "configmap not found\ninvalid configuration\n",
		},
		{
			name:   "container crash",
			wantID: "container-crash",
			log:    "CrashLoopBackOff\nBack-off restarting failed container\n",
		},
		{
			name:   "database lock",
			wantID: "database-lock",
			log:    "deadlock detected\nfailed to acquire advisory lock\n",
		},
		{
			name:   "ecs deployment failed",
			wantID: "ecs-deployment-failed",
			log:    "deployment circuit breaker\nservice could not stabilize\n",
		},
		{
			name:   "health check failure",
			wantID: "health-check-failure",
			log:    "readiness probe failed\nservice is not ready\n",
		},
		{
			name:   "helm chart failure",
			wantID: "helm-chart-failure",
			log:    "Error: INSTALLATION FAILED\nhelm upgrade\n",
		},
		{
			name:   "k8s insufficient resources",
			wantID: "k8s-insufficient-resources",
			log:    "Insufficient cpu\nFailedScheduling\n",
		},
		{
			name:   "port conflict",
			wantID: "port-conflict",
			log:    "port conflict\ncannot start service\n",
		},
		{
			name:   "terraform apply error",
			wantID: "terraform-apply-error",
			log:    "Error applying plan\nError: Reference to undeclared resource\n",
		},
		{
			name:   "terraform state lock",
			wantID: "terraform-state-lock",
			log:    "error acquiring the state lock\nlock id\n",
		},
		{
			name:   "connection refused",
			wantID: "connection-refused",
			log:    "connect: connection refused\nfailed to connect\n",
		},
		{
			name:   "dns resolution",
			wantID: "dns-resolution",
			log:    "dial tcp: lookup registry.example.com: no such host\n",
		},
		{
			name:   "network timeout",
			wantID: "network-timeout",
			log:    "connection timed out\nrequest timeout\n",
		},
		{
			name:   "proxy misconfiguration",
			wantID: "proxy-misconfiguration",
			log:    "407 Proxy Authentication Required\nproxy tunnel\n",
		},
		{
			name:   "rate limited",
			wantID: "rate-limited",
			log:    "429 Too Many Requests\nrate limit exceeded\n",
		},
		{
			name:   "ssl cert error",
			wantID: "ssl-cert-error",
			log:    "certificate verify failed\ncertificate signed by unknown authority\n",
		},
		{
			name:   "disk full",
			wantID: "disk-full",
			log:    "no space left on device\nENOSPC\n",
		},
		{
			name:   "oom killed",
			wantID: "oom-killed",
			log:    "Process exited with exit code 137\nout of memory\n",
		},
		{
			name:   "permission denied",
			wantID: "permission-denied",
			log:    "permission denied\nEACCES\n",
		},
		{
			name:   "port in use",
			wantID: "port-in-use",
			log:    "address already in use\nEADDRINUSE\n",
		},
		{
			name:   "resource limits",
			wantID: "resource-limits",
			log:    "too many open files\nfile descriptor limit\n",
		},
		{
			name:   "segfault",
			wantID: "segfault",
			log:    "Segmentation fault (core dumped)\nSIGSEGV\n",
		},
		{
			name:   "coverage threshold",
			wantID: "coverage-threshold",
			log:    "Coverage threshold not met\nbelow the minimum\n",
		},
		{
			name:   "missing test fixture",
			wantID: "missing-test-fixture",
			log:    "fixture not found\ntestdata\n",
		},
		{
			name:   "order dependency",
			wantID: "order-dependency",
			log:    "depends on previous test\nglobal state\n",
		},
		{
			name:   "parallelism conflict",
			wantID: "parallelism-conflict",
			log:    "test is not parallelizable\ndatabase locked\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lines, err := readLines(strings.NewReader(tc.log))
			if err != nil {
				t.Fatalf("read lines: %v", err)
			}
			ctx := ExtractContext(lines)
			results := matcher.Rank(pbs, lines, ctx)
			if len(results) == 0 {
				t.Fatalf("expected fixture to match at least one playbook")
			}
			if got := results[0].Playbook.ID; got != tc.wantID {
				t.Fatalf("expected top match %s, got %s", tc.wantID, got)
			}
			if len(results[0].Evidence) == 0 {
				t.Fatalf("expected evidence for %s", tc.wantID)
			}

			again := matcher.Rank(pbs, lines, ctx)
			if len(again) != len(results) {
				t.Fatalf("expected deterministic result count for %s", tc.wantID)
			}
			for i := range results {
				if results[i].Playbook.ID != again[i].Playbook.ID ||
					results[i].Score != again[i].Score ||
					results[i].Confidence != again[i].Confidence {
					t.Fatalf("expected deterministic ranking for %s", tc.wantID)
				}
			}
		})
	}
}

func TestBundledPlaybookNoneExclusions(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	t.Run("network timeout excludes test timeout phrases", func(t *testing.T) {
		lines, err := readLines(strings.NewReader(
			"go test -timeout 30s\nTest timed out after 30000ms\ncontext deadline exceeded\n",
		))
		if err != nil {
			t.Fatalf("read lines: %v", err)
		}
		ctx := ExtractContext(lines)
		results := matcher.Rank(pbs, lines, ctx)
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		if got := results[0].Playbook.ID; got != "test-timeout" {
			t.Fatalf("expected test-timeout first, got %s", got)
		}
		if containsPlaybook(results, "network-timeout") {
			t.Fatal("expected network-timeout to be excluded by match.none")
		}
	})

	t.Run("oom killed excludes container crash phrases", func(t *testing.T) {
		lines, err := readLines(strings.NewReader(
			"CrashLoopBackOff\nexit code 137\nOut of memory\nBack-off restarting failed container\n",
		))
		if err != nil {
			t.Fatalf("read lines: %v", err)
		}
		ctx := ExtractContext(lines)
		results := matcher.Rank(pbs, lines, ctx)
		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
		if got := results[0].Playbook.ID; got != "container-crash" {
			t.Fatalf("expected container-crash first, got %s", got)
		}
		if containsPlaybook(results, "oom-killed") {
			t.Fatal("expected oom-killed to be excluded by match.none")
		}
	})
}
