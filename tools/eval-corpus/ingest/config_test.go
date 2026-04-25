package ingest_test

import (
	"os"
	"path/filepath"
	"testing"

	"faultline/tools/eval-corpus/ingest"
)

func TestLoadConfigValid(t *testing.T) {
	path := filepath.Join("..", "testdata", "sample-config.yaml")
	cfg, err := ingest.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Parsing.LogField != "message" {
		t.Errorf("LogField = %q, want %q", cfg.Parsing.LogField, "message")
	}
	if cfg.Processing.Dedupe != true {
		t.Error("Dedupe should be true")
	}
	if cfg.Processing.Redact.Emails != true {
		t.Error("Redact.Emails should be true")
	}
}

func TestLoadConfigMissingLogField(t *testing.T) {
	content := "input:\n  type: csv\nparsing:\n  id_field: id\n"
	tmp := writeTempConfig(t, content)
	_, err := ingest.LoadConfig(tmp)
	if err == nil {
		t.Fatal("expected validation error for missing log_field")
	}
}

func TestLoadConfigUnsupportedType(t *testing.T) {
	content := "input:\n  type: parquet\nparsing:\n  log_field: msg\n"
	tmp := writeTempConfig(t, content)
	_, err := ingest.LoadConfig(tmp)
	if err == nil {
		t.Fatal("expected error for unsupported input type")
	}
}

func TestConfigEffectiveName(t *testing.T) {
	tests := []struct {
		name      string
		cfgName   string
		inputPath string
		want      string
	}{
		{"from config name", "mydata", "file.csv", "mydata"},
		{"from input path", "", "/data/ci_logs.csv", "ci_logs"},
		{"from input no ext", "", "/data/ci_logs", "ci_logs"},
		{"fallback", "", "", "dataset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ingest.Config{Name: tt.cfgName}
			got := cfg.EffectiveName(tt.inputPath)
			if got != tt.want {
				t.Errorf("EffectiveName = %q, want %q", got, tt.want)
			}
		})
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}
