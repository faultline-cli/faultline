package logdetector

import (
	"faultline/internal/detectors"
	"faultline/internal/matcher"
	"faultline/internal/model"
)

type Detector struct{}

func (Detector) Kind() detectors.Kind {
	return detectors.KindLog
}

func (Detector) Detect(playbooks []model.Playbook, target detectors.Target) []model.Result {
	weights := target.LogAnyWeights
	if weights == nil {
		weights = matcher.AnyWeights(playbooks)
	}
	return matcher.RankPrecomputed(playbooks, weights, target.LogLines, target.LogContext)
}
