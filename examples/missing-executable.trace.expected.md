# Faultline Trace

- Playbook: `missing-executable`
- Title: Required executable or runtime binary missing
- Source: `stdin`
- Outcome: matched and ranked #1

## Outcome

- Score: 2.10
- Confidence: 55%

## Rule Evaluation

- `MISSING` `match.any[0]`
  pattern: `executable file not found in $PATH`
  evidence: none
  note: trigger rule did not match
- `MATCHED` `match.any[1]`
  pattern: `exec /`
  evidence: line 1: exec /__e/node20/bin/node: no such file or directory
  note: trigger rule matched the log
- `MISSING` `match.any[2]`
  pattern: `exec: "`
  evidence: none
  note: trigger rule did not match
- `MISSING` `match.any[3]`
  pattern: `command not found`
  evidence: none
  note: trigger rule did not match
- `MISSING` `match.any[4]`
  pattern: `is not recognized as an internal or external command`
  evidence: none
  note: trigger rule did not match
- `MISSING` `match.any[5]`
  pattern: `exit status 127`
  evidence: none
  note: trigger rule did not match
- `CLEAR` `match.none[0]`
  pattern: `fixture`
  evidence: none
  note: exclusion rule stayed clear
- `CLEAR` `match.none[1]`
  pattern: `testdata`
  evidence: none
  note: exclusion rule stayed clear
- `CLEAR` `match.none[2]`
  pattern: `cannot stat`
  evidence: none
  note: exclusion rule stayed clear
- `CLEAR` `match.none[3]`
  pattern: `withBinaryFile:`
  evidence: none
  note: exclusion rule stayed clear
- `CLEAR` `match.none[4]`
  pattern: `package-lock.json`
  evidence: none
  note: exclusion rule stayed clear
- `CLEAR` `match.none[5]`
  pattern: `Dockerfile: no such file or directory`
  evidence: none
  note: exclusion rule stayed clear

## Why This Result

- 1 trigger rule(s) matched explicit log evidence
- 6 exclusion rule(s) stayed clear
- matched evidence was pulled directly from the input log

## Signature

- Hash: `4e9c13ba9d6a5ad10c761c94bcdf32b86abb858f75fc56874ec59c55b28bf11a`
- Version: `signature.v1`
- Payload:
```json
{"version":"signature.v1","failure_id":"missing-executable","detector":"log","evidence":["exec <runner>/node20/bin/node no such file or directory"]}
```
