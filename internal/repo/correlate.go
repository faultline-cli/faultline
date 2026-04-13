package repo

import (
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/model"
)

const (
	maxRecentFiles    = 10
	maxRelatedCommits = 5
	maxHotspotDirs    = 5
	maxCoChangeHints  = 3
	maxSignalHints    = 5
)

type correlationSpec struct {
	Patterns []string
}

// Correlate builds a model.RepoContext by mapping a failure category and playbook ID
// to the most relevant commits and file areas from history and signals.
func Correlate(root, category, playbookID string, commits []Commit, sigs Signals) model.RepoContext {
	spec := failureSpec(category, playbookID)
	ctx := model.RepoContext{RepoRoot: root}

	relevantFiles := rankRelevantFiles(spec, commits, sigs)
	ctx.RecentFiles = limitStrings(relevantFiles, maxRecentFiles)
	ctx.RelatedCommits = relatedCommits(spec, commits)
	ctx.HotspotDirectories = hotspotDirectories(spec, commits, sigs)
	ctx.CoChangeHints = coChangeHints(spec, sigs)
	ctx.HotfixSignals = hotfixSignals(spec, commits, sigs)
	ctx.DriftSignals = driftSignals(spec, commits, sigs)

	return ctx
}

func rankRelevantFiles(spec correlationSpec, commits []Commit, sigs Signals) []string {
	seen := make(map[string]struct{})
	files := make([]string, 0, maxRecentFiles)

	for _, commit := range commits {
		matches := matchedFiles(spec, commit.Files)
		sort.Strings(matches)
		for _, file := range matches {
			if _, ok := seen[file]; ok {
				continue
			}
			seen[file] = struct{}{}
			files = append(files, file)
		}
		if len(files) >= maxRecentFiles {
			break
		}
	}

	if len(files) > 0 {
		return files
	}

	// If no direct pattern matches exist, fall back to the highest-churn files.
	fallback := make([]string, 0, maxRecentFiles)
	for _, hotspot := range sigs.HotspotFiles {
		if _, ok := seen[hotspot.File]; ok {
			continue
		}
		seen[hotspot.File] = struct{}{}
		fallback = append(fallback, hotspot.File)
		if len(fallback) >= maxRecentFiles {
			break
		}
	}
	return fallback
}

func relatedCommits(spec correlationSpec, commits []Commit) []model.RepoCommit {
	results := make([]model.RepoCommit, 0, maxRelatedCommits)
	for _, commit := range commits {
		if len(matchedFiles(spec, commit.Files)) == 0 {
			continue
		}
		hash := commit.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		results = append(results, model.RepoCommit{
			Hash:    hash,
			Subject: commit.Subject,
			Date:    commit.Time.Format("2006-01-02"),
		})
		if len(results) >= maxRelatedCommits {
			break
		}
	}
	return results
}

func hotspotDirectories(spec correlationSpec, commits []Commit, sigs Signals) []string {
	relevantDirs := make(map[string]int)
	for _, commit := range commits {
		for _, file := range matchedFiles(spec, commit.Files) {
			dir := filepath.ToSlash(filepath.Dir(file))
			if dir == "." || dir == "" {
				continue
			}
			relevantDirs[dir]++
		}
	}

	if len(relevantDirs) == 0 {
		out := make([]string, 0, maxHotspotDirs)
		for _, dir := range sigs.HotspotDirs {
			if dir.Dir == "" {
				continue
			}
			out = append(out, dir.Dir)
			if len(out) >= maxHotspotDirs {
				break
			}
		}
		return out
	}

	dirs := make([]DirChurn, 0, len(relevantDirs))
	for dir, count := range relevantDirs {
		dirs = append(dirs, DirChurn{Dir: dir, Count: count})
	}
	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i].Count != dirs[j].Count {
			return dirs[i].Count > dirs[j].Count
		}
		return dirs[i].Dir < dirs[j].Dir
	})

	out := make([]string, 0, maxHotspotDirs)
	for _, dir := range dirs {
		out = append(out, dir.Dir)
		if len(out) >= maxHotspotDirs {
			break
		}
	}
	return out
}

