package signature

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"faultline/internal/model"
)

const Version = "signature.v1"

type Attributes struct {
	Files      []string `json:"files,omitempty"`
	ScopeNames []string `json:"scope_names,omitempty"`
	SignalIDs  []string `json:"signal_ids,omitempty"`
}

type Payload struct {
	Version    string      `json:"version"`
	FailureID  string      `json:"failure_id"`
	Detector   string      `json:"detector,omitempty"`
	Evidence   []string    `json:"evidence"`
	Attributes *Attributes `json:"attributes,omitempty"`
}

type ResultSignature struct {
	Hash       string
	Version    string
	Payload    Payload
	Normalized string
}

var (
	ansiEscapePattern     = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	timestampPattern      = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[tT ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[zZ]|[+-]\d{2}:?\d{2})?\b`)
	dateTimePattern       = regexp.MustCompile(`\b\d{4}/\d{2}/\d{2}[ t]\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)
	dateOnlyPattern       = regexp.MustCompile(`\b\d{4}[-/]\d{2}[-/]\d{2}\b`)
	timeOnlyPattern       = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)
	lineNumberPattern     = regexp.MustCompile(`(?i)([a-z0-9_./-]+\.[a-z0-9]+):\d+(?::\d+)?`)
	normalizedLinePattern = regexp.MustCompile(`:<n>(?::<n>)+`)
	longHexPattern        = regexp.MustCompile(`(?i)\b[0-9a-f]{7,64}\b`)
	ipv4Pattern           = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	longNumberPattern     = regexp.MustCompile(`\b\d{5,}\b`)
	jobLikeNumberPattern  = regexp.MustCompile(`\b(line|column|col|position|offset|pid|process|worker|attempt|build|run|job)\s+#?\d+\b`)
	pathTokenPattern      = regexp.MustCompile(`(?:[A-Za-z]:\\|/)[^\s"'()\\[\\]{}<>]+`)
	uuidLikePattern       = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	ciWorkspacePathPrefix = []string{
		"/home/runner/work/",
		"/__w/",
		"/workspace/",
		"/builds/",
		"/github/workspace/",
		"d:/a/",
		"c:/users/runneradmin/work/",
	}
	runnerToolPathPrefix = []string{
		"/__e/",
		"/opt/hostedtoolcache/",
		"c:/hostedtoolcache/",
		"d:/a/_temp/",
	}
	homePathPrefix = []string{
		"/home/",
		"/users/",
		"c:/users/",
	}
	tempPathPrefix = []string{
		"/tmp/",
		"/var/folders/",
		"/private/var/folders/",
	}
)

func ForResult(result model.Result) ResultSignature {
	payload := Payload{
		Version:   Version,
		FailureID: strings.TrimSpace(result.Playbook.ID),
		Detector:  strings.TrimSpace(result.Detector),
	}

	evidence := make([]string, 0, len(result.Evidence))
	seenEvidence := map[string]struct{}{}
	addEvidence := func(text string) {
		for _, line := range NormalizeEvidenceLines(text) {
			if _, ok := seenEvidence[line]; ok {
				continue
			}
			seenEvidence[line] = struct{}{}
			evidence = append(evidence, line)
		}
	}

	for _, line := range result.Evidence {
		addEvidence(line)
	}

	triggerAttrs := buildTriggerAttributes(result.EvidenceBy.Triggers, addEvidence)
	if len(evidence) == 0 {
		if fallback := NormalizeEvidenceLine(result.Playbook.Title); fallback != "" {
			evidence = append(evidence, fallback)
		}
	}
	if len(evidence) == 0 && payload.FailureID != "" {
		evidence = append(evidence, payload.FailureID)
	}

	sort.Strings(evidence)
	payload.Evidence = evidence
	if triggerAttrs.hasValues() {
		payload.Attributes = &triggerAttrs
	}

	normalized := mustMarshalPayload(payload)
	sum := sha256.Sum256(normalized)
	return ResultSignature{
		Hash:       hex.EncodeToString(sum[:]),
		Version:    Version,
		Payload:    payload,
		Normalized: string(normalized),
	}
}

func NormalizeEvidenceLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if line := NormalizeEvidenceLine(part); line != "" {
			out = append(out, line)
		}
	}
	return out
}

func NormalizeEvidenceLine(line string) string {
	line = ansiEscapePattern.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	line = timestampPattern.ReplaceAllString(line, "<timestamp>")
	line = dateTimePattern.ReplaceAllString(line, "<timestamp>")
	line = dateOnlyPattern.ReplaceAllString(line, "<date>")
	line = timeOnlyPattern.ReplaceAllString(line, "<time>")
	line = normalizePathFields(line)
	line = pathTokenPattern.ReplaceAllStringFunc(line, normalizePathToken)
	line = lineNumberPattern.ReplaceAllString(line, `$1:<n>`)
	line = normalizedLinePattern.ReplaceAllString(line, ":<n>")
	line = uuidLikePattern.ReplaceAllString(line, "<id>")
	line = longHexPattern.ReplaceAllString(line, "<hex>")
	line = ipv4Pattern.ReplaceAllString(line, "<ip>")
	line = jobLikeNumberPattern.ReplaceAllStringFunc(line, func(match string) string {
		parts := strings.Fields(match)
		if len(parts) == 0 {
			return match
		}
		return parts[0] + " <n>"
	})
	line = longNumberPattern.ReplaceAllString(line, "<n>")
	line = strings.ToLower(strings.Join(strings.Fields(line), " "))
	return line
}

