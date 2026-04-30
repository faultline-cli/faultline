# Translate Toolkit conversion test failure

**Playbook ID:** `translate-toolkit-test-failure`
**Category:** test
**Severity:** high
**Tags:** `translate-toolkit`, `python`, `po`, `xliff`, `localization`, `i18n`, `unittest`

## What this failure means

One or more test cases in the Translate Toolkit localisation library failed.
The toolkit covers PO, XLIFF, MO, and other translation format conversions.
Failures indicate a format-parsing regression, a broken conversion pipeline,
or a lxml/icu dependency version mismatch.

## Common log signals

```text
test_open_office_to_xliff
test_po_to_xliff
TestPO2Sub.test_subrip
TestCPOFile.test_obsolete
TestMOFile.test_output
TestDTD.test_invalid_quoting
```

## Diagnosis

The Translate Toolkit test suite exercises format converters and file handlers:

- `test_open_office_to_xliff` / `test_po_to_xliff` — bidirectional conversion
  between OpenOffice SDF and PO/XLIFF. A failure here is usually a parsing
  change in the source format or a namespace mismatch in the output XML.
- `TestPO2Sub.test_subrip` / `TestPO2SubCommand.test_subrip` — conversion of PO
  files to SubRip subtitle format. Fails when timestamp parsing or escaping
  changes.
- `TestCPOFile.test_obsolete` / `TestCPOFile.test_obsolete_with_prev_msgid` —
  handling of obsolete translation units in compiled PO (CPO) format.
- `TestMOFile.test_output` — binary MO (compiled gettext) file serialisation.
- `TestDTD.test_invalid_quoting` — DTD entity file round-trip for Firefox
  localisations.

Common root causes:

- **lxml or icu version change** — many converters depend on XML/Unicode
  libraries whose output format changed between versions.
- **Python 2→3 byte string difference** — the toolkit historically ran on both;
  some tests write bytes then read strings.
- **Fixture file encoding** — test fixtures stored with CRLF line endings break
  PO parsers that expect LF.

## Fix steps

1. Install the development dependencies:
   ```bash
   pip install -e ".[dev]"
   ```
2. Run only the failing tests to isolate the scope:
   ```bash
   python -m pytest translate/storage/test_pypo.py -k "test_obsolete" -v
   python -m pytest translate/convert/test_po2sub.py -v
   python -m pytest translate/convert/test_oo2xliff.py -v
   ```
3. Check the lxml version matches what the CI baseline used:
   ```bash
   python -c "import lxml; print(lxml.__version__)"
   ```
4. Run the full conversion test matrix to see the extent of breakage:
   ```bash
   python -m pytest translate/convert/ -v
   ```

## Validation

- `python -m pytest translate/convert/ translate/storage/ -v` all pass.
- `python -m pytest translate/convert/test_po2xliff.py -v` passes.

## Likely files to inspect

- `translate/convert/oo2xliff.py`
- `translate/convert/po2xliff.py`
- `translate/convert/po2sub.py`
- `translate/storage/pypo.py`
- `translate/storage/cpo.py`
- `translate/storage/dtd.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain translate-toolkit-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Translate Toolkit conversion test failure
- Test: translate toolkit conversion test failure
- TestDTD.test_invalid_quoting
- faultline explain translate-toolkit-test-failure
- Python translate toolkit conversion test failure


---

*Generated from [playbooks/bundled/log/test/translate-toolkit-test-failure.yaml](../../../playbooks/bundled/log/test/translate-toolkit-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
