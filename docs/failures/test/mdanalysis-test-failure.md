# MDAnalysis molecular dynamics library test failure

**Playbook ID:** `mdanalysis-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `mdanalysis`, `molecular-dynamics`, `python`, `scientific`, `coordinates`, `topology`

## What this failure means

A test failure in the MDAnalysis Python library for molecular dynamics trajectory
analysis. Failing tests typically cover coordinate readers/writers (DCD, DLP, GRO,
LAMMPS, PDB, XYZ), topology parsing, or atom group operations.

## Common log signals

```text
MDAnalysisTests.coordinates.
MDAnalysisTests.topology.
MDAnalysisTests.test_atomgroup.
MDAnalysisTests.test_atomselections.
MDAnalysisTests.test_modelling.
MDAnalysisTests.test_topologyattrs.
MDAnalysisTests.test_util.
```

## Diagnosis

MDAnalysis uses a modular test suite under `MDAnalysisTests`. Common failing
subsystems include:

- **Coordinate readers** (`MDAnalysisTests.coordinates.*`) — DCD, DLP, GRO,
  LAMMPS, NetCDF, PDB, TRZ, XYZ format readers and writers.
- **Topology parsers** (`MDAnalysisTests.topology.*`) — GRO parser and
  related topology attribute reading.
- **Atom group operations** (`MDAnalysisTests.test_atomgroup.*`) — Universe
  construction, atom selection, indexing.
- **Atom selections** (`MDAnalysisTests.test_atomselections.*`) — around,
  zone, and bonded selection algorithms.
- **Utility functions** (`MDAnalysisTests.test_util.*`) — `make_whole` and
  other geometry utilities.

Common causes:

- **Missing optional dependency** — a reader requires `netcdf4`, `h5py`, or
  `GridDataFormats` which is not installed.
- **NumPy/SciPy version incompatibility** — array shape or dtype assumptions
  in test fixtures diverge from the installed version.
- **Trajectory fixture missing or corrupt** — the test data files under
  `MDAnalysisTests/data/` were not installed correctly.

## Fix steps

1. Install all optional dependencies:
   ```bash
   pip install MDAnalysis[analysis] netcdf4 h5py GridDataFormats
   ```
2. Run the failing test file directly:
   ```bash
   python -m pytest MDAnalysisTests/coordinates/test_dlpoly.py -v
   ```
3. Verify that test data files are present:
   ```bash
   python -c "import MDAnalysisTests; print(MDAnalysisTests.__file__)"
   ls $(python -c "import MDAnalysisTests; import os; print(os.path.dirname(MDAnalysisTests.__file__))")/data/
   ```

## Validation

Re-run the coordinate and topology test suites:
```bash
python -m pytest MDAnalysisTests/coordinates/ MDAnalysisTests/topology/ -v
```

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain mdanalysis-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- MDAnalysis molecular dynamics library test failure
- Test: mdanalysis molecular dynamics library test failure
- MDAnalysisTests.test_atomselections.
- faultline explain mdanalysis-test-failure
- Python mdanalysis molecular dynamics library test failure


---

*Generated from [playbooks/bundled/log/test/mdanalysis-test-failure.yaml](../../../playbooks/bundled/log/test/mdanalysis-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
