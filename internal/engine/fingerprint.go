package engine

import (
	"fmt"
	"hash/fnv"

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
