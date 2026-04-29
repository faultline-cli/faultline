package sourcedetector

import (
	"path/filepath"
	"sort"
	"strings"

	"faultline/internal/detectors"
)

type preparedFile struct {
	path       string
	lines      []preparedLine
	moduleKey  string
	pathClass  string
	hotPath    bool
	critical   bool
	classBonus float64
}

type preparedLine struct {
	original   string
	normalized string
	number     int
	function   string
	depth      int
}

func prepareFiles(files []detectors.SourceFile) []preparedFile {
	out := make([]preparedFile, 0, len(files))
	for _, file := range files {
		lines := make([]preparedLine, 0, len(file.Lines))
		currentFunc := ""
		funcDepth := 0
		depth := 0
		for i, line := range file.Lines {
			trimmed := strings.TrimSpace(line)
			if fn := inferFunctionName(trimmed); fn != "" {
				currentFunc = fn
				funcDepth = depth
			}
			lines = append(lines, preparedLine{
				original:   line,
				normalized: normalize(line),
				number:     i + 1,
				function:   currentFunc,
				depth:      depth,
			})
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if currentFunc != "" && depth <= funcDepth {
				currentFunc = ""
			}
		}
		pathClass, hotPath, critical, classBonus := classifyPath(file.Path)
		out = append(out, preparedFile{
			path:       file.Path,
			lines:      lines,
			moduleKey:  filepath.Dir(file.Path),
			pathClass:  pathClass,
			hotPath:    hotPath,
			critical:   critical,
			classBonus: classBonus,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].path < out[j].path
	})
	return out
}
