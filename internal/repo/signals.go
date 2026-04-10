package repo

import (
	"path/filepath"
	"sort"
	"strings"
)

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

// DeriveSignals analyses a slice of commits and returns derived signals.
// Results are deterministic: ties are broken by file/directory path.
func DeriveSignals(commits []Commit) Signals {
	fileCounts := make(map[string]int)
	dirCounts := make(map[string]int)

	pairKey := func(a, b string) string {
		if a > b {
			a, b = b, a
		}
		return a + "\x00" + b
	}
	pairFiles := make(map[string][2]string)
	pairCounts := make(map[string]int)

	var hotfix, reverts []Commit

	for _, c := range commits {
		subLower := strings.ToLower(c.Subject)
		if isHotfix(subLower) {
			hotfix = append(hotfix, c)
		}
		if isRevert(subLower) {
			reverts = append(reverts, c)
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

	return Signals{
		HotspotFiles:  hotspotFiles,
		HotspotDirs:   hotspotDirs,
		RepeatedFiles: repeatedFiles,
		RepeatedDirs:  repeatedDirs,
		HotfixCommits: hotfix,
		RevertCommits: reverts,
		CoChangePairs: coChangePairs,
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
