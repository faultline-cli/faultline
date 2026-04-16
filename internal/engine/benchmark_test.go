package engine

import (
	"strings"
	"testing"

	"faultline/internal/matcher"
	"faultline/internal/model"
	"faultline/internal/playbooks"
)

func BenchmarkLoadBundledPlaybooks(b *testing.B) {
	dir := repoPlaybookDir(nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pbs, err := playbooks.LoadDir(dir)
		if err != nil {
			b.Fatalf("load playbooks: %v", err)
		}
		if len(pbs) == 0 {
			b.Fatal("expected bundled playbooks to load")
		}
	}
}

func BenchmarkAnalyzeRepresentativeCorpus(b *testing.B) {
	pbs, err := playbooks.LoadDir(repoPlaybookDir(nil))
	if err != nil {
		b.Fatalf("load playbooks: %v", err)
	}
	index := newPatternIndex(pbs)
	logs := representativeCorpus(pbs, index)
	if len(logs) == 0 {
		b.Fatal("expected representative corpus")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, log := range logs {
			lines, err := readLines(strings.NewReader(log))
			if err != nil {
				b.Fatalf("read lines: %v", err)
			}
			results := matcher.Rank(pbs, lines, ExtractContext(lines))
			if len(results) == 0 {
				b.Fatal("expected representative corpus log to match")
			}
		}
	}
}

func representativeCorpus(pbs []model.Playbook, index patternIndex) []string {
	logs := make([]string, 0, len(pbs))
	for _, pb := range pbs {
		logs = append(logs, representativeLogForPlaybook(pb, index, ""))
	}
	return logs
}
