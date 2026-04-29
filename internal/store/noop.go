package store

import "context"

type noopStore struct{}

func Noop() Store {
	return noopStore{}
}

func (noopStore) BeginRun(context.Context, BeginRunParams) (RunHandle, error) {
	return RunHandle{}, nil
}

func (noopStore) CompleteRun(context.Context, RunHandle, CompleteRunParams) error {
	return nil
}

func (noopStore) LookupSignatureHistory(context.Context, string) (SignatureHistory, error) {
	return SignatureHistory{}, nil
}

func (noopStore) CountSeenFailure(context.Context, string) (int, error) {
	return 0, nil
}

func (noopStore) RecentTopFailures(context.Context, int) ([]string, error) {
	return nil, nil
}

func (noopStore) ListSignatures(context.Context, int) ([]SignatureSummary, error) {
	return nil, nil
}

func (noopStore) GetRecentFindingsBySignature(context.Context, string, int) ([]FindingSummary, error) {
	return nil, nil
}

func (noopStore) ListPlaybookStats(context.Context, int) ([]PlaybookStats, error) {
	return nil, nil
}

func (noopStore) LookupHookHistory(context.Context, string, string) (*HookHistorySummary, error) {
	return nil, nil
}

func (noopStore) ListHookStats(context.Context, int) ([]HookStats, error) {
	return nil, nil
}

func (noopStore) VerifyDeterminismForInputHash(context.Context, string) (DeterminismSummary, error) {
	return DeterminismSummary{}, nil
}

func (noopStore) Close() error {
	return nil
}