func coChangeHints(spec correlationSpec, sigs Signals) []string {
	hints := make([]string, 0, maxCoChangeHints)
	for _, pair := range sigs.CoChangePairs {
		if len(matchedFiles(spec, []string{pair.FileA, pair.FileB})) == 0 {
			continue
		}
		hints = append(hints, pair.FileA+" <-> "+pair.FileB)
		if len(hints) >= maxCoChangeHints {
			break
		}
	}
	return hints
}

func hotfixSignals(spec correlationSpec, commits []Commit, sigs Signals) []string {
	var hints []string
	for _, commit := range sigs.HotfixCommits {
		if len(matchedFiles(spec, commit.Files)) == 0 {
			continue
		}
		hints = append(hints, commit.Subject)
		if len(hints) >= maxSignalHints {
			return hints
		}
	}

	if len(hints) > 0 {
		return hints
	}

	for _, commit := range sigs.HotfixCommits {
		hints = append(hints, commit.Subject)
		if len(hints) >= maxSignalHints {
			break
		}
	}
	return hints
}

func driftSignals(spec correlationSpec, commits []Commit, sigs Signals) []string {
	hints := make([]string, 0, maxSignalHints)

	for _, commit := range sigs.RevertCommits {
		if len(matchedFiles(spec, commit.Files)) == 0 {
			continue
		}
		hints = append(hints, commit.Subject)
		if len(hints) >= maxSignalHints {
			return hints
		}
	}

	for _, repeated := range sigs.RepeatedDirs {
		if repeated.Dir == "" || !dirLooksRelevant(spec, repeated.Dir, commits) {
			continue
		}
		hints = append(hints, "Repeated edits in "+repeated.Dir)
		if len(hints) >= maxSignalHints {
			return hints
		}
	}

	for _, repeated := range sigs.RepeatedFiles {
		if !matchesAny(repeated.File, spec.Patterns) {
			continue
		}
		hints = append(hints, "Repeated edits in "+repeated.File)
		if len(hints) >= maxSignalHints {
			return hints
		}
	}

	if len(hints) > 0 {
		return hints
	}

	for _, commit := range sigs.RevertCommits {
		hints = append(hints, commit.Subject)
		if len(hints) >= maxSignalHints {
			break
		}
	}
	return hints
}

func dirLooksRelevant(spec correlationSpec, dir string, commits []Commit) bool {
	prefix := dir + "/"
	for _, commit := range commits {
		for _, file := range commit.Files {
			if strings.HasPrefix(file, prefix) && matchesAny(file, spec.Patterns) {
				return true
			}
		}
	}
	return false
}

func matchedFiles(spec correlationSpec, files []string) []string {
	matches := make([]string, 0, len(files))
	for _, file := range files {
		if matchesAny(file, spec.Patterns) {
			matches = append(matches, file)
		}
	}
	return matches
}

func limitStrings(items []string, max int) []string {
	if max <= 0 || len(items) <= max {
		return items
	}
	return items[:max]
}

func failureSpec(category, playbookID string) correlationSpec {
	return correlationSpec{Patterns: failurePatterns(category, playbookID)}
}

