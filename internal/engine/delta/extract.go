package delta

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"faultline/internal/model"
)

var (
	goFailPattern     = regexp.MustCompile(`(?i)^--- FAIL: ([^\s(]+)`)
	pytestFailPattern = regexp.MustCompile(`(?i)^FAILED\s+([^\s]+)`)
	jestFailPattern   = regexp.MustCompile(`^\s*[●\*]\s+(.+)$`)
)

func buildSnapshot(provider string, currentLog string, baselineLog string, changedFiles []string, envDiff map[string]model.DeltaEnvChange) Snapshot {
	currentTests := failingTestsFromLog(currentLog)
	baselineTests := failingTestsFromLog(baselineLog)
	currentErrors := salientErrors(currentLog)
	baselineErrors := salientErrors(baselineLog)
	return Snapshot{
		Provider:          provider,
		FilesChanged:      dedupeStrings(changedFiles),
		TestsNewlyFailing: subtractStrings(currentTests, baselineTests),
		ErrorsAdded:       subtractStrings(currentErrors, baselineErrors),
		EnvDiff:           cloneEnvDiff(envDiff),
	}
}

func unzipLogs(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	type entry struct {
		name string
		body string
	}
	var entries []entry
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		entries = append(entries, entry{
			name: filepath.ToSlash(file.Name),
			body: strings.ReplaceAll(string(body), "\r\n", "\n"),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})
	var b strings.Builder
	for i, item := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(item.body)
		if !strings.HasSuffix(item.body, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String(), nil
}

func failingTestsFromLog(log string) []string {
	log = strings.ReplaceAll(log, "\r\n", "\n")
	var tests []string
	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if matches := goFailPattern.FindStringSubmatch(line); len(matches) == 2 {
			tests = append(tests, matches[1])
			continue
		}
		if matches := pytestFailPattern.FindStringSubmatch(line); len(matches) == 2 {
			tests = append(tests, matches[1])
			continue
		}
		if matches := jestFailPattern.FindStringSubmatch(line); len(matches) == 2 {
			tests = append(tests, strings.TrimSpace(matches[1]))
		}
	}
	return dedupeStrings(tests)
}

func salientErrors(log string) []string {
	log = strings.ReplaceAll(log, "\r\n", "\n")
	var out []string
	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "error"),
			strings.Contains(lower, "err!"),
			strings.Contains(lower, "failed"),
			strings.Contains(lower, "panic"),
			strings.Contains(lower, "fatal"),
			strings.Contains(lower, "exception"),
			strings.Contains(lower, "traceback"):
			out = append(out, normalizeLine(line))
		}
	}
	return dedupeStrings(out)
}

func normalizeLine(line string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
}

func subtractStrings(current, baseline []string) []string {
	seen := map[string]struct{}{}
	for _, item := range baseline {
		seen[item] = struct{}{}
	}
	var out []string
	for _, item := range current {
		if _, ok := seen[item]; ok {
			continue
		}
		out = append(out, item)
	}
	return dedupeStrings(out)
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func cloneEnvDiff(in map[string]model.DeltaEnvChange) map[string]model.DeltaEnvChange {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]model.DeltaEnvChange, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
