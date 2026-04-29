# Python parser or code-completion library test failure

**Playbook ID:** `python-parser-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `python`, `parser`, `jedi`, `code-completion`, `ast`

## What this failure means

A unit test failure in a Python parser library (such as jedi or a custom parser).
Tests cover string literal handling, comprehension expressions, call argument
parsing, attribute access, and doctest validation of public API.

## Common log signals

```text
test_parser.test_str_
test_parser.test_dict
test_parser.test_setcomp
test_parser.test_call_
test_parser.test_getattr
test_parser.test_for
[doctest] jedi
```

## Diagnosis

These test failures come from two overlapping sources:

- **`test_parser.*`** — unit tests for a Python expression parser, testing how
  string literals, dict/set comprehensions, function calls, and attribute chains
  are parsed and round-tripped.
- **`[doctest] jedi`** — doctest failures in the jedi code-completion library,
  covering `jedi.api`, `jedi.api.interpreter`, and `jedi.api.classes`.

Common causes:

- **Python version incompatibility** — the parser or jedi library has not been
  updated to handle syntax changes in the target Python version.
- **AST structure change** — an upstream CPython change altered the parsed AST
  tree structure expected by the tests.
- **Docstring drift** — a refactor changed a docstring example without updating
  the corresponding doctest.

## Fix steps

1. Run the failing parser tests:
   ```bash
   python -m pytest test/test_parser.py -v
   ```
2. For jedi doctest failures:
   ```bash
   python -m pytest --doctest-modules jedi/api/ -v
   ```
3. Check if the installed Python version matches the tested syntax features.

## Validation

Re-run the full test_parser module and jedi doctests:
```bash
python -m pytest test/test_parser.py --doctest-modules jedi/ -v
```

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain python-parser-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Python parser or code-completion library test failure
- Test: python parser or code-completion library test failure
- test_parser.test_setcomp
- faultline explain python-parser-test-failure
- Python python parser or code-completion library test failure


---

*Generated from [playbooks/bundled/log/test/python-parser-test-failure.yaml](../../playbooks/bundled/log/test/python-parser-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
