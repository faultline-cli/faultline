# Refactor Module

Use this prompt when improving an existing package without changing product behavior.

## Goal

Improve clarity, maintainability, or separation of concerns while preserving behavior.

## Required Approach

- Preserve public behavior unless the task explicitly requests otherwise.
- Keep the refactor focused on one module or one narrow set of files.
- Avoid introducing new layers unless they remove real complexity.
- Validate that output and ordering remain unchanged.

## Deliverable

- Refined code
- Regression checks or tests
- Notes on any structural decisions