// failurePatterns returns glob-style path patterns likely associated with the
// given failure category and playbook ID.
func failurePatterns(category, playbookID string) []string {
	category = strings.ToLower(strings.TrimSpace(category))
	playbookID = strings.ToLower(strings.TrimSpace(playbookID))

	commonCI := []string{
		".github*", ".gitlab-ci*", ".circleci*", "ci*", "Makefile", "*.sh",
	}

	switch playbookID {
	case "docker-auth":
		return append(commonCI, "Dockerfile", ".docker*", ".env*", "deploy*", "scripts/*deploy*")
	case "git-auth":
		return append(commonCI, ".gitconfig", ".git-credentials", ".netrc")
	case "kubectl-auth":
		return append(commonCI, "k8s*", "kubernetes*", "helm*", "*.yaml", "*.yml", "deploy*")
	case "oidc-token-failure":
		return append(commonCI, ".env*", "deploy*", "*.yaml", "*.yml", "infra*", "terraform*")
	case "vault-secret-missing":
		return append(commonCI, "vault*", ".env*", "deploy*", "*.yaml", "*.yml", "infra*")
	case "missing-env", "config-mismatch", "secrets-not-available":
		return append(commonCI, ".env*", "*.env", "*.json", "*.toml", "*.yaml", "*.yml", "config*", "deploy*")
	case "runtime-mismatch":
		return append(commonCI, "Dockerfile", "docker-compose*", "package.json", "go.mod", "go.sum", "pyproject.toml", "Gemfile", "*.tool-versions", ".nvmrc")
	case "node-version-mismatch":
		return append(commonCI, ".nvmrc", ".node-version", "package.json", "Dockerfile")
	case "java-version-mismatch":
		return append(commonCI, "pom.xml", "build.gradle", "build.gradle.kts", ".java-version", ".tool-versions", "Dockerfile")
	case "gradle-build", "gradle-wrapper-missing":
		return append(commonCI, "pom.xml", "build.gradle", "build.gradle.kts", "gradlew", "gradle*", ".java-version")
	case "health-check-failure":
		return append(commonCI, "Dockerfile", "cmd/*", "internal/*", "server*", "api*", "routes*", "deploy*", "k8s*", "*.yaml", "*.yml")
	case "flaky-test", "parallelism-conflict", "order-dependency":
		return []string{"*_test.go", "testdata*", "fixtures*", "migrations*", "db*", "internal/*test*", "pkg/*test*", "*.sql"}
	case "mock-assertion-failure":
		return []string{"*_test.go", "*.test.ts", "*.spec.ts", "*.test.js", "*.spec.js", "*_test.py", "testdata*", "fixtures*"}
	case "e2e-browser-failure":
		return []string{"playwright.config.*", "cypress.config.*", "cypress*", "e2e*", "tests/e2e*", "*.spec.ts", "*.spec.js"}
	case "testcontainer-startup":
		return append(commonCI, "docker-compose*", "src/test*", "*Test.java", "*_test.go", "Dockerfile")
	case "database-migration-test":
		return []string{"migrations*", "db*", "database*", "*.sql", "alembic*", "flyway*", "Gemfile", "*.rb"}
	case "go-sum-missing":
		return []string{"go.mod", "go.sum"}
	case "npm-ci-lockfile", "yarn-lockfile", "dependency-drift":
		return []string{"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", ".npmrc", ".nvmrc"}
	case "pnpm-lockfile":
		return []string{"pnpm-lock.yaml", "package.json", ".npmrc"}
	case "ruby-bundler":
		return []string{"Gemfile", "Gemfile.lock", ".ruby-version"}
	case "python-module-missing":
		return []string{"requirements*.txt", "pyproject.toml", "setup.py", "setup.cfg", "Pipfile*", "poetry.lock"}
	case "node-out-of-memory":
		return []string{"package.json", "webpack.config.*", "next.config.*", "jest.config.*", ".babelrc", "tsconfig.json"}
	case "cmake-configure-error":
		return append(commonCI, "CMakeLists.txt", "cmake*", "CMake*", "conanfile*", "vcpkg.json")
	case "protobuf-compile":
		return append(commonCI, "*.proto", "buf.gen.yaml", "buf.work.yaml", "Makefile")
	case "terraform-state-lock":
		return []string{"*.tf", "*.tfvars", "terraform*", "infra*", "deploy*"}
	case "container-crash", "oom-killed", "resource-limits":
		return append(commonCI, "Dockerfile", "k8s*", "kubernetes*", "deploy*", "*.yaml", "*.yml")
	case "container-entrypoint-missing":
		return append(commonCI, "Dockerfile", "docker-compose*", "deploy*", "*.yaml", "*.yml")
	case "shared-library-missing":
		return append(commonCI, "Dockerfile", "CMakeLists.txt", "*.c", "*.cpp", "*.h", "Makefile")
	case "file-descriptor-limit":
		return append(commonCI, "Dockerfile", "*.yaml", "*.yml", "config*")
	case "port-conflict", "port-in-use":
		return append(commonCI, "Dockerfile", "docker-compose*", "*.yaml", "*.yml", "*.json", "*.toml")
	case "database-lock":
		return []string{"migrations*", "db*", "database*", "*.sql", "*.yaml", "*.yml", "deploy*"}
	case "ssl-cert-error":
		return append(commonCI, "*.pem", "*.crt", "*.csr", "*.key", "tls*", "cert*", "nginx*")
	case "http-auth-failure":
		return append(commonCI, ".env*", "*.yaml", "*.yml", "config*", "deploy*")
	case "webhook-delivery-failure":
		return append(commonCI, "*.yaml", "*.yml", ".github*", "deploy*", "scripts*")
	case "github-actions-syntax":
		return []string{".github/workflows/*.yml", ".github/workflows/*.yaml", ".github*"}
	case "gitlab-ci-yaml-invalid":
		return []string{".gitlab-ci.yml", ".gitlab*", "*.gitlab-ci.yml"}
	case "circleci-config-invalid":
		return []string{".circleci/config.yml", ".circleci*"}
	case "jenkins-pipeline-syntax":
		return []string{"Jenkinsfile", "vars/*.groovy", "src/**/*.groovy", "jenkins*"}
	case "azure-pipeline-task-failure":
		return []string{"azure-pipelines.yml", ".azure*", "*.azure-pipelines.yml"}
	case "bitbucket-pipeline-failure":
		return []string{"bitbucket-pipelines.yml"}
	case "cache-restore-miss", "artifact-upload-failure":
		return append(commonCI, "Makefile", "package.json", "go.sum", "Cargo.lock", "pom.xml", "build.gradle")
	case "runner-disk-full":
		return append(commonCI, "Dockerfile", ".dockerignore", "docker-compose*")
	case "k8s-configmap-missing", "k8s-insufficient-resources", "k8s-rbac-forbidden":
		return append(commonCI, "k8s*", "kubernetes*", "helm*", "kustomize*", "deploy*", "*.yaml", "*.yml")
	case "helm-upgrade-failed", "helm-chart-failure":
		return append(commonCI, "helm*", "Chart.yaml", "values*.yaml", "templates*", "deploy*")
	case "argocd-sync-failed":
		return append(commonCI, "argocd*", "manifests*", "deploy*", "*.yaml", "*.yml")
	case "image-pull-backoff":
		return append(commonCI, "Dockerfile", "deploy*", "k8s*", "helm*", "*.yaml", "*.yml")
	case "ecs-deployment-failed":
		return append(commonCI, "*.tf", "*.json", "ecs*", "task-definition*", "deploy*")
	}

	switch category {
	case "auth":
		return append(commonCI, ".env*", "deploy*", "Dockerfile", "*.yaml", "*.yml", "vault*", "secrets*")
	case "build":
		return append(commonCI, "Dockerfile", "go.mod", "go.sum", "package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "pom.xml", "build.gradle", "pyproject.toml", "Gemfile", "requirements*.txt", "CMakeLists.txt")
	case "ci":
		return append(commonCI, ".github*", ".gitlab-ci.yml", ".gitlab*", ".circleci*", "Jenkinsfile", "azure-pipelines.yml", "bitbucket-pipelines.yml", "*.yml", "*.yaml")
	case "deploy":
		return append(commonCI, "deploy*", "k8s*", "kubernetes*", "helm*", "kustomize*", "argocd*", "*.yaml", "*.yml", "Dockerfile", "*.tf")
	case "test":
		return []string{"*_test.go", "testdata*", "fixtures*", "migrations*", "*.sql", "db*", "internal/*test*", "pkg/*test*", "playwright.config.*", "cypress.config.*", "*.spec.ts", "*.spec.js"}
	case "runtime":
		return append(commonCI, "Dockerfile", "*.yaml", "*.yml", "*.json", "*.toml", "config*")
	case "network":
		return append(commonCI, "*.pem", "*.crt", "*.key", "nginx*", "tls*", "cert*", "deploy*")
	default:
		return append(commonCI, "Dockerfile", "deploy*", "*.yaml", "*.yml")
	}
}

// matchesAny returns true if filePath matches any of the glob patterns.
// Patterns are matched against the full path, the basename, and each path
// segment individually.
func matchesAny(filePath string, patterns []string) bool {
	base := filepath.Base(filePath)
	for _, pat := range patterns {
		if m, _ := filepath.Match(pat, base); m {
			return true
		}
		if m, _ := filepath.Match(pat, filePath); m {
			return true
		}
		for _, part := range strings.Split(filePath, "/") {
			if m, _ := filepath.Match(pat, part); m {
				return true
			}
		}
	}
	return false
}
