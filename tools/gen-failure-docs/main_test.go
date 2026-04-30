package main

import "testing"

func TestRelPlaybookLinkFromCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/auth/aws-credentials.yaml", "auth")
	want := "../../../playbooks/bundled/log/auth/aws-credentials.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}

func TestRelPlaybookLinkFromNestedCategoryPage(t *testing.T) {
	got := relPlaybookLink("playbooks/bundled/log/silent/artifact-missing.yaml", "silent/failure")
	want := "../../../../playbooks/bundled/log/silent/artifact-missing.yaml"
	if got != want {
		t.Fatalf("relPlaybookLink() = %q, want %q", got, want)
	}
}
