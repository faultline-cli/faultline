package scoring

import "testing"

func TestIsDependencyFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{"go.mod", "go.mod", true},
		{"go.sum", "go.sum", true},
		{"package.json", "package.json", true},
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"requirements.txt", "requirements.txt", true},
		{"pyproject.toml", "pyproject.toml", true},
		{"cargo.lock", "cargo.lock", true},
		{"pom.xml", "pom.xml", true},
		{"build.gradle", "build.gradle", true},
		{"requirements-dev.txt", "requirements-dev.txt", true},
		{"gemfile.lock", "gemfile.lock", true},
		{"main.go", "main.go", false},
		{"app.py", "app.py", false},
		// path-based match
		{"deps.txt", "dev/requirements-ci.txt", true},
	}
	for _, tc := range cases {
		if got := isDependencyFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isDependencyFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsRuntimeFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{".nvmrc", ".nvmrc", true},
		{".python-version", ".python-version", true},
		{".ruby-version", ".ruby-version", true},
		{".tool-versions", ".tool-versions", true},
		{"dockerfile", "dockerfile", true},
		{"main.go", "main.go", false},
		// path-based match (lowercase)
		{"runtime.yaml", "config/runtime.yaml", true},
		{"Dockerfile.prod", "build/dockerfile.prod", true}, // lowercase path matches
		{"Dockerfile", "Dockerfile", false},                 // uppercase base, uppercase path — no match
	}
	for _, tc := range cases {
		if got := isRuntimeFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isRuntimeFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsCIFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{".gitlab-ci.yml", ".gitlab-ci.yml", true},
		{"jenkinsfile", "jenkinsfile", true},
		{"azure-pipelines.yml", "azure-pipelines.yml", true},
		{"bitbucket-pipelines.yml", "bitbucket-pipelines.yml", true},
		// prefix-based
		{"ci.yml", ".github/workflows/ci.yml", true},
		{"config.yml", ".circleci/config.yml", true},
		{"main.go", "main.go", false},
	}
	for _, tc := range cases {
		if got := isCIFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isCIFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsEnvironmentFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{".env", ".env", true},
		{".env.local", ".env.local", true},
		{"production.env", "production.env", true},
		{"secret.yaml", "secret.yaml", true},
		{"vault.json", "vault.json", true},
		{"app.toml", "app.toml", true},
		{"app.ini", "app.ini", true},
		{"app.cfg", "app.cfg", true},
		{"app.conf", "app.conf", true},
		{"config.yaml", "config.yaml", true},
		{"config.yml", "config.yml", true},
		{"config.json", "config.json", true},
		{".config", ".config", true},
		{"config.local", "config.local", true},
		// path-based
		{"settings.go", "internal/config/settings.go", true},
		{"creds.json", "deploy/secrets/creds.json", true},
		{"main.go", "main.go", false},
		{"app.go", "pkg/app.go", false},
	}
	for _, tc := range cases {
		if got := isEnvironmentFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isEnvironmentFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsDeployFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{"main.tf", "main.tf", true},
		{"vars.tfvars", "vars.tfvars", true},
		// prefix-based
		{"deploy.sh", "deploy/deploy.sh", true},
		{"chart.yaml", "helm/chart.yaml", true},
		{"app.yaml", "k8s/app.yaml", true},
		{"app.yaml", "infra/app.yaml", true},
		{"app.yaml", "argocd/app.yaml", true},
		// path-contains
		{"stack.yaml", "aws/kubernetes/stack.yaml", true},
		{"state.tf", "cloud/terraform/state.tf", true},
		{"main.go", "main.go", false},
	}
	for _, tc := range cases {
		if got := isDeployFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isDeployFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsTestFile(t *testing.T) {
	cases := []struct {
		base, file string
		want       bool
	}{
		{"main_test.go", "main_test.go", true},
		{"app.spec.ts", "app.spec.ts", true},
		{"app.test.js", "app.test.js", true},
		// path-based: the check uses strings.Contains with leading slash
		{"util.go", "pkg/tests/util.go", true},
		{"fixture.yaml", "pkg/testdata/fixture.yaml", true},
		{"case.yaml", "repo/fixtures/case.yaml", true},
		{"main.go", "main.go", false},
		{"service.go", "pkg/service.go", false},
	}
	for _, tc := range cases {
		if got := isTestFile(tc.base, tc.file); got != tc.want {
			t.Errorf("isTestFile(%q, %q) = %v, want %v", tc.base, tc.file, got, tc.want)
		}
	}
}

func TestIsSourceFile(t *testing.T) {
	cases := []struct {
		base string
		want bool
	}{
		{"main.go", true},
		{"app.js", true},
		{"server.ts", true},
		{"component.tsx", true},
		{"script.py", true},
		{"model.rb", true},
		{"Service.java", true},
		{"README.md", false},
		{"config.yaml", false},
		{"Dockerfile", false},
	}
	for _, tc := range cases {
		if got := isSourceFile(tc.base); got != tc.want {
			t.Errorf("isSourceFile(%q) = %v, want %v", tc.base, got, tc.want)
		}
	}
}
