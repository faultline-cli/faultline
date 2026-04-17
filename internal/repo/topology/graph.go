package topology

import (
	"io/fs"
	"path/filepath"
	"sort"
)

// Node represents a single path element in the repository topology graph.
// Each node corresponds to a top-level directory (or the root itself) and
// carries the ownership assignments derived from CODEOWNERS.
type Node struct {
	// Path is the slash-separated path relative to the repository root.
	Path string
	// Owners is the list of CODEOWNERS handles that own this path.
	Owners []string
	// Dependencies holds paths of other nodes that this node directly imports
	// or calls. Populated by language-specific analysis; empty by default.
	Dependencies []string
}

// Graph is the full ownership topology for a repository.
type Graph struct {
	// Nodes is a stable, path-ordered slice of all discovered nodes.
	Nodes []Node
	// index provides O(1) lookup by path.
	index map[string]*Node
}

// OwnersFor returns the owners for the given file path by walking up the
// graph's node tree until an owning node is found. Returns nil if no node
// covers the path.
func (g *Graph) OwnersFor(path string) []string {
	path = filepath.ToSlash(path)
	// Try direct match, then walk up the directory hierarchy.
	for {
		if n, ok := g.index[path]; ok && len(n.Owners) > 0 {
			return n.Owners
		}
		parent := filepath.ToSlash(filepath.Dir(path))
		if parent == path || parent == "." || parent == "" {
			break
		}
		path = parent
	}
	return nil
}

// NodeForPath returns the Node for the given path, or zero value if not found.
func (g *Graph) NodeForPath(path string) (Node, bool) {
	n, ok := g.index[filepath.ToSlash(path)]
	if !ok {
		return Node{}, false
	}
	return *n, true
}

// BuildGraph constructs a topology Graph for root using the provided CODEOWNERS
// rules. It scans the top-level directory structure (depth ≤ 2) to discover
// nodes and assigns ownership via the rules.
//
// fsys is used for directory traversal and exists to allow testing without a
// real filesystem; pass os.DirFS(root) in production.
func BuildGraph(root string, rules []OwnerRule, fsys fs.FS) Graph {
	graph := Graph{
		index: map[string]*Node{},
	}

	// Add the root node.
	rootOwners := OwnersFor(rules, ".")
	rootNode := &Node{Path: ".", Owners: rootOwners}
	graph.index["."] = rootNode
	graph.Nodes = append(graph.Nodes, *rootNode)

	// Walk up to depth 2 to discover ownership boundaries.
	_ = fs.WalkDir(fsys, ".", func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		depth := pathDepth(entryPath)
		if depth == 0 {
			return nil // skip "."
		}
		if depth > 2 {
			return fs.SkipDir
		}
		slashPath := filepath.ToSlash(entryPath)
		owners := OwnersFor(rules, slashPath+"/")
		if len(owners) == 0 {
			owners = OwnersFor(rules, slashPath)
		}
		n := &Node{Path: slashPath, Owners: owners}
		graph.index[slashPath] = n
		graph.Nodes = append(graph.Nodes, *n)
		return nil
	})

	// Stable order: sort by path.
	sort.Slice(graph.Nodes, func(i, j int) bool {
		return graph.Nodes[i].Path < graph.Nodes[j].Path
	})
	// Re-index after sort.
	for i := range graph.Nodes {
		graph.index[graph.Nodes[i].Path] = &graph.Nodes[i]
	}

	_ = root // root is used via fsys; kept for documentation.
	return graph
}

// pathDepth returns the directory depth relative to the repo root.
// "." → 0, "a" → 1, "a/b" → 2.
func pathDepth(path string) int {
	if path == "" || path == "." {
		return 0
	}
	path = filepath.ToSlash(path)
	n := 1
	for _, c := range path {
		if c == '/' {
			n++
		}
	}
	return n
}
