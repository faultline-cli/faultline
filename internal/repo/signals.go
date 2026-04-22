package repo

import (
	"path/filepath"
	"sort"
	"strings"
)

// LargeCommitFileThreshold is the minimum number of changed files in a single
// commit that qualifies it as a large (blast-radius) commit.
const LargeCommitFileThreshold = 10

// maxTopAuthors is the maximum number of top-author entries retained in Signals.
const maxTopAuthors = 5

// Signals holds derived signals from commit history.
type Signals struct {
	// HotspotFiles are files edited most frequently, ranked by edit count.
	HotspotFiles []FileChurn
	// HotspotDirs are directories with the most total file edits.
	HotspotDirs []DirChurn
	// RepeatedFiles are files touched at least twice in the commit window.
	RepeatedFiles []FileChurn
	// RepeatedDirs are directories touched at least twice in the commit window.
	RepeatedDirs []DirChurn
	// HotfixCommits are commits whose subjects match hotfix patterns.
	HotfixCommits []Commit
	// RevertCommits are commits whose subjects match revert patterns.
	RevertCommits []Commit
	// CoChangePairs are pairs of files frequently changed together.
	CoChangePairs []CoChangePair
	// LargeCommits are commits that touched at least LargeCommitFileThreshold
	// files, indicating a potentially risky blast-radius change.
	LargeCommits []Commit
	// ConfigChangedFiles are dependency or config files (go.mod, Dockerfile,
	// package.json, etc.) that were touched in the commit window, ranked by
	// edit count.
	ConfigChangedFiles []FileChurn
	// CIConfigChangedFiles are CI pipeline config files (.github/workflows,
	// Makefile, etc.) touched in the commit window, ranked by edit count.
	CIConfigChangedFiles []FileChurn
	// AuthorCount is the number of distinct commit authors in the window.
	AuthorCount int
	// TopAuthors are the most prolific authors in the commit window, ranked
	// by commit count. Ties are broken alphabetically by author.
	TopAuthors []AuthorChurn
}

// FileChurn records how many times a file was touched in the commit window.
type FileChurn struct {
	File  string
	Count int
}

// DirChurn records how many file edits occurred inside a directory.
type DirChurn struct {
	Dir   string
	Count int
}

// CoChangePair records two files that appeared together in commits.
type CoChangePair struct {
	FileA string
	FileB string
	Count int
}

// AuthorChurn records how many commits an author made in the commit window.
type AuthorChurn struct {
	Author string
	Count  int
}

// DeriveSignals analyses a slice of commits and returns derived signals.
// Results are deterministic: ties are broken by file/directory path.
func DeriveSignals(commits []Commit) Signals {
	fileCounts := make(map[string]int)
	dirCounts := make(map[string]int)
	authorCounts := make(map[string]int)

	pairKey := func(a, b string) string {
		if a > b {
			a, b = b, a
		}
		return a + "\x00" + b
	}
	pairFiles := make(map[string][2]string)
	pairCounts := make(map[string]int)

	var hotfix, reverts, large []Commit

	for _, c := range commits {
		subLower := strings.ToLower(c.Subject)
		if isHotfix(subLower) {
			hotfix = append(hotfix, c)
		}
		if isRevert(subLower) {
			reverts = append(reverts, c)
		}
		if len(c.Files) >= LargeCommitFileThreshold {
			large = append(large, c)
		}
		if c.Author != "" {
			authorCounts[c.Author]++
		}

		for _, f := range c.Files {
			fileCounts[f]++
			dir := filepath.ToSlash(filepath.Dir(f))
			if dir == "." {
				dir = ""
			}
			if dir != "" {
				dirCounts[dir]++
			}
		}

		for i := 0; i < len(c.Files); i++ {
			for j := i + 1; j < len(c.Files); j++ {
				k := pairKey(c.Files[i], c.Files[j])
				pairFiles[k] = [2]string{c.Files[i], c.Files[j]}
				pairCounts[k]++
			}
		}
	}

	// Build sorted file hotspot list.
	hotspotFiles := make([]FileChurn, 0, len(fileCounts))
	for f, n := range fileCounts {
		hotspotFiles = append(hotspotFiles, FileChurn{File: f, Count: n})
	}
	sort.Slice(hotspotFiles, func(i, j int) bool {
		if hotspotFiles[i].Count != hotspotFiles[j].Count {
			return hotspotFiles[i].Count > hotspotFiles[j].Count
		}
		return hotspotFiles[i].File < hotspotFiles[j].File
	})

	// Build config and CI config file lists from the full file churn data.
	var configFiles, ciFiles []FileChurn
	for f, n := range fileCounts {
		if isConfigFile(f) {
			configFiles = append(configFiles, FileChurn{File: f, Count: n})
		}
		if isCIConfigFile(f) {
			ciFiles = append(ciFiles, FileChurn{File: f, Count: n})
		}
	}
	sortFileChurn(configFiles)
	sortFileChurn(ciFiles)

	// Build sorted directory hotspot list.
	hotspotDirs := make([]DirChurn, 0, len(dirCounts))
	repeatedDirs := make([]DirChurn, 0, len(dirCounts))
	for d, n := range dirCounts {
		entry := DirChurn{Dir: d, Count: n}
		hotspotDirs = append(hotspotDirs, entry)
		if n >= 2 {
			repeatedDirs = append(repeatedDirs, entry)
		}
	}
	sort.Slice(hotspotDirs, func(i, j int) bool {
		if hotspotDirs[i].Count != hotspotDirs[j].Count {
			return hotspotDirs[i].Count > hotspotDirs[j].Count
		}
		return hotspotDirs[i].Dir < hotspotDirs[j].Dir
	})
	sort.Slice(repeatedDirs, func(i, j int) bool {
		if repeatedDirs[i].Count != repeatedDirs[j].Count {
			return repeatedDirs[i].Count > repeatedDirs[j].Count
		}
		return repeatedDirs[i].Dir < repeatedDirs[j].Dir
	})

	repeatedFiles := make([]FileChurn, 0, len(fileCounts))
	for _, file := range hotspotFiles {
		if file.Count >= 2 {
			repeatedFiles = append(repeatedFiles, file)
		}
	}

	// Build sorted co-change list (only pairs appearing together > 1 time).
	coChangePairs := make([]CoChangePair, 0, len(pairCounts))
	for k, n := range pairCounts {
		if n < 2 {
			continue
		}
		files := pairFiles[k]
		coChangePairs = append(coChangePairs, CoChangePair{
			FileA: files[0],
			FileB: files[1],
			Count: n,
		})
	}
	sort.Slice(coChangePairs, func(i, j int) bool {
		if coChangePairs[i].Count != coChangePairs[j].Count {
			return coChangePairs[i].Count > coChangePairs[j].Count
		}
		if coChangePairs[i].FileA != coChangePairs[j].FileA {
			return coChangePairs[i].FileA < coChangePairs[j].FileA
		}
		return coChangePairs[i].FileB < coChangePairs[j].FileB
	})

	// Build sorted author churn list.
	topAuthors := make([]AuthorChurn, 0, len(authorCounts))
	for a, n := range authorCounts {
		topAuthors = append(topAuthors, AuthorChurn{Author: a, Count: n})
	}
	sort.Slice(topAuthors, func(i, j int) bool {
		if topAuthors[i].Count != topAuthors[j].Count {
			return topAuthors[i].Count > topAuthors[j].Count
		}
		return topAuthors[i].Author < topAuthors[j].Author
	})
	if len(topAuthors) > maxTopAuthors {
		topAuthors = topAuthors[:maxTopAuthors]
	}

	return Signals{
		HotspotFiles:         hotspotFiles,
		HotspotDirs:          hotspotDirs,
		RepeatedFiles:        repeatedFiles,
		RepeatedDirs:         repeatedDirs,
		HotfixCommits:        hotfix,
		RevertCommits:        reverts,
		CoChangePairs:        coChangePairs,
		LargeCommits:         large,
		ConfigChangedFiles:   configFiles,
		CIConfigChangedFiles: ciFiles,
		AuthorCount:          len(authorCounts),
		TopAuthors:           topAuthors,
	}
}

