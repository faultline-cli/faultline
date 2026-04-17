package topology

import (
	"path/filepath"
	"sort"
	"strings"
)

const (
	// Signal names emitted by DeriveSignals.
	SignalBoundaryCrossed   = "topology.boundary.crossed"
	SignalUpstreamChanged   = "topology.upstream.changed"
	SignalOwnershipMismatch = "topology.ownership.mismatch"
	SignalFailureClustered  = "topology.failure.clustered"
)

// Signals holds the topology-derived signals for a given set of changed and
// failed files.
type Signals struct {
	// ActiveSignals is the deterministic, sorted list of signal names that
	// fired for the current analysis context.
	ActiveSignals []string

	// BoundaryCrossed is true when changed files and failure-related files
	// belong to different ownership zones.
	BoundaryCrossed bool
	// UpstreamChanged is true when a changed file's directory is a parent of
	// (or the same as) a hotspot directory containing the failure.
	UpstreamChanged bool
	// OwnershipMismatch is true when the top changed file and the top failure
	// file are owned by different teams.
	OwnershipMismatch bool
	// FailureClustered is true when all failure-related files share a common
	// directory prefix, suggesting a localised component failure.
	FailureClustered bool

	// OwnerZones is a deduplicated, sorted list of owner handles that appear
	// across the changed files.
	OwnerZones []string
}

// DeriveSignals analyses changedFiles (repo-relative paths of recently changed
// or modified files) and failureFiles (paths inferred from log evidence that
// might contain the root cause) against the topology graph and returns the
// fired signals.
//
// Results are deterministic for the same inputs.
func DeriveSignals(graph Graph, changedFiles []string, failureFiles []string) Signals {
	var sigs Signals

	changeOwners := ownersForFiles(graph, changedFiles)
	failureOwners := ownersForFiles(graph, failureFiles)

	// BoundaryCrossed: changed and failure files have non-overlapping owners.
	if len(changeOwners) > 0 && len(failureOwners) > 0 {
		if !hasIntersection(changeOwners, failureOwners) {
			sigs.BoundaryCrossed = true
		}
	}

	// OwnershipMismatch: top changed file's owner ≠ top failure file's owner.
	if len(changedFiles) > 0 && len(failureFiles) > 0 {
		topChangeOwner := graph.OwnersFor(changedFiles[0])
		topFailureOwner := graph.OwnersFor(failureFiles[0])
		if len(topChangeOwner) > 0 && len(topFailureOwner) > 0 &&
			!stringSlicesOverlap(topChangeOwner, topFailureOwner) {
			sigs.OwnershipMismatch = true
		}
	}

	// UpstreamChanged: a changed file's directory is an ancestor of a failure
	// file's directory (suggesting the change propagated downstream).
	if len(changedFiles) > 0 && len(failureFiles) > 0 {
		for _, changed := range changedFiles {
			changedDir := filepath.ToSlash(filepath.Dir(changed))
			for _, failed := range failureFiles {
				failedDir := filepath.ToSlash(filepath.Dir(failed))
				if isAncestorDir(changedDir, failedDir) {
					sigs.UpstreamChanged = true
					break
				}
			}
			if sigs.UpstreamChanged {
				break
			}
		}
	}

	// FailureClustered: all failure files share a common directory prefix.
	if len(failureFiles) >= 2 {
		if commonDirPrefix(failureFiles) != "." {
			sigs.FailureClustered = true
		}
	}

	// Collect owner zones from changed files.
	ownerSet := map[string]struct{}{}
	for owner := range changeOwners {
		ownerSet[owner] = struct{}{}
	}
	for owner := range ownerSet {
		sigs.OwnerZones = append(sigs.OwnerZones, owner)
	}
	sort.Strings(sigs.OwnerZones)

	// Build sorted active signal list.
	if sigs.BoundaryCrossed {
		sigs.ActiveSignals = append(sigs.ActiveSignals, SignalBoundaryCrossed)
	}
	if sigs.FailureClustered {
		sigs.ActiveSignals = append(sigs.ActiveSignals, SignalFailureClustered)
	}
	if sigs.OwnershipMismatch {
		sigs.ActiveSignals = append(sigs.ActiveSignals, SignalOwnershipMismatch)
	}
	if sigs.UpstreamChanged {
		sigs.ActiveSignals = append(sigs.ActiveSignals, SignalUpstreamChanged)
	}
	sort.Strings(sigs.ActiveSignals)

	return sigs
}

// ownersForFiles returns a set of all owner handles that cover the given
// file paths according to the graph.
func ownersForFiles(graph Graph, files []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, f := range files {
		for _, owner := range graph.OwnersFor(f) {
			set[owner] = struct{}{}
		}
	}
	return set
}

// hasIntersection returns true when the two owner sets share at least one
// element.
func hasIntersection(a, b map[string]struct{}) bool {
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

// stringSlicesOverlap returns true when the two slices share at least one
// element (case-sensitive).
func stringSlicesOverlap(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

// isAncestorDir returns true when ancestor is a directory prefix of descendant.
// Both values should use forward slashes.
func isAncestorDir(ancestor, descendant string) bool {
	if ancestor == "." || ancestor == "" {
		return true
	}
	return strings.HasPrefix(descendant, ancestor+"/") || descendant == ancestor
}

// commonDirPrefix returns the longest common directory prefix shared by all
// file paths. Returns "." when there is no meaningful common prefix.
func commonDirPrefix(files []string) string {
	if len(files) == 0 {
		return "."
	}
	dirs := make([]string, len(files))
	for i, f := range files {
		dirs[i] = filepath.ToSlash(filepath.Dir(f))
	}
	prefix := dirs[0]
	for _, d := range dirs[1:] {
		prefix = commonPrefix(prefix, d)
		if prefix == "." || prefix == "" {
			return "."
		}
	}
	return prefix
}

// commonPrefix returns the common directory prefix of two slash-separated paths.
func commonPrefix(a, b string) string {
	partsA := strings.Split(a, "/")
	partsB := strings.Split(b, "/")
	n := len(partsA)
	if len(partsB) < n {
		n = len(partsB)
	}
	common := []string{}
	for i := 0; i < n; i++ {
		if partsA[i] != partsB[i] {
			break
		}
		common = append(common, partsA[i])
	}
	if len(common) == 0 {
		return "."
	}
	return strings.Join(common, "/")
}