func buildTriggerAttributes(items []model.Evidence, addEvidence func(string)) Attributes {
	files := make([]string, 0, len(items))
	scopes := make([]string, 0, len(items))
	signals := make([]string, 0, len(items))
	seenFiles := map[string]struct{}{}
	seenScopes := map[string]struct{}{}
	seenSignals := map[string]struct{}{}

	for _, item := range items {
		addEvidence(item.Detail)
		if signalID := normalizeToken(item.SignalID); signalID != "" {
			if _, ok := seenSignals[signalID]; !ok {
				seenSignals[signalID] = struct{}{}
				signals = append(signals, signalID)
			}
		}
		if file := normalizeStructuredFile(item.File); file != "" {
			if _, ok := seenFiles[file]; !ok {
				seenFiles[file] = struct{}{}
				files = append(files, file)
			}
		}
		if scope := normalizeToken(item.ScopeName); scope != "" {
			if _, ok := seenScopes[scope]; !ok {
				seenScopes[scope] = struct{}{}
				scopes = append(scopes, scope)
			}
		}
	}

	sort.Strings(files)
	sort.Strings(scopes)
	sort.Strings(signals)
	return Attributes{
		Files:      files,
		ScopeNames: scopes,
		SignalIDs:  signals,
	}
}

func normalizeStructuredFile(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = strings.ReplaceAll(path, `\`, `/`)
	path = strings.Trim(path, `"'()[]{}<>,;:`)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) || looksLikeWindowsAbs(path) {
		return normalizePathToken(path)
	}
	return strings.Join(strings.Fields(path), " ")
}

func normalizePathToken(token string) string {
	trimmed := strings.Trim(token, `"'()[]{}<>,;:`)
	lineSuffix := ""
	for idx := strings.LastIndex(trimmed, ":"); idx > 0; idx = strings.LastIndex(trimmed, ":") {
		tail := trimmed[idx+1:]
		if !isDigits(tail) {
			break
		}
		lineSuffix = ":<n>" + lineSuffix
		trimmed = trimmed[:idx]
	}
	normalized := strings.ReplaceAll(trimmed, `\`, `/`)
	normalized = strings.TrimRight(normalized, "/")
	lower := strings.ToLower(normalized)
	switch {
	case isTempPath(lower):
		if tail := meaningfulPathTail(normalized, 2); tail != "" {
			return "<tmp>/" + tail + lineSuffix
		}
		return "<tmp>" + lineSuffix
	case isRunnerToolPath(lower):
		if tail := meaningfulPathTail(normalized, 3); tail != "" {
			return "<runner>/" + tail + lineSuffix
		}
		return "<runner>" + lineSuffix
	case isWorkspacePath(lower):
		if tail := meaningfulPathTail(normalized, 3); tail != "" {
			return "<workspace>/" + tail + lineSuffix
		}
		return "<workspace>" + lineSuffix
	case isHomePath(lower):
		if tail := meaningfulPathTail(normalized, 3); tail != "" {
			return "<home>/" + tail + lineSuffix
		}
		return "<home>" + lineSuffix
	}
	if filepath.IsAbs(normalized) || looksLikeWindowsAbs(normalized) {
		if tail := meaningfulPathTail(normalized, 3); tail != "" {
			return "<path>/" + tail + lineSuffix
		}
		return "<path>" + lineSuffix
	}
	return normalized + lineSuffix
}

func normalizePathFields(line string) string {
	fields := strings.Fields(line)
	for i, field := range fields {
		if !looksLikePathToken(field) {
			continue
		}
		fields[i] = normalizePathToken(field)
	}
	return strings.Join(fields, " ")
}

func meaningfulPathTail(path string, keep int) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) > 0 && strings.HasSuffix(parts[0], ":") {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) <= keep {
		return strings.Join(parts, "/")
	}
	return strings.Join(parts[len(parts)-keep:], "/")
}

func normalizeToken(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return ""
	}
	return strings.ToLower(value)
}

func isTempPath(path string) bool {
	if strings.HasPrefix(path, "c:/users/") && strings.Contains(path, "/appdata/local/temp/") {
		return true
	}
	for _, prefix := range tempPathPrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isWorkspacePath(path string) bool {
	for _, prefix := range ciWorkspacePathPrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isRunnerToolPath(path string) bool {
	for _, prefix := range runnerToolPathPrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isHomePath(path string) bool {
	for _, prefix := range homePathPrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func looksLikeWindowsAbs(path string) bool {
	return len(path) >= 3 && path[1] == ':' && path[2] == '/'
}

func looksLikePathToken(token string) bool {
	trimmed := strings.Trim(token, `"'()[]{}<>,;:`)
	normalized := strings.ReplaceAll(trimmed, `\`, `/`)
	return strings.HasPrefix(normalized, "/") || looksLikeWindowsAbs(normalized)
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func mustMarshalPayload(payload Payload) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(payload)
	if err != nil {
		panic(err)
	}
	return bytes.TrimSpace(buf.Bytes())
}

func (a Attributes) hasValues() bool {
	return len(a.Files) > 0 || len(a.ScopeNames) > 0 || len(a.SignalIDs) > 0
}
