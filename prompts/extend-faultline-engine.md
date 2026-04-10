# Extend Faultline Engine

Use this prompt when adding a new deterministic faultline rule.

## Goal

Add a rule that fits the existing engine and produces stable, explainable output.

## Required Approach

- Express the rule in explicit inputs, conditions, and evidence.
- Keep the rule deterministic and easy to test.
- Reuse existing indicator data instead of broadening scope unnecessarily.
- Ensure the new rule does not increase comment noise.

## Deliverable

- Rule implementation
- Test cases for positive and negative paths
- Any doc update needed to explain the new rule