// isHotfix returns true when a commit subject looks like a hotfix commit.
func isHotfix(subLower string) bool {
	for _, kw := range []string{
		"hotfix", "hot-fix", "hot fix", "fix:", "fixup", "urgent fix",
		"quick fix", "patch:", "patch ",
	} {
		if strings.Contains(subLower, kw) {
			return true
		}
	}
	return false
}

// isRevert returns true when a commit subject looks like a revert commit.
func isRevert(subLower string) bool {
	for _, kw := range []string{
		`revert "`, "revert:", "revert ", "this reverts", "rollback", "rolled back",
	} {
		if strings.Contains(subLower, kw) {
			return true
		}
	}
	return false
}

// isConfigFile returns true when the file path matches a known dependency or
// configuration file pattern that frequently contributes to CI failures.
func isConfigFile(name string) bool {
	base := filepath.Base(name)
	// Exact-name matches.
	switch base {
	case "go.mod", "go.sum",
		"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"Dockerfile",
		"pyproject.toml", "setup.py", "setup.cfg", "Pipfile", "Pipfile.lock",
		"Gemfile", "Gemfile.lock",
		"Cargo.toml", "Cargo.lock",
		"pom.xml":
		return true
	}
	// Glob-based matches (build.gradle*, requirements*.txt, docker-compose*, .env*).
	for _, pat := range []string{
		"build.gradle*", "requirements*.txt", "docker-compose*", ".env*",
	} {
		if m, _ := filepath.Match(pat, base); m {
			return true
		}
	}
	return false
}

// isCIConfigFile returns true when the file path is a CI pipeline or build
// system configuration file.
func isCIConfigFile(name string) bool {
	name = filepath.ToSlash(name)
	base := filepath.Base(name)
	// Paths that begin with known CI directories.
	for _, prefix := range []string{".github/", ".circleci/", ".gitlab", ".buildkite/", ".semaphore/", ".tekton/"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	// Exact-name matches.
	switch base {
	case "Makefile", "Jenkinsfile", "azure-pipelines.yml", "bitbucket-pipelines.yml",
		".gitlab-ci.yml", ".drone.yml", ".travis.yml", ".woodpecker.yml",
		"appveyor.yml", "bitrise.yml", "buildspec.yml", "cloudbuild.yaml", "cloudbuild.yml":
		return true
	}
	return false
}

// sortFileChurn sorts a FileChurn slice in-place: highest count first,
// ties broken alphabetically by file path.
func sortFileChurn(items []FileChurn) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].File < items[j].File
	})
}
