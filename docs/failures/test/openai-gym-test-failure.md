# OpenAI Gym environment test failure

**Playbook ID:** `openai-gym-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `openai-gym`, `reinforcement-learning`, `python`, `environment`, `atari`

## What this failure means

A test failure in the OpenAI Gym reinforcement learning library. These tests
exercise environment registration, monitor lifecycle, and parametrized
environment step/reset/render conformance across all registered environments.

## Common log signals

```text
gym.monitoring.tests.
gym.envs.tests.test_envs.
test_envs.test_env:
```

## Diagnosis

OpenAI Gym's test suite runs three main categories of tests:

- **`gym.envs.tests.test_envs`** — parametrized test that instantiates every
  registered Gym environment, calls `reset()`, takes a random `step()`, and
  verifies the observation/reward/done/info tuple shape.
- **`gym.monitoring.tests.test_monitor_envs`** — tests the `Monitor` wrapper
  that wraps environments to record episode statistics and video.
- **`test_envs.test_env`** — standalone environment conformance tests.

The tests are parametrized by environment ID (`:N` suffix in the log), so
failing tests point to a specific registered environment.

Common causes:

- **Missing renderer** — Atari or MuJoCo environments require system libraries
  (`libgl`, `libglew`, MuJoCo license key) that are not available in CI.
- **Unregistered environment** — a new environment was added but not registered
  in `gym/__init__.py` or `gym/envs/__init__.py`.
- **Observation space shape mismatch** — environment `reset()` or `step()`
  returns an observation that does not conform to the declared `observation_space`.

## Fix steps

1. Identify the failing environment IDs from the `:N` suffix in the log.
2. Check if the environment requires system libraries:
   ```bash
   apt-get install -y libgl1-mesa-glx libglew-dev
   pip install gym[atari]
   ```
3. Run the failing environment test individually:
   ```bash
   python -m pytest gym/envs/tests/test_envs.py -k "CartPole" -v
   ```
4. Verify the environment is properly registered:
   ```python
   import gym; gym.make('CartPole-v1')
   ```

## Validation

Run the environment conformance tests:
```bash
python -m pytest gym/envs/tests/ gym/monitoring/tests/ -v
```

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain openai-gym-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- OpenAI Gym environment test failure
- Test: openai gym environment test failure
- gym.envs.tests.test_envs.
- faultline explain openai-gym-test-failure
- Python openai gym environment test failure


---

*Generated from [playbooks/bundled/log/test/openai-gym-test-failure.yaml](../../../playbooks/bundled/log/test/openai-gym-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
