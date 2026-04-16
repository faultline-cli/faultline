package scoring

import "strings"

func isDependencyFile(base, file string) bool {
	switch base {
	case "go.mod", "go.sum", "package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "requirements.txt", "requirements-dev.txt", "pyproject.toml", "poetry.lock", "pipfile", "pipfile.lock", "gemfile", "gemfile.lock", "cargo.lock", "pom.xml", "build.gradle", "build.gradle.kts":
		return true
	}
	return strings.Contains(file, "requirements") || strings.Contains(file, "poetry.lock")
}

func isRuntimeFile(base, file string) bool {
	switch base {
	case ".nvmrc", ".node-version", ".python-version", ".ruby-version", ".tool-versions", "dockerfile":
		return true
	}
	return strings.Contains(file, "dockerfile") || strings.Contains(file, "runtime")
}

func isCIFile(base, file string) bool {
	switch base {
	case ".gitlab-ci.yml", "jenkinsfile", "azure-pipelines.yml", "bitbucket-pipelines.yml":
		return true
	}
	return strings.HasPrefix(file, ".github/workflows/") || strings.HasPrefix(file, ".circleci/")
}

func isEnvironmentFile(base, file string) bool {
	switch {
	case strings.HasPrefix(base, ".env"), strings.HasSuffix(base, ".env"),
		strings.Contains(base, "secret"), strings.Contains(base, "vault"):
		return true
	}
	// Known config file suffixes only — avoid matching arbitrary source files that
	// happen to contain the word "config" (e.g. appconfig.go, configparser.py).
	switch {
	case strings.HasSuffix(base, ".toml"), strings.HasSuffix(base, ".ini"),
		strings.HasSuffix(base, ".cfg"), strings.HasSuffix(base, ".conf"),
		base == "config.yaml", base == "config.yml", base == "config.json",
		base == ".config", strings.HasPrefix(base, "config."):
		return true
	}
	return strings.Contains(file, "/config/") || strings.Contains(file, "/secrets/")
}

func isDeployFile(base, file string) bool {
	switch {
	case strings.HasSuffix(base, ".tf"), strings.HasSuffix(base, ".tfvars"), strings.HasPrefix(file, "deploy/"), strings.HasPrefix(file, "infra/"), strings.HasPrefix(file, "k8s/"), strings.HasPrefix(file, "helm/"), strings.HasPrefix(file, "argocd/"):
		return true
	}
	return strings.Contains(file, "kubernetes") || strings.Contains(file, "terraform")
}

func isTestFile(base, file string) bool {
	switch {
	case strings.HasSuffix(base, "_test.go"), strings.Contains(base, ".spec."), strings.Contains(base, ".test."), strings.Contains(file, "/tests/"), strings.Contains(file, "/testdata/"), strings.Contains(file, "/fixtures/"):
		return true
	}
	return false
}

func isSourceFile(base string) bool {
	switch {
	case strings.HasSuffix(base, ".go"), strings.HasSuffix(base, ".js"), strings.HasSuffix(base, ".ts"), strings.HasSuffix(base, ".tsx"), strings.HasSuffix(base, ".py"), strings.HasSuffix(base, ".rb"), strings.HasSuffix(base, ".java"):
		return true
	}
	return false
}
