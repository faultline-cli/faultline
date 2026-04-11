package engine

import (
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/playbooks"
)

// TestRemainingBundledPlaybookFixtures covers the 36 bundled playbooks that are
// not yet exercised by TestBundledPlaybookFixtures or
// TestAdditionalBundledPlaybookFixtures. Each case supplies a minimal inline
// log that should rank the target playbook first.
func TestRemainingBundledPlaybookFixtures(t *testing.T) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(t))
	if err != nil {
		t.Fatalf("load playbooks: %v", err)
	}

	tests := []struct {
		name   string
		wantID string
		log    string
	}{
		// ── auth ─────────────────────────────────────────────────────────────
		{
			name:   "oidc token failure",
			wantID: "oidc-token-failure",
			log:    "Unable to get OIDC token\nOIDC token request failed\ntoken audience mismatch\ninvalid_grant\n",
		},
		{
			name:   "vault secret missing",
			wantID: "vault-secret-missing",
			log:    "VAULT_TOKEN\nVAULT_ADDR\nsecret does not exist\nSecret retrieval failed\n",
		},

		// ── build ─────────────────────────────────────────────────────────────
		{
			name:   "cmake configure error",
			wantID: "cmake-configure-error",
			log:    "CMake Error at CMakeLists.txt\nNo CMAKE_CXX_COMPILER could be found\nCould NOT find\nCMake Error\n",
		},
		{
			name:   "docker build context",
			wantID: "docker-build-context",
			log:    "failed to read dockerfile\nfailed to solve with frontend dockerfile.v0\nunable to prepare context\n",
		},
		{
			name:   "go sum missing",
			wantID: "go-sum-missing",
			log:    "missing go.sum entry for module providing package\ngo.sum file is not up to date\n",
		},
		{
			name:   "gradle build",
			wantID: "gradle-build",
			log:    "FAILURE: Build failed with an exception\nWhat went wrong:\nExecution failed for task ':compileJava'\n",
		},
		{
			name:   "gradle wrapper missing",
			wantID: "gradle-wrapper-missing",
			log:    "gradlew: command not found\nGradle wrapper not found\n./gradlew: No such file or directory\n",
		},
		{
			name:   "java version mismatch",
			wantID: "java-version-mismatch",
			log:    "UnsupportedClassVersionError\nhas been compiled by a more recent version of the Java Runtime\nerror: invalid source release\n",
		},
		{
			name:   "node out of memory",
			wantID: "node-out-of-memory",
			log:    "JavaScript heap out of memory\nFATAL ERROR: Reached heap limit Allocation failed\nIneffective mark-compacts near heap limit\n",
		},
		{
			name:   "node version mismatch",
			wantID: "node-version-mismatch",
			log:    "The engine \"node\" is incompatible with this module\nUnsupported engine\nRequired engine node\n",
		},
		{
			name:   "pnpm lockfile",
			wantID: "pnpm-lockfile",
			log:    "ERR_PNPM_OUTDATED_LOCKFILE\nERR_PNPM_FROZEN_LOCKFILE\npnpm-lock.yaml is not up to date\n",
		},
		{
			name:   "protobuf compile",
			wantID: "protobuf-compile",
			log:    "protoc: command not found\nprotoc-gen-go: program not found\nFailed to parse protobuf\n",
		},
		{
			name:   "python lint",
			wantID: "python-lint",
			log:    "flake8\nruff\nwould reformat\nImports are incorrectly sorted\n",
		},

		// ── ci ───────────────────────────────────────────────────────────────
		{
			name:   "artifact upload failure",
			wantID: "artifact-upload-failure",
			log:    "artifact upload failed\nfailed to upload artifact\nError uploading\n",
		},
		{
			name:   "bitbucket pipeline failure",
			wantID: "bitbucket-pipeline-failure",
			log:    "bitbucket-pipelines.yml\nBitbucket Pipelines\nBuild image failed\n",
		},
		{
			name:   "cache restore miss",
			wantID: "cache-restore-miss",
			log:    "Cache not found\ncache miss\nNo cache found\ncache restore failed\n",
		},
		{
			name:   "circleci config invalid",
			wantID: "circleci-config-invalid",
			log:    "ERROR IN CONFIG FILE\nError: .circleci/config.yml\nConfig processing error\n",
		},
		{
			name:   "github actions syntax",
			wantID: "github-actions-syntax",
			log:    "Invalid workflow file\nYou have an error in your yaml syntax\n",
		},
		{
			name:   "gitlab ci yaml invalid",
			wantID: "gitlab-ci-yaml-invalid",
			log:    ".gitlab-ci.yml\nERROR: Config file is not valid\nERROR: Job is invalid\n",
		},
		{
			name:   "jenkins pipeline syntax",
			wantID: "jenkins-pipeline-syntax",
			log:    "groovy.lang.MissingPropertyException\norg.codehaus.groovy\nJenkinsfile\n",
		},
		{
			name:   "runner disk full",
			wantID: "runner-disk-full",
			log:    "disk quota exceeded\nWrite failed: No space left\nnot enough disk space\n",
		},
		{
			name:   "secrets not available",
			wantID: "secrets-not-available",
			log:    "Context access might be invalid: secrets\nNo secret named\n",
		},

		// ── deploy ───────────────────────────────────────────────────────────
		{
			name:   "argocd sync failed",
			wantID: "argocd-sync-failed",
			log:    "SyncFailed\nfailed to sync app\nerror syncing application\nOutOfSync\n",
		},
		{
			name:   "helm upgrade failed",
			wantID: "helm-upgrade-failed",
			log:    "UPGRADE FAILED\nRollback was triggered following\nError: UPGRADE FAILED\nhas been rolled back\n",
		},
		{
			name:   "k8s configmap missing",
			wantID: "k8s-configmap-missing",
			log:    "Error from server (NotFound): configmaps\nError from server (NotFound): secrets\n",
		},
		{
			name:   "k8s rbac forbidden",
			wantID: "k8s-rbac-forbidden",
			log:    "RBAC: access denied\nUser cannot\nerror from server (Forbidden)\ndoes not have get access\n",
		},
		{
			name:   "terraform init",
			wantID: "terraform-init",
			log:    "Error: Failed to install provider\nFailed to query available provider packages\nError: Could not load plugin\n",
		},

		// ── network ──────────────────────────────────────────────────────────
		{
			name:   "http auth failure",
			wantID: "http-auth-failure",
			log:    "HTTP 401\nstatus 401\nHTTP 403\nstatus 403\n401 status code\n403 status code\n",
		},
		{
			name:   "webhook delivery failure",
			wantID: "webhook-delivery-failure",
			log:    "Failed to deliver webhook\nwebhook delivery failed\nsignature verification failed\n",
		},

		// ── runtime ──────────────────────────────────────────────────────────
		{
			name:   "container entrypoint missing",
			wantID: "container-entrypoint-missing",
			log:    "exec format error\nOCI runtime exec failed\nexecutable file not found in $PATH\n",
		},
		{
			name:   "file descriptor limit",
			wantID: "file-descriptor-limit",
			log:    "open: too many open files\nOSError: [Errno 24]\nsocket: too many open files\n",
		},
		{
			name:   "shared library missing",
			wantID: "shared-library-missing",
			log:    "error while loading shared libraries\ncannot open shared object file\nsymbol lookup error\n",
		},

		// ── test ─────────────────────────────────────────────────────────────
		{
			name:   "database migration test",
			wantID: "database-migration-test",
			log:    "ActiveRecord::PendingMigrationError\nMigration failed\nrake db:migrate\n",
		},
		{
			name:   "e2e browser failure",
			wantID: "e2e-browser-failure",
			log:    "browserType.launch\nFailed to launch the browser process\nCypress could not start\n",
		},
		{
			name:   "mock assertion failure",
			wantID: "mock-assertion-failure",
			log:    "WantedButNotInvoked\nExpected call was not received\nMockException\n",
		},
		{
			name:   "testcontainer startup",
			wantID: "testcontainer-startup",
			log:    "Could not find a valid Docker environment\norg.testcontainers\ntestcontainers\ndocker.sock\n",
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
				t.Fatalf("expected fixture to match at least one playbook, got none")
			}
			if got := results[0].Playbook.ID; got != tc.wantID {
				t.Fatalf("expected top match %s, got %s (all results: %v)", tc.wantID, got, resultIDs(results))
			}
			if len(results[0].Evidence) == 0 {
				t.Fatalf("expected evidence for %s", tc.wantID)
			}
		})
	}
}
