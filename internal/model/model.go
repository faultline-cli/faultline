// Package model defines the shared data types used across all Faultline packages.
// No other internal packages should be imported here.
//
// Types are organised across thematic files:
//
//   - model_playbook.go   — Playbook, MatchSpec, SourceSpec, and related scoring/workflow types
//   - model_hooks.go      — Hook kinds, definitions, results, and reports
//   - model_hypothesis.go — Hypothesis specs, signals, and assessments
//   - model_repo.go       — RepoContext, TopologySignals, PackProvenance
//   - model_analysis.go   — Line, Context, Evidence, Result, Analysis, Metrics, Policy
//   - model_workflow.go   — WorkflowExecutionRecord and related execution types
package model
