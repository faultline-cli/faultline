package scoring

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed weights/bayes_v1.json
var weightsFS embed.FS

var (
	loadWeightsOnce sync.Once
	loadedWeights   weightsFile
	loadWeightsErr  error
)

func defaultWeights() (weightsFile, error) {
	loadWeightsOnce.Do(func() {
		data, err := weightsFS.ReadFile("weights/bayes_v1.json")
		if err != nil {
			loadWeightsErr = fmt.Errorf("read scoring weights: %w", err)
			return
		}
		if err := json.Unmarshal(data, &loadedWeights); err != nil {
			loadWeightsErr = fmt.Errorf("parse scoring weights: %w", err)
			return
		}
		if loadedWeights.PriorSmoothing <= 0 {
			loadedWeights.PriorSmoothing = 1
		}
		if loadedWeights.FeatureWeights == nil {
			loadedWeights.FeatureWeights = map[string]float64{}
		}
		if loadedWeights.PlaybookCounts == nil {
			loadedWeights.PlaybookCounts = map[string]int{}
		}
	})
	return loadedWeights, loadWeightsErr
}
