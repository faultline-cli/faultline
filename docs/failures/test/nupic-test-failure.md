# NuPIC region or encoder test failure

**Playbook ID:** `nupic-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `nupic`, `numenta`, `machine-learning`, `cpp`, `python`, `encoder`

## What this failure means

A test failure in the Numenta NuPIC (Numenta Platform for Intelligent Computing)
project. This may indicate a region I/O validation error (unknown output name on
a region) or a failure in the spatial/temporal encoder tests.

## Common log signals

```text
getOutputData -- unknown output 'doesnotexist'
CoordinateEncoderTest.testEncodeAdjacentPositions
extensions/core/src/main/engine/RegionIo.cpp
```

## Diagnosis

NuPIC test failures in this pattern come from two areas:

- **Region I/O tests** — The C++ engine raises an error when `getOutputData` is
  called with an output name that does not exist on the specified region. The error
  message is emitted from `RegionIo.cpp`.
- **Encoder unit tests** — `CoordinateEncoderTest.testEncodeAdjacentPositions`
  and similar tests validate the spatial pooler's coordinate encoder.

Common causes:

- **C++ extension build failure** — the NuPIC C++ core was not compiled or is
  incompatible with the installed Python version.
- **API change** — an output name was renamed or removed in the region schema.
- **Test data regression** — encoder sensitivity parameters changed.

## Fix steps

1. Rebuild the NuPIC C++ extension:
   ```bash
   python setup.py build_ext --inplace
   ```
2. Run the failing encoder test:
   ```bash
   python -m pytest tests/unit/encoders/coordinate_encoder_test.py -v
   ```
3. Check the region XML schema for the output name referenced in the error.

## Validation

Run the affected test file:
```bash
python -m pytest tests/unit/ -k "CoordinateEncoder" -v
```

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain nupic-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- NuPIC region or encoder test failure
- Test: nupic region or encoder test failure
- CoordinateEncoderTest.testEncodeAdjacentPositions
- faultline explain nupic-test-failure
- Python nupic region or encoder test failure


---

*Generated from [playbooks/bundled/log/test/nupic-test-failure.yaml](../../playbooks/bundled/log/test/nupic-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
