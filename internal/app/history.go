package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"faultline/internal/store"
)

type historyStoreJSON struct {
	Mode     store.Mode `json:"mode,omitempty"`
	Backend  string     `json:"backend,omitempty"`
	Path     string     `json:"path,omitempty"`
	Degraded bool       `json:"degraded,omitempty"`
	Warning  string     `json:"warning,omitempty"`
}

type historyOverviewJSON struct {
	Store      historyStoreJSON         `json:"store"`
	Signatures []store.SignatureSummary `json:"signatures,omitempty"`
	Playbooks  []store.PlaybookStats    `json:"playbooks,omitempty"`
	Hooks      []store.HookStats        `json:"hooks,omitempty"`
}

type historyDetailJSON struct {
	Store       historyStoreJSON          `json:"store"`
	Signature   store.SignatureSummary    `json:"signature"`
	Findings    []store.FindingSummary    `json:"findings,omitempty"`
	HookHistory *store.HookHistorySummary `json:"hook_history,omitempty"`
}

type determinismJSON struct {
	Store       historyStoreJSON         `json:"store"`
	Source      string                   `json:"source,omitempty"`
	InputHash   string                   `json:"input_hash,omitempty"`
	Determinism store.DeterminismSummary `json:"determinism"`
}

func (Service) History(signatureHash, storePath string, limit int, jsonOut bool, w io.Writer) error {
	st, info, err := openHistoryStore(storePath)
	if err != nil {
		return err
	}
	defer st.Close()

	ctx := context.Background()
	if strings.TrimSpace(signatureHash) != "" {
		return writeHistorySignature(ctx, st, info, signatureHash, limit, jsonOut, w)
	}
	return writeHistoryOverview(ctx, st, info, limit, jsonOut, w)
}

func (Service) Signatures(storePath string, limit int, jsonOut bool, w io.Writer) error {
	st, info, err := openHistoryStore(storePath)
	if err != nil {
		return err
	}
	defer st.Close()

	ctx := context.Background()
	items, err := st.ListSignatures(ctx, limit)
	if err != nil {
		return err
	}
	if jsonOut {
		payload := struct {
			Store      historyStoreJSON         `json:"store"`
			Signatures []store.SignatureSummary `json:"signatures,omitempty"`
		}{
			Store:      historyStorePayload(info),
			Signatures: items,
		}
		return writeJSON(w, payload)
	}

	var b strings.Builder
	b.WriteString("Signatures\n")
	b.WriteString("----------\n\n")
	writeStoreInfoText(&b, info)
	if len(items) == 0 {
		b.WriteString("No stored signatures yet.\n")
		_, err := fmt.Fprint(w, b.String())
		return err
	}
	for _, item := range items {
		fmt.Fprintf(&b, "- %s  %s  seen %d time(s)\n", shortHash(item.SignatureHash), item.FailureID, item.OccurrenceCount)
		if item.Title != "" {
			fmt.Fprintf(&b, "  %s\n", item.Title)
		}
		if item.LastSeenAt != "" {
			fmt.Fprintf(&b, "  last seen: %s\n", item.LastSeenAt)
		}
	}
	_, err = fmt.Fprint(w, b.String())
	return err
}

func (Service) VerifyDeterminism(r io.Reader, source, storePath string, jsonOut bool, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read log input: %w", err)
	}
	inputHash := store.InputHashForLog(string(data))

	st, info, err := openHistoryStore(storePath)
	if err != nil {
		return err
	}
	defer st.Close()

	summary, err := st.VerifyDeterminismForInputHash(context.Background(), inputHash)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(w, determinismJSON{
			Store:       historyStorePayload(info),
			Source:      source,
			InputHash:   inputHash,
			Determinism: summary,
		})
	}

	var b strings.Builder
	b.WriteString("Determinism Check\n")
	b.WriteString("-----------------\n\n")
	writeStoreInfoText(&b, info)
	fmt.Fprintf(&b, "Source: %s\n", fallbackSource(source))
	fmt.Fprintf(&b, "Input hash: %s\n", inputHash)
	switch {
	case summary.RunCount == 0:
		b.WriteString("Status: no stored runs for this input hash\n")
	case summary.Stable:
		b.WriteString("Status: stable across stored runs\n")
	default:
		b.WriteString("Status: output drift detected across stored runs\n")
	}
	fmt.Fprintf(&b, "Stored runs: %d\n", summary.RunCount)
	fmt.Fprintf(&b, "Distinct output hashes: %d\n", summary.DistinctOutputHashes)
	if summary.FirstSeenAt != "" {
		fmt.Fprintf(&b, "First seen: %s\n", summary.FirstSeenAt)
	}
	if summary.LastSeenAt != "" {
		fmt.Fprintf(&b, "Last seen: %s\n", summary.LastSeenAt)
	}
	_, err = fmt.Fprint(w, b.String())
	return err
}

