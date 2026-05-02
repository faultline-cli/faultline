# Evaluation Workspace

`eval-work/` is a generated workspace for large-corpus evaluation runs.

The Makefile targets under `eval-*` recreate JSONL corpora, result files,
coverage reports, and gap-analysis samples here. Those artifacts can be very
large and are intentionally ignored by Git.

Keep durable evaluation source inputs under `fixtures/datasets/` and reusable
evaluation code under `tools/eval-corpus/`.
