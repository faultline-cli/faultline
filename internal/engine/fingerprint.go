package engine

import (
	"fmt"
	"hash/fnv"
	"strings"

	"faultline/internal/model"
)

// fingerprint returns a short deterministic token derived from the playbook ID
// and the first evidence lines. It remains a stable top-level analysis field
// for backwards-compatible output, but it is not used as the durable store key.
func fingerprint(result model.Result) string {
	h := fnv.New32a()
	h.Write([]byte(result.Playbook.ID))
	limit := 3
	if len(result.Evidence) < limit {
		limit = len(result.Evidence)
	}
	for _, evidence := range result.Evidence[:limit] {
		h.Write([]byte(evidence))
	}
	return fmt.Sprintf("%08x", h.Sum32())
}

func fingerprintUnknown(signals []string, ctx model.Context) string {
	h := fnv.New32a()
	h.Write([]byte("unknown"))
	h.Write([]byte(ctx.Stage))
	h.Write([]byte(ctx.CommandHint))
	h.Write([]byte(ctx.Step))
	limit := 3
	if len(signals) < limit {
		limit = len(signals)
	}
	for _, signal := range signals[:limit] {
		h.Write([]byte(strings.TrimSpace(signal)))
	}
	return fmt.Sprintf("%08x", h.Sum32())
}