func openHistoryStore(storePath string) (store.Store, store.Info, error) {
	cfg, err := store.ResolveConfig(storePath, false)
	if err != nil {
		return nil, store.Info{}, err
	}
	return store.OpenBestEffort(cfg)
}

func writeHistoryOverview(ctx context.Context, st store.Store, info store.Info, limit int, jsonOut bool, w io.Writer) error {
	signatures, err := st.ListSignatures(ctx, limit)
	if err != nil {
		return err
	}
	playbooks, err := st.ListPlaybookStats(ctx, limit)
	if err != nil {
		return err
	}
	hooks, err := st.ListHookStats(ctx, limit)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(w, historyOverviewJSON{
			Store:      historyStorePayload(info),
			Signatures: signatures,
			Playbooks:  playbooks,
			Hooks:      hooks,
		})
	}

	var b strings.Builder
	b.WriteString("History\n")
	b.WriteString("-------\n\n")
	writeStoreInfoText(&b, info)
	if len(signatures) == 0 && len(playbooks) == 0 && len(hooks) == 0 {
		b.WriteString("No stored history yet.\n")
		_, err := fmt.Fprint(w, b.String())
		return err
	}

	if len(signatures) > 0 {
		b.WriteString("Recurring Signatures\n")
		b.WriteString("~~~~~~~~~~~~~~~~~~~~\n")
		for _, item := range signatures {
			fmt.Fprintf(&b, "- %s  %s  seen %d time(s)\n", shortHash(item.SignatureHash), item.FailureID, item.OccurrenceCount)
			if item.Title != "" {
				fmt.Fprintf(&b, "  %s\n", item.Title)
			}
			if item.FirstSeenAt != "" || item.LastSeenAt != "" {
				fmt.Fprintf(&b, "  first seen: %s  last seen: %s\n", emptyDash(item.FirstSeenAt), emptyDash(item.LastSeenAt))
			}
		}
		b.WriteString("\n")
	}

	if len(playbooks) > 0 {
		b.WriteString("Playbook Quality Snapshot\n")
		b.WriteString("~~~~~~~~~~~~~~~~~~~~~~~~\n")
		for _, item := range playbooks {
			fmt.Fprintf(&b, "- %s  selected %d of %d matched run(s), avg selected confidence %d%%\n",
				item.FailureID,
				item.SelectedCount,
				maxInt(item.MatchedCount, item.SelectedCount),
				int(item.AvgConfidence*100+0.5),
			)
			if item.NonSelectedCount > 0 || item.AvgRank > 1 {
				fmt.Fprintf(&b, "  non-selected matches: %d  avg rank: %.2f\n", item.NonSelectedCount, item.AvgRank)
			}
			if item.RecurringRunCount > 0 || item.RecurringSignatures > 0 {
				fmt.Fprintf(&b, "  recurring runs: %d  recurring signatures: %d\n", item.RecurringRunCount, item.RecurringSignatures)
			}
			if item.LastSeenAt != "" {
				fmt.Fprintf(&b, "  last seen: %s\n", item.LastSeenAt)
			}
		}
		b.WriteString("\n")
	}

	if len(hooks) > 0 {
		b.WriteString("Hook Quality Snapshot\n")
		b.WriteString("~~~~~~~~~~~~~~~~~~~~~\n")
		for _, item := range hooks {
			fmt.Fprintf(&b, "- %s/%s  total %d, executed %d, passed %d, failed %d, blocked %d\n",
				item.PlaybookID,
				item.HookID,
				item.TotalCount,
				item.ExecutedCount,
				item.PassedCount,
				item.FailedCount,
				item.BlockedCount,
			)
			fmt.Fprintf(&b, "  skipped: %d  avg delta: %+.2f\n", item.SkippedCount, item.AvgConfidenceDelta)
			if item.LastSeenAt != "" {
				fmt.Fprintf(&b, "  last seen: %s\n", item.LastSeenAt)
			}
		}
	}

	_, err = fmt.Fprint(w, strings.TrimRight(b.String(), "\n")+"\n")
	return err
}

