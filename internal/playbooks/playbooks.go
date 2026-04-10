package playbooks

// Package playbooks loads and validates YAML failure playbooks.
// It replaces the former internal/loader package and adds support for
// recursive directory trees so playbooks can be organised into sub-directories
// by category.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"faultline/internal/model"
)

const envKey = "FAULTLINE_PLAYBOOK_DIR"

// raw is the on-disk YAML shape.  The legacy "explanation" field is accepted
// as a fallback for the canonical "explain" field so that older custom
// playbooks continue to load without modification.
type raw struct {
	ID         string   `yaml:"id"`
	Title      string   `yaml:"title"`
	Category   string   `yaml:"category"`
	Severity   string   `yaml:"severity"`
	BaseScore  float64  `yaml:"base_score"`
	Tags       []string `yaml:"tags"`
	StageHints []string `yaml:"stage_hints"`
	Match      struct {
		Any  []string `yaml:"any"`
		All  []string `yaml:"all"`
		None []string `yaml:"none"`
	} `yaml:"match"`
	Explain     string   `yaml:"explain"`
	Explanation string   `yaml:"explanation"` // legacy alias
	Why         string   `yaml:"why"`
	Fix         []string `yaml:"fix"`
	Prevent     []string `yaml:"prevent"`
	Workflow    struct {
		LikelyFiles []string `yaml:"likely_files"`
		LocalRepro  []string `yaml:"local_repro"`
		Verify      []string `yaml:"verify"`
	} `yaml:"workflow"`
}

// LoadDefault loads playbooks from the default directory resolved by
// DefaultDir.
func LoadDefault() ([]model.Playbook, error) {
	dir, err := DefaultDir()
	if err != nil {
		return nil, err
	}
	return LoadDir(dir)
}

// DefaultDir resolves the playbook directory using the following priority:
//  1. FAULTLINE_PLAYBOOK_DIR environment variable
//  2. A "playbooks" directory found by walking upward from the working
//     directory or the executable directory
//  3. /playbooks (Docker container convention)
func DefaultDir() (string, error) {
	if envDir := strings.TrimSpace(os.Getenv(envKey)); envDir != "" {
		return validateDir(envDir)
	}
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, upwardDirs(cwd)...)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, upwardDirs(filepath.Dir(exe))...)
	}
	candidates = append(candidates, "/playbooks")
	seen := make(map[string]struct{})
	for _, c := range candidates {
		if c == "" {
			continue
		}
		c = filepath.Clean(c)
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		if dir, err := validateDir(c); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf(
		"playbook directory not found; set %s or add a playbooks/ directory",
		envKey,
	)
}

// LoadDir loads all .yaml/.yml files found recursively under dir.
// Files are processed in lexical order to ensure deterministic loading.
// Duplicate playbook IDs are treated as a hard error.
func LoadDir(dir string) ([]model.Playbook, error) {
	dir, err := validateDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk playbook directory: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no playbook files found in %s", dir)
	}
	playbooks := make([]model.Playbook, 0, len(files))
	seenIDs := make(map[string]string, len(files))
	for _, path := range files {
		pb, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		if prev, ok := seenIDs[pb.ID]; ok {
			return nil, fmt.Errorf(
				"duplicate playbook id %q in %s and %s",
				pb.ID, prev, path,
			)
		}
		seenIDs[pb.ID] = path
		playbooks = append(playbooks, pb)
	}
	// Secondary sort by ID for fully deterministic ordering.
	sort.Slice(playbooks, func(i, j int) bool {
		return playbooks[i].ID < playbooks[j].ID
	})
	return playbooks, nil
}

func loadFile(path string) (model.Playbook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Playbook{}, fmt.Errorf("read playbook %s: %w", path, err)
	}
	var r raw
	if err := yaml.Unmarshal(data, &r); err != nil {
		return model.Playbook{}, fmt.Errorf("parse playbook %s: %w", path, err)
	}
	if err := validate(r, path); err != nil {
		return model.Playbook{}, err
	}
	// Accept legacy "explanation" if the canonical "explain" is absent.
	explain := r.Explain
	if explain == "" {
		explain = r.Explanation
	}
	return model.Playbook{
		ID:         r.ID,
		Title:      r.Title,
		Category:   r.Category,
		Severity:   r.Severity,
		BaseScore:  r.BaseScore,
		Tags:       r.Tags,
		StageHints: r.StageHints,
		Match: model.MatchSpec{
			Any:  r.Match.Any,
			All:  r.Match.All,
			None: r.Match.None,
		},
		Explain: explain,
		Why:     r.Why,
		Fix:     r.Fix,
		Prevent: r.Prevent,
		Workflow: model.WorkflowSpec{
			LikelyFiles: r.Workflow.LikelyFiles,
			LocalRepro:  r.Workflow.LocalRepro,
			Verify:      r.Workflow.Verify,
		},
	}, nil
}

func validate(r raw, path string) error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("playbook %s: missing required field 'id'", path)
	}
	if strings.TrimSpace(r.Title) == "" {
		return fmt.Errorf("playbook %s: missing required field 'title'", path)
	}
	if len(r.Match.Any) == 0 && len(r.Match.All) == 0 {
		return fmt.Errorf(
			"playbook %s: must define at least one pattern in match.any or match.all",
			path,
		)
	}
	return nil
}

func validateDir(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("playbook directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return dir, nil
}

// upwardDirs returns a list of "playbooks" directory candidates by walking
// upward from dir toward the filesystem root.
func upwardDirs(dir string) []string {
	var result []string
	for {
		result = append(result, filepath.Join(dir, "playbooks"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return result
}
