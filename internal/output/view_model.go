package output

import "faultline/internal/model"

// AnalysisView is the shared, human-output-facing projection of an analysis.
// JSON formatting intentionally stays on model.Analysis directly so the public
// machine schema remains explicit and stable.
type AnalysisView struct {
	Analysis *model.Analysis
	Results  []model.Result
}

func NewAnalysisView(a *model.Analysis, top int) AnalysisView {
	if a == nil {
		return AnalysisView{}
	}
	return AnalysisView{
		Analysis: a,
		Results:  topN(a.Results, top),
	}
}

func (v AnalysisView) Empty() bool {
	return v.Analysis == nil || len(v.Results) == 0
}

func (v AnalysisView) TopResult() (model.Result, bool) {
	if len(v.Results) == 0 {
		return model.Result{}, false
	}
	return v.Results[0], true
}