func writeHistorySignature(ctx context.Context, st store.Store, info store.Info, signatureHash string, limit int, jsonOut bool, w io.Writer) error {
	signatureHash = strings.TrimSpace(signatureHash)
	history, err := st.LookupSignatureHistory(ctx, signatureHash)
	if err != nil {
		return err
	}
	findings, err := st.GetRecentFindingsBySignature(ctx, signatureHash, limit)
	if err != nil {
		return err
	}

	summary := store.SignatureSummary{
		SignatureHash:   signatureHash,
		OccurrenceCount: history.OccurrenceCount,
		FirstSeenAt:     history.FirstSeenAt,
		LastSeenAt:      history.LastSeenAt,
	}
	if len(findings) > 0 {
		summary.FailureID = findings[0].FailureID
		summary.Title = findings[0].Title
		summary.Category = findings[0].Category
	}

	var hookHistory *store.HookHistorySummary
	if summary.FailureID != "" {
		hookHistory, err = st.LookupHookHistory(ctx, signatureHash, summary.FailureID)
		if err != nil {
			return err
		}
	}

	if jsonOut {
		return writeJSON(w, historyDetailJSON{
			Store:       historyStorePayload(info),
			Signature:   summary,
			Findings:    findings,
			HookHistory: hookHistory,
		})
	}

	var b strings.Builder
	b.WriteString("Signature History\n")
	b.WriteString("-----------------\n\n")
	writeStoreInfoText(&b, info)
	fmt.Fprintf(&b, "Signature: %s\n", summary.SignatureHash)
	if summary.FailureID != "" {
		fmt.Fprintf(&b, "Failure: %s\n", summary.FailureID)
	}
	if summary.Title != "" {
		fmt.Fprintf(&b, "Title: %s\n", summary.Title)
	}
	switch summary.OccurrenceCount {
	case 0:
		b.WriteString("Seen: no stored occurrences yet\n")
	case 1:
		b.WriteString("Seen: 1 recorded occurrence\n")
	default:
		if span := historyWindow(summary.FirstSeenAt, summary.LastSeenAt); span != "" {
			fmt.Fprintf(&b, "Seen: %d recorded occurrences over %s\n", summary.OccurrenceCount, span)
		} else {
			fmt.Fprintf(&b, "Seen: %d recorded occurrences\n", summary.OccurrenceCount)
		}
	}
	if summary.FirstSeenAt != "" {
		fmt.Fprintf(&b, "First seen: %s\n", summary.FirstSeenAt)
	}
	if summary.LastSeenAt != "" {
		fmt.Fprintf(&b, "Last seen: %s\n", summary.LastSeenAt)
	}
	if hookHistory != nil {
		fmt.Fprintf(&b, "Hook history: total %d, executed %d, passed %d, failed %d, blocked %d, skipped %d\n",
			hookHistory.TotalCount,
			hookHistory.ExecutedCount,
			hookHistory.PassedCount,
			hookHistory.FailedCount,
			hookHistory.BlockedCount,
			hookHistory.SkippedCount,
		)
	}
	if len(findings) > 0 {
		b.WriteString("\nRecent Findings\n")
		b.WriteString("~~~~~~~~~~~~~~~\n")
		for _, item := range findings {
			fmt.Fprintf(&b, "- %s  %s\n", item.SeenAt, item.FailureID)
			if item.Title != "" {
				fmt.Fprintf(&b, "  %s\n", item.Title)
			}
		}
	}
	_, err = fmt.Fprint(w, b.String())
	return err
}

func historyStorePayload(info store.Info) historyStoreJSON {
	return historyStoreJSON{
		Mode:     info.Mode,
		Backend:  info.Backend,
		Path:     info.Path,
		Degraded: info.Degraded,
		Warning:  info.Warning,
	}
}

func writeStoreInfoText(b *strings.Builder, info store.Info) {
	if info.Backend == "" && info.Path == "" && !info.Degraded {
		b.WriteString("Store: disabled\n\n")
		return
	}
	label := strings.TrimSpace(info.Backend)
	if label == "" {
		label = "store"
	}
	if info.Path != "" {
		fmt.Fprintf(b, "Store: %s (%s)\n", label, info.Path)
	} else {
		fmt.Fprintf(b, "Store: %s\n", label)
	}
	if info.Degraded && info.Warning != "" {
		fmt.Fprintf(b, "Warning: %s\n", info.Warning)
	}
	b.WriteString("\n")
}

func writeJSON(w io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

func shortHash(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func historyWindow(firstSeenAt, lastSeenAt string) string {
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(firstSeenAt))
	if err != nil {
		return ""
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(lastSeenAt))
	if err != nil || end.Before(start) {
		return ""
	}
	duration := end.Sub(start)
	switch {
	case duration >= 48*time.Hour:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	case duration >= time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	case duration >= time.Minute:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	default:
		return ""
	}
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func fallbackSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "stdin"
	}
	return source
}
