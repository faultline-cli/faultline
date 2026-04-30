# False Positive Analysis — github-actions-2026-04-29
Dataset: 30,094 fixtures | Matched: 29499 | Suspect low-confidence: 17767

## Confidence Distribution (all matched)
| Range | Count | % |
|---|---|---|
| <0.40 | 17609 | 59.7% |
| 0.40-0.50 | 1680 | 5.7% |
| 0.50-0.60 | 4738 | 16.1% |
| 0.60-0.70 | 3126 | 10.6% |
| 0.70-0.85 | 1814 | 6.1% |
| >=0.85 | 532 | 1.8% |

## Suspect Playbooks (≥50% matches below conf 0.5)
| Playbook | Matches | Low-conf | Low-conf% | Median conf | Max conf |
|---|---|---|---|---|---|
| git-auth | 6184 | 5961 | 96% | 0.38 | 0.63 |
| pip-install-failure | 4328 | 4151 | 96% | 0.37 | 0.87 |
| yarn-lockfile | 2279 | 2118 | 93% | 0.37 | 0.64 |
| runtime-mismatch | 1405 | 1261 | 90% | 0.37 | 0.87 |
| arch-mismatch | 896 | 896 | 100% | 0.37 | 0.39 |
| database-test-isolation | 648 | 647 | 100% | 0.32 | 0.82 |
| poetry-lockfile-drift | 545 | 538 | 99% | 0.37 | 0.69 |
| node-version-mismatch | 508 | 508 | 100% | 0.37 | 0.38 |
| artifact-upload-failure | 485 | 289 | 60% | 0.39 | 0.87 |
| ipv6-ipv4-resolution | 334 | 298 | 89% | 0.46 | 0.58 |
| dotnet-restore | 292 | 208 | 71% | 0.37 | 0.87 |
| permission-denied | 276 | 170 | 62% | 0.41 | 0.88 |
| formatting-failure | 111 | 111 | 100% | 0.47 | 0.47 |
| alpine-debian-incompatibility | 100 | 94 | 94% | 0.45 | 0.86 |
| install-failure | 85 | 85 | 100% | 0.37 | 0.37 |
| docker-manifest-not-found | 65 | 42 | 65% | 0.37 | 0.53 |
| npm-eresolve-conflict | 64 | 33 | 52% | 0.42 | 0.52 |
| go-sum-missing | 47 | 47 | 100% | 0.37 | 0.37 |
| missing-test-fixture | 43 | 43 | 100% | 0.34 | 0.41 |
| git-shallow-checkout | 40 | 22 | 55% | 0.49 | 0.62 |
| zero-tests-executed | 40 | 40 | 100% | 0.38 | 0.38 |
| build-input-file-missing | 32 | 26 | 81% | 0.37 | 0.53 |
| artifact-missing | 29 | 26 | 90% | 0.37 | 0.87 |
| working-directory | 18 | 17 | 94% | 0.40 | 0.85 |
| docker-daemon-unavailable | 17 | 17 | 100% | 0.37 | 0.38 |
| ignored-exit-code | 16 | 9 | 56% | 0.46 | 0.53 |
| cache-corruption | 15 | 15 | 100% | 0.35 | 0.35 |
| package-manager-mismatch | 14 | 8 | 57% | 0.49 | 0.52 |
| dependency-drift | 14 | 14 | 100% | 0.35 | 0.35 |
| invalid-config-schema | 12 | 12 | 100% | 0.45 | 0.49 |
| clock-drift | 11 | 11 | 100% | 0.38 | 0.38 |
| proxy-configuration | 10 | 9 | 90% | 0.47 | 0.59 |
| dependency-removed-upstream | 9 | 9 | 100% | 0.47 | 0.47 |
| aws-credentials | 7 | 7 | 100% | 0.37 | 0.37 |
| path-case-mismatch | 7 | 7 | 100% | 0.40 | 0.46 |
| test-timeout | 7 | 7 | 100% | 0.33 | 0.33 |
| process-killed-no-logs | 6 | 6 | 100% | 0.46 | 0.46 |
| terraform-state-lock | 5 | 5 | 100% | 0.42 | 0.42 |

## Per-Playbook Log Samples (low-confidence matches)

### `git-auth` — 6184 matches, 5961 low-conf (96%)

**Fixture:** `e6c2765a96f46397c21d...` conf=0.38
**Evidence:** ['<timestamp>.2648734Z ##[group]Fetching the repository']
```
2026-03-24T17:35:08.6600042Z [ERROR] Failed to execute goal org.apache.maven.plugins:maven-checkstyle-plugin:3.3.0:check (default) on project oop-example: Failed during checkstyle execution: There are 3 errors reported by Checkstyle 9.3 with https://raw.githubusercontent.com/mate-academy/style-guides/master/java/checkstyle.xml ruleset. -> [Help 1]
2026-03-24T17:35:08.6602703Z [ERROR] 
2026-03-24T17:35:08.6603663Z [ERROR] To see the full stack trace of the errors, re-run Maven with the -e switch.
2026-03-24T17:35:08.6605107Z [ERROR] Re-run Maven using the -X switch to enable full debug logging.
2026-03-24T17:35:08.6606509Z [ERROR] 
2026-03-24T17:35:08.6607609Z [ERROR] For more information about the errors and possible solutions, please read the following articles:
2026-03-24T17:35:08.6609057Z [ERROR] [Help 1] http://cwiki.apache.org/confluence/display/MAVEN/MojoExecutionException
2026-03-24T17:35:08.6902442Z ##[error]Process completed with exit code 1.
2026-03-24T17:35:08.7027439Z Post job cleanup.
2026-03-24T17:35:08.8885306Z Post job cleanup.
2026-03-24T17:35:08.9844337Z [command]/usr/bin/git version
2026-03-24T17:35:08.9890958Z git version 2.53.0
2026-03-24T17:35:08.9933923Z Temporarily overriding HOME='/home/runner/work/_temp/aa14590d-865f-4f50-ad95-44b2539ddce5' before making global git config changes
2026-03-24T17:35:08.9935237Z Adding repository directory to the temporary git global config as a safe directory
2026-03-24T17:35:08.9939808Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/jv-oop/jv-oop
2026-03-24T17:35:08.9975452Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-24T17:35:09.0009167Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-24T17:35:09.0257280Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-24T17:35:09.0280706Z http.https://github.com/.extraheader
2026-03-24T17:35:09.0293949Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-24T17:35:09.0328029Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-24T17:35:09.0564054Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-24T17:35:09.0597526Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-24T17:35:09.0944656Z Cleaning up orphan processes
2026-03-24T17:35:09.1249229Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-java@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `3531a17b43cbf4617be6...` conf=0.38
**Evidence:** ['<timestamp>.6820049Z ##[group]Fetching the repository']
```
2026-04-10T14:40:47.2760992Z Removing includeIf entries pointing to credentials config files
2026-04-10T14:40:47.2768405Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-10T14:40:47.2791265Z includeif.gitdir:/home/runner/work/okr/okr/.git.path
2026-04-10T14:40:47.2792254Z includeif.gitdir:/home/runner/work/okr/okr/.git/worktrees/*.path
2026-04-10T14:40:47.2793318Z includeif.gitdir:/github/workspace/.git.path
2026-04-10T14:40:47.2794137Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-10T14:40:47.2801683Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/okr/okr/.git.path
2026-04-10T14:40:47.2823103Z /home/runner/work/_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2832414Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/okr/okr/.git.path /home/runner/work/_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2866420Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/okr/okr/.git/worktrees/*.path
2026-04-10T14:40:47.2887814Z /home/runner/work/_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2896358Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/okr/okr/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2927315Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-10T14:40:47.2949104Z /github/runner_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2959347Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.2992084Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-10T14:40:47.3012938Z /github/runner_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.3021635Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config
2026-04-10T14:40:47.3056229Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-10T14:40:47.3273598Z Removing credentials config '/home/runner/work/_temp/git-credentials-be5861ee-aedb-4367-87cb-02df99900de4.config'
2026-04-10T14:40:47.3409331Z Cleaning up orphan processes
2026-04-10T14:40:47.3847079Z Terminate orphan process: pid (3162) (npm run start)
2026-04-10T14:40:47.3872483Z Terminate orphan process: pid (3177) (sh)
2026-04-10T14:40:47.3896043Z Terminate orphan process: pid (3178) (ng serve (frontend))
2026-04-10T14:40:47.3909809Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: cypress-io/github-action@v6. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `854394b19130ea040896...` conf=0.38
**Evidence:** ['<timestamp>.3696989Z ##[group]Fetching the repository']
```
2026-01-30T07:21:55.5166255Z   BRANCH_NAME: renovate/org.springdoc-springdoc-openapi-starter-webmvc-ui-3.x
2026-01-30T07:21:55.5166795Z   COMMIT_REF: refs/heads/renovate/org.springdoc-springdoc-openapi-starter-webmvc-ui-3.x
2026-01-30T07:21:55.5167310Z   JAVA_HOME: /opt/hostedtoolcache/Java_Temurin-Hotspot_jdk/21.0.10-7/x64
2026-01-30T07:21:55.5167749Z   JAVA_HOME_21_X64: /opt/hostedtoolcache/Java_Temurin-Hotspot_jdk/21.0.10-7/x64
2026-01-30T07:21:55.5168109Z ##[endgroup]
2026-01-30T07:21:55.7299050Z ##[warning]No files were found with the provided path: frontend/cypress/screenshots. No artifacts will be uploaded.
2026-01-30T07:21:55.7453053Z Post job cleanup.
2026-01-30T07:21:55.8805126Z Post job cleanup.
2026-01-30T07:21:55.9712787Z [command]/usr/bin/git version
2026-01-30T07:21:55.9762282Z git version 2.52.0
2026-01-30T07:21:55.9816839Z Temporarily overriding HOME='/home/runner/work/_temp/013f1dbf-a161-4e6f-8a23-312cb5931ac4' before making global git config changes
2026-01-30T07:21:55.9819285Z Adding repository directory to the temporary git global config as a safe directory
2026-01-30T07:21:55.9824306Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/okr/okr
2026-01-30T07:21:55.9867067Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-01-30T07:21:55.9904137Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-01-30T07:21:56.0159709Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-01-30T07:21:56.0186487Z http.https://github.com/.extraheader
2026-01-30T07:21:56.0198996Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-01-30T07:21:56.0233160Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-01-30T07:21:56.0477141Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-01-30T07:21:56.0509181Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-01-30T07:21:56.0871160Z Cleaning up orphan processes
2026-01-30T07:21:56.1208083Z Terminate orphan process: pid (2999) (npm run start)
2026-01-30T07:21:56.1246087Z Terminate orphan process: pid (3013) (sh)
2026-01-30T07:21:56.1283732Z Terminate orphan process: pid (3014) (ng serve (frontend))
```

### `pip-install-failure` — 4328 matches, 4151 low-conf (96%)

**Fixture:** `0a25beb6030e0c9add0c...` conf=0.37
**Evidence:** ['<timestamp>.8695779Z #8 [3/3] RUN pip install -r /code/requirements.txt']
```
2026-03-23T23:54:26.7527849Z Terminate orphan process: pid (4349) (python)
2026-03-23T23:54:26.7561991Z Terminate orphan process: pid (4387) (python)
2026-03-23T23:54:26.7594503Z Terminate orphan process: pid (4424) (python)
2026-03-23T23:54:26.7648298Z Terminate orphan process: pid (4460) (python)
2026-03-23T23:54:26.7681496Z Terminate orphan process: pid (4498) (python)
2026-03-23T23:54:26.7714162Z Terminate orphan process: pid (4535) (python)
2026-03-23T23:54:26.7748590Z Terminate orphan process: pid (4572) (python)
2026-03-23T23:54:26.7783002Z Terminate orphan process: pid (4612) (python)
2026-03-23T23:54:26.7838273Z Terminate orphan process: pid (4652) (python)
2026-03-23T23:54:26.7871984Z Terminate orphan process: pid (4689) (python)
2026-03-23T23:54:26.7997010Z Terminate orphan process: pid (4731) (python)
2026-03-23T23:54:26.8229432Z Terminate orphan process: pid (4768) (python)
2026-03-23T23:54:26.8444759Z Terminate orphan process: pid (4807) (python)
2026-03-23T23:54:26.8708813Z Terminate orphan process: pid (4844) (python)
2026-03-23T23:54:26.8952077Z Terminate orphan process: pid (4882) (python)
2026-03-23T23:54:26.9136676Z Terminate orphan process: pid (4919) (python)
2026-03-23T23:54:26.9373724Z Terminate orphan process: pid (4957) (python)
2026-03-23T23:54:26.9590014Z Terminate orphan process: pid (4993) (python)
2026-03-23T23:54:26.9754813Z Terminate orphan process: pid (5030) (python)
2026-03-23T23:54:26.9996592Z Terminate orphan process: pid (5069) (python)
2026-03-23T23:54:27.0180788Z Terminate orphan process: pid (5398) (python)
2026-03-23T23:54:27.0245800Z Terminate orphan process: pid (5399) (python)
2026-03-23T23:54:27.0304925Z Terminate orphan process: pid (5400) (python)
2026-03-23T23:54:27.0338025Z Terminate orphan process: pid (5401) (python)
2026-03-23T23:54:27.0411978Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v3, actions/setup-python@v3. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `61546ae9aa1e7ba6eec6...` conf=0.42
**Evidence:** ['<timestamp>.0867347Z nox > /home/runner/work/bx_django_utils/bx_django_utils/.venv/bin/uv pip install uv']
```
2026-04-21T11:40:09.0352855Z +* [`get_log_entry_qs()`](https://github.com/boxine/bx_django_utils/blob/master/bx_django_utils/admin_utils/log_entry.py#L91-L114) - Get a QuerySet of LogEntry objects, with optional filtering by model, object ID, and action flag.
2026-04-21T11:40:09.0354269Z +* [`get_log_message_data()`](https://github.com/boxine/bx_django_utils/blob/master/bx_django_utils/admin_utils/log_entry.py#L71-L88) - Get `LogEntry` change messages data structure as list.
2026-04-21T11:40:09.0355458Z  * [`validate_action_flag()`](https://github.com/boxine/bx_django_utils/blob/master/bx_django_utils/admin_utils/log_entry.py#L17-L22) - Validate that the action flag is one of the allowed values.
2026-04-21T11:40:09.0356174Z  
2026-04-21T11:40:09.0356348Z  ### bx_django_utils.approve_workflow
2026-04-21T11:40:09.0356516Z 
2026-04-21T11:40:09.0356655Z ----------------------------------------------------------------------
2026-04-21T11:40:09.0357313Z Ran 191 tests in 11.074s
2026-04-21T11:40:09.0357473Z 
2026-04-21T11:40:09.0357549Z FAILED (failures=1)
2026-04-21T11:40:09.0357772Z Used shuffle seed: 650271287 (generated)
2026-04-21T11:40:09.0358048Z Destroying test database for alias 'default'...
2026-04-21T11:40:09.0358351Z Destroying test database for alias 'default'...
2026-04-21T11:40:09.0358646Z Destroying test database for alias 'default'...
2026-04-21T11:40:09.0358942Z Destroying test database for alias 'default'...
2026-04-21T11:40:09.0359222Z Destroying test database for alias 'default'...
2026-04-21T11:40:09.0359507Z Destroying test database for alias 'second'...
2026-04-21T11:40:09.0359801Z Destroying test database for alias 'second'...
2026-04-21T11:40:09.0360232Z Destroying test database for alias 'second'...
2026-04-21T11:40:09.0360522Z Destroying test database for alias 'second'...
2026-04-21T11:40:09.0360803Z Destroying test database for alias 'second'...
2026-04-21T11:40:09.2075704Z nox > Command python -m coverage run --context py3.12-django6.0 failed with exit code 1
2026-04-21T11:40:09.2076647Z nox > Session tests-3.12(django='6.0') failed.
2026-04-21T11:40:09.2256303Z ##[error]Process completed with exit code 1.
2026-04-21T11:40:09.2344537Z Cleaning up orphan processes
```

**Fixture:** `88e58d661159a123271d...` conf=0.37
**Evidence:** ['<timestamp>.3035750Z ##[group]Run pip install coverage tox']
```
2026-04-04T20:57:01.4369780Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-04T20:57:01.4537871Z Removing HTTP extra header
2026-04-04T20:57:01.4542091Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-04T20:57:01.4570570Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-04T20:57:01.4732131Z Removing includeIf entries pointing to credentials config files
2026-04-04T20:57:01.4737259Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-04T20:57:01.4756333Z includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git.path
2026-04-04T20:57:01.4757404Z includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git/worktrees/*.path
2026-04-04T20:57:01.4758189Z includeif.gitdir:/github/workspace/.git.path
2026-04-04T20:57:01.4758676Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-04T20:57:01.4765305Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git.path
2026-04-04T20:57:01.4782358Z /home/runner/work/_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4790558Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git.path /home/runner/work/_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4820136Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git/worktrees/*.path
2026-04-04T20:57:01.4839413Z /home/runner/work/_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4846948Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/mopidy-podcast-itunes/mopidy-podcast-itunes/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4871563Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-04T20:57:01.4887827Z /github/runner_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4894827Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4919244Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-04T20:57:01.4935689Z /github/runner_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4974091Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config
2026-04-04T20:57:01.4998427Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-04T20:57:01.5158657Z Removing credentials config '/home/runner/work/_temp/git-credentials-9051b125-083d-4026-ac2e-6871ad6c10e2.config'
2026-04-04T20:57:01.5269710Z Cleaning up orphan processes
```

### `yarn-lockfile` — 2279 matches, 2118 low-conf (93%)

**Fixture:** `d548a75d2fa7bca6871c...` conf=0.37
**Evidence:** ['<timestamp>.0609022Z ##[group]Run yarn install --immutable']
```
2026-04-21T16:49:10.6528438Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-21T16:49:10.6750132Z Removing HTTP extra header
2026-04-21T16:49:10.6754474Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-21T16:49:10.6784461Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-21T16:49:10.6972181Z Removing includeIf entries pointing to credentials config files
2026-04-21T16:49:10.6978036Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-21T16:49:10.6998898Z includeif.gitdir:/home/runner/work/nightingale/nightingale/.git.path
2026-04-21T16:49:10.6999726Z includeif.gitdir:/home/runner/work/nightingale/nightingale/.git/worktrees/*.path
2026-04-21T16:49:10.7000326Z includeif.gitdir:/github/workspace/.git.path
2026-04-21T16:49:10.7000839Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-21T16:49:10.7009219Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/nightingale/nightingale/.git.path
2026-04-21T16:49:10.7029144Z /home/runner/work/_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7037649Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/nightingale/nightingale/.git.path /home/runner/work/_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7067291Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/nightingale/nightingale/.git/worktrees/*.path
2026-04-21T16:49:10.7086415Z /home/runner/work/_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7094639Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/nightingale/nightingale/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7120486Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-21T16:49:10.7139133Z /github/runner_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7145939Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7171266Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-21T16:49:10.7188773Z /github/runner_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7195664Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config
2026-04-21T16:49:10.7222250Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-21T16:49:10.7408574Z Removing credentials config '/home/runner/work/_temp/git-credentials-f195a531-9b11-404e-ba68-22965cb9ad8a.config'
2026-04-21T16:49:10.7524887Z Cleaning up orphan processes
```

**Fixture:** `653c2a494c572de3356a...` conf=0.37
**Evidence:** ['<timestamp>.1156323Z ##[group]Run pnpm install --frozen-lockfile']
```
2026-03-10T12:38:53.3290786Z ##[error]src/tui/commands/sessions.ts(24,26): error TS7006: Parameter 's' implicitly has an 'any' type.
2026-03-10T12:38:53.3293218Z ##[error]src/tui/commands/sessions.ts(36,15): error TS7006: Parameter 'm' implicitly has an 'any' type.
2026-03-10T12:38:53.3295923Z ##[error]src/tui/commands/team.ts(7,15): error TS2305: Module '"../../core/types.js"' has no exported member 'AgentConfig'.
2026-03-10T12:38:53.3298620Z ##[error]src/vault/encrypted-store.ts(17,15): error TS2305: Module '"../core/types.js"' has no exported member 'VaultEntry'.
2026-03-10T12:38:53.3301164Z ##[error]src/vault/resolver.ts(10,15): error TS2305: Module '"../core/types.js"' has no exported member 'VaultEntry'.
2026-03-10T12:38:53.3303969Z ##[error]src/vault/resolver.ts(34,35): error TS2345: Argument of type 'unknown' is not assignable to parameter of type 'string'.
2026-03-10T12:38:53.3388923Z  ELIFECYCLE  Command failed with exit code 2.
2026-03-10T12:38:53.3588862Z ##[error]Process completed with exit code 2.
2026-03-10T12:38:53.3658136Z Post job cleanup.
2026-03-10T12:38:53.4248718Z Pruning is unnecessary.
2026-03-10T12:38:53.4376386Z Post job cleanup.
2026-03-10T12:38:53.5423734Z [command]/usr/bin/git version
2026-03-10T12:38:53.5464306Z git version 2.53.0
2026-03-10T12:38:53.5511533Z Temporarily overriding HOME='/home/runner/work/_temp/b0365e4f-b7b8-446e-b61c-70eab6a27617' before making global git config changes
2026-03-10T12:38:53.5513162Z Adding repository directory to the temporary git global config as a safe directory
2026-03-10T12:38:53.5526187Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/polpo/polpo
2026-03-10T12:38:53.5568703Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-10T12:38:53.5608999Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-10T12:38:53.5881115Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-10T12:38:53.5908062Z http.https://github.com/.extraheader
2026-03-10T12:38:53.5922455Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-10T12:38:53.5959965Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-10T12:38:53.6219841Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-10T12:38:53.6258456Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-10T12:38:53.6658104Z Cleaning up orphan processes
```

**Fixture:** `45f29cbb2a6649eb4e53...` conf=0.37
**Evidence:** ['<timestamp>.4204478Z ##[group]Run pnpm install --frozen-lockfile']
```
2026-03-09T16:10:26.8781284Z ##[error]src/tui/commands/sessions.ts(24,26): error TS7006: Parameter 's' implicitly has an 'any' type.
2026-03-09T16:10:26.8782591Z ##[error]src/tui/commands/sessions.ts(36,15): error TS7006: Parameter 'm' implicitly has an 'any' type.
2026-03-09T16:10:26.8784148Z ##[error]src/tui/commands/team.ts(7,15): error TS2305: Module '"../../core/types.js"' has no exported member 'AgentConfig'.
2026-03-09T16:10:26.8785737Z ##[error]src/vault/encrypted-store.ts(17,15): error TS2305: Module '"../core/types.js"' has no exported member 'VaultEntry'.
2026-03-09T16:10:26.8787192Z ##[error]src/vault/resolver.ts(10,15): error TS2305: Module '"../core/types.js"' has no exported member 'VaultEntry'.
2026-03-09T16:10:26.8788975Z ##[error]src/vault/resolver.ts(34,35): error TS2345: Argument of type 'unknown' is not assignable to parameter of type 'string'.
2026-03-09T16:10:26.8891090Z  ELIFECYCLE  Command failed with exit code 2.
2026-03-09T16:10:26.9078590Z ##[error]Process completed with exit code 2.
2026-03-09T16:10:26.9142669Z Post job cleanup.
2026-03-09T16:10:26.9683956Z Pruning is unnecessary.
2026-03-09T16:10:26.9799792Z Post job cleanup.
2026-03-09T16:10:27.0746325Z [command]/usr/bin/git version
2026-03-09T16:10:27.0783150Z git version 2.53.0
2026-03-09T16:10:27.0827364Z Temporarily overriding HOME='/home/runner/work/_temp/3fda6407-03d8-4d72-88ef-ca4f756fa3a1' before making global git config changes
2026-03-09T16:10:27.0829062Z Adding repository directory to the temporary git global config as a safe directory
2026-03-09T16:10:27.0841535Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/polpo/polpo
2026-03-09T16:10:27.0883163Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-09T16:10:27.0924276Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-09T16:10:27.1202485Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-09T16:10:27.1226620Z http.https://github.com/.extraheader
2026-03-09T16:10:27.1242400Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-09T16:10:27.1277520Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-09T16:10:27.1529690Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-09T16:10:27.1561820Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-09T16:10:27.1906847Z Cleaning up orphan processes
```

### `runtime-mismatch` — 1405 matches, 1261 low-conf (90%)

**Fixture:** `4bd377e5bf4b01b9cce7...` conf=0.4
**Evidence:** ["<timestamp>.7689216Z updater | <timestamp> INFO <job_1277027779> Dependabot is using Python version '3.10.19'."]
```
2026-03-12T20:15:59.2722512Z updater | 2026/03/12 20:15:59 INFO <job_1277027779> Finished job processing
2026-03-12T20:15:59.2731746Z updater | 2026/03/12 20:15:59 INFO Results:
2026-03-12T20:15:59.2732544Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-12T20:15:59.2733485Z +------------------------------------------------------------------------------------------------+
2026-03-12T20:15:59.2734352Z |                                 Dependencies failed to update                                  |
2026-03-12T20:15:59.2735200Z +------------+-----------------------+-----------------------------------------------------------+
2026-03-12T20:15:59.2736087Z | Dependency | Error Type            | Error Details                                             |
2026-03-12T20:15:59.2736968Z +------------+-----------------------+-----------------------------------------------------------+
2026-03-12T20:15:59.2737843Z | black      | illformed_requirement | {                                                         |
2026-03-12T20:15:59.2739125Z |            |                       |   "message": "Illformed requirement [\">= 3.10 | ^3.9\"]" |
2026-03-12T20:15:59.2740063Z |            |                       | }                                                         |
2026-03-12T20:15:59.2740826Z +------------+-----------------------+-----------------------------------------------------------+
2026-03-12T20:15:59.4055293Z Failure running container 5c9dd859197c9474cb32ee317ed8328949a8deec3fb93743a4b58f245bd7d5dc: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-12T20:15:59.7644560Z Cleaned up container 5c9dd859197c9474cb32ee317ed8328949a8deec3fb93743a4b58f245bd7d5dc
2026-03-12T20:15:59.7753095Z   proxy | 2026/03/12 20:15:59 Posting metrics to remote API endpoint
2026-03-12T20:15:59.7753807Z   proxy | 2026/03/12 20:15:59 0/13 calls cached (0%)
2026-03-12T20:16:00.1289529Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/bitkarrot/scheduler/network/updates/1277027779 (write access to the repository is required to view the log)
2026-03-12T20:16:00.1299873Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-12T20:16:00.1410515Z Post job cleanup.
2026-03-12T20:16:00.3108103Z Cleaning up orphan processes
2026-03-12T20:16:00.3614348Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `4db85feefa7c47778b9b...` conf=0.37
**Evidence:** ["<timestamp>.4818992Z updater | <timestamp> INFO <job_1282559222> Dependabot is using Python version '3.14.2'."]
```
2026-03-17T18:09:22.4973211Z updater | 2026/03/17 18:09:22 INFO <job_1282559222> Finished job processing
2026-03-17T18:09:22.4982367Z updater | 2026/03/17 18:09:22 INFO Results:
2026-03-17T18:09:22.4983476Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-17T18:09:22.4984832Z +-----------------------------------------------------------------------------------------------------------------------------------+
2026-03-17T18:09:22.4986190Z |                                                   Dependencies failed to update                                                   |
2026-03-17T18:09:22.4987204Z +------------+--------------------------+-------------------------------------------------------------------------------------------+
2026-03-17T18:09:22.4988475Z | Dependency | Error Type               | Error Details                                                                             |
2026-03-17T18:09:22.4989301Z +------------+--------------------------+-------------------------------------------------------------------------------------------+
2026-03-17T18:09:22.4990108Z | pyasn1     | private_source_timed_out | {                                                                                         |
2026-03-17T18:09:22.4990980Z |            |                          |   "source": "https://basic.artifactory.example.com/api/pypi/pypi-linuxbuild-prod/simple/" |
2026-03-17T18:09:22.4991742Z |            |                          | }                                                                                         |
2026-03-17T18:09:22.4992446Z +------------+--------------------------+-------------------------------------------------------------------------------------------+
2026-03-17T18:09:22.6209334Z Failure running container 3a48243aa09ea23dfb7e6e53ab7d9e640dacd004d908833fb4477567e0424884: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-17T18:09:22.7743769Z Cleaned up container 3a48243aa09ea23dfb7e6e53ab7d9e640dacd004d908833fb4477567e0424884
2026-03-17T18:09:22.7839099Z   proxy | 2026/03/17 18:09:22 0/19 calls cached (0%)
2026-03-17T18:09:22.7839976Z 2026/03/17 18:09:22 Posting metrics to remote API endpoint
2026-03-17T18:09:23.1761167Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/pulp/Pulp-manager/network/updates/1282559222 (write access to the repository is required to view the log)
2026-03-17T18:09:23.1772769Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-17T18:09:23.1867956Z Post job cleanup.
2026-03-17T18:09:23.3517822Z Cleaning up orphan processes
2026-03-17T18:09:23.4012351Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `ff2c0ff6c048ebf6cd2c...` conf=0.4
**Evidence:** ["<timestamp>.7139783Z updater | <timestamp> INFO <job_1289786409> Dependabot is using Python version '3.14.2'."]
```
2026-03-23T19:51:47.6416993Z   proxy | 2026/03/23 19:51:47 [022] PATCH /update_jobs/1289786409/mark_as_processed
2026-03-23T19:51:47.7191692Z   proxy | 2026/03/23 19:51:47 [022] 204 /update_jobs/1289786409/mark_as_processed
2026-03-23T19:51:47.7240647Z updater | 2026/03/23 19:51:47 INFO <job_1289786409> Finished job processing
2026-03-23T19:51:47.7283183Z updater | 2026/03/23 19:51:47 INFO Results:
2026-03-23T19:51:47.7283981Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-23T19:51:47.7284739Z +--------------------------------+
2026-03-23T19:51:47.7285149Z |             Errors             |
2026-03-23T19:51:47.7285521Z +----------------------+---------+
2026-03-23T19:51:47.7285916Z | Type                 | Details |
2026-03-23T19:51:47.7286292Z +----------------------+---------+
2026-03-23T19:51:47.7286701Z | all_versions_ignored | null    |
2026-03-23T19:51:47.7287159Z +----------------------+---------+
2026-03-23T19:51:47.8541905Z Failure running container 4a0bc6f1ae6ee28a50f4bf30771f2157d487fd8e67f6f0abce817f25e258384c: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-23T19:51:48.0661807Z Cleaned up container 4a0bc6f1ae6ee28a50f4bf30771f2157d487fd8e67f6f0abce817f25e258384c
2026-03-23T19:51:48.0755270Z   proxy | 2026/03/23 19:51:48 0/11 calls cached (0%)
2026-03-23T19:51:48.0756284Z 2026/03/23 19:51:48 Posting metrics to remote API endpoint
2026-03-23T19:51:48.4433609Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/wazo-platform/wazo-confd/network/updates/1289786409 (write access to the repository is required to view the log)
2026-03-23T19:51:48.4445096Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-23T19:51:48.4525625Z Post job cleanup.
2026-03-23T19:51:48.6179407Z Cleaning up orphan processes
2026-03-23T19:51:48.6641116Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `arch-mismatch` — 896 matches, 896 low-conf (100%)

**Fixture:** `63122d11433ba856fcdb...` conf=0.39
**Evidence:** ["<timestamp>.4639650Z Download action repository 'docker/setup-qemu-action@v3' (SHA:c7c53464625b32c7a7e944ae62b3e17d2b600130)", '<timestamp>.0854173Z "qemu-aarch64",']
```
2026-02-21T23:09:54.8186645Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-21T23:09:54.8423334Z Removing HTTP extra header
2026-02-21T23:09:54.8428445Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-21T23:09:54.8464838Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-21T23:09:54.8694856Z Removing includeIf entries pointing to credentials config files
2026-02-21T23:09:54.8701683Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-21T23:09:54.8726088Z includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path
2026-02-21T23:09:54.8726832Z includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path
2026-02-21T23:09:54.8727580Z includeif.gitdir:/github/workspace/.git.path
2026-02-21T23:09:54.8728184Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-21T23:09:54.8737567Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path
2026-02-21T23:09:54.8760155Z /home/runner/work/_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8771828Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path /home/runner/work/_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8806369Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path
2026-02-21T23:09:54.8829045Z /home/runner/work/_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8839365Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8872672Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-21T23:09:54.8896746Z /github/runner_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8905550Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8938243Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-21T23:09:54.8959328Z /github/runner_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.8968800Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config
2026-02-21T23:09:54.9000689Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-21T23:09:54.9227608Z Removing credentials config '/home/runner/work/_temp/git-credentials-ed8dc429-870e-43e0-aa05-093f4525c41c.config'
2026-02-21T23:09:54.9352580Z Cleaning up orphan processes
```

**Fixture:** `8355b3a4207aaa039f48...` conf=0.37
**Evidence:** ['<timestamp>.3262013Z /home/runner/work/stepflow/stepflow/examples/kubernetes-batch-demo/config/lima-k3s-qemu.yaml: 0/0 passed']
```
2026-03-18T22:36:16.6649727Z {"timestamp": "2026-03-18T22:36:16.660718+00:00", "level": "INFO", "message": "Graceful shutdown: no in-flight tasks", "logger": "stepflow_py.worker.grpc_worker", "module": "grpc_worker", "function": "_graceful_shutdown", "line": 631}
2026-03-18T22:36:16.9812332Z   Workflow tests       ✗
2026-03-18T22:36:16.9823731Z     Fix:   Fix failing workflow tests
2026-03-18T22:36:16.9825284Z 
2026-03-18T22:36:16.9825721Z ❌ 1 check failed
2026-03-18T22:36:16.9825958Z 
2026-03-18T22:36:16.9826099Z   Workflow tests:
2026-03-18T22:36:16.9827339Z     Check: /home/runner/work/stepflow/stepflow/stepflow-rs/target/debug/stepflow test /home/runner/work/stepflow/stepflow/tests /home/runner/work/stepflow/stepflow/examples
2026-03-18T22:36:16.9862808Z ##[error]Process completed with exit code 1.
2026-03-18T22:36:16.9963653Z Post job cleanup.
2026-03-18T22:36:17.0922343Z [command]/usr/bin/git version
2026-03-18T22:36:17.0966020Z git version 2.53.0
2026-03-18T22:36:17.1007374Z Temporarily overriding HOME='/home/runner/work/_temp/ca88fa01-48f7-4687-b43c-cdfac4af062d' before making global git config changes
2026-03-18T22:36:17.1008774Z Adding repository directory to the temporary git global config as a safe directory
2026-03-18T22:36:17.1012672Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/stepflow/stepflow
2026-03-18T22:36:17.1047422Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-18T22:36:17.1082024Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-18T22:36:17.1325157Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-18T22:36:17.1347079Z http.https://github.com/.extraheader
2026-03-18T22:36:17.1360134Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-18T22:36:17.1391959Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-18T22:36:17.1625137Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-18T22:36:17.1657600Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-18T22:36:17.2017412Z Cleaning up orphan processes
2026-03-18T22:36:17.2324018Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, arduino/setup-protoc@v3, astral-sh/setup-uv@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `148733b5b094fd628baf...` conf=0.39
**Evidence:** ["<timestamp>.6172252Z Download action repository 'docker/setup-qemu-action@ce360397dd3f832beb865e1373c09c0e9f86d70a' (SHA:ce360397dd3f832beb865e1373c09c0e9f86d70a)", '<timestamp>.0308650Z "qemu-aarch64",']
```
2026-04-16T13:44:26.0551147Z ##[group]Post cache
2026-04-16T13:44:26.0552460Z State not set
2026-04-16T13:44:26.0553390Z ##[endgroup]
2026-04-16T13:44:26.0694126Z Post job cleanup.
2026-04-16T13:44:26.2077927Z ##[group]Post cache
2026-04-16T13:44:26.2079560Z Caching docker.io--tonistiigi--binfmt-latest-linux-x64 to GitHub Actions cache
2026-04-16T13:44:26.2241811Z [command]/usr/bin/tar --posix -cf cache.tzst --exclude cache.tzst -P -C /home/runner/work/docker-kali/docker-kali --files-from manifest.txt --use-compress-program zstdmt
2026-04-16T13:44:27.2957894Z Sent 32159220 of 32159220 (100.0%), 42.8 MBs/sec
2026-04-16T13:44:27.4839786Z ##[endgroup]
2026-04-16T13:44:27.4984101Z Post job cleanup.
2026-04-16T13:44:27.5789323Z [command]/usr/bin/git version
2026-04-16T13:44:27.5860040Z git version 2.53.0
2026-04-16T13:44:27.5901346Z Temporarily overriding HOME='/home/runner/work/_temp/ef4fd549-2ff6-4a8d-a54f-173ffe6be744' before making global git config changes
2026-04-16T13:44:27.5902597Z Adding repository directory to the temporary git global config as a safe directory
2026-04-16T13:44:27.5911153Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/docker-kali/docker-kali
2026-04-16T13:44:27.5944988Z Removing SSH command configuration
2026-04-16T13:44:27.5953490Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-16T13:44:27.5993262Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-16T13:44:27.6239959Z Removing HTTP extra header
2026-04-16T13:44:27.6245589Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-16T13:44:27.6280082Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-16T13:44:27.6507846Z Removing includeIf entries pointing to credentials config files
2026-04-16T13:44:27.6514828Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-16T13:44:27.6549439Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-16T13:44:27.6922043Z Cleaning up orphan processes
```

### `database-test-isolation` — 648 matches, 647 low-conf (100%)

**Fixture:** `097a47d3975fbc21d543...` conf=0.32
**Evidence:** ['<timestamp>.5914834Z ReleaseAsset.name already exists']
```
2026-04-05T23:18:21.1698228Z [36;1m  fi[0m
2026-04-05T23:18:21.1698426Z [36;1mfi[0m
2026-04-05T23:18:21.1723421Z shell: /usr/bin/bash -e {0}
2026-04-05T23:18:21.1723701Z env:
2026-04-05T23:18:21.1724203Z   GITHUB_TOKEN: ***
2026-04-05T23:18:21.1724435Z ##[endgroup]
2026-04-05T23:18:29.5913228Z HTTP 422: Validation Failed (https://uploads.github.com/repos/BetterSEQTA/DesQTA/releases/305489767/assets?label=&name=DesQTA.app.tar.gz.sig)
2026-04-05T23:18:29.5914834Z ReleaseAsset.name already exists
2026-04-05T23:18:29.5957605Z ##[error]Process completed with exit code 1.
2026-04-05T23:18:29.6071702Z Post job cleanup.
2026-04-05T23:18:29.7010775Z [command]/usr/bin/git version
2026-04-05T23:18:29.7047671Z git version 2.53.0
2026-04-05T23:18:29.7090535Z Temporarily overriding HOME='/home/runner/work/_temp/c9dc4a8f-1493-4fbc-bdb4-58f291df6e8a' before making global git config changes
2026-04-05T23:18:29.7092161Z Adding repository directory to the temporary git global config as a safe directory
2026-04-05T23:18:29.7096217Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/DesQTA/DesQTA
2026-04-05T23:18:29.7130230Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-05T23:18:29.7162155Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-05T23:18:29.7379279Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-05T23:18:29.7399008Z http.https://github.com/.extraheader
2026-04-05T23:18:29.7410867Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-05T23:18:29.7439718Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-05T23:18:29.7654617Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-05T23:18:29.7683832Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-05T23:18:29.8005306Z Cleaning up orphan processes
2026-04-05T23:18:29.8300399Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/download-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `25f821cc8d39f157db60...` conf=0.32
**Evidence:** ['<timestamp>.7886337Z Error: sql/migrate: read migration directory state: sql/migrate: executing statement "CREATE TABLE \\"public\\".\\"verification_claim\\" (\\n \\"id\\" uuid NOT NULL DEFAULT uuidv7(),\\n \\"parsed_prediction_id\\" uuid NOT NULL DEFAULT uuidv7(),\\n \\"verifier_agent_id\\" character varying(256) NOT NULL,\\n \\"verifier_agent_signature\\" text NOT NULL,\\n \\"claim_outcome\\" boolean NOT NULL,\\n \\"confidence\\" numeric NOT NULL,\\n \\"reasoning\\" text NOT NULL,\\n \\"sources\\" jsonb NULL,\\n \\"timeframe_start_utc\\" timestamptz NULL,\\n \\"timeframe_end_utc\\" timestamptz NULL,\\n \\"timeframe_precision\\" character varying(32) NULL,\\n \\"created_at\\" timestamptz NOT NULL DEFAULT now(),\\n \\"updated_at\\" timestamptz NOT NULL DEFAULT now(),\\n \\"deleted_at\\" timestamptz NULL,\\n PRIMARY KEY (\\"id\\"),\\n CONSTRAINT \\"verification_claim_unique_verifier\\" UNIQUE (\\"parsed_prediction_id\\", \\"verifier_agent_id\\"),\\n CONSTRAINT \\"verification_claim_parsed_prediction_id_parsed_prediction_id_fk\\" FOREIGN KEY (\\"parsed_prediction_id\\") REFERENCES \\"public\\".\\"parsed_prediction\\" (\\"id\\") ON UPDATE NO ACTION ON DELETE CASCADE\\n);" from version "20260105202138": pq: relation "verification_claim" already exists']
```
2026-02-02T21:26:46.9146448Z sed -E 's#(postgresql://)[^:@]+:[^@]+@#\1****:****@#' migration_status_prediction_swarm.txt >> /home/runner/work/_temp/_runner_file_commands/step_summary_4bfdc780-2d61-4c5c-97e4-299407090a18
2026-02-02T21:26:46.9149980Z echo : No such file or directory
2026-02-02T21:26:46.9192385Z Post job cleanup.
2026-02-02T21:26:46.9215207Z Post job cleanup.
2026-02-02T21:26:46.9645399Z Pruning is unnecessary.
2026-02-02T21:26:46.9726781Z Post job cleanup.
2026-02-02T21:26:47.0498255Z [command]/usr/bin/git version
2026-02-02T21:26:47.0545445Z git version 2.52.0
2026-02-02T21:26:47.0579749Z Temporarily overriding HOME='/home/runner/work/_temp/e95f750a-60bb-4a7b-8827-2f3b255df85f' before making global git config changes
2026-02-02T21:26:47.0581991Z Adding repository directory to the temporary git global config as a safe directory
2026-02-02T21:26:47.0585951Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/torus-ts/torus-ts
2026-02-02T21:26:47.0638891Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-02T21:26:47.0679893Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-02T21:26:47.1006643Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-02T21:26:47.1026650Z http.https://github.com/.extraheader
2026-02-02T21:26:47.1038978Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-02T21:26:47.1077170Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-02T21:26:47.1289409Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-02T21:26:47.1315591Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-02T21:26:47.1596054Z A job completed hook has been configured by the self-hosted runner administrator
2026-02-02T21:26:47.1656810Z ##[group]Run '/home/runner/actions-runner/complete-hook.sh'
2026-02-02T21:26:47.1714024Z shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
2026-02-02T21:26:47.1714335Z ##[endgroup]
2026-02-02T21:26:47.1841503Z Evaluate and set job outputs
2026-02-02T21:26:47.1846345Z Cleaning up orphan processes
```

**Fixture:** `272fd18a3c4d1687c052...` conf=0.32
**Evidence:** ['<timestamp>.4353347Z updater | <timestamp> INFO <job_1297242912> Pull request already exists for handlebars with latest version 4.7.9']
```
2026-03-28T00:14:15.9391872Z updater | 2026/03/28 00:14:15 INFO Results:
2026-03-28T00:14:15.9392947Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-28T00:14:15.9393943Z +------------------------------------------------------------------------------------------+
2026-03-28T00:14:15.9394704Z |                              Dependencies failed to update                               |
2026-03-28T00:14:15.9395388Z +------------+----------------------------------------+------------------------------------+
2026-03-28T00:14:15.9396509Z | Dependency | Error Type                             | Error Details                      |
2026-03-28T00:14:15.9397202Z +------------+----------------------------------------+------------------------------------+
2026-03-28T00:14:15.9397915Z | handlebars | pull_request_exists_for_latest_version | {                                  |
2026-03-28T00:14:15.9398695Z |            |                                        |   "dependency-name": "handlebars", |
2026-03-28T00:14:15.9399279Z |            |                                        |   "dependency-version": "4.7.9"    |
2026-03-28T00:14:15.9399832Z |            |                                        | }                                  |
2026-03-28T00:14:15.9400480Z +------------+----------------------------------------+------------------------------------+
2026-03-28T00:14:16.0650112Z Failure running container e1b265ecb1be744863d5f3f8e90c5dcb9c158cfd91fcaf129cac6565ac9b3b82: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-28T00:14:16.1615260Z Cleaned up container e1b265ecb1be744863d5f3f8e90c5dcb9c158cfd91fcaf129cac6565ac9b3b82
2026-03-28T00:14:16.1697748Z   proxy | 2026/03/28 00:14:16 0/16 calls cached (0%)
2026-03-28T00:14:16.1698773Z 2026/03/28 00:14:16 Posting metrics to remote API endpoint
2026-03-28T00:14:16.5361416Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/kodustech/kodus-service-ast/network/updates/1297242912 (write access to the repository is required to view the log)
2026-03-28T00:14:16.5372809Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-28T00:14:16.5522351Z Post job cleanup.
2026-03-28T00:14:16.7128536Z Cleaning up orphan processes
2026-03-28T00:14:16.7594721Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `poetry-lockfile-drift` — 545 matches, 538 low-conf (99%)

**Fixture:** `e124f0e24525298e7eaa...` conf=0.37
**Evidence:** ['<timestamp>.2481135Z Installing dependencies from lock file (including require-dev)']
```
2026-02-23T11:46:48.1448288Z  <error line="21" column="5" severity="error" message="MissingOverrideAttribute: Method Spiral\Stempler\Transform\Visitor\DefineHidden::leavenode should have the &quot;Override&quot; attribute"/>
2026-02-23T11:46:48.1449491Z </file>
2026-02-23T11:46:48.1449803Z <file name="src/Transform/Visitor/DefineStacks.php">
2026-02-23T11:46:48.1450685Z  <error line="20" column="5" severity="error" message="MissingOverrideAttribute: Method Spiral\Stempler\Transform\Visitor\DefineStacks::enternode should have the &quot;Override&quot; attribute"/>
2026-02-23T11:46:48.1451442Z </file>
2026-02-23T11:46:48.1451877Z <file name="src/Transform/Visitor/DefineStacks.php">
2026-02-23T11:46:48.1452732Z  <error line="29" column="5" severity="error" message="MissingOverrideAttribute: Method Spiral\Stempler\Transform\Visitor\DefineStacks::leavenode should have the &quot;Override&quot; attribute"/>
2026-02-23T11:46:48.1453591Z </file>
2026-02-23T11:46:48.1510064Z </checkstyle>
2026-02-23T11:46:49.9728833Z 🐑 results sent to shepherd.dev 🐑
2026-02-23T11:46:50.0226204Z ##[error]Process completed with exit code 2.
2026-02-23T11:46:50.0289591Z Post job cleanup.
2026-02-23T11:46:50.1026634Z [command]/usr/bin/git version
2026-02-23T11:46:50.1062895Z git version 2.52.0
2026-02-23T11:46:50.1097427Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/433c3e17-7912-49fb-a198-2cd42d9b1cef/.gitconfig'
2026-02-23T11:46:50.1107077Z Temporarily overriding HOME='/home/runner/work/_temp/433c3e17-7912-49fb-a198-2cd42d9b1cef' before making global git config changes
2026-02-23T11:46:50.1108200Z Adding repository directory to the temporary git global config as a safe directory
2026-02-23T11:46:50.1111696Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/stempler/stempler
2026-02-23T11:46:50.1143201Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-23T11:46:50.1172995Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-23T11:46:50.1464229Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-23T11:46:50.1484327Z http.https://github.com/.extraheader
2026-02-23T11:46:50.1495690Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-23T11:46:50.1525278Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-23T11:46:50.1864315Z Cleaning up orphan processes
```

**Fixture:** `a874f3b971f83c218f99...` conf=0.37
**Evidence:** ['<timestamp>.2956113Z Installing dependencies from lock file (including require-dev)']
```
2026-02-02T19:13:36.9435885Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-02T19:13:36.9656279Z Removing HTTP extra header
2026-02-02T19:13:36.9661166Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-02T19:13:36.9693458Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-02T19:13:36.9907964Z Removing includeIf entries pointing to credentials config files
2026-02-02T19:13:36.9914796Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-02T19:13:36.9936696Z includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git.path
2026-02-02T19:13:36.9938257Z includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git/worktrees/*.path
2026-02-02T19:13:36.9939388Z includeif.gitdir:/github/workspace/.git.path
2026-02-02T19:13:36.9940186Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-02T19:13:36.9947386Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git.path
2026-02-02T19:13:36.9968047Z /home/runner/work/_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:36.9979276Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git.path /home/runner/work/_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0011597Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git/worktrees/*.path
2026-02-02T19:13:37.0032211Z /home/runner/work/_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0041477Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0070630Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-02T19:13:37.0091511Z /github/runner_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0099715Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0130417Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-02T19:13:37.0150446Z /github/runner_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0160320Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config
2026-02-02T19:13:37.0191627Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-02T19:13:37.0401900Z Removing credentials config '/home/runner/work/_temp/git-credentials-53f0f797-101d-4cf4-83b5-60f857b81ee5.config'
2026-02-02T19:13:37.0518348Z Cleaning up orphan processes
```

**Fixture:** `0c7dc02ac7f02d1d4021...` conf=0.37
**Evidence:** ['<timestamp>.4041118Z Installing dependencies from lock file (including require-dev)']
```
2026-03-16T18:02:32.3163514Z      40▕     }
2026-03-16T18:02:32.3163697Z      41▕ }
2026-03-16T18:02:32.3163787Z 
2026-03-16T18:02:32.3163938Z       [2m+1 vendor frames [22m
2026-03-16T18:02:32.3164157Z   2   [internal]:0
2026-03-16T18:02:32.3164385Z       Whoops\Run::handleShutdown()
2026-03-16T18:02:32.3246805Z Script @php vendor/bin/testbench package:discover --ansi handling the prepare event returned with error code 255
2026-03-16T18:02:32.3405857Z Script @composer run prepare handling the post-autoload-dump event returned with error code 255
2026-03-16T18:02:32.3919851Z ##[error]Process completed with exit code 255.
2026-03-16T18:02:32.4036980Z Post job cleanup.
2026-03-16T18:02:32.5054619Z [command]/usr/bin/git version
2026-03-16T18:02:32.5094179Z git version 2.53.0
2026-03-16T18:02:32.5140150Z Temporarily overriding HOME='/home/runner/work/_temp/695a3a0d-9ed6-43fa-bd17-92d1815d013a' before making global git config changes
2026-03-16T18:02:32.5141876Z Adding repository directory to the temporary git global config as a safe directory
2026-03-16T18:02:32.5146768Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/laravel-geetest-captcha/laravel-geetest-captcha
2026-03-16T18:02:32.5187890Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-16T18:02:32.5224200Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-16T18:02:32.5481643Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-16T18:02:32.5507544Z http.https://github.com/.extraheader
2026-03-16T18:02:32.5524209Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-16T18:02:32.5560405Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-16T18:02:32.5816637Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-16T18:02:32.5856900Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-16T18:02:32.6231909Z Cleaning up orphan processes
2026-03-16T18:02:32.6563284Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `node-version-mismatch` — 508 matches, 508 low-conf (100%)

**Fixture:** `145e8b699ade2beecb26...` conf=0.37
**Evidence:** ['<timestamp>.1645524Z npm WARN EBADENGINE Unsupported engine {']
```
2026-03-30T04:33:58.9134233Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-30T04:33:58.9354978Z Removing HTTP extra header
2026-03-30T04:33:58.9360540Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-30T04:33:58.9395058Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-30T04:33:58.9612515Z Removing includeIf entries pointing to credentials config files
2026-03-30T04:33:58.9619812Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-30T04:33:58.9642744Z includeif.gitdir:/home/runner/work/error/error/.git.path
2026-03-30T04:33:58.9644081Z includeif.gitdir:/home/runner/work/error/error/.git/worktrees/*.path
2026-03-30T04:33:58.9644985Z includeif.gitdir:/github/workspace/.git.path
2026-03-30T04:33:58.9645862Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-03-30T04:33:58.9653717Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/error/error/.git.path
2026-03-30T04:33:58.9675240Z /home/runner/work/_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9685103Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/error/error/.git.path /home/runner/work/_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9718408Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/error/error/.git/worktrees/*.path
2026-03-30T04:33:58.9740705Z /home/runner/work/_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9749762Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/error/error/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9781443Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-03-30T04:33:58.9803376Z /github/runner_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9812295Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9843381Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-03-30T04:33:58.9864455Z /github/runner_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9874180Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config
2026-03-30T04:33:58.9905473Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-30T04:33:59.0127239Z Removing credentials config '/home/runner/work/_temp/git-credentials-6052d1fe-9b32-4cc7-8d14-39355e328f24.config'
2026-03-30T04:33:59.0295143Z Cleaning up orphan processes
```

**Fixture:** `533888e93cc0ae40f50d...` conf=0.37
**Evidence:** ['<timestamp>.7620129Z error star-wars-wiki@2.1.0: The engine "node" is incompatible with this module. Expected version ">=16.12.0". Got "12.22.12"']
```
2026-04-08T19:43:32.4788164Z [36;1myarn build[0m
2026-04-08T19:43:32.4788366Z [36;1myarn start:ci[0m
2026-04-08T19:43:32.4788574Z [36;1myarn cypress:run[0m
2026-04-08T19:43:32.4816327Z shell: /usr/bin/bash -e {0}
2026-04-08T19:43:32.4816597Z ##[endgroup]
2026-04-08T19:43:32.6974554Z yarn install v1.22.22
2026-04-08T19:43:32.7598445Z [1/5] Validating package.json...
2026-04-08T19:43:32.7620129Z error star-wars-wiki@2.1.0: The engine "node" is incompatible with this module. Expected version ">=16.12.0". Got "12.22.12"
2026-04-08T19:43:32.7689517Z error Found incompatible module.
2026-04-08T19:43:32.7690606Z info Visit https://yarnpkg.com/en/docs/cli/install for documentation about this command.
2026-04-08T19:43:32.7790522Z ##[error]Process completed with exit code 1.
2026-04-08T19:43:32.7892905Z Post job cleanup.
2026-04-08T19:43:32.8793304Z [command]/usr/bin/git version
2026-04-08T19:43:32.8839025Z git version 2.53.0
2026-04-08T19:43:32.8882721Z Temporarily overriding HOME='/home/runner/work/_temp/d5ac3e18-3ffa-4c46-abe5-7fc9f9800978' before making global git config changes
2026-04-08T19:43:32.8883856Z Adding repository directory to the temporary git global config as a safe directory
2026-04-08T19:43:32.8886177Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/carbonite/carbonite
2026-04-08T19:43:32.8918059Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-08T19:43:32.8948245Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-08T19:43:32.9163322Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-08T19:43:32.9183784Z http.https://github.com/.extraheader
2026-04-08T19:43:32.9193818Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-08T19:43:32.9223775Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-08T19:43:32.9556242Z Cleaning up orphan processes
2026-04-08T19:43:32.9816686Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v2, actions/setup-node@v2. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `707224fb904a824c94be...` conf=0.37
**Evidence:** ['<timestamp>.4980017Z npm warn EBADENGINE Unsupported engine {']
```
2026-03-07T16:52:29.5041216Z Uploaded bytes 16777216
2026-03-07T16:52:29.6952398Z Uploaded bytes 25165824
2026-03-07T16:52:30.0187272Z Uploaded bytes 33554432
2026-03-07T16:52:30.2812623Z Uploaded bytes 40274568
2026-03-07T16:52:30.2940295Z Finished uploading artifact content to blob storage!
2026-03-07T16:52:30.2943720Z SHA256 digest of uploaded artifact zip is ebb5efd7625f0e86d2038ad5430fc566bba134b95b052410870f26b6bc2aea97
2026-03-07T16:52:30.2945953Z Finalizing artifact upload
2026-03-07T16:52:30.3625760Z Artifact playwright-artifacts.zip successfully finalized. Artifact ID 5812283687
2026-03-07T16:52:30.3627840Z Artifact playwright-artifacts has been successfully uploaded! Final size is 40274568 bytes. Artifact ID is 5812283687
2026-03-07T16:52:30.3635085Z Artifact download URL: https://github.com/trybick/tv-minder/actions/runs/22803057142/artifacts/5812283687
2026-03-07T16:52:30.3834509Z Post job cleanup.
2026-03-07T16:52:30.4768946Z [command]/usr/bin/git version
2026-03-07T16:52:30.4805806Z git version 2.53.0
2026-03-07T16:52:30.4847585Z Temporarily overriding HOME='/home/runner/work/_temp/87f75279-7f8f-47e8-99a8-65a94dacade9' before making global git config changes
2026-03-07T16:52:30.4848957Z Adding repository directory to the temporary git global config as a safe directory
2026-03-07T16:52:30.4859814Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/tv-minder/tv-minder
2026-03-07T16:52:30.4894724Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-07T16:52:30.4928103Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-07T16:52:30.5169346Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-07T16:52:30.5190167Z http.https://github.com/.extraheader
2026-03-07T16:52:30.5202304Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-07T16:52:30.5233265Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-07T16:52:30.5464310Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-07T16:52:30.5496511Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-07T16:52:30.5842086Z Cleaning up orphan processes
```

### `artifact-upload-failure` — 485 matches, 289 low-conf (60%)

**Fixture:** `edb45e1fa2b7f67cf70f...` conf=0.39
**Evidence:** ['<timestamp>.4879997Z ##[warning]No files were found with the provided path: temp/tests-report-fluvius_process.html. No artifacts will be uploaded.']
```
2026-03-15T03:45:07.7451350Z   done
2026-03-15T03:45:07.7451500Z  server started
2026-03-15T03:45:07.7451671Z  CREATE DATABASE
2026-03-15T03:45:07.7451826Z  
2026-03-15T03:45:07.7451964Z  
2026-03-15T03:45:07.7452260Z  /usr/local/bin/docker-entrypoint.sh: ignoring /docker-entrypoint-initdb.d/*
2026-03-15T03:45:07.7452616Z  
2026-03-15T03:45:07.7453109Z  2026-03-15 03:44:46.770 UTC [48] LOG:  received fast shutdown request
2026-03-15T03:45:07.7453617Z  waiting for server to shut down....2026-03-15 03:44:46.770 UTC [48] LOG:  aborting any active transactions
2026-03-15T03:45:07.7454289Z  2026-03-15 03:44:46.773 UTC [48] LOG:  background worker "logical replication launcher" (PID 55) exited with exit code 1
2026-03-15T03:45:07.7454829Z  2026-03-15 03:44:46.773 UTC [50] LOG:  shutting down
2026-03-15T03:45:07.7455163Z  2026-03-15 03:44:46.780 UTC [48] LOG:  database system is shut down
2026-03-15T03:45:07.7455467Z   done
2026-03-15T03:45:07.7455620Z  server stopped
2026-03-15T03:45:07.7455775Z  
2026-03-15T03:45:07.7455983Z  PostgreSQL init process complete; ready for start up.
2026-03-15T03:45:07.7456595Z  
2026-03-15T03:45:07.7461869Z Stop and remove container: fb8cb79fe7b74ad49efe3d90d70bed0b_postgres14_6677c8
2026-03-15T03:45:07.7467745Z ##[command]/usr/bin/docker rm --force 33b99de683ed8d6467adbc2039bfdfa3d74e5c99fc85dc1b5000a0021ece91e1
2026-03-15T03:45:08.4963200Z 33b99de683ed8d6467adbc2039bfdfa3d74e5c99fc85dc1b5000a0021ece91e1
2026-03-15T03:45:08.4995183Z Remove container network: github_network_8751b85002b34c339281de64d87e48f5
2026-03-15T03:45:08.4999747Z ##[command]/usr/bin/docker network rm github_network_8751b85002b34c339281de64d87e48f5
2026-03-15T03:45:08.6583906Z github_network_8751b85002b34c339281de64d87e48f5
2026-03-15T03:45:08.6645278Z Cleaning up orphan processes
2026-03-15T03:45:08.7078428Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-python@v5, actions/upload-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `b503dcf89f7c0d2a4fd2...` conf=0.39
**Evidence:** ['<timestamp>.1294247Z ##[warning]No files were found with the provided path: ./build/reports/ci-info.txt', '**/build/test-results/**/*. No artifacts will be uploaded.']
```
2026-02-21T22:52:21.1294247Z ##[warning]No files were found with the provided path: ./build/reports/ci-info.txt
**/build/reports/**/*
**/build/test-results/**/*. No artifacts will be uploaded.
2026-02-21T22:52:21.1499554Z Post job cleanup.
2026-02-21T22:52:21.4694401Z In post-action step
2026-02-21T22:52:21.4703488Z Cache is read-only: will not save state for use in subsequent builds.
2026-02-21T22:52:21.4709416Z Generating Job Summary
2026-02-21T22:52:21.4731392Z Completed post-action step
2026-02-21T22:52:21.4870518Z Post job cleanup.
2026-02-21T22:52:21.6674496Z Post job cleanup.
2026-02-21T22:52:21.7623500Z [command]/usr/bin/git version
2026-02-21T22:52:21.7660052Z git version 2.52.0
2026-02-21T22:52:21.7705225Z Temporarily overriding HOME='/home/runner/work/_temp/cafd9ea3-b43a-49f0-bb7b-47d211b3ccc2' before making global git config changes
2026-02-21T22:52:21.7706596Z Adding repository directory to the temporary git global config as a safe directory
2026-02-21T22:52:21.7719566Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/backend/backend
2026-02-21T22:52:21.7755527Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-21T22:52:21.7791915Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-21T22:52:21.8080756Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-21T22:52:21.8106996Z http.https://github.com/.extraheader
2026-02-21T22:52:21.8121283Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-21T22:52:21.8154102Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-21T22:52:21.8401687Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-21T22:52:21.8435689Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-21T22:52:21.8768942Z Cleaning up orphan processes
2026-02-21T22:52:21.9051969Z Terminate orphan process: pid (2427) (java)
```

**Fixture:** `ff759bbde651c280b248...` conf=0.37
**Evidence:** ['<timestamp>.4385437Z Download artifact has finished successfully']
```
2026-03-31T10:29:59.5061481Z [36;1mfi[0m
2026-03-31T10:29:59.5061723Z [36;1mecho "COVERAGE=${coverage%.*}%" >> $GITHUB_ENV[0m
2026-03-31T10:29:59.5062069Z [36;1mecho "COLOR=$COLOR" >> $GITHUB_ENV[0m
2026-03-31T10:29:59.5091963Z shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
2026-03-31T10:29:59.5092317Z env:
2026-03-31T10:29:59.5092550Z   REPO_NAME: DialogueKit
2026-03-31T10:29:59.5092789Z ##[endgroup]
2026-03-31T10:29:59.5693498Z ##[group]Run schneegans/dynamic-badges-action@v1.7.0
2026-03-31T10:29:59.5693836Z with:
2026-03-31T10:29:59.5694229Z   auth: ***
2026-03-31T10:29:59.5694440Z   gistID: 35bb996459f0949b38da651c66cf95cb
2026-03-31T10:29:59.5694745Z   filename: coverage.DialogueKit.278.json
2026-03-31T10:29:59.5695018Z   label: coverage
2026-03-31T10:29:59.5695209Z   message: 88%
2026-03-31T10:29:59.5695392Z   color: red
2026-03-31T10:29:59.5695602Z   host: https://api.github.com/gists/
2026-03-31T10:29:59.5695867Z   forceUpdate: false
2026-03-31T10:29:59.5696070Z env:
2026-03-31T10:29:59.5696244Z   REPO_NAME: DialogueKit
2026-03-31T10:29:59.5696457Z   COVERAGE: 88%
2026-03-31T10:29:59.5696638Z   COLOR: red
2026-03-31T10:29:59.5696823Z ##[endgroup]
2026-03-31T10:29:59.8371215Z ##[error]Failed to get gist: 401 Unauthorized
2026-03-31T10:29:59.9212627Z Cleaning up orphan processes
2026-03-31T10:29:59.9532609Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/download-artifact@v4, schneegans/dynamic-badges-action@v1.7.0. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `ipv6-ipv4-resolution` — 334 matches, 298 low-conf (89%)

**Fixture:** `c388c445ef7bec899aa1...` conf=0.46
**Evidence:** ['<timestamp>.2818914Z * IPv6: 2600:9000:201f:c00:c:7ed3:240:93a1, 2600:9000:201f:d200:c:7ed3:240:93a1, 2600:9000:201f:9c00:c:7ed3:240:93a1, 2600:9000:201f:aa00:c:7ed3:240:93a1, 2600:9000:201f:5400:c:7ed3:240:93a1, 2600:9000:201f:9600:c:7ed3:240:93a1, 2600:9000:201f:d800:c:7ed3:240:93a1, 2600:9000:201f:7a00:c:7ed3:240:93a1']
```
2026-02-10T03:09:43.1031354Z [32;1m        PASS[0m [   0.005s] (838/838) [35;1merltf_serde::serde_tests[0m [34;1mtest_u64_max_roundtrip[0m
2026-02-10T03:09:43.1032140Z ────────────
2026-02-10T03:09:43.1033063Z [31;1m     Summary[0m [   4.238s] [1m838[0m tests run: [1m834[0m [32;1mpassed[0m, [1m4[0m [31;1mfailed[0m, [1m0[0m [33;1mskipped[0m
2026-02-10T03:09:43.1034405Z [31;1m        FAIL[0m [   0.575s] (318/838) [35;1medp_examples_elixir::test_elixir_interop[0m [34;1mtest_elixir_echo_atom[0m
2026-02-10T03:09:43.1035543Z [31;1m        FAIL[0m [   0.580s] (319/838) [35;1medp_examples_elixir::test_elixir_interop[0m [34;1mtest_elixir_echo_option_string[0m
2026-02-10T03:09:43.1036399Z [31;1m        FAIL[0m [   0.593s] (320/838) [35;1medp_examples_elixir::test_elixir_interop[0m [34;1mtest_elixir_echo_nil[0m
2026-02-10T03:09:43.1037204Z [31;1m        FAIL[0m [   0.585s] (330/838) [35;1medp_examples_elixir::test_elixir_interop[0m [34;1mtest_elixir_echo_some[0m
2026-02-10T03:09:43.1043802Z [31;1merror[0m: test run failed
2026-02-10T03:09:43.1067853Z ##[error]Process completed with exit code 100.
2026-02-10T03:09:43.1144492Z Post job cleanup.
2026-02-10T03:09:43.2071966Z [command]/usr/bin/git version
2026-02-10T03:09:43.2107001Z git version 2.52.0
2026-02-10T03:09:43.2150522Z Temporarily overriding HOME='/home/runner/work/_temp/f8e3f0e3-6d7e-4375-bb82-3b82d2fd6744' before making global git config changes
2026-02-10T03:09:43.2151857Z Adding repository directory to the temporary git global config as a safe directory
2026-02-10T03:09:43.2156856Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/edp-rs/edp-rs
2026-02-10T03:09:43.2192256Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-10T03:09:43.2224458Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-10T03:09:43.2451908Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-10T03:09:43.2472015Z http.https://github.com/.extraheader
2026-02-10T03:09:43.2484526Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-10T03:09:43.2515101Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-10T03:09:43.2732927Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-10T03:09:43.2766038Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-10T03:09:43.3101043Z Cleaning up orphan processes
2026-02-10T03:09:43.3363250Z Terminate orphan process: pid (2328) (epmd)
```

**Fixture:** `f1cbb38090537205b998...` conf=0.46
**Evidence:** ['<timestamp>.5106380Z * IPv6: 2a04:4e42:83::347']
```
2026-04-21T04:05:39.9673140Z copying path '/nix/store/ab3753m6i7isgvzphlar0a8xb84gl96i-gcc-15.2.0-lib' from 'https://cache.nixos.org'...
2026-04-21T04:05:40.0647173Z copying path '/nix/store/cynd1a544hd6j4s44xawchgdki5b4xwb-alejandra-4.0.0' from 'https://cache.nixos.org'...
2026-04-21T04:05:40.1167554Z Checking style in 43 files using 2 threads.
2026-04-21T04:05:40.1168580Z 
2026-04-21T04:05:40.1270041Z Requires formatting: ./machines/framework/configuration.nix
2026-04-21T04:05:40.1305141Z Requires formatting: ./scripts/tmux-switch-ssh-session.nix
2026-04-21T04:05:40.1323165Z 
2026-04-21T04:05:40.1323552Z Alert! 2 files require formatting.
2026-04-21T04:05:40.1340285Z ##[error]Process completed with exit code 2.
2026-04-21T04:05:40.1458960Z Post job cleanup.
2026-04-21T04:05:40.2380847Z [command]/usr/bin/git version
2026-04-21T04:05:40.2417169Z git version 2.53.0
2026-04-21T04:05:40.2460579Z Temporarily overriding HOME='/home/runner/work/_temp/149e9f28-e58c-438e-bdd1-bce8010c4cbb' before making global git config changes
2026-04-21T04:05:40.2461937Z Adding repository directory to the temporary git global config as a safe directory
2026-04-21T04:05:40.2467346Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/nixos-config/nixos-config
2026-04-21T04:05:40.2502217Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-21T04:05:40.2534523Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-21T04:05:40.2749938Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-21T04:05:40.2769183Z http.https://github.com/.extraheader
2026-04-21T04:05:40.2781869Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-21T04:05:40.2811308Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-21T04:05:40.3020187Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-21T04:05:40.3048775Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-21T04:05:40.3366006Z Cleaning up orphan processes
2026-04-21T04:05:40.3663025Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `678f8883029547dedac9...` conf=0.46
**Evidence:** ['<timestamp>.9547709Z <timestamp>.939475Z 0 [Note] IPv6 is available.']
```
2026-03-03T01:44:52.9543381Z  
2026-03-03T01:44:52.9543612Z  2026-03-03 01:44:36+00:00 [Note] [Entrypoint]: Stopping temporary server
2026-03-03T01:44:52.9543973Z  2026-03-03 01:44:38+00:00 [Note] [Entrypoint]: Temporary server stopped
2026-03-03T01:44:52.9544242Z  
2026-03-03T01:44:52.9544508Z  2026-03-03 01:44:38+00:00 [Note] [Entrypoint]: MySQL init process done. Ready for start up.
2026-03-03T01:44:52.9544843Z  
2026-03-03T01:44:52.9545204Z  2026-03-03T01:44:38.938765Z 0 [Warning] A deprecated TLS version TLSv1 is enabled. Please use TLSv1.2 or higher.
2026-03-03T01:44:52.9545807Z  2026-03-03T01:44:38.938767Z 0 [Warning] A deprecated TLS version TLSv1.1 is enabled. Please use TLSv1.2 or higher.
2026-03-03T01:44:52.9546312Z  2026-03-03T01:44:38.939203Z 0 [Warning] CA certificate ca.pem is self signed.
2026-03-03T01:44:52.9546815Z  2026-03-03T01:44:38.939233Z 0 [Note] Skipping generation of RSA key pair as key files are present in data directory.
2026-03-03T01:44:52.9547343Z  2026-03-03T01:44:38.939451Z 0 [Note] Server hostname (bind-address): '*'; port: 3306
2026-03-03T01:44:52.9547709Z  2026-03-03T01:44:38.939475Z 0 [Note] IPv6 is available.
2026-03-03T01:44:52.9548027Z  2026-03-03T01:44:38.939487Z 0 [Note]   - '::' resolves to '::';
2026-03-03T01:44:52.9548381Z  2026-03-03T01:44:38.939498Z 0 [Note] Server socket created on IP: '::'.
2026-03-03T01:44:52.9549224Z  2026-03-03T01:44:38.939990Z 0 [Warning] Insecure configuration for --pid-file: Location '/var/run/mysqld' in the path is accessible to all OS users. Consider choosing a different directory.
2026-03-03T01:44:52.9549945Z  2026-03-03T01:44:38.944405Z 0 [Note] Event Scheduler: Loaded 0 events
2026-03-03T01:44:52.9550305Z  2026-03-03T01:44:38.944570Z 0 [Note] mysqld: ready for connections.
2026-03-03T01:44:52.9550739Z  Version: '5.7.44'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server (GPL)
2026-03-03T01:44:52.9556206Z Stop and remove container: e5f125973be34661b8de0c5bd00179c5_mysql57_2f9830
2026-03-03T01:44:52.9560663Z ##[command]/usr/bin/docker rm --force 111bead36ad2a39cf1db3956455b61f685695af1b6f299bc813c479cbdbbaf71
2026-03-03T01:44:54.1918755Z 111bead36ad2a39cf1db3956455b61f685695af1b6f299bc813c479cbdbbaf71
2026-03-03T01:44:54.1939849Z Remove container network: github_network_e59fb49720df4d0385404e9fd9f7973c
2026-03-03T01:44:54.1943951Z ##[command]/usr/bin/docker network rm github_network_e59fb49720df4d0385404e9fd9f7973c
2026-03-03T01:44:54.2895990Z github_network_e59fb49720df4d0385404e9fd9f7973c
2026-03-03T01:44:54.2949883Z Cleaning up orphan processes
```

### `dotnet-restore` — 292 matches, 208 low-conf (71%)

**Fixture:** `9a0dfa14dffbbad16693...` conf=0.37
**Evidence:** ['<timestamp>.7244713Z [ERROR] Non-resolvable import POM: The following artifacts could not be resolved: tools.jackson:jackson-bom:pom:2.21.1 (absent): Could not find artifact tools.jackson:jackson-bom:pom:2.21.1 in central (https://repo.maven.apache.org/maven2) @ org.springframework.boot:spring-boot-dependencies:4.0.5, /home/runner/.m2/repository/org/springframework/boot/spring-boot-dependencies/4.0.5/spring-boot-dependencies-4.0.5.pom, line 3272, column 19']
```
2026-04-09T09:47:39.9799680Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-09T09:47:40.0016349Z Removing HTTP extra header
2026-04-09T09:47:40.0020956Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-09T09:47:40.0055681Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-09T09:47:40.0266499Z Removing includeIf entries pointing to credentials config files
2026-04-09T09:47:40.0272796Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-09T09:47:40.0325909Z includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git.path
2026-04-09T09:47:40.0326703Z includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git/worktrees/*.path
2026-04-09T09:47:40.0327302Z includeif.gitdir:/github/workspace/.git.path
2026-04-09T09:47:40.0327881Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-09T09:47:40.0334983Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git.path
2026-04-09T09:47:40.0355786Z /home/runner/work/_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0366605Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git.path /home/runner/work/_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0398842Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git/worktrees/*.path
2026-04-09T09:47:40.0419366Z /home/runner/work/_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0428788Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/RecordPlatform/RecordPlatform/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0460167Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-09T09:47:40.0481708Z /github/runner_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0489233Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0519021Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-09T09:47:40.0539676Z /github/runner_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0551292Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config
2026-04-09T09:47:40.0586928Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-09T09:47:40.0800175Z Removing credentials config '/home/runner/work/_temp/git-credentials-eb1ef087-b26a-45b5-b8ed-6540ff3e65cd.config'
2026-04-09T09:47:40.0934536Z Cleaning up orphan processes
```

**Fixture:** `5f3d0e7d8acfa2318e6e...` conf=0.37
**Evidence:** ['<timestamp>.8992407Z [ERROR] Caused by: The following artifacts could not be resolved: com.liquibase:liquibase-commercial:pom:master-989eda1 (absent): Could not transfer artifact com.liquibase:liquibase-commercial:pom:master-989eda1 from/to github-pro (https://maven.pkg.github.com/liquibase/liquibase-pro): status code: 401, reason phrase: Unauthorized (401)']
```
2026-04-20T06:50:53.4529390Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-20T06:50:53.4785658Z Removing HTTP extra header
2026-04-20T06:50:53.4794117Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-20T06:50:53.4835446Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-20T06:50:53.5092889Z Removing includeIf entries pointing to credentials config files
2026-04-20T06:50:53.5102481Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-20T06:50:53.5130300Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-20T06:50:53.5132574Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-20T06:50:53.5134536Z includeif.gitdir:/github/workspace/.git.path
2026-04-20T06:50:53.5135892Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-20T06:50:53.5144831Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-20T06:50:53.5172706Z /home/runner/work/_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5187637Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path /home/runner/work/_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5236791Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-20T06:50:53.5262552Z /home/runner/work/_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5275427Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5315070Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-20T06:50:53.5340760Z /github/runner_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5352714Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5391510Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-20T06:50:53.5416367Z /github/runner_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5431464Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config
2026-04-20T06:50:53.5477935Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-20T06:50:53.5750341Z Removing credentials config '/home/runner/work/_temp/git-credentials-13b4d105-1321-4dfc-9ba6-4adb1a8632fd.config'
2026-04-20T06:50:53.5928797Z Cleaning up orphan processes
```

**Fixture:** `2d35ef072f3361136d14...` conf=0.46
**Evidence:** ['<timestamp>.5560768Z ##[error]Unable to resolve action opendatacube/eo3-validate, repository not found']
```
2026-02-04T22:48:29.2187984Z Image Release: https://github.com/actions/runner-images/releases/tag/ubuntu24%2F20260126.10
2026-02-04T22:48:29.2189460Z ##[endgroup]
2026-02-04T22:48:29.2194229Z ##[group]GITHUB_TOKEN Permissions
2026-02-04T22:48:29.2197292Z Actions: write
2026-02-04T22:48:29.2198228Z ArtifactMetadata: write
2026-02-04T22:48:29.2199074Z Attestations: write
2026-02-04T22:48:29.2199830Z Checks: write
2026-02-04T22:48:29.2200768Z Contents: write
2026-02-04T22:48:29.2201506Z Deployments: write
2026-02-04T22:48:29.2202333Z Discussions: write
2026-02-04T22:48:29.2203227Z Issues: write
2026-02-04T22:48:29.2203977Z Metadata: read
2026-02-04T22:48:29.2204936Z Models: read
2026-02-04T22:48:29.2205714Z Packages: write
2026-02-04T22:48:29.2206527Z Pages: write
2026-02-04T22:48:29.2207465Z PullRequests: write
2026-02-04T22:48:29.2208408Z RepositoryProjects: write
2026-02-04T22:48:29.2209304Z SecurityEvents: write
2026-02-04T22:48:29.2210124Z Statuses: write
2026-02-04T22:48:29.2211044Z ##[endgroup]
2026-02-04T22:48:29.2213797Z Secret source: Actions
2026-02-04T22:48:29.2215058Z Prepare workflow directory
2026-02-04T22:48:29.2662965Z Prepare all required actions
2026-02-04T22:48:29.2717563Z Getting action download info
2026-02-04T22:48:29.5560768Z ##[error]Unable to resolve action opendatacube/eo3-validate, repository not found
```

### `permission-denied` — 276 matches, 170 low-conf (62%)

**Fixture:** `a294bec6d3b6064f8361...` conf=0.41
**Evidence:** ['<timestamp>.0009189Z "Host": "unix:///var/run/docker.sock",', '<timestamp>.7781215Z apparmor']
```
2026-02-28T09:25:50.6216439Z Post job cleanup.
2026-02-28T09:25:50.8736152Z ##[group]Removing builder
2026-02-28T09:25:50.9745275Z [command]/usr/bin/docker buildx rm builder-d57c5620-9cda-4889-b712-4d544fd859ef
2026-02-28T09:25:51.1345033Z builder-d57c5620-9cda-4889-b712-4d544fd859ef removed
2026-02-28T09:25:51.1382216Z ##[endgroup]
2026-02-28T09:25:51.1382980Z ##[group]Cleaning up certificates
2026-02-28T09:25:51.1389449Z ##[endgroup]
2026-02-28T09:25:51.1390076Z ##[group]Post cache
2026-02-28T09:25:51.1393171Z State not set
2026-02-28T09:25:51.1393820Z ##[endgroup]
2026-02-28T09:25:51.1530786Z Post job cleanup.
2026-02-28T09:25:51.2492138Z [command]/usr/bin/git version
2026-02-28T09:25:51.2529738Z git version 2.53.0
2026-02-28T09:25:51.2574870Z Temporarily overriding HOME='/home/runner/work/_temp/55d6faa5-fc85-41ac-8875-3122b67fd2be' before making global git config changes
2026-02-28T09:25:51.2576249Z Adding repository directory to the temporary git global config as a safe directory
2026-02-28T09:25:51.2582027Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/todoki/todoki
2026-02-28T09:25:51.2625003Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-28T09:25:51.2659166Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-28T09:25:51.2908705Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-28T09:25:51.2932593Z http.https://github.com/.extraheader
2026-02-28T09:25:51.2946132Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-28T09:25:51.2979824Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-28T09:25:51.3228986Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-28T09:25:51.3266781Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-28T09:25:51.3619652Z Cleaning up orphan processes
```

**Fixture:** `31468179185f6f2f8da2...` conf=0.41
**Evidence:** ['<timestamp>.2497820Z "Host": "unix:///var/run/docker.sock",', '<timestamp>.0745602Z apparmor']
```
2026-02-13T12:03:28.5576393Z Post job cleanup.
2026-02-13T12:03:28.7892470Z ##[group]Removing builder
2026-02-13T12:03:28.8858660Z [command]/usr/bin/docker buildx rm builder-e5b7064d-5267-4c87-a176-209e4c846ef9
2026-02-13T12:03:29.0255211Z builder-e5b7064d-5267-4c87-a176-209e4c846ef9 removed
2026-02-13T12:03:29.0282368Z ##[endgroup]
2026-02-13T12:03:29.0283314Z ##[group]Cleaning up certificates
2026-02-13T12:03:29.0290317Z ##[endgroup]
2026-02-13T12:03:29.0291066Z ##[group]Post cache
2026-02-13T12:03:29.0291942Z State not set
2026-02-13T12:03:29.0292998Z ##[endgroup]
2026-02-13T12:03:29.0407433Z Post job cleanup.
2026-02-13T12:03:29.1322903Z [command]/usr/bin/git version
2026-02-13T12:03:29.1358258Z git version 2.52.0
2026-02-13T12:03:29.1397542Z Temporarily overriding HOME='/home/runner/work/_temp/1fa4c8ec-f268-4f58-9d70-8ec7ee0125ce' before making global git config changes
2026-02-13T12:03:29.1398913Z Adding repository directory to the temporary git global config as a safe directory
2026-02-13T12:03:29.1404264Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/todoki/todoki
2026-02-13T12:03:29.1444882Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-13T12:03:29.1474921Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-13T12:03:29.1672420Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-13T12:03:29.1692249Z http.https://github.com/.extraheader
2026-02-13T12:03:29.1704382Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-13T12:03:29.1733176Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-13T12:03:29.1919558Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-13T12:03:29.1948332Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-13T12:03:29.2249014Z Cleaning up orphan processes
```

**Fixture:** `de1cced0c80f04a22bc6...` conf=0.41
**Evidence:** ['<timestamp>.1343317Z Endpoint: unix:///var/run/docker.sock', '<timestamp>.6874043Z apparmor']
```
2026-02-17T15:01:43.4230627Z ERROR: failed to build: failed to solve: process "/bin/sh -c GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0     TAG_NAME=$TAG_NAME COMMIT=$COMMIT ./build.sh &&     cp /bin/sh /app/sh && chmod +x /app/sh" did not complete successfully: exit code: 1
2026-02-17T15:01:43.4281832Z ❌ Failed to build image with tag: ghcr.io/morpheusais/morpheus-lumerin-node:v5.11.3-dev
2026-02-17T15:01:43.4299845Z ##[error]Process completed with exit code 1.
2026-02-17T15:01:43.4393540Z Post job cleanup.
2026-02-17T15:01:43.5810242Z ##[group]Removing builder
2026-02-17T15:01:43.7037165Z [command]/usr/bin/docker buildx rm builder-9ac6d67a-e844-4e21-8770-a22af15aedf8
2026-02-17T15:01:46.7511704Z builder-9ac6d67a-e844-4e21-8770-a22af15aedf8 removed
2026-02-17T15:01:46.7560983Z ##[endgroup]
2026-02-17T15:01:46.7561619Z ##[group]Cleaning up certificates
2026-02-17T15:01:46.7562377Z ##[endgroup]
2026-02-17T15:01:46.7694321Z Post job cleanup.
2026-02-17T15:01:46.8653697Z [command]/usr/bin/git version
2026-02-17T15:01:46.8690932Z git version 2.52.0
2026-02-17T15:01:46.8734441Z Temporarily overriding HOME='/home/runner/work/_temp/1b457459-b9d3-4dae-81b3-21df7e3ee7e0' before making global git config changes
2026-02-17T15:01:46.8736023Z Adding repository directory to the temporary git global config as a safe directory
2026-02-17T15:01:46.8740380Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/Morpheus-Lumerin-Node/Morpheus-Lumerin-Node
2026-02-17T15:01:46.8777109Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-17T15:01:46.8811088Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-17T15:01:46.9054642Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-17T15:01:46.9075742Z http.https://github.com/.extraheader
2026-02-17T15:01:46.9088355Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-17T15:01:46.9121261Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-17T15:01:46.9355876Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-17T15:01:46.9387235Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-17T15:01:46.9736191Z Cleaning up orphan processes
```

### `formatting-failure` — 111 matches, 111 low-conf (100%)

**Fixture:** `723244ee656b80573a95...` conf=0.47
**Evidence:** ['<timestamp>.6595511Z Would reformat: stepflow-py/src/stepflow_py/__init__.py', '<timestamp>.6597125Z 3 files would be reformatted, 45 files already formatted']
```
2026-03-23T16:58:59.2602325Z 
2026-03-23T16:58:59.2602453Z   Formatting:
2026-03-23T16:58:59.2602854Z     Check: cd sdks/python && uv run poe fmt-check
2026-03-23T16:58:59.2603439Z     Fix:   cd sdks/python && uv run poe fmt-fix
2026-03-23T16:58:59.2603883Z   Linting:
2026-03-23T16:58:59.2604257Z     Check: cd sdks/python && uv run poe lint-check
2026-03-23T16:58:59.2605002Z     Fix:   cd sdks/python && uv run poe lint-fix
2026-03-23T16:58:59.2633331Z ##[error]Process completed with exit code 1.
2026-03-23T16:58:59.2729202Z Post job cleanup.
2026-03-23T16:58:59.2810054Z Post job cleanup.
2026-03-23T16:58:59.3781538Z [command]/usr/bin/git version
2026-03-23T16:58:59.3822206Z git version 2.53.0
2026-03-23T16:58:59.3872064Z Temporarily overriding HOME='/home/runner/work/_temp/a40dcc41-cbce-4d2b-b51b-40f29bb402d6' before making global git config changes
2026-03-23T16:58:59.3873092Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T16:58:59.3879240Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/stepflow/stepflow
2026-03-23T16:58:59.3915720Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-23T16:58:59.3950214Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-23T16:58:59.4194232Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T16:58:59.4216247Z http.https://github.com/.extraheader
2026-03-23T16:58:59.4229328Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-23T16:58:59.4260787Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-23T16:58:59.4497641Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T16:58:59.4529272Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-23T16:58:59.4887365Z Cleaning up orphan processes
2026-03-23T16:58:59.5193429Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, astral-sh/setup-uv@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `ccea80adbff27ed97b19...` conf=0.47
**Evidence:** ['<timestamp>.9167957Z Would reformat: workshops/dcc26/participant/00_setup_api_warmup.ipynb', '<timestamp>.9242565Z 6 files would be reformatted, 64 files already formatted']
```
2026-04-07T15:40:34.9166437Z error: Failed to read workshops/dcc26/solutions/00_setup_api_warmup.ipynb: This file does not match the schema expected of Jupyter Notebooks: unknown field `execution_count`, expected one of `attachments`, `id`, `metadata`, `source` at line 505 column 3
2026-04-07T15:40:34.9167957Z Would reformat: workshops/dcc26/participant/00_setup_api_warmup.ipynb
2026-04-07T15:40:34.9168525Z Would reformat: workshops/dcc26/participant/01_train_generate.ipynb
2026-04-07T15:40:34.9168981Z Would reformat: workshops/dcc26/participant/02_evaluate_metrics.ipynb
2026-04-07T15:40:34.9169821Z Would reformat: workshops/dcc26/solutions/01_train_generate.ipynb
2026-04-07T15:40:34.9170368Z Would reformat: workshops/dcc26/solutions/02_evaluate_metrics.ipynb
2026-04-07T15:40:34.9170794Z Would reformat: workshops/dcc26/utils/notebook_helpers.py
2026-04-07T15:40:34.9242565Z 6 files would be reformatted, 64 files already formatted
2026-04-07T15:40:34.9263291Z ##[error]Process completed with exit code 2.
2026-04-07T15:40:34.9378653Z Post job cleanup.
2026-04-07T15:40:35.0382104Z [command]/usr/bin/git version
2026-04-07T15:40:35.0420528Z git version 2.53.0
2026-04-07T15:40:35.0473981Z Temporarily overriding HOME='/home/runner/work/_temp/f86a665f-ddfc-48d3-a745-7b7d90091509' before making global git config changes
2026-04-07T15:40:35.0475131Z Adding repository directory to the temporary git global config as a safe directory
2026-04-07T15:40:35.0479298Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/EngiOpt/EngiOpt
2026-04-07T15:40:35.0515284Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-07T15:40:35.0550438Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-07T15:40:35.0782137Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-07T15:40:35.0805041Z http.https://github.com/.extraheader
2026-04-07T15:40:35.0818034Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-07T15:40:35.0853114Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-07T15:40:35.1087386Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-07T15:40:35.1123816Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-07T15:40:35.1479052Z Cleaning up orphan processes
2026-04-07T15:40:35.1742163Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, astral-sh/ruff-action@v3. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `9b58d42d694c27f2e8de...` conf=0.47
**Evidence:** ['<timestamp>.2715446Z Would reformat: stepflow-py/src/stepflow_py/__init__.py', '<timestamp>.2716530Z 2 files would be reformatted, 46 files already formatted']
```
2026-03-23T16:34:32.1773009Z 
2026-03-23T16:34:32.1773131Z   Formatting:
2026-03-23T16:34:32.1773512Z     Check: cd sdks/python && uv run poe fmt-check
2026-03-23T16:34:32.1774029Z     Fix:   cd sdks/python && uv run poe fmt-fix
2026-03-23T16:34:32.1774469Z   Linting:
2026-03-23T16:34:32.1774816Z     Check: cd sdks/python && uv run poe lint-check
2026-03-23T16:34:32.1775732Z     Fix:   cd sdks/python && uv run poe lint-fix
2026-03-23T16:34:32.1805606Z ##[error]Process completed with exit code 1.
2026-03-23T16:34:32.1897546Z Post job cleanup.
2026-03-23T16:34:32.1985914Z Post job cleanup.
2026-03-23T16:34:32.2975777Z [command]/usr/bin/git version
2026-03-23T16:34:32.3023204Z git version 2.53.0
2026-03-23T16:34:32.3068315Z Temporarily overriding HOME='/home/runner/work/_temp/c8e1511c-be30-4f73-8eb8-b5c151cdd23c' before making global git config changes
2026-03-23T16:34:32.3069621Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T16:34:32.3074360Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/stepflow/stepflow
2026-03-23T16:34:32.3114777Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-23T16:34:32.3150686Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-23T16:34:32.3402095Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T16:34:32.3426319Z http.https://github.com/.extraheader
2026-03-23T16:34:32.3439592Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-23T16:34:32.3472542Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-23T16:34:32.3734355Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T16:34:32.3776327Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-23T16:34:32.4274399Z Cleaning up orphan processes
2026-03-23T16:34:32.4633754Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, astral-sh/setup-uv@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `alpine-debian-incompatibility` — 100 matches, 94 low-conf (94%)

**Fixture:** `e45dfdd9021d13a1b976...` conf=0.45
**Evidence:** ['<timestamp>.4161369Z proxy | <timestamp> [180] GET https://registry.npmjs.org:443/@rollup%2Frollup-linux-x64-musl']
```
2026-04-05T12:52:44.5831700Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-04-05T12:52:44.5832626Z +-------------------------------------------------------------------------------------------+
2026-04-05T12:52:44.5833666Z |                               Dependencies failed to update                               |
2026-04-05T12:52:44.5834459Z +------------+----------------------------+-------------------------------------------------+
2026-04-05T12:52:44.5835238Z | Dependency | Error Type                 | Error Details                                   |
2026-04-05T12:52:44.5836023Z +------------+----------------------------+-------------------------------------------------+
2026-04-05T12:52:44.5836838Z | picomatch  | tool_feature_not_supported | {                                               |
2026-04-05T12:52:44.5837637Z |            |                            |   "tool-name": "pnpm",                          |
2026-04-05T12:52:44.5838259Z |            |                            |   "tool-type": "package_manager",               |
2026-04-05T12:52:44.5838862Z |            |                            |   "feature": "updating transitive dependencies" |
2026-04-05T12:52:44.5839426Z |            |                            | }                                               |
2026-04-05T12:52:44.5840029Z +------------+----------------------------+-------------------------------------------------+
2026-04-05T12:52:44.7044648Z Failure running container 4041f58c8db8edd5ec6a379e20969cfbd34608b5208155baf82d7efa88ceb157: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-04-05T12:52:44.8334922Z Cleaned up container 4041f58c8db8edd5ec6a379e20969cfbd34608b5208155baf82d7efa88ceb157
2026-04-05T12:52:44.8431366Z   proxy | 2026/04/05 12:52:44 2/282 calls cached (0%)
2026-04-05T12:52:44.8438957Z   proxy | 2026/04/05 12:52:44 Posting metrics to remote API endpoint
2026-04-05T12:52:45.5725072Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/InsightSoftwareConsortium/ITKIOScanco/network/updates/1307988336 (write access to the repository is required to view the log)
2026-04-05T12:52:45.5733836Z 🤖 ~ finished: error reported to Dependabot ~
2026-04-05T12:52:45.5836246Z Post job cleanup.
2026-04-05T12:52:45.7438360Z Cleaning up orphan processes
2026-04-05T12:52:45.7767411Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `821b6f03438b4b33cb94...` conf=0.45
**Evidence:** ['<timestamp>.1101881Z proxy | <timestamp> [415] GET https://registry.npmjs.org:443/@rollup%2Frollup-linux-x64-musl']
```
2026-04-09T09:04:03.6218754Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-04-09T09:04:03.6219687Z +-------------------------------------------------------------------------------------------+
2026-04-09T09:04:03.6221108Z |                               Dependencies failed to update                               |
2026-04-09T09:04:03.6222018Z +------------+----------------------------+-------------------------------------------------+
2026-04-09T09:04:03.6222930Z | Dependency | Error Type                 | Error Details                                   |
2026-04-09T09:04:03.6223833Z +------------+----------------------------+-------------------------------------------------+
2026-04-09T09:04:03.6224730Z | dompurify  | tool_feature_not_supported | {                                               |
2026-04-09T09:04:03.6225537Z |            |                            |   "tool-name": "pnpm",                          |
2026-04-09T09:04:03.6226214Z |            |                            |   "tool-type": "package_manager",               |
2026-04-09T09:04:03.6226933Z |            |                            |   "feature": "updating transitive dependencies" |
2026-04-09T09:04:03.6227618Z |            |                            | }                                               |
2026-04-09T09:04:03.6228347Z +------------+----------------------------+-------------------------------------------------+
2026-04-09T09:04:03.7533829Z Failure running container 1e3638427fb045a59281215ef772d8e82e4b9983d3b19d2d73b3ace56baa5e8e: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-04-09T09:04:03.8287673Z Cleaned up container 1e3638427fb045a59281215ef772d8e82e4b9983d3b19d2d73b3ace56baa5e8e
2026-04-09T09:04:03.8403274Z   proxy | 2026/04/09 09:04:03 20/340 calls cached (5%)
2026-04-09T09:04:03.8404832Z 2026/04/09 09:04:03 Posting metrics to remote API endpoint
2026-04-09T09:04:04.3434498Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/SoarCollab/RecordPlatform/network/updates/1315345222 (write access to the repository is required to view the log)
2026-04-09T09:04:04.3443809Z 🤖 ~ finished: error reported to Dependabot ~
2026-04-09T09:04:04.3516071Z Post job cleanup.
2026-04-09T09:04:04.5035048Z Cleaning up orphan processes
2026-04-09T09:04:04.5408702Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `6f5ca81db834b6eee573...` conf=0.45
**Evidence:** ['<timestamp>.9371636Z proxy | <timestamp> [207] GET https://registry.npmjs.org:443/@rollup%2Frollup-linux-x64-musl']
```
2026-03-25T22:11:06.4717987Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-25T22:11:06.4719425Z +-------------------------------------------------------------------------------------------+
2026-03-25T22:11:06.4720119Z |                               Dependencies failed to update                               |
2026-03-25T22:11:06.4720796Z +------------+----------------------------+-------------------------------------------------+
2026-03-25T22:11:06.4721535Z | Dependency | Error Type                 | Error Details                                   |
2026-03-25T22:11:06.4722219Z +------------+----------------------------+-------------------------------------------------+
2026-03-25T22:11:06.4722951Z | picomatch  | tool_feature_not_supported | {                                               |
2026-03-25T22:11:06.4723641Z |            |                            |   "tool-name": "pnpm",                          |
2026-03-25T22:11:06.4724207Z |            |                            |   "tool-type": "package_manager",               |
2026-03-25T22:11:06.4724788Z |            |                            |   "feature": "updating transitive dependencies" |
2026-03-25T22:11:06.4725335Z |            |                            | }                                               |
2026-03-25T22:11:06.4725913Z +------------+----------------------------+-------------------------------------------------+
2026-03-25T22:11:06.6028415Z Failure running container f66ea0a55ff09c088cd8c603e6abc3e3edd2bade52d67d5be7d58c4bd95893aa: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-25T22:11:06.7513983Z Cleaned up container f66ea0a55ff09c088cd8c603e6abc3e3edd2bade52d67d5be7d58c4bd95893aa
2026-03-25T22:11:06.7621281Z   proxy | 2026/03/25 22:11:06 Posting metrics to remote API endpoint
2026-03-25T22:11:06.7622010Z   proxy | 2026/03/25 22:11:06 2/282 calls cached (0%)
2026-03-25T22:11:07.5232864Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/InsightSoftwareConsortium/ITKIOScanco/network/updates/1292551178 (write access to the repository is required to view the log)
2026-03-25T22:11:07.5242239Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-25T22:11:07.5344302Z Post job cleanup.
2026-03-25T22:11:07.7057169Z Cleaning up orphan processes
2026-03-25T22:11:07.7592650Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `install-failure` — 85 matches, 85 low-conf (100%)

**Fixture:** `9e1a601cf7950660da61...` conf=0.37
**Evidence:** ['<timestamp>.5012882Z E: Unable to locate package libegl1-mesa']
```
2026-02-26T16:05:18.8340673Z Get:36 http://azure.archive.ubuntu.com/ubuntu noble-security/restricted Translation-en [596 kB]
2026-02-26T16:05:18.8379895Z Get:37 http://azure.archive.ubuntu.com/ubuntu noble-security/restricted amd64 Components [212 B]
2026-02-26T16:05:18.8395194Z Get:38 http://azure.archive.ubuntu.com/ubuntu noble-security/multiverse amd64 Components [208 B]
2026-02-26T16:05:24.5200066Z Fetched 14.7 MB in 2s (8190 kB/s)
2026-02-26T16:05:25.2857051Z Reading package lists...
2026-02-26T16:05:25.3200901Z Reading package lists...
2026-02-26T16:05:25.4842230Z Building dependency tree...
2026-02-26T16:05:25.4849123Z Reading state information...
2026-02-26T16:05:25.5012882Z E: Unable to locate package libegl1-mesa
2026-02-26T16:05:25.5046036Z ##[error]Process completed with exit code 100.
2026-02-26T16:05:25.5194881Z Post job cleanup.
2026-02-26T16:05:25.6139166Z [command]/usr/bin/git version
2026-02-26T16:05:25.6175746Z git version 2.53.0
2026-02-26T16:05:25.6225839Z Temporarily overriding HOME='/home/runner/work/_temp/c30fcafb-0b4d-4a47-8878-7bc1fffb0f93' before making global git config changes
2026-02-26T16:05:25.6227276Z Adding repository directory to the temporary git global config as a safe directory
2026-02-26T16:05:25.6231199Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/iBridges-GUI/iBridges-GUI
2026-02-26T16:05:25.6265908Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-26T16:05:25.6299321Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-26T16:05:25.6540782Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-26T16:05:25.6562000Z http.https://github.com/.extraheader
2026-02-26T16:05:25.6574445Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-26T16:05:25.6605322Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-26T16:05:25.6838599Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-26T16:05:25.6869518Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-26T16:05:25.7220416Z Cleaning up orphan processes
```

**Fixture:** `e7d36af2b3c0358b8a0e...` conf=0.37
**Evidence:** ['<timestamp>.5165086Z E: Unable to locate package libc6:i386']
```
2026-02-11T12:29:09.7096928Z     Building for DEVICE:am62dx PROFILE:debug completed!!
2026-02-11T12:29:09.7097494Z Build for SOC am62dx failed ....
2026-02-11T12:29:09.7097757Z 
2026-02-11T12:29:09.7097763Z 
2026-02-11T12:29:09.7106485Z make[3]: /home/runner/ti/sysconfig_1.26.2/nodejs/node: No such file or directory
2026-02-11T12:29:09.7107420Z make[3]: *** [makefile:226: syscfg] Error 127
2026-02-11T12:29:09.7109173Z make[2]: *** [makefile.am62dx:5103: sbl] Error 2
2026-02-11T12:29:09.7109703Z make[1]: *** [makefile.am62dx:8: all] Error 2
2026-02-11T12:29:09.7110647Z make: *** [makefile:76: all] Error 2
2026-02-11T12:29:09.7128201Z ##[error]Process completed with exit code 1.
2026-02-11T12:29:09.7232874Z Post job cleanup.
2026-02-11T12:29:09.8276824Z [command]/usr/bin/git version
2026-02-11T12:29:09.8313678Z git version 2.52.0
2026-02-11T12:29:09.8360653Z Temporarily overriding HOME='/home/runner/work/_temp/c28b42ed-98e5-43f6-9330-a34ff81ab5bb' before making global git config changes
2026-02-11T12:29:09.8361897Z Adding repository directory to the temporary git global config as a safe directory
2026-02-11T12:29:09.8374644Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/mcupsdk-core-k3/mcupsdk-core-k3/pr_checkout
2026-02-11T12:29:09.8408216Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-11T12:29:09.8440117Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-11T12:29:09.8784873Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-11T12:29:09.8807611Z http.https://github.com/.extraheader
2026-02-11T12:29:09.8820059Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-11T12:29:09.8851571Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-11T12:29:09.9119462Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-11T12:29:09.9151898Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-11T12:29:09.9526514Z Cleaning up orphan processes
```

**Fixture:** `536ca3005fec7e89d098...` conf=0.37
**Evidence:** ['<timestamp>.8594568Z #12 2.085 E: Unable to locate package openjdk-17-jdk']
```
2026-02-20T18:45:15.2064479Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-20T18:45:15.2289370Z Removing HTTP extra header
2026-02-20T18:45:15.2294235Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-20T18:45:15.2325175Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-20T18:45:15.2545604Z Removing includeIf entries pointing to credentials config files
2026-02-20T18:45:15.2546785Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-20T18:45:15.2567713Z includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path
2026-02-20T18:45:15.2568573Z includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path
2026-02-20T18:45:15.2569308Z includeif.gitdir:/github/workspace/.git.path
2026-02-20T18:45:15.2569917Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-20T18:45:15.2578476Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path
2026-02-20T18:45:15.2599316Z /home/runner/work/_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2609932Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/pantheon/pantheon/.git.path /home/runner/work/_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2641543Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path
2026-02-20T18:45:15.2662033Z /home/runner/work/_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2671830Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/pantheon/pantheon/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2702043Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-20T18:45:15.2723119Z /github/runner_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2731244Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2759812Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-20T18:45:15.2779323Z /github/runner_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2787904Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config
2026-02-20T18:45:15.2817727Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-20T18:45:15.3029271Z Removing credentials config '/home/runner/work/_temp/git-credentials-6434fb53-26cb-4bd7-9809-b08d459aa990.config'
2026-02-20T18:45:15.3167083Z Cleaning up orphan processes
```

### `docker-manifest-not-found` — 65 matches, 42 low-conf (65%)

**Fixture:** `63d56de627005499b452...` conf=0.37
**Evidence:** ['<timestamp>.0962270Z [WARNING] Could not transfer metadata org.apache.flink:flink-connector-kafka:3.1-SNAPSHOT/maven-metadata.xml from/to example-repo (file:///usr/local/google/home/runner/.m2/repository): Repository path /usr/local/google/home/runner/.m2/repository does not exist, and cannot be created.']
```
2026-02-18T01:55:34.6838535Z   JAVA_HOME: /opt/hostedtoolcache/Java_Temurin-Hotspot_jdk/11.0.30-7/x64
2026-02-18T01:55:34.6838994Z   JAVA_HOME_11_X64: /opt/hostedtoolcache/Java_Temurin-Hotspot_jdk/11.0.30-7/x64
2026-02-18T01:55:34.6839758Z   MVN_DEPENDENCY_CONVERGENCE: -Dflink.convergence.phase=install -Pcheck-convergence
2026-02-18T01:55:34.6840360Z   binary_url: https://archive.apache.org/dist/flink/flink-1.18.0/flink-1.18.0-bin-scala_2.12.tgz
2026-02-18T01:55:34.6840909Z   cache_binary: true
2026-02-18T01:55:34.6841102Z ##[endgroup]
2026-02-18T01:55:34.6909639Z linux-gnu
2026-02-18T01:55:34.6911879Z Setting up JVM thread dumps
2026-02-18T01:55:34.9129118Z /tmp/jattach: OK
2026-02-18T01:55:34.9396395Z Post job cleanup.
2026-02-18T01:55:35.1157415Z Post job cleanup.
2026-02-18T01:55:35.2183582Z [command]/usr/bin/git version
2026-02-18T01:55:35.2235853Z git version 2.52.0
2026-02-18T01:55:35.2280868Z Temporarily overriding HOME='/home/runner/work/_temp/f8d9bc2c-8bce-46da-bfe6-042aed88931f' before making global git config changes
2026-02-18T01:55:35.2282829Z Adding repository directory to the temporary git global config as a safe directory
2026-02-18T01:55:35.2287992Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/flink-connector-gcp/flink-connector-gcp
2026-02-18T01:55:35.2325823Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-18T01:55:35.2362766Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-18T01:55:35.2628759Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-18T01:55:35.2653823Z http.https://github.com/.extraheader
2026-02-18T01:55:35.2666962Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-18T01:55:35.2704476Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-18T01:55:35.2958935Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-18T01:55:35.2993859Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-18T01:55:35.3376024Z Cleaning up orphan processes
```

**Fixture:** `ac7a737bcea9f3ddf4fa...` conf=0.37
**Evidence:** ['<timestamp>.0489126Z 5 : ┆ ┆ ┆ [0.0s] | Error response from daemon: No such image: registry.dagger.io/engine:v0.19.11']
```
2026-02-09T08:28:16.4851419Z main.go:6:2: var `version` is unused (unused)
2026-02-09T08:28:16.4851790Z 	version = "dev"
2026-02-09T08:28:16.4855909Z 	^
2026-02-09T08:28:16.4856207Z main.go:7:2: var `commit` is unused (unused)
2026-02-09T08:28:16.4856594Z 	commit  = "none"
2026-02-09T08:28:16.4856829Z 	^
2026-02-09T08:28:16.4858628Z main.go:8:2: var `date` is unused (unused)
2026-02-09T08:28:16.4859011Z 	date    = "unknown"
2026-02-09T08:28:16.4859262Z 	^
2026-02-09T08:28:16.4883189Z ##[error]Process completed with exit code 1.
2026-02-09T08:28:16.5039535Z Post job cleanup.
2026-02-09T08:28:16.5931879Z [command]/usr/bin/git version
2026-02-09T08:28:16.5965706Z git version 2.52.0
2026-02-09T08:28:16.6004728Z Temporarily overriding HOME='/home/runner/work/_temp/88313d61-0062-4419-90df-2ff1acfd1673' before making global git config changes
2026-02-09T08:28:16.6006027Z Adding repository directory to the temporary git global config as a safe directory
2026-02-09T08:28:16.6018750Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/anker/anker
2026-02-09T08:28:16.6049709Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-09T08:28:16.6078443Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-09T08:28:16.6261460Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-09T08:28:16.6281047Z http.https://github.com/.extraheader
2026-02-09T08:28:16.6291782Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-09T08:28:16.6319913Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-09T08:28:16.6496241Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-09T08:28:16.6523024Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-09T08:28:16.6793643Z Cleaning up orphan processes
```

**Fixture:** `48ad21554e1404d7c357...` conf=0.37
**Evidence:** ['<timestamp>.3357722Z 5 : ┆ ┆ ┆ [0.0s] | Error response from daemon: No such image: registry.dagger.io/engine:v0.19.11']
```
2026-02-09T09:47:48.3103465Z main.go:6:2: var `version` is unused (unused)
2026-02-09T09:47:48.3103900Z 	version = "dev"
2026-02-09T09:47:48.3104189Z 	^
2026-02-09T09:47:48.3104775Z main.go:7:2: var `commit` is unused (unused)
2026-02-09T09:47:48.3105202Z 	commit  = "none"
2026-02-09T09:47:48.3105490Z 	^
2026-02-09T09:47:48.3105774Z main.go:8:2: var `date` is unused (unused)
2026-02-09T09:47:48.3106187Z 	date    = "unknown"
2026-02-09T09:47:48.3106484Z 	^
2026-02-09T09:47:48.3143214Z ##[error]Process completed with exit code 1.
2026-02-09T09:47:48.3302488Z Post job cleanup.
2026-02-09T09:47:48.4232604Z [command]/usr/bin/git version
2026-02-09T09:47:48.4268230Z git version 2.52.0
2026-02-09T09:47:48.4311682Z Temporarily overriding HOME='/home/runner/work/_temp/2332eb76-df7c-451c-9b94-a71062e0d340' before making global git config changes
2026-02-09T09:47:48.4313324Z Adding repository directory to the temporary git global config as a safe directory
2026-02-09T09:47:48.4318032Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/anker/anker
2026-02-09T09:47:48.4358580Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-09T09:47:48.4390226Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-09T09:47:48.4615501Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-09T09:47:48.4635202Z http.https://github.com/.extraheader
2026-02-09T09:47:48.4647242Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-09T09:47:48.4676792Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-09T09:47:48.4893135Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-09T09:47:48.4922350Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-09T09:47:48.5252032Z Cleaning up orphan processes
```

### `npm-eresolve-conflict` — 64 matches, 33 low-conf (52%)

**Fixture:** `a4dfb4b57643528e2833...` conf=0.42
**Evidence:** ['<timestamp>.3290836Z npm ERR! code ERESOLVE', '<timestamp>.3303548Z npm ERR! ERESOLVE could not resolve', '<timestamp>.3315356Z npm ERR! Could not resolve dependency:']
```
2026-02-09T18:44:07.3319255Z npm ERR! 
2026-02-09T18:44:07.3319537Z npm ERR! Fix the upstream dependency conflict, or retry
2026-02-09T18:44:07.3319947Z npm ERR! this command with --force, or --legacy-peer-deps
2026-02-09T18:44:07.3320451Z npm ERR! to accept an incorrect (and potentially broken) dependency resolution.
2026-02-09T18:44:07.3320814Z npm ERR! 
2026-02-09T18:44:07.3321142Z npm ERR! See /home/runner/.npm/eresolve-report.txt for a full report.
2026-02-09T18:44:07.3326065Z 
2026-02-09T18:44:07.3327744Z npm ERR! A complete log of this run can be found in:
2026-02-09T18:44:07.3328605Z npm ERR!     /home/runner/.npm/_logs/2026-02-09T18_44_06_166Z-debug-0.log
2026-02-09T18:44:07.3456362Z ##[error]Process completed with exit code 1.
2026-02-09T18:44:07.3649025Z Post job cleanup.
2026-02-09T18:44:07.4609533Z [command]/usr/bin/git version
2026-02-09T18:44:07.4646525Z git version 2.52.0
2026-02-09T18:44:07.4692148Z Temporarily overriding HOME='/home/runner/work/_temp/008f67e5-5fcc-48a7-990c-bd178bb15879' before making global git config changes
2026-02-09T18:44:07.4693940Z Adding repository directory to the temporary git global config as a safe directory
2026-02-09T18:44:07.4707598Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/homebridge-flexom/homebridge-flexom
2026-02-09T18:44:07.4745981Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-09T18:44:07.4780541Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-09T18:44:07.5014587Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-09T18:44:07.5036970Z http.https://github.com/.extraheader
2026-02-09T18:44:07.5049524Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-09T18:44:07.5080168Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-09T18:44:07.5296839Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-09T18:44:07.5326469Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-09T18:44:07.5646958Z Cleaning up orphan processes
```

**Fixture:** `8737b61d999994935f86...` conf=0.42
**Evidence:** ['<timestamp>.9806524Z npm ERR! code ERESOLVE', '<timestamp>.9821658Z npm ERR! ERESOLVE could not resolve', '<timestamp>.9834661Z npm ERR! Could not resolve dependency:']
```
2026-03-30T17:36:56.9845175Z npm ERR! Fix the upstream dependency conflict, or retry
2026-03-30T17:36:56.9846064Z npm ERR! this command with --force, or --legacy-peer-deps
2026-03-30T17:36:56.9847194Z npm ERR! to accept an incorrect (and potentially broken) dependency resolution.
2026-03-30T17:36:56.9849626Z npm ERR! 
2026-03-30T17:36:56.9851520Z npm ERR! See /home/runner/.npm/eresolve-report.txt for a full report.
2026-03-30T17:36:56.9852135Z 
2026-03-30T17:36:56.9852540Z npm ERR! A complete log of this run can be found in:
2026-03-30T17:36:56.9853597Z npm ERR!     /home/runner/.npm/_logs/2026-03-30T17_36_55_332Z-debug-0.log
2026-03-30T17:36:56.9956190Z ##[error]Process completed with exit code 1.
2026-03-30T17:36:57.0149055Z Post job cleanup.
2026-03-30T17:36:57.1110928Z [command]/usr/bin/git version
2026-03-30T17:36:57.1151417Z git version 2.53.0
2026-03-30T17:36:57.1196823Z Temporarily overriding HOME='/home/runner/work/_temp/8df42b5b-4b65-41eb-bbc7-fe0664f503ef' before making global git config changes
2026-03-30T17:36:57.1199848Z Adding repository directory to the temporary git global config as a safe directory
2026-03-30T17:36:57.1201334Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/homebridge-flexom/homebridge-flexom
2026-03-30T17:36:57.1236253Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-30T17:36:57.1268393Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-30T17:36:57.1513664Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-30T17:36:57.1538388Z http.https://github.com/.extraheader
2026-03-30T17:36:57.1550757Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-30T17:36:57.1583511Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-30T17:36:57.1827651Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-30T17:36:57.1861097Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-30T17:36:57.2215204Z Cleaning up orphan processes
2026-03-30T17:36:57.2488983Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-node@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `82088405f0e2fd854621...` conf=0.42
**Evidence:** ['<timestamp>.9424239Z npm ERR! code ERESOLVE', '<timestamp>.9438984Z npm ERR! ERESOLVE could not resolve', '<timestamp>.9449288Z npm ERR! Could not resolve dependency:']
```
2026-03-09T19:05:26.9457229Z npm ERR! 
2026-03-09T19:05:26.9457781Z npm ERR! Fix the upstream dependency conflict, or retry
2026-03-09T19:05:26.9458577Z npm ERR! this command with --force, or --legacy-peer-deps
2026-03-09T19:05:26.9459596Z npm ERR! to accept an incorrect (and potentially broken) dependency resolution.
2026-03-09T19:05:26.9460381Z npm ERR! 
2026-03-09T19:05:26.9461066Z npm ERR! See /home/runner/.npm/eresolve-report.txt for a full report.
2026-03-09T19:05:26.9467193Z 
2026-03-09T19:05:26.9468698Z npm ERR! A complete log of this run can be found in:
2026-03-09T19:05:26.9469701Z npm ERR!     /home/runner/.npm/_logs/2026-03-09T19_05_25_419Z-debug-0.log
2026-03-09T19:05:26.9623356Z ##[error]Process completed with exit code 1.
2026-03-09T19:05:26.9757111Z Post job cleanup.
2026-03-09T19:05:27.0830547Z [command]/usr/bin/git version
2026-03-09T19:05:27.0882438Z git version 2.53.0
2026-03-09T19:05:27.0929855Z Temporarily overriding HOME='/home/runner/work/_temp/449fa23e-6aa0-4d01-8c37-327a15df9c37' before making global git config changes
2026-03-09T19:05:27.0931483Z Adding repository directory to the temporary git global config as a safe directory
2026-03-09T19:05:27.0937669Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/homebridge-flexom/homebridge-flexom
2026-03-09T19:05:27.1992349Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-09T19:05:27.2054624Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-09T19:05:27.2928681Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-09T19:05:27.2966044Z http.https://github.com/.extraheader
2026-03-09T19:05:27.2995051Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-09T19:05:27.3051339Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-09T19:05:27.3434452Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-09T19:05:27.3496414Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-09T19:05:27.4009764Z Cleaning up orphan processes
```

### `go-sum-missing` — 47 matches, 47 low-conf (100%)

**Fixture:** `920f07b55bbb6ba96090...` conf=0.37
**Evidence:** ['<timestamp>.8443980Z ##[error]../../../go/pkg/mod/github.com/argoproj/argo-cd/v3@v3.1.2/util/rbac/rbac.go:26:2: missing go.sum entry for module providing package k8s.io/api/core/v1 (imported by github.com/peak-scale/capsule-argo-addon/internal/controllers/repositories); to add:']
```
2026-03-27T18:15:55.8933600Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "runtime"
2026-03-27T18:15:55.8934348Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "record"
2026-03-27T18:15:55.8935082Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "rest"
2026-03-27T18:15:55.8935841Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "runtime"
2026-03-27T18:15:55.8936762Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "record"
2026-03-27T18:15:55.8938448Z ##[error]cmd/main.go:26:2: missing go.sum entry for module providing package k8s.io/client-go/plugin/pkg/client/auth (imported by github.com/peak-scale/capsule-argo-addon/cmd); to add:
2026-03-27T18:15:55.8939684Z 	go get github.com/peak-scale/capsule-argo-addon/cmd
2026-03-27T18:15:55.8940028Z Error: not all generators ran successfully
2026-03-27T18:15:55.8940879Z run `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -w` to see all available markers, or `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -h` for usage
2026-03-27T18:15:55.8941674Z make: *** [Makefile:70: generate] Error 1
2026-03-27T18:15:55.8946548Z ##[error]Process completed with exit code 2.
2026-03-27T18:15:56.0316399Z Post job cleanup.
2026-03-27T18:15:56.1154194Z [command]/usr/bin/git version
2026-03-27T18:15:56.1221778Z git version 2.53.0
2026-03-27T18:15:56.1259514Z Temporarily overriding HOME='/home/runner/work/_temp/80e6a068-0ea4-4597-8f8a-7444ec3365cc' before making global git config changes
2026-03-27T18:15:56.1260450Z Adding repository directory to the temporary git global config as a safe directory
2026-03-27T18:15:56.1265593Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/capsule-argo-addon/capsule-argo-addon
2026-03-27T18:15:56.1300178Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-27T18:15:56.1331930Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-27T18:15:56.1547055Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-27T18:15:56.1568704Z http.https://github.com/.extraheader
2026-03-27T18:15:56.1579942Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-27T18:15:56.1610642Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-27T18:15:56.1952985Z Cleaning up orphan processes
2026-03-27T18:15:56.2353546Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `fba116edf9686a9af6a8...` conf=0.37
**Evidence:** ['<timestamp>.8179921Z ##[error]../../../go/pkg/mod/github.com/argoproj/argo-cd/v3@v3.1.2/util/rbac/rbac.go:26:2: missing go.sum entry for module providing package k8s.io/api/core/v1 (imported by github.com/peak-scale/capsule-argo-addon/internal/controllers/repositories); to add:']
```
2026-04-07T05:47:05.8680188Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "runtime"
2026-04-07T05:47:05.8680929Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "record"
2026-04-07T05:47:05.8681640Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "rest"
2026-04-07T05:47:05.8682381Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "runtime"
2026-04-07T05:47:05.8683149Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "record"
2026-04-07T05:47:05.8684743Z ##[error]cmd/main.go:26:2: missing go.sum entry for module providing package k8s.io/client-go/plugin/pkg/client/auth (imported by github.com/peak-scale/capsule-argo-addon/cmd); to add:
2026-04-07T05:47:05.8686120Z 	go get github.com/peak-scale/capsule-argo-addon/cmd
2026-04-07T05:47:05.8686452Z Error: not all generators ran successfully
2026-04-07T05:47:05.8687438Z run `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -w` to see all available markers, or `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -h` for usage
2026-04-07T05:47:05.8688240Z make: *** [Makefile:70: generate] Error 1
2026-04-07T05:47:05.8693091Z ##[error]Process completed with exit code 2.
2026-04-07T05:47:05.8811297Z Post job cleanup.
2026-04-07T05:47:05.9605270Z [command]/usr/bin/git version
2026-04-07T05:47:05.9642958Z git version 2.53.0
2026-04-07T05:47:05.9681292Z Temporarily overriding HOME='/home/runner/work/_temp/404c002f-3492-4dc6-9510-0196e920afaa' before making global git config changes
2026-04-07T05:47:05.9683502Z Adding repository directory to the temporary git global config as a safe directory
2026-04-07T05:47:05.9687630Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/capsule-argo-addon/capsule-argo-addon
2026-04-07T05:47:05.9725925Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-07T05:47:05.9761222Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-07T05:47:05.9988842Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-07T05:47:06.0012681Z http.https://github.com/.extraheader
2026-04-07T05:47:06.0023271Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-07T05:47:06.0054672Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-07T05:47:06.0442905Z Cleaning up orphan processes
2026-04-07T05:47:06.0705909Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `fefb224d1171359563de...` conf=0.37
**Evidence:** ['<timestamp>.8458292Z ##[error]../../../go/pkg/mod/github.com/argoproj/argo-cd/v3@v3.1.2/util/rbac/rbac.go:26:2: missing go.sum entry for module providing package k8s.io/api/core/v1 (imported by github.com/peak-scale/capsule-argo-addon/internal/controllers/repositories); to add:']
```
2026-03-27T14:37:27.8954098Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "runtime"
2026-03-27T14:37:27.8954873Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "record"
2026-03-27T14:37:27.8955585Z github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant:-: use of unimported package "rest"
2026-03-27T14:37:27.8956353Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "runtime"
2026-03-27T14:37:27.8957114Z github.com/peak-scale/capsule-argo-addon/internal/controllers/translator:-: use of unimported package "record"
2026-03-27T14:37:27.8958705Z ##[error]cmd/main.go:26:2: missing go.sum entry for module providing package k8s.io/client-go/plugin/pkg/client/auth (imported by github.com/peak-scale/capsule-argo-addon/cmd); to add:
2026-03-27T14:37:27.8960261Z 	go get github.com/peak-scale/capsule-argo-addon/cmd
2026-03-27T14:37:27.8960615Z Error: not all generators ran successfully
2026-03-27T14:37:27.8961457Z run `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -w` to see all available markers, or `controller-gen object:headerFile=hack/boilerplate.go.txt paths=./... -h` for usage
2026-03-27T14:37:27.8962235Z make: *** [Makefile:70: generate] Error 1
2026-03-27T14:37:27.8966459Z ##[error]Process completed with exit code 2.
2026-03-27T14:37:27.9083475Z Post job cleanup.
2026-03-27T14:37:27.9862751Z [command]/usr/bin/git version
2026-03-27T14:37:27.9899406Z git version 2.53.0
2026-03-27T14:37:27.9937153Z Temporarily overriding HOME='/home/runner/work/_temp/ca31a6fb-5a33-4b77-b583-5087c0895790' before making global git config changes
2026-03-27T14:37:27.9938546Z Adding repository directory to the temporary git global config as a safe directory
2026-03-27T14:37:27.9943604Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/capsule-argo-addon/capsule-argo-addon
2026-03-27T14:37:27.9978347Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-27T14:37:28.0036596Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-27T14:37:28.0255738Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-27T14:37:28.0276855Z http.https://github.com/.extraheader
2026-03-27T14:37:28.0287244Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-27T14:37:28.0317808Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-27T14:37:28.0681582Z Cleaning up orphan processes
2026-03-27T14:37:28.0950530Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `missing-test-fixture` — 43 matches, 43 low-conf (100%)

**Fixture:** `14e86fcf034f0f1bcb6c...` conf=0.38
**Evidence:** ['<timestamp>.5017939Z > genomeinfo : https://raw.githubusercontent.com/nf-core/test-datasets/magmap/testdata/duplicates_genomes.csv']
```
2026-04-16T04:03:51.7028764Z 
2026-04-16T04:03:51.7028892Z  2634 ┤                                                                 [32m╭───[0m
2026-04-16T04:03:51.7029197Z  2371 ┤                                                           [32m╭─────╯[0m
2026-04-16T04:03:51.7029851Z  2107 ┤                                                     [32m╭─────╯[0m
2026-04-16T04:03:51.7030177Z  1844 ┤                                            [32m╭────────╯[0m
2026-04-16T04:03:51.7030467Z  1581 ┤                                        [32m╭───╯[0m
2026-04-16T04:03:51.7030737Z  1317 ┤                                    [32m╭───╯[0m
2026-04-16T04:03:51.7031025Z  1054 ┤                                 [32m╭──╯[0m
2026-04-16T04:03:51.7031383Z   790 ┤                             [32m╭───╯[0m
2026-04-16T04:03:51.7031647Z   527 ┤                          [32m╭──╯[0m
2026-04-16T04:03:51.7031889Z   264 ┤    [32m╭─────────────────────╯[0m
2026-04-16T04:03:51.7032187Z     0 ┼[32m────╯[91m────────────────────────────────────────────────────────────────[0m
2026-04-16T04:03:51.7032579Z        Network I/O by Direction (MB) (min: 0.24, max: 2634.27, avg: 580.26 MB)
2026-04-16T04:03:51.7032811Z 
2026-04-16T04:03:51.7032952Z                      [91m■[0m enp39s0 (transmit)   [32m■[0m enp39s0 (receive)
2026-04-16T04:03:51.7033166Z 
2026-04-16T04:03:51.7033333Z 📈 Summary Statistics:
2026-04-16T04:03:51.7033617Z --------------------------------------------------------------------------------
2026-04-16T04:03:51.7034013Z   system.cpu.load_average.1m                     min: 0.08  max: 3.17  avg: 1.88 
2026-04-16T04:03:51.7034408Z   system.cpu.load_average.5m                     min: 0.02  max: 1.16  avg: 0.58 
2026-04-16T04:03:51.7034825Z   system.memory.utilization (used)               min: 2.09  max: 29.12  avg: 11.96 %
2026-04-16T04:03:51.7035192Z ================================================================================
2026-04-16T04:03:52.2755106Z 📤 metrics.jsonl uploaded (144970 bytes) to s3://runs-on-s3bucketcache-ld4yrjfcrlzx/cache/metrics/v1/nf-core/magmap/24490261558/nf-test/i-0f493f6d9c1e8187e/metrics.jsonl
2026-04-16T04:03:52.2853299Z Cleaning up orphan processes
2026-04-16T04:03:52.3251415Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: ./setup-nextflow/subaction, actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683, actions/setup-java@8df1039502a15bceb9433410b1a100fbe190c53b, nf-core/setup-nf-test@v1. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `0b20e78d68389bd1ee91...` conf=0.37
**Evidence:** ['<timestamp>.3972192Z > genomeinfo : https://raw.githubusercontent.com/nf-core/test-datasets/magmap/testdata/genometest.csv']
```
2026-04-18T01:34:00.2129722Z  5258 ┤                  [91m╭╯[0m
2026-04-18T01:34:00.2129998Z  4674 ┤               [91m╭──╯[0m
2026-04-18T01:34:00.2130208Z  4090 ┤              [91m╭╯[0m
2026-04-18T01:34:00.2130413Z  3506 ┤             [91m╭╯[0m
2026-04-18T01:34:00.2130600Z  2921 ┤      [91m╭──────╯[0m
2026-04-18T01:34:00.2130788Z  2337 ┤     [91m╭╯[0m
2026-04-18T01:34:00.2130964Z  1753 ┤   [91m╭─╯[0m
2026-04-18T01:34:00.2131135Z  1169 ┤  [91m╭╯[0m
2026-04-18T01:34:00.2131296Z   584 ┤[91m╭─╯[0m
2026-04-18T01:34:00.2131529Z     0 ┼[32m─────────────────────────────────────────────────────────────────────[0m
2026-04-18T01:34:00.2131880Z        Network I/O by Direction (MB) (min: 0.26, max: 5842.34, avg: 2476.56 MB)
2026-04-18T01:34:00.2132095Z 
2026-04-18T01:34:00.2132252Z                      [91m■[0m enp39s0 (receive)   [32m■[0m enp39s0 (transmit)
2026-04-18T01:34:00.2132455Z 
2026-04-18T01:34:00.2132538Z 📈 Summary Statistics:
2026-04-18T01:34:00.2132787Z --------------------------------------------------------------------------------
2026-04-18T01:34:00.2133162Z   system.cpu.load_average.1m                     min: 0.23  max: 7.81  avg: 2.35 
2026-04-18T01:34:00.2133542Z   system.cpu.load_average.5m                     min: 0.05  max: 3.72  avg: 2.13 
2026-04-18T01:34:00.2133924Z   system.memory.utilization (used)               min: 1.99  max: 44.31  avg: 30.82 %
2026-04-18T01:34:00.2134252Z ================================================================================
2026-04-18T01:34:00.8520142Z 📤 metrics.jsonl uploaded (2738930 bytes) to s3://runs-on-s3bucketcache-ld4yrjfcrlzx/cache/metrics/v1/nf-core/magmap/24593020827/nf-test/i-0764dad37bb9ee1c5/metrics.jsonl
2026-04-18T01:34:00.8652017Z Cleaning up orphan processes
2026-04-18T01:34:00.9022705Z Terminate orphan process: pid (177220) (conda)
2026-04-18T01:34:00.9044049Z Terminate orphan process: pid (177261) (python)
2026-04-18T01:34:00.9057395Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: ./setup-nextflow/subaction, actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683, actions/setup-java@8df1039502a15bceb9433410b1a100fbe190c53b, conda-incubator/setup-miniconda@505e6394dae86d6a5c7fbb6e3fb8938e3e863830, nf-core/setup-nf-test@v1. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `4e5d6b7eefd0ee04ec81...` conf=0.33
**Evidence:** ['<timestamp>.6923430Z ✅ test/testdata/erm/DO1.json']
```
2026-02-12T14:40:03.5150942Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-12T14:40:03.5387373Z Removing HTTP extra header
2026-02-12T14:40:03.5392693Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-12T14:40:03.5427162Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-12T14:40:03.5653320Z Removing includeIf entries pointing to credentials config files
2026-02-12T14:40:03.5659968Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-12T14:40:03.5685676Z includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git.path
2026-02-12T14:40:03.5686566Z includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git/worktrees/*.path
2026-02-12T14:40:03.5687312Z includeif.gitdir:/github/workspace/.git.path
2026-02-12T14:40:03.5687919Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-12T14:40:03.5697157Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git.path
2026-02-12T14:40:03.5719796Z /home/runner/work/_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5731376Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git.path /home/runner/work/_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5764520Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git/worktrees/*.path
2026-02-12T14:40:03.5786560Z /home/runner/work/_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5796777Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/vcmi-wasm/vcmi-wasm/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5826366Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-12T14:40:03.5846318Z /github/runner_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5854424Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5884474Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-12T14:40:03.5904493Z /github/runner_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5913952Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config
2026-02-12T14:40:03.5949564Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-12T14:40:03.6195565Z Removing credentials config '/home/runner/work/_temp/git-credentials-5a8f4f45-bdcb-4199-bfad-b9d8fd1297a4.config'
2026-02-12T14:40:03.6354155Z Cleaning up orphan processes
```

### `git-shallow-checkout` — 40 matches, 22 low-conf (55%)

**Fixture:** `d498d05a6f7574b605e7...` conf=0.37
**Evidence:** ['<timestamp>.6884522Z fetch-depth: 0']
```
2026-02-06T21:34:17.6705343Z ##[group]Run actions/upload-artifact@v4
2026-02-06T21:34:17.6706255Z with:
2026-02-06T21:34:17.6706921Z   name: buildlog.txt
2026-02-06T21:34:17.6707655Z   path: ./buildroot/buildlog.txt
2026-02-06T21:34:17.6708468Z   if-no-files-found: warn
2026-02-06T21:34:17.6709232Z   compression-level: 6
2026-02-06T21:34:17.6709955Z   overwrite: false
2026-02-06T21:34:17.6710681Z   include-hidden-files: false
2026-02-06T21:34:17.6711425Z env:
2026-02-06T21:34:17.6712062Z   CI: true
2026-02-06T21:34:17.6712718Z ##[endgroup]
2026-02-06T21:34:18.0472582Z ##[warning]No files were found with the provided path: ./buildroot/buildlog.txt. No artifacts will be uploaded.
2026-02-06T21:34:18.0739682Z ##[group]Run rm -rf ./*
2026-02-06T21:34:18.0740523Z [36;1mrm -rf ./*[0m
2026-02-06T21:34:18.0741459Z [36;1moutput_check=/home/ec2-user/actions-runner/cache/output[0m
2026-02-06T21:34:18.0742561Z [36;1mrm -rf ${output_check:-/does/not/exist}[0m
2026-02-06T21:34:18.0760290Z shell: /usr/bin/bash -e {0}
2026-02-06T21:34:18.0761044Z env:
2026-02-06T21:34:18.0761693Z   CI: true
2026-02-06T21:34:18.0762383Z ##[endgroup]
2026-02-06T21:34:22.7022780Z Post job cleanup.
2026-02-06T21:34:22.8435689Z Post job cleanup.
2026-02-06T21:34:22.9828842Z Post job cleanup.
2026-02-06T21:34:23.1160159Z Evaluate and set job outputs
2026-02-06T21:34:23.1170894Z Cleaning up orphan processes
```

**Fixture:** `4cb8dec01bc71e65c22a...` conf=0.49
**Evidence:** ['<timestamp>.5697843Z fetch-depth: 0']
```
2026-01-29T13:03:47.6517248Z Could not load layer Surface reflectance (Sentinel-2): Could not find product s2_l2a in datacube for layer s2_l2a
2026-01-29T13:03:47.6523876Z Could not load layer Surface reflectance (Landsat 9): Could not find product ls9_sr in datacube for layer ls9_sr
2026-01-29T13:03:47.6530557Z Could not load layer Surface reflectance (Landsat 8): Could not find product ls8_sr in datacube for layer ls8_sr
2026-01-29T13:03:47.6536638Z Could not load layer Surface reflectance (Landsat 7): Could not find product ls7_sr in datacube for layer ls7_sr
2026-01-29T13:03:47.6543039Z Could not load layer Surface reflectance (Landsat 5): Could not find product ls5_sr in datacube for layer ls5_sr
2026-01-29T13:03:47.6549725Z Could not load layer Surface temperature (Landsat 9): Could not find product ls9_st in datacube for layer ls9_st
2026-01-29T13:03:47.6555950Z Could not load layer Surface temperature (Landsat 8): Could not find product ls8_st in datacube for layer ls8_st
2026-01-29T13:03:47.6562300Z Could not load layer Surface temperature (Landsat 7): Could not find product ls7_st in datacube for layer ls7_st
2026-01-29T13:03:47.6568940Z Could not load layer Surface temperature (Landsat 5): Could not find product ls5_st in datacube for layer ls5_st
2026-01-29T13:03:47.6575711Z Could not load layer Monthly mosaic (Sentinel-1): Could not find product s1_monthly_mosaic in datacube for layer s1_monthly_mosaic
2026-01-29T13:03:47.6582491Z Could not load layer Annual mosaic (ALOS/PALSAR): Could not find product alos_palsar_mosaic in datacube for layer alos_palsar_mosaic
2026-01-29T13:03:47.6588525Z Could not load layer Annual mosaic (JERS): Could not find product jers_sar_mosaic in datacube for layer jers_sar_mosaic
2026-01-29T13:03:47.6595048Z Could not load layer World Settlement Footprint 2015: Could not find product wsf_2015 in datacube for layer wsf_2015
2026-01-29T13:03:47.6601702Z Could not load layer World Settlement Footprint 2019: Could not find product wsf_2019 in datacube for layer wsf_2019
2026-01-29T13:03:47.6608703Z Could not load layer World Settlement Footprint Evolution: Could not find product wsf_evolution in datacube for layer wsf_evolution
2026-01-29T13:03:47.6614791Z Could not load layer Global Mangrove Watch: Could not find product gmw in datacube for layer gmw
2026-01-29T13:03:47.6622225Z {'name': 'prob', 'title': 'Probability of cropping', 'abstract': 'Probability of cropping', 'needed_bands': ['prob'], 'index_function': {'function': 'datacube_ows.band_utils.single_band', 'mapped_bands': True, 'kwargs': {'band': 'prob'}}, 'color_ramp': [{'value': 0, 'color': 'black'}, {'value': 1, 'color': '#010007'}, {'value': 10, 'color': '#170b3b'}, {'value': 20, 'color': '#410967'}, {'value': 30, 'color': '#6b176e'}, {'value': 40, 'color': '#952666'}, {'value': 50, 'color': '#bb3754'}, {'value': 60, 'color': '#dd5238'}, {'value': 70, 'color': '#f37719'}, {'value': 80, 'color': '#fba60b'}, {'value': 90, 'color': '#f5d948'}, {'value': 100, 'color': '#fcfea4'}], 'range': [0, 100], 'legend': {'begin': '0', 'end': '100', 'ticks_every': '20'}}
2026-01-29T13:03:47.6630117Z {'name': 'prob', 'title': 'Probability of cropping', 'abstract': 'Probability of cropping', 'needed_bands': ['prob'], 'index_function': {'function': 'datacube_ows.band_utils.single_band', 'mapped_bands': True, 'kwargs': {'band': 'prob'}}, 'color_ramp': [{'value': 0, 'color': 'black', 'alpha': 0.0}, {'value': 1, 'color': '#010007'}, {'value': 10, 'color': '#170b3b'}, {'value': 20, 'color': '#410967'}, {'value': 30, 'color': '#6b176e'}, {'value': 40, 'color': '#952666'}, {'value': 50, 'color': '#bb3754'}, {'value': 60, 'color': '#dd5238'}, {'value': 70, 'color': '#f37719'}, {'value': 80, 'color': '#fba60b'}, {'value': 90, 'color': '#f5d948'}, {'value': 100, 'color': '#fcfea4'}], 'range': [0, 100], 'legend': {'begin': '0', 'end': '100', 'ticks_every': '20'}}
2026-01-29T13:03:47.6634005Z Configuration parsed OK
2026-01-29T13:03:47.6634420Z Configured message file location: None
2026-01-29T13:03:47.6635156Z Configured translations directory location: /env/config/ows_refactored/translations
2026-01-29T13:03:47.6756220Z {'values_changed': {"root['total_layers_count']": {'new_value': 47, 'old_value': 50}}, 'iterable_item_added': {"root['layers'][14]": {'layer': 'esa_worldcereal_temporarycrops', 'product': ['esa_worldcereal_temporarycrops'], 'styles_count': 1, 'styles_list': ['style_temporarycrops']}, "root['layers'][16]": {'layer': 'esa_worldcereal_maize_main', 'product': ['esa_worldcereal_maize_main'], 'styles_count': 1, 'styles_list': ['style_maize']}, "root['layers'][17]": {'layer': 'esa_worldcereal_maize_active', 'product': ['esa_worldcereal_maize_active'], 'styles_count': 1, 'styles_list': ['style_activecropland']}, "root['layers'][18]": {'layer': 'esa_worldcereal_maize_irrigation', 'product': ['esa_worldcereal_maize_irrigation'], 'styles_count': 1, 'styles_list': ['style_irrigated']}, "root['layers'][19]": {'layer': 'esa_worldcereal_wintercereals', 'product': ['esa_worldcereal_wintercereals'], 'styles_count': 1, 'styles_list': ['style_wintercereals']}, "root['layers'][20]": {'layer': 'esa_worldcereal_wintercereals_irrigation', 'product': ['esa_worldcereal_wintercereals_irrigation'], 'styles_count': 1, 'styles_list': ['style_irrigated']}}}
2026-01-29T13:03:48.0655757Z 
2026-01-29T13:03:48.0701914Z ##[error]Process completed with exit code 1.
2026-01-29T13:03:48.0767711Z Cleaning up orphan processes
```

**Fixture:** `49ce45fc4b6fe77faca3...` conf=0.37
**Evidence:** ['<timestamp>.5286500Z fetch-depth: 0']
```
2026-03-19T04:37:03.2591481Z With the provided path, there will be 1 file uploaded
2026-03-19T04:37:03.2599511Z Artifact name is valid!
2026-03-19T04:37:03.2612708Z Root directory input is valid!
2026-03-19T04:37:03.3878777Z Beginning upload of artifact content to blob storage
2026-03-19T04:37:04.1535996Z Uploaded bytes 2224929
2026-03-19T04:37:04.1783044Z Finished uploading artifact content to blob storage!
2026-03-19T04:37:04.1787671Z SHA256 digest of uploaded artifact zip is 497852302602798f44d91cb30d46cd8cbfd0dc457aad6211bf5a716031b3a1e7
2026-03-19T04:37:04.1789624Z Finalizing artifact upload
2026-03-19T04:37:04.2661890Z Artifact buildlog.txt.zip successfully finalized. Artifact ID 5999700869
2026-03-19T04:37:04.2663961Z Artifact buildlog.txt has been successfully uploaded! Final size is 2224929 bytes. Artifact ID is 5999700869
2026-03-19T04:37:04.2673897Z Artifact download URL: https://github.com/Opentrons/buildroot/actions/runs/23276641613/artifacts/5999700869
2026-03-19T04:37:04.2990900Z ##[group]Run rm -rf ./*
2026-03-19T04:37:04.2991724Z [36;1mrm -rf ./*[0m
2026-03-19T04:37:04.2992833Z [36;1moutput_check=/home/ec2-user/actions-runner/cache/output[0m
2026-03-19T04:37:04.2994036Z [36;1mrm -rf ${output_check:-/does/not/exist}[0m
2026-03-19T04:37:04.3013424Z shell: /usr/bin/bash -e {0}
2026-03-19T04:37:04.3014186Z env:
2026-03-19T04:37:04.3014831Z   CI: true
2026-03-19T04:37:04.3015664Z ##[endgroup]
2026-03-19T04:37:32.6398505Z Post job cleanup.
2026-03-19T04:37:32.7832048Z Post job cleanup.
2026-03-19T04:37:32.9250842Z Post job cleanup.
2026-03-19T04:37:33.0596399Z Evaluate and set job outputs
2026-03-19T04:37:33.0618688Z Cleaning up orphan processes
2026-03-19T04:37:33.1428598Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v3, actions/upload-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `zero-tests-executed` — 40 matches, 40 low-conf (100%)

**Fixture:** `93f83e9132944adb032e...` conf=0.38
**Evidence:** ['<timestamp>.3444754Z running 0 tests', '<timestamp>.3445692Z test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s']
```
2026-02-25T15:56:41.4969592Z      if free.len() > 1 {
2026-02-25T15:56:41.4969889Z -        return Err(format!(
2026-02-25T15:56:41.4970224Z -            "unexpected extra positional argument: {}",
2026-02-25T15:56:41.4970570Z -            free[1]
2026-02-25T15:56:41.4970798Z -        ));
2026-02-25T15:56:41.4971256Z +        return Err(format!("unexpected extra positional argument: {}", free[1]));
2026-02-25T15:56:41.4971706Z      }
2026-02-25T15:56:41.4971925Z      let integer = free.pop();
2026-02-25T15:56:41.4972187Z      Ok(Cli {
2026-02-25T15:56:41.5442647Z ##[error]Process completed with exit code 1.
2026-02-25T15:56:41.5551121Z Post job cleanup.
2026-02-25T15:56:41.6502618Z [command]/usr/bin/git version
2026-02-25T15:56:41.6540492Z git version 2.53.0
2026-02-25T15:56:41.6585343Z Temporarily overriding HOME='/home/runner/work/_temp/5a010132-d736-4aee-8b3c-777aacb63615' before making global git config changes
2026-02-25T15:56:41.6586684Z Adding repository directory to the temporary git global config as a safe directory
2026-02-25T15:56:41.6592300Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/rust-number-theory/rust-number-theory
2026-02-25T15:56:41.6630078Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-25T15:56:41.6664183Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-25T15:56:41.7170517Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-25T15:56:41.7193064Z http.https://github.com/.extraheader
2026-02-25T15:56:41.7206489Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-25T15:56:41.7238803Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-25T15:56:41.7470190Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-25T15:56:41.7502016Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-25T15:56:41.7845853Z Cleaning up orphan processes
```

**Fixture:** `f9a6d770083bae0cfed3...` conf=0.38
**Evidence:** ['<timestamp>.0839239Z running 0 tests', '<timestamp>.0840026Z test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s']
```
2026-03-01T11:05:29.0797179Z 
2026-03-01T11:05:29.0797334Z --
2026-03-01T11:05:29.0797750Z 
2026-03-01T11:05:29.0797927Z ********************
2026-03-01T11:05:36.4638866Z PASS: Churchroad tests :: wide_mul_yosys.v (2 of 4)
2026-03-01T11:05:40.3786343Z PASS: Churchroad tests :: simple_mul.v (3 of 4)
2026-03-01T11:05:41.2988180Z PASS: Churchroad tests :: wide_mul.v (4 of 4)
2026-03-01T11:05:41.3162233Z ********************
2026-03-01T11:05:41.3162610Z Failed Tests (1):
2026-03-01T11:05:41.3163116Z   Churchroad tests :: nextmap_eval_signed_mac.sv
2026-03-01T11:05:41.3163474Z 
2026-03-01T11:05:41.3163480Z 
2026-03-01T11:05:41.3163591Z Testing Time: 12.91s
2026-03-01T11:05:41.3163912Z   Passed: 3
2026-03-01T11:05:41.3164168Z   Failed: 1
2026-03-01T11:05:41.7483498Z ##[error]Process completed with exit code 1.
2026-03-01T11:05:41.7550349Z Post job cleanup.
2026-03-01T11:05:42.0565347Z ##[group]Logout from ghcr.io
2026-03-01T11:05:42.0612572Z [command]/usr/bin/docker logout ghcr.io
2026-03-01T11:05:42.0752878Z Removing login credentials for ghcr.io
2026-03-01T11:05:42.0785076Z ##[endgroup]
2026-03-01T11:05:42.0785789Z ##[group]Post cache
2026-03-01T11:05:42.0787170Z State not set
2026-03-01T11:05:42.0787873Z ##[endgroup]
2026-03-01T11:05:42.0912245Z Cleaning up orphan processes
```

**Fixture:** `558cb5e0623e7445ed9d...` conf=0.38
**Evidence:** ['<timestamp>.3146670Z running 0 tests', '<timestamp>.3147687Z test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s']
```
2026-03-02T03:10:55.7058872Z     parse::test_parse_3
2026-03-02T03:10:55.7059056Z     sema::test_sema_1
2026-03-02T03:10:55.7059230Z     sema::test_sema_2
2026-03-02T03:10:55.7059406Z     sema::test_sema_3
2026-03-02T03:10:55.7059580Z     sema::test_sema_4
2026-03-02T03:10:55.7059748Z     sema::test_sema_5
2026-03-02T03:10:55.7059852Z 
2026-03-02T03:10:55.7060099Z test result: FAILED. 4 passed; 8 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.36s
2026-03-02T03:10:55.7060441Z 
2026-03-02T03:10:55.7071223Z ##[error]Process completed with exit code 101.
2026-03-02T03:10:55.7192433Z Post job cleanup.
2026-03-02T03:10:55.8161119Z [command]/usr/bin/git version
2026-03-02T03:10:55.8198796Z git version 2.53.0
2026-03-02T03:10:55.8244859Z Temporarily overriding HOME='/home/runner/work/_temp/f54398b8-7998-4eb0-817e-a6c8875499b6' before making global git config changes
2026-03-02T03:10:55.8246219Z Adding repository directory to the temporary git global config as a safe directory
2026-03-02T03:10:55.8259857Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/tick/tick
2026-03-02T03:10:55.8297233Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-02T03:10:55.8330883Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-02T03:10:55.8579737Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-02T03:10:55.8604379Z http.https://github.com/.extraheader
2026-03-02T03:10:55.8617676Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-02T03:10:55.8650753Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-02T03:10:55.8898663Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-02T03:10:55.8945293Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-02T03:10:55.9350600Z Cleaning up orphan processes
```

### `build-input-file-missing` — 32 matches, 26 low-conf (81%)

**Fixture:** `bdde53eb23e86d138442...` conf=0.37
**Evidence:** ['<timestamp>.0486589Z failed to run protoc and produce descriptor set: Custom { kind: Other, error: "protoc failed with exit code: 1\\nstderr: stepflow/v1/common.proto: File not found.\\r\\n" }']
```
2026-03-23T22:10:06.8741657Z   process didn't exit successfully: `D:\a\stepflow\stepflow\stepflow-rs\target\release\build\stepflow-proto-abf8ef4e6b729061\build-script-build` (exit code: 101)
2026-03-23T22:10:06.9221330Z   --- stderr
2026-03-23T22:10:07.0481022Z 
2026-03-23T22:10:07.0484034Z   thread 'main' (488) panicked at C:\Users\runneradmin\.cargo\registry\src\index.crates.io-1949cf8c6b5b557f\tonic-rest-build-0.1.5\src\helpers.rs:220:10:
2026-03-23T22:10:07.0486589Z   failed to run protoc and produce descriptor set: Custom { kind: Other, error: "protoc failed with exit code: 1\nstderr: stepflow/v1/common.proto: File not found.\r\n" }
2026-03-23T22:10:07.0645298Z   note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
2026-03-23T22:10:07.0693235Z [1m[93mwarning[0m: build failed, waiting for other jobs to finish...
2026-03-23T22:10:14.8952407Z ##[error]Process completed with exit code 101.
2026-03-23T22:10:14.9137946Z Post job cleanup.
2026-03-23T22:10:15.1072209Z [command]"C:\Program Files\Git\bin\git.exe" version
2026-03-23T22:10:15.1325004Z git version 2.53.0.windows.2
2026-03-23T22:10:15.1394012Z Temporarily overriding HOME='D:\a\_temp\45028f2f-a443-4110-8406-81f37db8d738' before making global git config changes
2026-03-23T22:10:15.1394704Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T22:10:15.1405141Z [command]"C:\Program Files\Git\bin\git.exe" config --global --add safe.directory D:\a\stepflow\stepflow
2026-03-23T22:10:15.1654491Z [command]"C:\Program Files\Git\bin\git.exe" config --local --name-only --get-regexp core\.sshCommand
2026-03-23T22:10:15.1894636Z [command]"C:\Program Files\Git\bin\git.exe" submodule foreach --recursive "sh -c \"git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :\""
2026-03-23T22:10:15.6170445Z [command]"C:\Program Files\Git\bin\git.exe" config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T22:10:15.6367353Z http.https://github.com/.extraheader
2026-03-23T22:10:15.6409125Z [command]"C:\Program Files\Git\bin\git.exe" config --local --unset-all http.https://github.com/.extraheader
2026-03-23T22:10:15.6646434Z [command]"C:\Program Files\Git\bin\git.exe" submodule foreach --recursive "sh -c \"git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :\""
2026-03-23T22:10:16.1153807Z [command]"C:\Program Files\Git\bin\git.exe" config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T22:10:16.1383650Z [command]"C:\Program Files\Git\bin\git.exe" submodule foreach --recursive "git config --local --show-origin --name-only --get-regexp remote.origin.url"
2026-03-23T22:10:16.5851753Z Cleaning up orphan processes
2026-03-23T22:10:16.6243711Z Terminate orphan process: pid (7336) (vctip)
2026-03-23T22:10:16.6258050Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/cache@v4, actions/checkout@v4, arduino/setup-protoc@v3. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `4b8ef93927191a99dec5...` conf=0.37
**Evidence:** ['<timestamp>.9356769Z <timestamp>.935510Z INFO release-plz config file not found, using default configuration']
```
2026-02-24T01:27:58.6901363Z     [1m[92m    Updating[0m crates.io index
2026-02-24T01:27:58.6901885Z     [1m[91merror[0m: 1 files in the working directory contain changes that were not yet committed into git:
2026-02-24T01:27:58.6902339Z     
2026-02-24T01:27:58.6902494Z     Cargo.lock
2026-02-24T01:27:58.6902653Z     
2026-02-24T01:27:58.6902980Z     to proceed despite this and include the uncommitted changes, pass the `--allow-dirty` flag
2026-02-24T01:27:58.6903373Z     
2026-02-24T01:27:58.6921601Z ##[error]Process completed with exit code 1.
2026-02-24T01:27:58.7004921Z Post job cleanup.
2026-02-24T01:27:58.7764064Z [command]/usr/bin/git version
2026-02-24T01:27:58.7798627Z git version 2.52.0
2026-02-24T01:27:58.7831396Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/53615a0c-351a-4a28-8490-fbe1b0ac35fd/.gitconfig'
2026-02-24T01:27:58.7840194Z Temporarily overriding HOME='/home/runner/work/_temp/53615a0c-351a-4a28-8490-fbe1b0ac35fd' before making global git config changes
2026-02-24T01:27:58.7841194Z Adding repository directory to the temporary git global config as a safe directory
2026-02-24T01:27:58.7845753Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/mdbook-pagecrypt/mdbook-pagecrypt
2026-02-24T01:27:58.7875179Z Removing SSH command configuration
2026-02-24T01:27:58.7881221Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-24T01:27:58.7913782Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-24T01:27:58.8128714Z Removing HTTP extra header
2026-02-24T01:27:58.8132750Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-24T01:27:58.8162943Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-24T01:27:58.8375336Z Removing includeIf entries pointing to credentials config files
2026-02-24T01:27:58.8380894Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-24T01:27:58.8411088Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-24T01:27:58.8733030Z Cleaning up orphan processes
```

**Fixture:** `cd0fdde99c83df5371f4...` conf=0.37
**Evidence:** ['<timestamp>.6795271Z ❌ Config file not found: modrinth_deps.json']
```
2026-02-27T18:44:57.8176888Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-27T18:44:57.8431142Z Removing HTTP extra header
2026-02-27T18:44:57.8437001Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-27T18:44:57.8475650Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-27T18:44:57.8726768Z Removing includeIf entries pointing to credentials config files
2026-02-27T18:44:57.8734024Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-27T18:44:57.8760143Z includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git.path
2026-02-27T18:44:57.8761169Z includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git/worktrees/*.path
2026-02-27T18:44:57.8761912Z includeif.gitdir:/github/workspace/.git.path
2026-02-27T18:44:57.8762566Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-27T18:44:57.8771232Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git.path
2026-02-27T18:44:57.8794059Z /home/runner/work/_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8804687Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git.path /home/runner/work/_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8839034Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git/worktrees/*.path
2026-02-27T18:44:57.8862422Z /home/runner/work/_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8872179Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/ChestESP/ChestESP/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8906151Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-27T18:44:57.8930878Z /github/runner_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8939819Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.8973354Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-27T18:44:57.8995447Z /github/runner_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.9004987Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config
2026-02-27T18:44:57.9041578Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-27T18:44:57.9299127Z Removing credentials config '/home/runner/work/_temp/git-credentials-0b16965a-5475-448f-8269-a57abded0df8.config'
2026-02-27T18:44:57.9426982Z Cleaning up orphan processes
```

### `artifact-missing` — 29 matches, 26 low-conf (90%)

**Fixture:** `1c21544c3e2bcb31df9b...` conf=0.37
**Evidence:** ['<timestamp>.8264793Z Deleted no artifacts, repos, packages or scratchspaces']
```
2026-03-09T16:54:04.6216429Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-09T16:54:04.6471150Z Removing HTTP extra header
2026-03-09T16:54:04.6476350Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-09T16:54:04.6515784Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-09T16:54:04.6764094Z Removing includeIf entries pointing to credentials config files
2026-03-09T16:54:04.6772540Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-09T16:54:04.6800192Z includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git.path
2026-03-09T16:54:04.6801269Z includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git/worktrees/*.path
2026-03-09T16:54:04.6801970Z includeif.gitdir:/github/workspace/.git.path
2026-03-09T16:54:04.6802623Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-03-09T16:54:04.6815566Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git.path
2026-03-09T16:54:04.6840098Z /home/runner/work/_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.6851867Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git.path /home/runner/work/_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.6890036Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git/worktrees/*.path
2026-03-09T16:54:04.6914610Z /home/runner/work/_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.6924792Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/MGVI.jl/MGVI.jl/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.6959943Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-03-09T16:54:04.6983412Z /github/runner_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.6992534Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.7027902Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-03-09T16:54:04.7051992Z /github/runner_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.7062280Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config
2026-03-09T16:54:04.7099402Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-09T16:54:04.7340231Z Removing credentials config '/home/runner/work/_temp/git-credentials-48741082-79d8-45c7-accd-0a1adb313173.config'
2026-03-09T16:54:04.7504092Z Cleaning up orphan processes
```

**Fixture:** `cb144b0790bfa200b089...` conf=0.37
**Evidence:** ['<timestamp>.3478175Z Deleted no artifacts, repos, packages or scratchspaces']
```
2026-02-04T16:16:44.4774363Z [90m    @[39m [90mBase[39m [90m./[39m[90m[4mclient.jl:550[24m[39m
2026-02-04T16:16:44.4775149Z in expression starting at /home/runner/work/BayesDensity.jl/BayesDensity.jl/docs/make.jl:24
2026-02-04T16:16:44.9956408Z ##[error]Process completed with exit code 1.
2026-02-04T16:16:45.0128151Z Post job cleanup.
2026-02-04T16:16:45.0153526Z Post job cleanup.
2026-02-04T16:16:46.2787967Z       Active manifest files: 1 found
2026-02-04T16:16:46.3317720Z       Active artifact files: 98 found
2026-02-04T16:16:46.3344299Z       Active scratchspaces: 0 found
2026-02-04T16:16:46.3478175Z      Deleted no artifacts, repos, packages or scratchspaces
2026-02-04T16:16:47.3515994Z No existing caches found on ref `refs/pull/4/merge` matching restore key `julia-cache;workflow=Documentation;job=build;os=Linux;`
2026-02-04T16:16:47.3849862Z Post job cleanup.
2026-02-04T16:16:47.4807923Z [command]/usr/bin/git version
2026-02-04T16:16:47.4845675Z git version 2.52.0
2026-02-04T16:16:47.4890253Z Temporarily overriding HOME='/home/runner/work/_temp/abae0148-423b-4fc3-9592-180a53a39c1e' before making global git config changes
2026-02-04T16:16:47.4891609Z Adding repository directory to the temporary git global config as a safe directory
2026-02-04T16:16:47.4903927Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/BayesDensity.jl/BayesDensity.jl
2026-02-04T16:16:47.4937987Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-04T16:16:47.4970020Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-04T16:16:47.5198031Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-04T16:16:47.5218576Z http.https://github.com/.extraheader
2026-02-04T16:16:47.5231503Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-04T16:16:47.5260988Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-04T16:16:47.5481029Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-04T16:16:47.5511187Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-04T16:16:47.5834055Z Cleaning up orphan processes
```

**Fixture:** `1903fef74fc025f7ca43...` conf=0.37
**Evidence:** ['<timestamp>.3297954Z Deleted no artifacts, repos, packages or scratchspaces']
```
2026-02-16T05:11:12.1237146Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-16T05:11:12.1596773Z Removing HTTP extra header
2026-02-16T05:11:12.1605194Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-16T05:11:12.1657464Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-16T05:11:12.1998802Z Removing includeIf entries pointing to credentials config files
2026-02-16T05:11:12.2009595Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-16T05:11:12.2054679Z includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git.path
2026-02-16T05:11:12.2055712Z includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git/worktrees/*.path
2026-02-16T05:11:12.2056544Z includeif.gitdir:/github/workspace/.git.path
2026-02-16T05:11:12.2057134Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-16T05:11:12.2063723Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git.path
2026-02-16T05:11:12.2095981Z /home/runner/work/_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2113093Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git.path /home/runner/work/_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2155353Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git/worktrees/*.path
2026-02-16T05:11:12.2180340Z /home/runner/work/_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2190827Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/MaybeInplace.jl/MaybeInplace.jl/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2225513Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-02-16T05:11:12.2250602Z /github/runner_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2259838Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2295317Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-02-16T05:11:12.2318880Z /github/runner_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2331001Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config
2026-02-16T05:11:12.2375958Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-16T05:11:12.2625689Z Removing credentials config '/home/runner/work/_temp/git-credentials-d71f9645-03c1-458e-a4d6-d1b4339a2433.config'
2026-02-16T05:11:12.2747287Z Cleaning up orphan processes
```

### `working-directory` — 18 matches, 17 low-conf (94%)

**Fixture:** `d5bb6943e5014abc7e55...` conf=0.4
**Evidence:** ['<timestamp>.1699494Z failed to run git: fatal: not a git repository (or any of the parent directories): .git']
```
2026-04-28T09:37:05.1789040Z   token: ***
2026-04-28T09:37:05.1789928Z   single_check: 47df720c-b45b-4b55-b070-1f14c9c4f7e8
2026-04-28T09:37:05.1791087Z   dispatch_only: false
2026-04-28T09:37:05.1791897Z ##[endgroup]
2026-04-28T09:37:05.2699461Z Single check mode enabled, checking status of workflow ID 47df720c-b45b-4b55-b070-1f14c9c4f7e8...
2026-04-28T09:37:05.2702433Z Checking workflow status for ID 47df720c-b45b-4b55-b070-1f14c9c4f7e8...
2026-04-28T09:37:06.0571801Z Workflow not completed yet.
2026-04-28T09:37:06.1046682Z ##[group]Run NEXT=$(( 1 + 1 ))
2026-04-28T09:37:06.1047037Z [36;1mNEXT=$(( 1 + 1 ))[0m
2026-04-28T09:37:06.1047438Z [36;1mecho "Signing not complete yet (attempt 1); scheduling attempt ${NEXT}"[0m
2026-04-28T09:37:06.1047994Z [36;1mgh workflow run windows-ossign-wait-signature.yml \[0m
2026-04-28T09:37:06.1048604Z [36;1m  --ref "${GITHUB_REF}" \[0m
2026-04-28T09:37:06.1048974Z [36;1m  -f workflow_id="47df720c-b45b-4b55-b070-1f14c9c4f7e8" \[0m
2026-04-28T09:37:06.1049372Z [36;1m  -f release_name="DesQTA v1.0.0-11" \[0m
2026-04-28T09:37:06.1049720Z [36;1m  -f attempt="${NEXT}" \[0m
2026-04-28T09:37:06.1050044Z [36;1m  -f max_attempts="100"[0m
2026-04-28T09:37:06.1070744Z shell: /usr/bin/bash -e {0}
2026-04-28T09:37:06.1071019Z env:
2026-04-28T09:37:06.1071472Z   GH_TOKEN: ***
2026-04-28T09:37:06.1071699Z ##[endgroup]
2026-04-28T09:37:06.1112680Z Signing not complete yet (attempt 1); scheduling attempt 2
2026-04-28T09:37:06.1699494Z failed to run git: fatal: not a git repository (or any of the parent directories): .git
2026-04-28T09:37:06.1700177Z 
2026-04-28T09:37:06.1723931Z ##[error]Process completed with exit code 1.
2026-04-28T09:37:06.1790925Z Cleaning up orphan processes
```

**Fixture:** `b5578e35b9b941ecb663...` conf=0.4
**Evidence:** ["<timestamp>.4377204Z <timestamp> ERROR <job_1326701334> /home/dependabot/common/lib/dependabot/shared_helpers.rb:81:in 'Dir.chdir'"]
```
2026-04-16T22:29:42.7098239Z   proxy | 2026/04/16 22:29:42 [603] PATCH /update_jobs/1326701334/mark_as_processed
2026-04-16T22:29:42.8626244Z   proxy | 2026/04/16 22:29:42 [603] 204 /update_jobs/1326701334/mark_as_processed
2026-04-16T22:29:42.8702428Z updater | 2026/04/16 22:29:42 INFO <job_1326701334> Finished job processing
2026-04-16T22:29:42.8711713Z updater | 2026/04/16 22:29:42 INFO Results:
2026-04-16T22:29:42.8712673Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-04-16T22:29:42.8713734Z +-----------------------------------------------+
2026-04-16T22:29:42.8714260Z |         Dependencies failed to update         |
2026-04-16T22:29:42.8714779Z +---------------+---------------+---------------+
2026-04-16T22:29:42.8715257Z | Dependency    | Error Type    | Error Details |
2026-04-16T22:29:42.8715737Z +---------------+---------------+---------------+
2026-04-16T22:29:42.8716217Z | rustls-webpki | unknown_error | null          |
2026-04-16T22:29:42.8716699Z +---------------+---------------+---------------+
2026-04-16T22:29:43.0025469Z Failure running container 8ad6bf7001befccdc8169f66dcf1f545bca457ebf7fdb920ebd73ed756ad7967: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-04-16T22:29:43.1489943Z Cleaned up container 8ad6bf7001befccdc8169f66dcf1f545bca457ebf7fdb920ebd73ed756ad7967
2026-04-16T22:29:43.1588310Z   proxy | 2026/04/16 22:29:43 1/574 calls cached (0%)
2026-04-16T22:29:43.1593569Z   proxy | 2026/04/16 22:29:43 Posting metrics to remote API endpoint
2026-04-16T22:29:43.8738148Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/RealViper8/Steamtools/network/updates/1326701334 (write access to the repository is required to view the log)
2026-04-16T22:29:43.8749261Z 🤖 ~ finished: error reported to Dependabot ~
2026-04-16T22:29:43.8832507Z Post job cleanup.
2026-04-16T22:29:44.0425410Z Cleaning up orphan processes
2026-04-16T22:29:44.0761107Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `b98008095ac378288b8c...` conf=0.4
**Evidence:** ['<timestamp>.2242812Z failed to determine base repo: failed to run git: fatal: not a git repository (or any of the parent directories): .git']
```
2026-02-15T00:21:12.0170355Z ##[group]Runner Image
2026-02-15T00:21:12.0170963Z Image: ubuntu-24.04
2026-02-15T00:21:12.0171489Z Version: 20260209.23.1
2026-02-15T00:21:12.0172683Z Included Software: https://github.com/actions/runner-images/blob/ubuntu24/20260209.23/images/ubuntu/Ubuntu2404-Readme.md
2026-02-15T00:21:12.0174196Z Image Release: https://github.com/actions/runner-images/releases/tag/ubuntu24%2F20260209.23
2026-02-15T00:21:12.0175067Z ##[endgroup]
2026-02-15T00:21:12.0176169Z ##[group]GITHUB_TOKEN Permissions
2026-02-15T00:21:12.0178056Z Contents: read
2026-02-15T00:21:12.0178876Z Metadata: read
2026-02-15T00:21:12.0179456Z Packages: read
2026-02-15T00:21:12.0180095Z ##[endgroup]
2026-02-15T00:21:12.0182094Z Secret source: Actions
2026-02-15T00:21:12.0182794Z Prepare workflow directory
2026-02-15T00:21:12.0587616Z Prepare all required actions
2026-02-15T00:21:12.0725347Z Complete job name: rerun-failed-jobs
2026-02-15T00:21:12.1824128Z ##[group]Run gh run rerun 22026698619 --failed
2026-02-15T00:21:12.1825780Z [36;1mgh run rerun 22026698619 --failed[0m
2026-02-15T00:21:12.2760186Z shell: /usr/bin/bash -e {0}
2026-02-15T00:21:12.2761849Z env:
2026-02-15T00:21:12.2763033Z   GH_TOKEN: ***
2026-02-15T00:21:12.2763932Z ##[endgroup]
2026-02-15T00:21:15.2242812Z failed to determine base repo: failed to run git: fatal: not a git repository (or any of the parent directories): .git
2026-02-15T00:21:15.2244231Z 
2026-02-15T00:21:15.2262652Z ##[error]Process completed with exit code 1.
2026-02-15T00:21:15.2385126Z Cleaning up orphan processes
```

### `docker-daemon-unavailable` — 17 matches, 17 low-conf (100%)

**Fixture:** `c6ece9e88b624c08ee81...` conf=0.37
**Evidence:** ['<timestamp>.3723278Z Mar 21 <timestamp> runnervm46oaq sudo[2482]: root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker']
```
2026-03-21T09:10:27.3688800Z Mar 21 09:10:23 runnervm46oaq systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-21T09:10:27.3694504Z Mar 21 09:10:23 runnervm46oaq systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-21T09:10:27.3696388Z Mar 21 09:10:23 runnervm46oaq systemd[1]: Started agent.service - Agent.
2026-03-21T09:10:27.3697866Z Mar 21 09:10:23 runnervm46oaq sudo[2388]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl stop systemd-resolved
2026-03-21T09:10:27.3699180Z Mar 21 09:10:23 runnervm46oaq sudo[2388]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3700568Z Mar 21 09:10:23 runnervm46oaq sudo[2388]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3701946Z Mar 21 09:10:23 runnervm46oaq sudo[2394]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart systemd-resolved
2026-03-21T09:10:27.3703242Z Mar 21 09:10:23 runnervm46oaq sudo[2394]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3704367Z Mar 21 09:10:23 runnervm46oaq sudo[2394]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3705596Z Mar 21 09:10:23 runnervm46oaq sudo[2401]:     root : *** ; USER=root ; COMMAND=/usr/bin/resolvectl flush-caches
2026-03-21T09:10:27.3706803Z Mar 21 09:10:23 runnervm46oaq sudo[2401]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3707940Z Mar 21 09:10:23 runnervm46oaq sudo[2401]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3709343Z Mar 21 09:10:23 runnervm46oaq sudo[2404]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl reload docker
2026-03-21T09:10:27.3710880Z Mar 21 09:10:23 runnervm46oaq sudo[2404]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3712049Z Mar 21 09:10:23 runnervm46oaq sudo[2404]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3713254Z Mar 21 09:10:23 runnervm46oaq sudo[2413]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl daemon-reload
2026-03-21T09:10:27.3714474Z Mar 21 09:10:23 runnervm46oaq sudo[2413]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3716580Z Mar 21 09:10:24 runnervm46oaq systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-21T09:10:27.3719420Z Mar 21 09:10:24 runnervm46oaq systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-21T09:10:27.3722030Z Mar 21 09:10:24 runnervm46oaq sudo[2413]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3723278Z Mar 21 09:10:24 runnervm46oaq sudo[2482]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker
2026-03-21T09:10:27.3724488Z Mar 21 09:10:24 runnervm46oaq sudo[2482]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-21T09:10:27.3725619Z Mar 21 09:10:24 runnervm46oaq sudo[2482]: pam_unix(sudo:session): session closed for user root
2026-03-21T09:10:27.3726225Z 
2026-03-21T09:10:27.7080980Z Cleaning up orphan processes
```

**Fixture:** `53135808a37013848d4e...` conf=0.38
**Evidence:** ['<timestamp>.2018812Z Mar 27 <timestamp> runnervmrg6be sudo[2317]: root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker']
```
2026-03-27T14:03:26.2000088Z Mar 27 14:03:19 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-27T14:03:26.2002518Z Mar 27 14:03:19 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-27T14:03:26.2003488Z Mar 27 14:03:19 runnervmrg6be systemd[1]: Started agent.service - Agent.
2026-03-27T14:03:26.2004590Z Mar 27 14:03:20 runnervmrg6be sudo[2225]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl stop systemd-resolved
2026-03-27T14:03:26.2005319Z Mar 27 14:03:20 runnervmrg6be sudo[2225]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2005915Z Mar 27 14:03:20 runnervmrg6be sudo[2225]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2006635Z Mar 27 14:03:20 runnervmrg6be sudo[2231]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart systemd-resolved
2026-03-27T14:03:26.2007337Z Mar 27 14:03:20 runnervmrg6be sudo[2231]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2008199Z Mar 27 14:03:20 runnervmrg6be sudo[2231]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2008930Z Mar 27 14:03:20 runnervmrg6be sudo[2237]:     root : *** ; USER=root ; COMMAND=/usr/bin/resolvectl flush-caches
2026-03-27T14:03:26.2009610Z Mar 27 14:03:20 runnervmrg6be sudo[2237]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2010223Z Mar 27 14:03:20 runnervmrg6be sudo[2237]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2010870Z Mar 27 14:03:20 runnervmrg6be sudo[2240]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl reload docker
2026-03-27T14:03:26.2012021Z Mar 27 14:03:20 runnervmrg6be sudo[2240]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2012894Z Mar 27 14:03:20 runnervmrg6be sudo[2240]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2013600Z Mar 27 14:03:20 runnervmrg6be sudo[2249]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl daemon-reload
2026-03-27T14:03:26.2014255Z Mar 27 14:03:20 runnervmrg6be sudo[2249]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2015474Z Mar 27 14:03:20 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-27T14:03:26.2017072Z Mar 27 14:03:20 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-27T14:03:26.2018136Z Mar 27 14:03:20 runnervmrg6be sudo[2249]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2018812Z Mar 27 14:03:20 runnervmrg6be sudo[2317]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker
2026-03-27T14:03:26.2019475Z Mar 27 14:03:20 runnervmrg6be sudo[2317]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-27T14:03:26.2020117Z Mar 27 14:03:23 runnervmrg6be sudo[2317]: pam_unix(sudo:session): session closed for user root
2026-03-27T14:03:26.2020454Z 
2026-03-27T14:03:26.2119114Z Cleaning up orphan processes
```

**Fixture:** `7e132712edb5e37eb54f...` conf=0.37
**Evidence:** ['<timestamp>.9205629Z Mar 28 <timestamp> runnervmrg6be sudo[2369]: root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker']
```
2026-03-28T08:35:32.9173580Z Mar 28 08:35:26 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-28T08:35:32.9176234Z Mar 28 08:35:26 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-28T08:35:32.9178061Z Mar 28 08:35:26 runnervmrg6be systemd[1]: Started agent.service - Agent.
2026-03-28T08:35:32.9179804Z Mar 28 08:35:26 runnervmrg6be sudo[2274]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl stop systemd-resolved
2026-03-28T08:35:32.9181134Z Mar 28 08:35:26 runnervmrg6be sudo[2274]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9182324Z Mar 28 08:35:26 runnervmrg6be sudo[2274]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9183630Z Mar 28 08:35:26 runnervmrg6be sudo[2280]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart systemd-resolved
2026-03-28T08:35:32.9184878Z Mar 28 08:35:26 runnervmrg6be sudo[2280]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9186011Z Mar 28 08:35:27 runnervmrg6be sudo[2280]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9187280Z Mar 28 08:35:27 runnervmrg6be sudo[2287]:     root : *** ; USER=root ; COMMAND=/usr/bin/resolvectl flush-caches
2026-03-28T08:35:32.9188530Z Mar 28 08:35:27 runnervmrg6be sudo[2287]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9190650Z Mar 28 08:35:27 runnervmrg6be sudo[2287]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9191904Z Mar 28 08:35:27 runnervmrg6be sudo[2290]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl reload docker
2026-03-28T08:35:32.9193120Z Mar 28 08:35:27 runnervmrg6be sudo[2290]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9194307Z Mar 28 08:35:27 runnervmrg6be sudo[2290]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9195621Z Mar 28 08:35:27 runnervmrg6be sudo[2296]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl daemon-reload
2026-03-28T08:35:32.9196828Z Mar 28 08:35:27 runnervmrg6be sudo[2296]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9199143Z Mar 28 08:35:27 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:9: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-28T08:35:32.9202032Z Mar 28 08:35:27 runnervmrg6be systemd[1]: /etc/systemd/system/agent.service:10: Standard output type syslog is obsolete, automatically updating to journal. Please update your unit file, and consider removing the setting altogether.
2026-03-28T08:35:32.9204310Z Mar 28 08:35:27 runnervmrg6be sudo[2296]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9205629Z Mar 28 08:35:27 runnervmrg6be sudo[2369]:     root : *** ; USER=root ; COMMAND=/usr/bin/systemctl restart docker
2026-03-28T08:35:32.9206897Z Mar 28 08:35:27 runnervmrg6be sudo[2369]: pam_unix(sudo:session): session opened for user root(uid=0) by (uid=0)
2026-03-28T08:35:32.9208044Z Mar 28 08:35:30 runnervmrg6be sudo[2369]: pam_unix(sudo:session): session closed for user root
2026-03-28T08:35:32.9208654Z 
2026-03-28T08:35:33.2715223Z Cleaning up orphan processes
```

### `ignored-exit-code` — 16 matches, 9 low-conf (56%)

**Fixture:** `ea2be2b8a97d1ed6fc8e...` conf=0.46
**Evidence:** ['<timestamp>.2617039Z sudo rm -rf /usr/local/lib/android || true', '<timestamp>.0241348Z set +e']
```
2026-04-27T12:29:56.7673111Z   CP_PR_CODE: 0
2026-04-27T12:29:56.7673288Z   CODE_BASE_CHECK: 1
2026-04-27T12:29:56.7673473Z ##[endgroup]
2026-04-27T12:29:56.7711564Z repo sync after cherry-pick manifest PR success.
2026-04-27T12:29:56.7712148Z PR code base need to rebase with source repo.
2026-04-27T12:29:56.7723160Z ##[error]Process completed with exit code 1.
2026-04-27T12:29:56.7793043Z ##[group]Run echo "status=${JOB_STATUS}" >> $GITHUB_OUTPUT
2026-04-27T12:29:56.7793457Z [36;1mecho "status=${JOB_STATUS}" >> $GITHUB_OUTPUT[0m
2026-04-27T12:29:56.7810730Z shell: /usr/bin/bash -e {0}
2026-04-27T12:29:56.7810959Z env:
2026-04-27T12:29:56.7811121Z   DOCKER_BUILDKIT: 1
2026-04-27T12:29:56.7811318Z   REPO_INIT: true
2026-04-27T12:29:56.7811526Z   REPO_ROOT: /home/runner/work/tests/tests
2026-04-27T12:29:56.7811783Z   REPO_EXIST: 0
2026-04-27T12:29:56.7811962Z   REPO_SYNC_CODE: 0
2026-04-27T12:29:56.7812144Z   FETCH_PR_CODE: 0
2026-04-27T12:29:56.7812322Z   CP_PR_CODE: 0
2026-04-27T12:29:56.7812486Z   CODE_BASE_CHECK: 1
2026-04-27T12:29:56.7812685Z   JOB_STATUS: failure
2026-04-27T12:29:56.7812881Z ##[endgroup]
2026-04-27T12:29:56.7912413Z Evaluate and set job outputs
2026-04-27T12:29:56.7918492Z Set output 'status'
2026-04-27T12:29:56.7920526Z Set output 'CURRENT_ACTION_URL'
2026-04-27T12:29:56.7921530Z Cleaning up orphan processes
2026-04-27T12:29:56.8751878Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/download-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `ecd7e1d85c15377ae45f...` conf=0.46
**Evidence:** ['<timestamp>.2161548Z sudo rm -rf /usr/local/lib/android || true', '<timestamp>.9129541Z set +e']
```
2026-04-27T12:29:00.1641935Z   CP_PR_CODE: 0
2026-04-27T12:29:00.1642112Z   CODE_BASE_CHECK: 1
2026-04-27T12:29:00.1642294Z ##[endgroup]
2026-04-27T12:29:00.1688796Z repo sync after cherry-pick manifest PR success.
2026-04-27T12:29:00.1689185Z PR code base need to rebase with source repo.
2026-04-27T12:29:00.1700338Z ##[error]Process completed with exit code 1.
2026-04-27T12:29:00.1770176Z ##[group]Run echo "status=${JOB_STATUS}" >> $GITHUB_OUTPUT
2026-04-27T12:29:00.1770561Z [36;1mecho "status=${JOB_STATUS}" >> $GITHUB_OUTPUT[0m
2026-04-27T12:29:00.1788021Z shell: /usr/bin/bash -e {0}
2026-04-27T12:29:00.1788245Z env:
2026-04-27T12:29:00.1788415Z   DOCKER_BUILDKIT: 1
2026-04-27T12:29:00.1788602Z   REPO_INIT: true
2026-04-27T12:29:00.1788807Z   REPO_ROOT: /home/runner/work/tests/tests
2026-04-27T12:29:00.1789056Z   REPO_EXIST: 0
2026-04-27T12:29:00.1789227Z   REPO_SYNC_CODE: 0
2026-04-27T12:29:00.1789398Z   FETCH_PR_CODE: 0
2026-04-27T12:29:00.1789567Z   CP_PR_CODE: 0
2026-04-27T12:29:00.1789744Z   CODE_BASE_CHECK: 1
2026-04-27T12:29:00.1789941Z   JOB_STATUS: failure
2026-04-27T12:29:00.1790129Z ##[endgroup]
2026-04-27T12:29:00.1889489Z Evaluate and set job outputs
2026-04-27T12:29:00.1895611Z Set output 'status'
2026-04-27T12:29:00.1897274Z Set output 'CURRENT_ACTION_URL'
2026-04-27T12:29:00.1898307Z Cleaning up orphan processes
2026-04-27T12:29:00.2198992Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/download-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `6d40376d697dcbb54e39...` conf=0.46
**Evidence:** ['<timestamp>.5726950Z sudo rm -rf /usr/local/lib/android || true', '<timestamp>.9971527Z set +e']
```
2026-04-21T09:37:30.8389816Z   FETCH_PR_CODE: 0
2026-04-21T09:37:30.8390086Z   CP_PR_CODE: 1
2026-04-21T09:37:30.8390252Z   CODE_BASE_CHECK: 0
2026-04-21T09:37:30.8390427Z ##[endgroup]
2026-04-21T09:37:30.8428009Z repo sync after cherry-pick manifest PR success.
2026-04-21T09:37:30.8458136Z ##[error]Process completed with exit code 1.
2026-04-21T09:37:30.8528866Z ##[group]Run echo "status=${JOB_STATUS}" >> $GITHUB_OUTPUT
2026-04-21T09:37:30.8529273Z [36;1mecho "status=${JOB_STATUS}" >> $GITHUB_OUTPUT[0m
2026-04-21T09:37:30.8547280Z shell: /usr/bin/bash -e {0}
2026-04-21T09:37:30.8547540Z env:
2026-04-21T09:37:30.8547708Z   DOCKER_BUILDKIT: 1
2026-04-21T09:37:30.8547906Z   REPO_INIT: true
2026-04-21T09:37:30.8548116Z   REPO_ROOT: /home/runner/work/tests/tests
2026-04-21T09:37:30.8548365Z   REPO_EXIST: 0
2026-04-21T09:37:30.8548537Z   REPO_SYNC_CODE: 0
2026-04-21T09:37:30.8548707Z   FETCH_PR_CODE: 0
2026-04-21T09:37:30.8548878Z   CP_PR_CODE: 1
2026-04-21T09:37:30.8549046Z   CODE_BASE_CHECK: 0
2026-04-21T09:37:30.8549224Z   JOB_STATUS: failure
2026-04-21T09:37:30.8549415Z ##[endgroup]
2026-04-21T09:37:30.8655743Z Evaluate and set job outputs
2026-04-21T09:37:30.8662407Z Set output 'status'
2026-04-21T09:37:30.8664013Z Set output 'CURRENT_ACTION_URL'
2026-04-21T09:37:30.8665646Z Cleaning up orphan processes
2026-04-21T09:37:30.9093322Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/download-artifact@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `cache-corruption` — 15 matches, 15 low-conf (100%)

**Fixture:** `e14da5395fc1a42271b9...` conf=0.35
**Evidence:** ['<timestamp>.5214586Z <timestamp> WARN c.f.s.i.IntegrityCheckService - [integrity-check] chain hash mismatch: fileId=1, dbHash=db_hash_value, chainHash=different_chain_hash']
```
2026-04-10T06:30:42.9831198Z warning - 2026-04-10 06:30:42,982 -- Some files were not found --- {"not_found_files": ["platform-backend/backend-common/target/site/jacoco/jacoco.xml\nplatform-backend/backend-service/target/site/jacoco/jacoco.xml\nplatform-backend/backend-web/target/site/jacoco/jacoco.xml\n"]}
2026-04-10T06:30:43.0419762Z info - 2026-04-10 06:30:43,041 -- Found 2 coverage files to report
2026-04-10T06:30:43.0421337Z info - 2026-04-10 06:30:43,041 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-common/target/site/jacoco/jacoco.xml
2026-04-10T06:30:43.0423518Z info - 2026-04-10 06:30:43,042 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-service/target/site/jacoco/jacoco.xml
2026-04-10T06:30:43.5386290Z info - 2026-04-10 06:30:43,538 -- Your upload is now queued for processing. When finished, results will be available at: https://app.codecov.io/github/soarcollab/recordplatform/commit/99e2567044701396ce2650ffa961bc6d35b27964
2026-04-10T06:30:43.5388170Z info - 2026-04-10 06:30:43,538 -- Sending upload (156500 bytes) to storage
2026-04-10T06:30:43.6423814Z info - 2026-04-10 06:30:43,642 -- Upload queued for processing complete
2026-04-10T06:30:43.7399849Z Post job cleanup.
2026-04-10T06:30:43.9119388Z Post job cleanup.
2026-04-10T06:30:44.0128653Z [command]/usr/bin/git version
2026-04-10T06:30:44.0170987Z git version 2.53.0
2026-04-10T06:30:44.0208661Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/4a0915f4-b3af-48fa-b5b5-e79f0ddc7bb0/.gitconfig'
2026-04-10T06:30:44.0218968Z Temporarily overriding HOME='/home/runner/work/_temp/4a0915f4-b3af-48fa-b5b5-e79f0ddc7bb0' before making global git config changes
2026-04-10T06:30:44.0220294Z Adding repository directory to the temporary git global config as a safe directory
2026-04-10T06:30:44.0224895Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/RecordPlatform/RecordPlatform
2026-04-10T06:30:44.0259248Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-10T06:30:44.0291458Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-10T06:30:44.0513610Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-10T06:30:44.0534126Z http.https://github.com/.extraheader
2026-04-10T06:30:44.0547475Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-10T06:30:44.0577403Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-10T06:30:44.0791392Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-10T06:30:44.0820801Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-10T06:30:44.1159338Z Cleaning up orphan processes
2026-04-10T06:30:44.1527970Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-java@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `eabb4c2e549b8a396132...` conf=0.35
**Evidence:** ['<timestamp>.5301032Z <timestamp> WARN c.f.s.i.IntegrityCheckService - [integrity-check] chain hash mismatch: fileId=1, dbHash=db_hash_value, chainHash=different_chain_hash']
```
2026-04-11T05:20:15.0204466Z warning - 2026-04-11 05:20:15,019 -- Some files were not found --- {"not_found_files": ["platform-backend/backend-common/target/site/jacoco/jacoco.xml\nplatform-backend/backend-service/target/site/jacoco/jacoco.xml\nplatform-backend/backend-web/target/site/jacoco/jacoco.xml\n"]}
2026-04-11T05:20:15.0874588Z info - 2026-04-11 05:20:15,087 -- Found 2 coverage files to report
2026-04-11T05:20:15.0876111Z info - 2026-04-11 05:20:15,087 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-common/target/site/jacoco/jacoco.xml
2026-04-11T05:20:15.0877993Z info - 2026-04-11 05:20:15,087 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-service/target/site/jacoco/jacoco.xml
2026-04-11T05:20:15.6701371Z info - 2026-04-11 05:20:15,669 -- Your upload is now queued for processing. When finished, results will be available at: https://app.codecov.io/github/soarcollab/recordplatform/commit/b6c14136dd5fdeba2c707b8a7a3a7484081494ce
2026-04-11T05:20:15.6703380Z info - 2026-04-11 05:20:15,669 -- Sending upload (158142 bytes) to storage
2026-04-11T05:20:16.0291019Z info - 2026-04-11 05:20:16,028 -- Upload queued for processing complete
2026-04-11T05:20:16.1285720Z Post job cleanup.
2026-04-11T05:20:16.3027694Z Post job cleanup.
2026-04-11T05:20:16.3976560Z [command]/usr/bin/git version
2026-04-11T05:20:16.4021189Z git version 2.53.0
2026-04-11T05:20:16.4061741Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/c7e7355d-855d-4434-9574-2a03f98a4fbd/.gitconfig'
2026-04-11T05:20:16.4071308Z Temporarily overriding HOME='/home/runner/work/_temp/c7e7355d-855d-4434-9574-2a03f98a4fbd' before making global git config changes
2026-04-11T05:20:16.4072681Z Adding repository directory to the temporary git global config as a safe directory
2026-04-11T05:20:16.4077504Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/RecordPlatform/RecordPlatform
2026-04-11T05:20:16.4111301Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-11T05:20:16.4142212Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-11T05:20:16.4363989Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-11T05:20:16.4384230Z http.https://github.com/.extraheader
2026-04-11T05:20:16.4396778Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-11T05:20:16.4426020Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-11T05:20:16.4640069Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-11T05:20:16.4677912Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-11T05:20:16.5011504Z Cleaning up orphan processes
2026-04-11T05:20:16.5382955Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-java@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `ec63b0e4ced86afbcded...` conf=0.35
**Evidence:** ['<timestamp>.4940812Z <timestamp> WARN c.f.s.i.IntegrityCheckService - [integrity-check] chain hash mismatch: fileId=1, dbHash=db_hash_value, chainHash=different_chain_hash']
```
2026-04-11T09:36:12.6221644Z warning - 2026-04-11 09:36:12,621 -- Some files were not found --- {"not_found_files": ["platform-backend/backend-common/target/site/jacoco/jacoco.xml\nplatform-backend/backend-service/target/site/jacoco/jacoco.xml\nplatform-backend/backend-web/target/site/jacoco/jacoco.xml\n"]}
2026-04-11T09:36:12.6819830Z info - 2026-04-11 09:36:12,681 -- Found 2 coverage files to report
2026-04-11T09:36:12.6821513Z info - 2026-04-11 09:36:12,681 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-service/target/site/jacoco/jacoco.xml
2026-04-11T09:36:12.6823328Z info - 2026-04-11 09:36:12,682 -- > /home/runner/work/RecordPlatform/RecordPlatform/platform-backend/backend-common/target/site/jacoco/jacoco.xml
2026-04-11T09:36:13.0686505Z info - 2026-04-11 09:36:13,068 -- Your upload is now queued for processing. When finished, results will be available at: https://app.codecov.io/github/soarcollab/recordplatform/commit/8b84df7892ce89c41fa1e175bfe3be39e07b8157
2026-04-11T09:36:13.0688270Z info - 2026-04-11 09:36:13,068 -- Sending upload (158278 bytes) to storage
2026-04-11T09:36:13.1852325Z info - 2026-04-11 09:36:13,185 -- Upload queued for processing complete
2026-04-11T09:36:13.3266698Z Post job cleanup.
2026-04-11T09:36:13.5022068Z Post job cleanup.
2026-04-11T09:36:13.5953661Z [command]/usr/bin/git version
2026-04-11T09:36:13.5988884Z git version 2.53.0
2026-04-11T09:36:13.6024524Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/95e8568a-ea79-4f81-88d4-17e2cddf1554/.gitconfig'
2026-04-11T09:36:13.6034655Z Temporarily overriding HOME='/home/runner/work/_temp/95e8568a-ea79-4f81-88d4-17e2cddf1554' before making global git config changes
2026-04-11T09:36:13.6035884Z Adding repository directory to the temporary git global config as a safe directory
2026-04-11T09:36:13.6039508Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/RecordPlatform/RecordPlatform
2026-04-11T09:36:13.6071874Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-11T09:36:13.6104104Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-11T09:36:13.6321306Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-11T09:36:13.6340901Z http.https://github.com/.extraheader
2026-04-11T09:36:13.6353087Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-11T09:36:13.6382214Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-11T09:36:13.6596517Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-11T09:36:13.6633690Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-11T09:36:13.6952635Z Cleaning up orphan processes
2026-04-11T09:36:13.7317096Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4, actions/setup-java@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `package-manager-mismatch` — 14 matches, 8 low-conf (57%)

**Fixture:** `1f662ec62c1392c6fd4c...` conf=0.49
**Evidence:** ['<timestamp>.5001070Z updater | <timestamp> ERROR <job_1249911966> Error during file fetching; aborting: /package-lock.json not parseable']
```
2026-02-19T21:48:31.1037791Z updater | 2026/02/19 21:48:31 INFO <job_1249911966> Finished job processing
2026-02-19T21:48:31.1056459Z updater | 2026/02/19 21:48:31 INFO Results:
2026-02-19T21:48:31.1057371Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-02-19T21:48:31.1058415Z +----------------------------------------------------------------------------------+
2026-02-19T21:48:31.1059280Z |                                      Errors                                      |
2026-02-19T21:48:31.1059793Z +-------------------------------+--------------------------------------------------+
2026-02-19T21:48:31.1060587Z | Type                          | Details                                          |
2026-02-19T21:48:31.1061105Z +-------------------------------+--------------------------------------------------+
2026-02-19T21:48:31.1061632Z | dependency_file_not_parseable | {                                                |
2026-02-19T21:48:31.1062212Z |                               |   "message": "/package-lock.json not parseable", |
2026-02-19T21:48:31.1062741Z |                               |   "file-path": "/package-lock.json"              |
2026-02-19T21:48:31.1063177Z |                               | }                                                |
2026-02-19T21:48:31.1063985Z +-------------------------------+--------------------------------------------------+
2026-02-19T21:48:31.2324217Z Failure running container e139783646656a1a6ddbcd323fbc9be1aac9c8f4f531d02ccda5ac9ff876455b: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-02-19T21:48:31.3457241Z Cleaned up container e139783646656a1a6ddbcd323fbc9be1aac9c8f4f531d02ccda5ac9ff876455b
2026-02-19T21:48:31.3564252Z   proxy | 2026/02/19 21:48:31 0/5 calls cached (0%)
2026-02-19T21:48:31.3567877Z 2026/02/19 21:48:31 Posting metrics to remote API endpoint
2026-02-19T21:48:32.1296461Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/Karaka-Management/jsOMS/network/updates/1249911966 (write access to the repository is required to view the log)
2026-02-19T21:48:32.1308261Z 🤖 ~ finished: error reported to Dependabot ~
2026-02-19T21:48:32.1401922Z Post job cleanup.
2026-02-19T21:48:32.3049083Z Cleaning up orphan processes
```

**Fixture:** `840176d2de63edc4c05b...` conf=0.49
**Evidence:** ['<timestamp>.9498366Z updater | <timestamp> ERROR <job_1289148685> Error during file fetching; aborting: /package-lock.json not parseable']
```
2026-03-23T11:15:02.3787604Z updater | 2026/03/23 11:15:02 INFO Results:
2026-03-23T11:15:02.3788468Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-23T11:15:02.3789126Z +----------------------------------------------------------------------------------+
2026-03-23T11:15:02.3789579Z |                                      Errors                                      |
2026-03-23T11:15:02.3790075Z +-------------------------------+--------------------------------------------------+
2026-03-23T11:15:02.3790522Z | Type                          | Details                                          |
2026-03-23T11:15:02.3790962Z +-------------------------------+--------------------------------------------------+
2026-03-23T11:15:02.3791458Z | dependency_file_not_parseable | {                                                |
2026-03-23T11:15:02.3791952Z |                               |   "message": "/package-lock.json not parseable", |
2026-03-23T11:15:02.3792577Z |                               |   "file-path": "/package-lock.json"              |
2026-03-23T11:15:02.3793027Z |                               | }                                                |
2026-03-23T11:15:02.3793705Z +-------------------------------+--------------------------------------------------+
2026-03-23T11:15:02.4953123Z Failure running container eac534ec9552048473cbbe0768e0dd36eb6add4b697bfa37a4e2549940354c63: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-23T11:15:02.5688541Z Cleaned up container eac534ec9552048473cbbe0768e0dd36eb6add4b697bfa37a4e2549940354c63
2026-03-23T11:15:02.5794147Z   proxy | 2026/03/23 11:15:02 0/6 calls cached (0%)
2026-03-23T11:15:02.5794794Z 2026/03/23 11:15:02 Posting metrics to remote API endpoint
2026-03-23T11:15:03.0386086Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/Karaka-Management/jsOMS/network/updates/1289148685 (write access to the repository is required to view the log)
2026-03-23T11:15:03.0396566Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-23T11:15:03.0462705Z Post job cleanup.
2026-03-23T11:15:03.2004075Z Cleaning up orphan processes
2026-03-23T11:15:03.2303948Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `a63761eba3ff0046fa21...` conf=0.49
**Evidence:** ['<timestamp>.0813717Z updater | <timestamp> ERROR <job_1249802600> Error during file fetching; aborting: /package-lock.json not parseable']
```
2026-02-19T20:37:02.6937453Z updater | 2026/02/19 20:37:02 INFO <job_1249802600> Finished job processing
2026-02-19T20:37:02.6991222Z updater | 2026/02/19 20:37:02 INFO Results:
2026-02-19T20:37:02.6992176Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-02-19T20:37:02.6993388Z +----------------------------------------------------------------------------------+
2026-02-19T20:37:02.6994298Z |                                      Errors                                      |
2026-02-19T20:37:02.6995248Z +-------------------------------+--------------------------------------------------+
2026-02-19T20:37:02.6996004Z | Type                          | Details                                          |
2026-02-19T20:37:02.6996865Z +-------------------------------+--------------------------------------------------+
2026-02-19T20:37:02.6997657Z | dependency_file_not_parseable | {                                                |
2026-02-19T20:37:02.6998511Z |                               |   "message": "/package-lock.json not parseable", |
2026-02-19T20:37:02.6999342Z |                               |   "file-path": "/package-lock.json"              |
2026-02-19T20:37:02.7000042Z |                               | }                                                |
2026-02-19T20:37:02.7001179Z +-------------------------------+--------------------------------------------------+
2026-02-19T20:37:02.8197087Z Failure running container 3b5871d7354530e00ad46d95216530fe22144e40cb86b38560d6447d6f84ecfb: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-02-19T20:37:02.9227552Z Cleaned up container 3b5871d7354530e00ad46d95216530fe22144e40cb86b38560d6447d6f84ecfb
2026-02-19T20:37:02.9307069Z   proxy | 2026/02/19 20:37:02 0/5 calls cached (0%)
2026-02-19T20:37:02.9307920Z 2026/02/19 20:37:02 Posting metrics to remote API endpoint
2026-02-19T20:37:03.3631111Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/Karaka-Management/jsOMS/network/updates/1249802600 (write access to the repository is required to view the log)
2026-02-19T20:37:03.3639922Z 🤖 ~ finished: error reported to Dependabot ~
2026-02-19T20:37:03.3737653Z Post job cleanup.
2026-02-19T20:37:03.5371065Z Cleaning up orphan processes
```

### `dependency-drift` — 14 matches, 14 low-conf (100%)

**Fixture:** `cae18e533ba9543a0dba...` conf=0.35
**Evidence:** ['<timestamp>.8420720Z * any::sessioninfo: dependency conflict']
```
2026-02-16T20:50:58.8423548Z 2. get("lockfile_create_internal", asNamespace("pak"))(...)
2026-02-16T20:50:58.8424104Z 3. prop$stop_for_solution_error()
2026-02-16T20:50:58.8424411Z 4. private$plan$stop_for_solve_error()
2026-02-16T20:50:58.8424796Z 5. pkgdepends:::pkgplan_stop_for_solve_error(self, private)
2026-02-16T20:50:58.8425467Z 6. base::throw(new_error("Could not solve package dependencies:\n", msg, …
2026-02-16T20:50:58.8425875Z 7. | base::signalCondition(cond)
2026-02-16T20:50:58.8426318Z 8. global (function (e) …
2026-02-16T20:50:58.8426542Z Execution halted
2026-02-16T20:50:58.8813131Z ##[error]Process completed with exit code 1.
2026-02-16T20:50:58.9080865Z Post job cleanup.
2026-02-16T20:50:58.9133677Z Post job cleanup.
2026-02-16T20:50:59.0099076Z [command]/usr/bin/git version
2026-02-16T20:50:59.0137187Z git version 2.52.0
2026-02-16T20:50:59.0180976Z Temporarily overriding HOME='/home/runner/work/_temp/2ee56e2b-fba4-43d7-bf23-98f8df9f269a' before making global git config changes
2026-02-16T20:50:59.0182217Z Adding repository directory to the temporary git global config as a safe directory
2026-02-16T20:50:59.0194477Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/spammR/spammR
2026-02-16T20:50:59.0230176Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-16T20:50:59.0263827Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-16T20:50:59.0522158Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-16T20:50:59.0549638Z http.https://github.com/.extraheader
2026-02-16T20:50:59.0564355Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-16T20:50:59.0602153Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-16T20:50:59.0869762Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-16T20:50:59.0906909Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-16T20:50:59.1360316Z Cleaning up orphan processes
```

**Fixture:** `183ab24d04ee7e938ca7...` conf=0.35
**Evidence:** ['<timestamp>.3192565Z * any::sessioninfo: dependency conflict']
```
2026-02-05T16:14:25.3197463Z 2. get("lockfile_create_internal", asNamespace("pak"))(...)
2026-02-05T16:14:25.3198053Z 3. prop$stop_for_solution_error()
2026-02-05T16:14:25.3198526Z 4. private$plan$stop_for_solve_error()
2026-02-05T16:14:25.3199193Z 5. pkgdepends:::pkgplan_stop_for_solve_error(self, private)
2026-02-05T16:14:25.3200363Z 6. base::throw(new_error("Could not solve package dependencies:\n", msg, …
2026-02-05T16:14:25.3201043Z 7. | base::signalCondition(cond)
2026-02-05T16:14:25.3201637Z 8. global (function (e) …
2026-02-05T16:14:25.3201993Z Execution halted
2026-02-05T16:14:25.3640015Z ##[error]Process completed with exit code 1.
2026-02-05T16:14:25.3907955Z Post job cleanup.
2026-02-05T16:14:25.3962930Z Post job cleanup.
2026-02-05T16:14:25.4927826Z [command]/usr/bin/git version
2026-02-05T16:14:25.4964654Z git version 2.52.0
2026-02-05T16:14:25.5011365Z Temporarily overriding HOME='/home/runner/work/_temp/fd6ec976-bfbb-47bb-b932-52d4f0a717ca' before making global git config changes
2026-02-05T16:14:25.5012893Z Adding repository directory to the temporary git global config as a safe directory
2026-02-05T16:14:25.5018142Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/spammR/spammR
2026-02-05T16:14:25.5064910Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-05T16:14:25.5099743Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-05T16:14:25.5347926Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-05T16:14:25.5375137Z http.https://github.com/.extraheader
2026-02-05T16:14:25.5389043Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-05T16:14:25.5424737Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-05T16:14:25.5677508Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-05T16:14:25.5714908Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-05T16:14:25.6090005Z Cleaning up orphan processes
```

**Fixture:** `d8877167d9df669d86cf...` conf=0.35
**Evidence:** ["<timestamp>.5518797Z ERROR: pip's dependency resolver does not currently take into account all the packages that are installed. This behaviour is the source of the following dependency conflicts."]
```
2026-03-08T09:11:14.9837173Z   File "/opt/hostedtoolcache/Python/3.10.19/x64/lib/python3.10/importlib/__init__.py", line 126, in import_module
2026-03-08T09:11:14.9837717Z     return _bootstrap._gcd_import(name[level:], package, level)
2026-03-08T09:11:14.9838137Z   File "<frozen importlib._bootstrap>", line 1050, in _gcd_import
2026-03-08T09:11:14.9838558Z   File "<frozen importlib._bootstrap>", line 1027, in _find_and_load
2026-03-08T09:11:14.9839246Z   File "<frozen importlib._bootstrap>", line 1006, in _find_and_load_unlocked
2026-03-08T09:11:14.9839983Z   File "<frozen importlib._bootstrap>", line 688, in _load_unlocked
2026-03-08T09:11:14.9840439Z   File "<frozen importlib._bootstrap_external>", line 883, in exec_module
2026-03-08T09:11:14.9841127Z   File "<frozen importlib._bootstrap>", line 241, in _call_with_frames_removed
2026-03-08T09:11:14.9841773Z   File "/opt/hostedtoolcache/Python/3.10.19/x64/lib/python3.10/site-packages/tox_conda/plugin.py", line 13, in <module>
2026-03-08T09:11:14.9842348Z     from tox.config import DepConfig, DepOption, TestenvConfig
2026-03-08T09:11:14.9843036Z ImportError: cannot import name 'DepConfig' from 'tox.config' (/opt/hostedtoolcache/Python/3.10.19/x64/lib/python3.10/site-packages/tox/config/__init__.py)
2026-03-08T09:11:15.0107107Z ##[error]Process completed with exit code 1.
2026-03-08T09:11:15.0211874Z Post job cleanup.
2026-03-08T09:11:15.1102588Z [command]/usr/bin/git version
2026-03-08T09:11:15.1142185Z git version 2.53.0
2026-03-08T09:11:15.1187947Z Temporarily overriding HOME='/home/runner/work/_temp/b5b6fa50-6821-43ad-9260-47efdf57da26' before making global git config changes
2026-03-08T09:11:15.1189294Z Adding repository directory to the temporary git global config as a safe directory
2026-03-08T09:11:15.1193168Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/eFISHent/eFISHent
2026-03-08T09:11:15.1228100Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-08T09:11:15.1260656Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-08T09:11:15.1502054Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-08T09:11:15.1525317Z http.https://github.com/.extraheader
2026-03-08T09:11:15.1535624Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-08T09:11:15.1567090Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-08T09:11:15.1930482Z Cleaning up orphan processes
```

### `invalid-config-schema` — 12 matches, 12 low-conf (100%)

**Fixture:** `25c9cde955a23689d750...` conf=0.33
**Evidence:** ["<timestamp>.9619013Z Standard output: Invalid configuration for provider 'Hang dump' (UID: HangDumpCommandLineProvider). Error: You specified one or more hang dump parameters but did not enable it, add --hangdump to the command line"]
```
2026-03-25T13:05:05.6825514Z ##[endgroup]
2026-03-25T13:05:05.9523416Z Post job cleanup.
2026-03-25T13:05:06.0192725Z [command]/usr/bin/git version
2026-03-25T13:05:06.0229449Z git version 2.53.0
2026-03-25T13:05:06.0259521Z Copying '/home/runner/.gitconfig' to '/home/runner/work/_temp/4547d720-d205-4cc9-baae-d68697cb620a/.gitconfig'
2026-03-25T13:05:06.0271628Z Temporarily overriding HOME='/home/runner/work/_temp/4547d720-d205-4cc9-baae-d68697cb620a' before making global git config changes
2026-03-25T13:05:06.0272515Z Adding repository directory to the temporary git global config as a safe directory
2026-03-25T13:05:06.0274669Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/alexa-london-travel/alexa-london-travel
2026-03-25T13:05:06.0303882Z Removing SSH command configuration
2026-03-25T13:05:06.0310781Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-25T13:05:06.0342970Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-25T13:05:06.0546569Z Removing HTTP extra header
2026-03-25T13:05:06.0551774Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-25T13:05:06.0583549Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-25T13:05:06.0783653Z Removing includeIf entries pointing to credentials config files
2026-03-25T13:05:06.0791056Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-25T13:05:06.0823169Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-25T13:05:06.1143239Z Post job cleanup.
2026-03-25T13:05:06.1718556Z [harden-runner] post-step
2026-03-25T13:05:06.4821532Z Evaluate and set job outputs
2026-03-25T13:05:06.4826942Z Set output 'archive-name'
2026-03-25T13:05:06.4828507Z Set output 'artifact-name'
2026-03-25T13:05:06.4829194Z Set output 'artifact-run-id'
2026-03-25T13:05:06.4829877Z Cleaning up orphan processes
2026-03-25T13:05:06.5254349Z Terminate orphan process: pid (2915) (VBCSCompiler)
```

**Fixture:** `20805455a4a0afbe78b0...` conf=0.49
**Evidence:** ['<timestamp>.0913610Z ✓ data-rx-disable-in-flight with invalid value warns and disables 11ms', '<timestamp>.0978050Z stderr | test/razorx.test.ts > RazorX Framework API Surface Tests > File Upload Feature > File Selection & Size Validation > file input value cleared on size validation error']
```
2026-02-02T08:30:21.1803460Z Beginning upload of artifact content to blob storage
2026-02-02T08:30:21.6706730Z Uploaded bytes 192976
2026-02-02T08:30:21.7473540Z Finished uploading artifact content to blob storage!
2026-02-02T08:30:21.7487420Z SHA256 digest of uploaded artifact zip is d3967633bceffabd260928c04652957b45626124126be9fbfedec5febf66c954
2026-02-02T08:30:21.7489710Z Finalizing artifact upload
2026-02-02T08:30:21.8985280Z Artifact test-results-macos-latest.zip successfully finalized. Artifact ID 5339993734
2026-02-02T08:30:21.8986830Z Artifact test-results-macos-latest has been successfully uploaded! Final size is 192976 bytes. Artifact ID is 5339993734
2026-02-02T08:30:21.8987900Z Artifact download URL: https://github.com/ranzlee/razorx-framework/actions/runs/21582745662/artifacts/5339993734
2026-02-02T08:30:21.9136590Z Post job cleanup.
2026-02-02T08:30:22.3604490Z [command]/opt/homebrew/bin/git version
2026-02-02T08:30:22.5293580Z git version 2.52.0
2026-02-02T08:30:22.5435020Z Copying '/Users/runner/.gitconfig' to '/Users/runner/work/_temp/d99ab1b8-68fe-4f2b-bcb5-e3f3f69a4ca9/.gitconfig'
2026-02-02T08:30:22.5517900Z Temporarily overriding HOME='/Users/runner/work/_temp/d99ab1b8-68fe-4f2b-bcb5-e3f3f69a4ca9' before making global git config changes
2026-02-02T08:30:22.5524520Z Adding repository directory to the temporary git global config as a safe directory
2026-02-02T08:30:22.5525640Z [command]/opt/homebrew/bin/git config --global --add safe.directory /Users/runner/work/razorx-framework/razorx-framework
2026-02-02T08:30:22.5628120Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-02T08:30:22.5852000Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-02T08:30:22.7832880Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-02T08:30:22.7891830Z http.https://github.com/.extraheader
2026-02-02T08:30:22.7898380Z [command]/opt/homebrew/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-02T08:30:22.7969880Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-02T08:30:22.9222610Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-02T08:30:22.9525910Z [command]/opt/homebrew/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-02T08:30:23.1355450Z Cleaning up orphan processes
2026-02-02T08:30:23.3477240Z Terminate orphan process: pid (2844) (VBCSCompiler)
```

**Fixture:** `b7c61efeb74f764af66f...` conf=0.41
**Evidence:** ['<timestamp>.3140608Z ERROR Invalid configuration file', '<timestamp>.3142924Z unknown field `multilingual`, expected one of `title`, `authors`, `description`, `src`, `language`, `text-direction`']
```
2026-02-24T01:33:51.3068214Z   GITHUB_PAGES: true
2026-02-24T01:33:51.3068393Z ##[endgroup]
2026-02-24T01:33:51.3140608Z ERROR Invalid configuration file
2026-02-24T01:33:51.3141130Z 	Caused by: TOML parse error at line 4, column 1
2026-02-24T01:33:51.3141591Z   |
2026-02-24T01:33:51.3141851Z 4 | multilingual = false
2026-02-24T01:33:51.3142209Z   | ^^^^^^^^^^^^
2026-02-24T01:33:51.3142924Z unknown field `multilingual`, expected one of `title`, `authors`, `description`, `src`, `language`, `text-direction`
2026-02-24T01:33:51.3143645Z 
2026-02-24T01:33:51.3155713Z ##[error]Process completed with exit code 101.
2026-02-24T01:33:51.3254392Z Post job cleanup.
2026-02-24T01:33:51.4199256Z [command]/usr/bin/git version
2026-02-24T01:33:51.4244175Z git version 2.52.0
2026-02-24T01:33:51.4287037Z Temporarily overriding HOME='/home/runner/work/_temp/ba579e7e-1153-4e55-910e-11aa2fc49af5' before making global git config changes
2026-02-24T01:33:51.4287986Z Adding repository directory to the temporary git global config as a safe directory
2026-02-24T01:33:51.4292723Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/mdbook-pagecrypt/mdbook-pagecrypt
2026-02-24T01:33:51.4327455Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-24T01:33:51.4359890Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-24T01:33:51.4589883Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-24T01:33:51.4610930Z http.https://github.com/.extraheader
2026-02-24T01:33:51.4624070Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-24T01:33:51.4655456Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-24T01:33:51.4879634Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-24T01:33:51.4911924Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-24T01:33:51.5287665Z Cleaning up orphan processes
```

### `clock-drift` — 11 matches, 11 low-conf (100%)

**Fixture:** `ad3c578067cc558d11c9...` conf=0.38
**Evidence:** ['<timestamp>.3934975Z If you get an error: Clock skew detected, on Windows, then try setting the clock exactly correct on your computer using ntp or some clock software.', '<timestamp>.3291979Z Checking for function "clock_gettime" : YES']
```
2026-02-25T15:16:03.1759581Z 
2026-02-25T15:16:03.1759780Z -- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2026-02-25T15:16:03.1760189Z =========================== short test summary info ============================
2026-02-25T15:16:03.1761113Z FAILED tests/test_integration.py::TestLLMIntegration::test_normalize_actor_id_for_unit_and_player_actions - AttributeError: 'LLMWSHandler' object has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1762682Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_invalid_target_raises_error - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1764282Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_string_target - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1765829Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1767373Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_name_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1769058Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_value_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:16:03.1770035Z ============= 6 failed, 741 passed, 5 skipped, 1 warning in 14.31s =============
2026-02-25T15:16:03.3471689Z ##[error]Process completed with exit code 1.
2026-02-25T15:16:03.3555226Z Post job cleanup.
2026-02-25T15:16:03.5343849Z Post job cleanup.
2026-02-25T15:16:03.6092457Z [command]/usr/bin/git version
2026-02-25T15:16:03.6132405Z git version 2.52.0
2026-02-25T15:16:03.6175964Z Temporarily overriding HOME='/home/runner/work/_temp/24dfed10-e37b-4581-9fb1-033fc7e94633' before making global git config changes
2026-02-25T15:16:03.6177419Z Adding repository directory to the temporary git global config as a safe directory
2026-02-25T15:16:03.6180975Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/freeciv-llm/freeciv-llm
2026-02-25T15:16:03.6214442Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-25T15:16:03.6242813Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-25T15:16:03.6481548Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-25T15:16:03.6501829Z http.https://github.com/.extraheader
2026-02-25T15:16:03.6513132Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-25T15:16:03.6541638Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-25T15:16:03.6877930Z Cleaning up orphan processes
```

**Fixture:** `df26201b9f84d03c61e4...` conf=0.38
**Evidence:** ['<timestamp>.1638443Z If you get an error: Clock skew detected, on Windows, then try setting the clock exactly correct on your computer using ntp or some clock software.', '<timestamp>.3741607Z Checking for function "clock_gettime" : YES']
```
2026-02-25T15:15:33.9457443Z 
2026-02-25T15:15:33.9457652Z -- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2026-02-25T15:15:33.9458072Z =========================== short test summary info ============================
2026-02-25T15:15:33.9459035Z FAILED tests/test_integration.py::TestLLMIntegration::test_normalize_actor_id_for_unit_and_player_actions - AttributeError: 'LLMWSHandler' object has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9461029Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_invalid_target_raises_error - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9462734Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_string_target - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9464346Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9466130Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_name_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9467719Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_value_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:33.9468724Z ============= 6 failed, 741 passed, 5 skipped, 1 warning in 14.29s =============
2026-02-25T15:15:34.0833778Z ##[error]Process completed with exit code 1.
2026-02-25T15:15:34.0917614Z Post job cleanup.
2026-02-25T15:15:34.2673344Z Post job cleanup.
2026-02-25T15:15:34.3398983Z [command]/usr/bin/git version
2026-02-25T15:15:34.3433559Z git version 2.52.0
2026-02-25T15:15:34.3477941Z Temporarily overriding HOME='/home/runner/work/_temp/fde4e4cf-6666-46ec-a7db-5baf1a65f996' before making global git config changes
2026-02-25T15:15:34.3479355Z Adding repository directory to the temporary git global config as a safe directory
2026-02-25T15:15:34.3483318Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/freeciv-llm/freeciv-llm
2026-02-25T15:15:34.3515485Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-25T15:15:34.3544034Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-25T15:15:34.3779815Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-25T15:15:34.3799956Z http.https://github.com/.extraheader
2026-02-25T15:15:34.3811584Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-25T15:15:34.3840868Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-25T15:15:34.4194118Z Cleaning up orphan processes
```

**Fixture:** `ade2419d2e47a8afe475...` conf=0.38
**Evidence:** ['<timestamp>.1847230Z If you get an error: Clock skew detected, on Windows, then try setting the clock exactly correct on your computer using ntp or some clock software.', '<timestamp>.0188165Z Checking for function "clock_gettime" : YES']
```
2026-02-25T15:15:24.5567915Z 
2026-02-25T15:15:24.5568121Z -- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2026-02-25T15:15:24.5568544Z =========================== short test summary info ============================
2026-02-25T15:15:24.5569494Z FAILED tests/test_integration.py::TestLLMIntegration::test_normalize_actor_id_for_unit_and_player_actions - AttributeError: 'LLMWSHandler' object has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5571119Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_invalid_target_raises_error - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5572782Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_string_target - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5574385Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5576089Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_tech_name_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5577696Z FAILED tests/test_llm_handler.py::TestTechActionFormatNormalization::test_normalize_tech_research_with_value_key - AttributeError: type object 'LLMWSHandler' has no attribute '_normalize_llm_action'. Did you mean: '_normalize_agent_action'?
2026-02-25T15:15:24.5578707Z ============= 6 failed, 741 passed, 5 skipped, 1 warning in 14.05s =============
2026-02-25T15:15:24.6861910Z ##[error]Process completed with exit code 1.
2026-02-25T15:15:24.6955973Z Post job cleanup.
2026-02-25T15:15:24.8668189Z Post job cleanup.
2026-02-25T15:15:24.9385022Z [command]/usr/bin/git version
2026-02-25T15:15:24.9418583Z git version 2.52.0
2026-02-25T15:15:24.9460796Z Temporarily overriding HOME='/home/runner/work/_temp/c856f35b-cc1e-4c18-bbbb-0e85c8288a22' before making global git config changes
2026-02-25T15:15:24.9462248Z Adding repository directory to the temporary git global config as a safe directory
2026-02-25T15:15:24.9466191Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/freeciv-llm/freeciv-llm
2026-02-25T15:15:24.9497184Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-25T15:15:24.9525603Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-25T15:15:24.9761818Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-25T15:15:24.9781727Z http.https://github.com/.extraheader
2026-02-25T15:15:24.9792703Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-25T15:15:24.9820350Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-25T15:15:25.0187867Z Cleaning up orphan processes
```

### `proxy-configuration` — 10 matches, 9 low-conf (90%)

**Fixture:** `3bf9a77149474d73f023...` conf=0.47
**Evidence:** ['<timestamp>.4936515Z DEBUG craft_providers.lxd.lxc:lxc.py:525 Executing in container: lxc --project testcraft exec local:testcraft-full-project-foo-71 -- env CRAFT_MANAGED_MODE=1 DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true DEBIAN_PRIORITY=critical GOPROXY=direct http_proxy=***10.157.194.1:13444/ https_proxy=***10.157.194.1:13444/ REQUESTS_CA_BUNDLE=/usr/local/share/ca-certificates/local-ca.crt CARGO_HTTP_CAINFO=/usr/local/share/ca-certificates/local-ca.crt apt install -y hello']
```
2026-03-17T17:56:54.8682023Z tests/integration/services/test_provider.py::test_provider_lifecycle[lxd-ubuntu@24.04]
2026-03-17T17:56:54.8682837Z tests/integration/services/test_provider.py::test_run_managed[True]
2026-03-17T17:56:54.8684065Z   /home/runner/work/craft-application/craft-application/.venv/lib/python3.12/site-packages/craft_providers/bases/ubuntu.py:363: DeprecationWarning: path is deprecated. Use files() instead. Refer to https://importlib-resources.readthedocs.io/en/latest/using.html#migrating-from-legacy for migration advice.
2026-03-17T17:56:54.8685203Z     with importlib.resources.path(
2026-03-17T17:56:54.8685352Z 
2026-03-17T17:56:54.8685520Z -- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2026-03-17T17:56:54.8685945Z = 3 failed, 50 passed, 3 skipped, 4870 deselected, 3 warnings in 668.72s (0:11:08) =
2026-03-17T17:56:55.2205542Z make: *** [common.mk:223: test-coverage] Error 1
2026-03-17T17:56:55.2220436Z ##[error]Process completed with exit code 2.
2026-03-17T17:56:55.2323355Z Post job cleanup.
2026-03-17T17:56:55.2370155Z Post job cleanup.
2026-03-17T17:56:55.3042613Z [command]/usr/bin/git version
2026-03-17T17:56:55.3078000Z git version 2.52.0
2026-03-17T17:56:55.3110516Z Temporarily overriding HOME='/home/runner/work/_temp/80b62aaa-8570-49a8-b670-8ed754acd347' before making global git config changes
2026-03-17T17:56:55.3111216Z Adding repository directory to the temporary git global config as a safe directory
2026-03-17T17:56:55.3115846Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/craft-application/craft-application
2026-03-17T17:56:55.3150336Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-17T17:56:55.3180997Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-17T17:56:55.3383091Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-17T17:56:55.3404950Z http.https://github.com/.extraheader
2026-03-17T17:56:55.3415467Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-17T17:56:55.3444877Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-17T17:56:55.3638904Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-17T17:56:55.3666815Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-17T17:56:55.3984226Z Cleaning up orphan processes
```

**Fixture:** `92285e8bf27166890f7b...` conf=0.47
**Evidence:** ['<timestamp>.9835313Z DEBUG craft_providers.lxd.lxc:lxc.py:525 Executing in container: lxc --project testcraft exec local:testcraft-full-project-foo-83 -- env CRAFT_MANAGED_MODE=1 DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true DEBIAN_PRIORITY=critical GOPROXY=direct http_proxy=***10.172.215.1:13444/ https_proxy=***10.172.215.1:13444/ REQUESTS_CA_BUNDLE=/usr/local/share/ca-certificates/local-ca.crt CARGO_HTTP_CAINFO=/usr/local/share/ca-certificates/local-ca.crt apt install -y hello']
```
2026-03-17T16:58:23.5230735Z DEBUG    craft_providers.lxd.lxc:lxc.py:192 Executing on host: lxc --project testcraft config device remove local:testcraft-full-project-foo-83 disk-/tmp/craft-state
2026-03-17T16:58:23.5232130Z DEBUG    craft_providers.lxd.lxc:lxc.py:192 Executing on host: lxc --project testcraft stop local:testcraft-full-project-foo-83
2026-03-17T16:58:23.5233600Z DEBUG    craft_providers.lxd.lxc:lxc.py:192 Executing on host: lxc --project testcraft list local: --format=yaml
2026-03-17T16:58:23.5235018Z DEBUG    craft_providers.lxd.lxc:lxc.py:192 Executing on host: lxc --project testcraft config set local:testcraft-full-project-foo-83 user.craft_providers.status FINISHED
2026-03-17T16:58:23.5236884Z DEBUG    craft_providers.lxd.lxc:lxc.py:192 Executing on host: lxc --project testcraft delete local:testcraft-full-project-foo-83 --force
2026-03-17T16:58:23.5237977Z ===== 1 failed, 52 passed, 3 skipped, 4876 deselected in 599.16s (0:09:59) =====
2026-03-17T16:58:23.8892437Z make: *** [common.mk:223: test-coverage] Error 1
2026-03-17T16:58:23.8906627Z ##[error]Process completed with exit code 2.
2026-03-17T16:58:23.9002821Z Post job cleanup.
2026-03-17T16:58:23.9049261Z Post job cleanup.
2026-03-17T16:58:23.9795955Z [command]/usr/bin/git version
2026-03-17T16:58:23.9827524Z git version 2.53.0
2026-03-17T16:58:23.9859528Z Temporarily overriding HOME='/home/runner/work/_temp/4818b1a0-e443-489e-8dec-b45fb0b0162d' before making global git config changes
2026-03-17T16:58:23.9860222Z Adding repository directory to the temporary git global config as a safe directory
2026-03-17T16:58:23.9864788Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/craft-application/craft-application
2026-03-17T16:58:23.9894143Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-17T16:58:23.9921592Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-17T16:58:24.0101441Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-17T16:58:24.0122714Z http.https://github.com/.extraheader
2026-03-17T16:58:24.0132326Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-17T16:58:24.0157108Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-17T16:58:24.0323687Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-17T16:58:24.0348517Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-17T16:58:24.0634160Z Cleaning up orphan processes
2026-03-17T16:58:24.1076308Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: astral-sh/setup-uv@v6. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `a17e425e1fe651c910dc...` conf=0.47
**Evidence:** ['<timestamp>.7808193Z DEBUG craft_providers.lxd.lxc:lxc.py:525 Executing in container: lxc --project testcraft exec local:testcraft-full-project-foo-71 -- env CRAFT_MANAGED_MODE=1 DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true DEBIAN_PRIORITY=critical GOPROXY=direct http_proxy=***10.106.133.1:13444/ https_proxy=***10.106.133.1:13444/ REQUESTS_CA_BUNDLE=/usr/local/share/ca-certificates/local-ca.crt CARGO_HTTP_CAINFO=/usr/local/share/ca-certificates/local-ca.crt apt install -y hello']
```
2026-03-17T17:42:36.1650145Z tests/integration/services/test_provider.py::test_run_managed[True]
2026-03-17T17:42:36.1651364Z   /home/runner/work/craft-application/craft-application/.venv/lib/python3.12/site-packages/craft_providers/bases/ubuntu.py:363: DeprecationWarning: path is deprecated. Use files() instead. Refer to https://importlib-resources.readthedocs.io/en/latest/using.html#migrating-from-legacy for migration advice.
2026-03-17T17:42:36.1652504Z     with importlib.resources.path(
2026-03-17T17:42:36.1652653Z 
2026-03-17T17:42:36.1652820Z -- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
2026-03-17T17:42:36.1653256Z = 3 failed, 50 passed, 3 skipped, 4876 deselected, 3 warnings in 722.39s (0:12:02) =
2026-03-17T17:42:36.4897368Z make: *** [common.mk:223: test-coverage] Error 1
2026-03-17T17:42:36.4912332Z ##[error]Process completed with exit code 2.
2026-03-17T17:42:36.5016927Z Post job cleanup.
2026-03-17T17:42:36.5065160Z Post job cleanup.
2026-03-17T17:42:36.5752336Z [command]/usr/bin/git version
2026-03-17T17:42:36.5789539Z git version 2.52.0
2026-03-17T17:42:36.5822840Z Temporarily overriding HOME='/home/runner/work/_temp/91947c8a-d06a-4938-a6c5-cb7a0da373c5' before making global git config changes
2026-03-17T17:42:36.5823689Z Adding repository directory to the temporary git global config as a safe directory
2026-03-17T17:42:36.5827920Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/craft-application/craft-application
2026-03-17T17:42:36.5863738Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-17T17:42:36.5894880Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-17T17:42:36.6105123Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-17T17:42:36.6128665Z http.https://github.com/.extraheader
2026-03-17T17:42:36.6139840Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-17T17:42:36.6172349Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-17T17:42:36.6377796Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-17T17:42:36.6409493Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-17T17:42:36.6749826Z Cleaning up orphan processes
2026-03-17T17:42:36.7334738Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: astral-sh/setup-uv@v6. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `dependency-removed-upstream` — 9 matches, 9 low-conf (100%)

**Fixture:** `cce15184b39f0536a4e4...` conf=0.47
**Evidence:** ['<timestamp>.5106412Z <timestamp> ERROR <job_1279987305> 404 Not Found']
```
2026-03-15T22:04:42.8891390Z   proxy | 2026/03/15 22:04:42 [248] PATCH /update_jobs/1279987305/mark_as_processed
2026-03-15T22:04:42.9508817Z   proxy | 2026/03/15 22:04:42 [248] 204 /update_jobs/1279987305/mark_as_processed
2026-03-15T22:04:42.9556937Z updater | 2026/03/15 22:04:42 INFO <job_1279987305> Finished job processing
2026-03-15T22:04:42.9569139Z updater | 2026/03/15 22:04:42 INFO Results:
2026-03-15T22:04:42.9570692Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-15T22:04:42.9572073Z +--------------------------------------------+
2026-03-15T22:04:42.9572811Z |       Dependencies failed to update        |
2026-03-15T22:04:42.9573882Z +------------+---------------+---------------+
2026-03-15T22:04:42.9574667Z | Dependency | Error Type    | Error Details |
2026-03-15T22:04:42.9575475Z +------------+---------------+---------------+
2026-03-15T22:04:42.9576192Z | golang     | unknown_error | null          |
2026-03-15T22:04:42.9577639Z +------------+---------------+---------------+
2026-03-15T22:04:43.0920560Z Failure running container 06f88e76b00ecd815557dfb7b7eefca933be890cfe9f06bbda0e48c63a771161: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-15T22:04:43.2168201Z Cleaned up container 06f88e76b00ecd815557dfb7b7eefca933be890cfe9f06bbda0e48c63a771161
2026-03-15T22:04:43.2270235Z   proxy | 2026/03/15 22:04:43 44/124 calls cached (35%)
2026-03-15T22:04:43.2271104Z 2026/03/15 22:04:43 Posting metrics to remote API endpoint
2026-03-15T22:04:43.5572768Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/MaterializeInc/terraform-provider-materialize/network/updates/1279987305 (write access to the repository is required to view the log)
2026-03-15T22:04:43.5584543Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-15T22:04:43.5682233Z Post job cleanup.
2026-03-15T22:04:43.7408990Z Cleaning up orphan processes
2026-03-15T22:04:43.8001944Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `b9e305cb0154cc0a6037...` conf=0.47
**Evidence:** ['<timestamp>.4121727Z <timestamp> ERROR <job_1263736123> 404 Not Found']
```
2026-03-01T22:12:28.7524048Z   proxy | 2026/03/01 22:12:28 [690] 204 /update_jobs/1263736123/record_cooldown_meta
2026-03-01T22:12:28.8802541Z   proxy | 2026/03/01 22:12:28 [692] PATCH /update_jobs/1263736123/mark_as_processed
2026-03-01T22:12:28.9626631Z   proxy | 2026/03/01 22:12:28 [692] 204 /update_jobs/1263736123/mark_as_processed
2026-03-01T22:12:28.9665610Z updater | 2026/03/01 22:12:28 INFO <job_1263736123> Finished job processing
2026-03-01T22:12:28.9675247Z updater | 2026/03/01 22:12:28 INFO Results:
2026-03-01T22:12:28.9676500Z Dependabot encountered '1' error(s) during execution, please check the logs for more details.
2026-03-01T22:12:28.9677290Z +--------------------------------------------+
2026-03-01T22:12:28.9677779Z |       Dependencies failed to update        |
2026-03-01T22:12:28.9678216Z +------------+---------------+---------------+
2026-03-01T22:12:28.9678651Z | Dependency | Error Type    | Error Details |
2026-03-01T22:12:28.9679143Z +------------+---------------+---------------+
2026-03-01T22:12:28.9679493Z | postgres   | unknown_error | null          |
2026-03-01T22:12:28.9679832Z +------------+---------------+---------------+
2026-03-01T22:12:29.0998521Z Failure running container 2221f8bcc109b9dfb5921e1b81b4657650a477ef16777ca527bb053d60322560: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-01T22:12:29.2062136Z Cleaned up container 2221f8bcc109b9dfb5921e1b81b4657650a477ef16777ca527bb053d60322560
2026-03-01T22:12:29.2145148Z   proxy | 2026/03/01 22:12:29 114/346 calls cached (32%)
2026-03-01T22:12:29.2146180Z 2026/03/01 22:12:29 Posting metrics to remote API endpoint
2026-03-01T22:12:29.8469703Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/MaterializeInc/terraform-provider-materialize/network/updates/1263736123 (write access to the repository is required to view the log)
2026-03-01T22:12:29.8480842Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-01T22:12:29.8584648Z Post job cleanup.
2026-03-01T22:12:30.0244165Z Cleaning up orphan processes
```

**Fixture:** `ccfd13d7e39a6a3c2ed7...` conf=0.47
**Evidence:** ['<timestamp>.0773072Z updater | <timestamp> INFO <job_1294730692> VulnerabilityAuditor: npm:vulnerabilityAuditor failed after 2.59s while auditing yaml: No matching version found for undefined@undefined.']
```
2026-03-26T18:16:31.9600298Z |                              |       "version": "2.241.0",                                                                                           |
2026-03-26T18:16:31.9601398Z |                              |       "requirement": "1.10.2"                                                                                         |
2026-03-26T18:16:31.9602418Z |                              |     },                                                                                                                |
2026-03-26T18:16:31.9603457Z |                              |     {                                                                                                                 |
2026-03-26T18:16:31.9604843Z |                              |       "explanation": "cdk-nag@2.37.55 requires yaml@1.10.2 via aws-cdk-lib@2.241.0",                                  |
2026-03-26T18:16:31.9606076Z |                              |       "name": "aws-cdk-lib",                                                                                          |
2026-03-26T18:16:31.9607233Z |                              |       "version": "2.241.0",                                                                                           |
2026-03-26T18:16:31.9608284Z |                              |       "requirement": "1.10.2"                                                                                         |
2026-03-26T18:16:31.9677706Z |                              |     }                                                                                                                 |
2026-03-26T18:16:31.9678898Z |                              |   ]                                                                                                                   |
2026-03-26T18:16:31.9679822Z |                              | }                                                                                                                     |
2026-03-26T18:16:31.9680840Z +------------------------------+-----------------------------------------------------------------------------------------------------------------------+
2026-03-26T18:16:32.0841841Z Failure running container 1099d8544e9ccad462f88b2e549a21004099f82e1a4abe5406400247eb0c7b87: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-03-26T18:16:32.1896668Z Cleaned up container 1099d8544e9ccad462f88b2e549a21004099f82e1a4abe5406400247eb0c7b87
2026-03-26T18:16:32.2010022Z   proxy | 2026/03/26 18:16:32 0/18 calls cached (0%)
2026-03-26T18:16:32.2010708Z 2026/03/26 18:16:32 Posting metrics to remote API endpoint
2026-03-26T18:16:32.5316130Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/aws-samples/genai-knowledge-capture-webapp/network/updates/1294730692 (write access to the repository is required to view the log)
2026-03-26T18:16:32.5326855Z 🤖 ~ finished: error reported to Dependabot ~
2026-03-26T18:16:32.5416196Z Post job cleanup.
2026-03-26T18:16:32.7059522Z Cleaning up orphan processes
2026-03-26T18:16:32.7510849Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: github/dependabot-action@main. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `aws-credentials` — 7 matches, 7 low-conf (100%)

**Fixture:** `e44f5208940183967f2d...` conf=0.37
**Evidence:** ['<timestamp>.3623919Z ##[error]Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers']
```
2026-04-17T10:07:30.5788500Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-17T10:07:30.6011931Z Removing HTTP extra header
2026-04-17T10:07:30.6017531Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-17T10:07:30.6050175Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-17T10:07:30.6268494Z Removing includeIf entries pointing to credentials config files
2026-04-17T10:07:30.6275036Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-17T10:07:30.6298652Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-17T10:07:30.6299912Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-17T10:07:30.6300556Z includeif.gitdir:/github/workspace/.git.path
2026-04-17T10:07:30.6300911Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-17T10:07:30.6309371Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-17T10:07:30.6330139Z /home/runner/work/_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6341104Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path /home/runner/work/_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6373969Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-17T10:07:30.6394997Z /home/runner/work/_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6404329Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6435322Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-17T10:07:30.6456297Z /github/runner_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6465115Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6495287Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-17T10:07:30.6515626Z /github/runner_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6525470Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config
2026-04-17T10:07:30.6557368Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-17T10:07:30.6775208Z Removing credentials config '/home/runner/work/_temp/git-credentials-17f9b0e7-8317-404d-b787-e78200217cd3.config'
2026-04-17T10:07:30.6918213Z Cleaning up orphan processes
```

**Fixture:** `d1e62754535468a0b91a...` conf=0.37
**Evidence:** ['<timestamp>.1197751Z ##[error]Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers']
```
2026-04-17T10:07:45.3466861Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-17T10:07:45.3751821Z Removing HTTP extra header
2026-04-17T10:07:45.3758347Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-17T10:07:45.3798859Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-17T10:07:45.4044215Z Removing includeIf entries pointing to credentials config files
2026-04-17T10:07:45.4053524Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-17T10:07:45.4078159Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-17T10:07:45.4079844Z includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-17T10:07:45.4080956Z includeif.gitdir:/github/workspace/.git.path
2026-04-17T10:07:45.4081686Z includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-17T10:07:45.4090600Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path
2026-04-17T10:07:45.4112530Z /home/runner/work/_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4125878Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git.path /home/runner/work/_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4162172Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path
2026-04-17T10:07:45.4189743Z /home/runner/work/_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4202113Z [command]/usr/bin/git config --local --unset includeif.gitdir:/home/runner/work/liquibase-test-harness/liquibase-test-harness/.git/worktrees/*.path /home/runner/work/_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4248049Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git.path
2026-04-17T10:07:45.4264988Z /github/runner_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4275850Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git.path /github/runner_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4308879Z [command]/usr/bin/git config --local --get-all includeif.gitdir:/github/workspace/.git/worktrees/*.path
2026-04-17T10:07:45.4333717Z /github/runner_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4343101Z [command]/usr/bin/git config --local --unset includeif.gitdir:/github/workspace/.git/worktrees/*.path /github/runner_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config
2026-04-17T10:07:45.4377531Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-17T10:07:45.4623007Z Removing credentials config '/home/runner/work/_temp/git-credentials-d4c7b855-acba-4bd0-b10a-d280507c1f07.config'
2026-04-17T10:07:45.4765265Z Cleaning up orphan processes
```

**Fixture:** `fefdc2b64f5f0a3b682e...` conf=0.37
**Evidence:** ['<timestamp>.2645256Z ##[error]Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers']
```
2026-02-02T19:11:49.4983705Z Metadata: read
2026-02-02T19:11:49.4984319Z Packages: read
2026-02-02T19:11:49.4984859Z ##[endgroup]
2026-02-02T19:11:49.4987005Z Secret source: Actions
2026-02-02T19:11:49.4987858Z Prepare workflow directory
2026-02-02T19:11:49.5335530Z Prepare all required actions
2026-02-02T19:11:49.5392957Z Getting action download info
2026-02-02T19:11:50.0510468Z Download action repository 'aws-actions/configure-aws-credentials@v5' (SHA:61815dcd50bd041e203e49132bacad1fd04d2708)
2026-02-02T19:11:50.6816727Z Download action repository 'actions/checkout@v4' (SHA:34e114876b0b11c390a56381ad16ebd13914f8d5)
2026-02-02T19:11:50.7172429Z Download action repository 'actions/setup-node@v4' (SHA:49933ea5288caeca8642d1e84afbd3f7d6820020)
2026-02-02T19:11:50.8059959Z Download action repository 'actions/cache@v4' (SHA:0057852bfaa89a56745cba8c7296529d2fc39830)
2026-02-02T19:11:51.0291959Z Uses: wakeuplabs-io/optimism-making-impact/.github/workflows/deploy-staging.yml@refs/heads/feat/github-oidc (8c3d5a4bfe81defa5708271f6c019e3b5f7a3ec2)
2026-02-02T19:11:51.0297675Z Complete job name: call-deploy-api-workflow / deploy
2026-02-02T19:11:51.1046888Z ##[group]Run aws-actions/configure-aws-credentials@v5
2026-02-02T19:11:51.1047857Z with:
2026-02-02T19:11:51.1048587Z   role-to-assume: ***
2026-02-02T19:11:51.1049158Z   role-session-name: github-actions-deploy-op-retro-impact
2026-02-02T19:11:51.1049805Z   aws-region: sa-east-1
2026-02-02T19:11:51.1050270Z   audience: sts.amazonaws.com
2026-02-02T19:11:51.1050776Z   output-env-credentials: true
2026-02-02T19:11:51.1051495Z ##[endgroup]
2026-02-02T19:11:51.2417471Z It looks like you might be trying to authenticate with OIDC. Did you mean to set the `id-token` permission? If you are not trying to authenticate with OIDC and the action is working successfully, you can ignore this message.
2026-02-02T19:11:51.2645256Z ##[error]Credentials could not be loaded, please check your action inputs: Could not load credentials from any providers
2026-02-02T19:11:51.2896233Z Post job cleanup.
2026-02-02T19:11:51.4081156Z Cleaning up orphan processes
```

### `path-case-mismatch` — 7 matches, 7 low-conf (100%)

**Fixture:** `365554fb0dc13bd082fc...` conf=0.4
**Evidence:** ['<timestamp>.1847643Z proxy | <timestamp> [774] GET https://registry.npmjs.org:443/postcss-attribute-case-insensitive']
```
2026-02-06T00:08:26.7776967Z |                              |       "requirement": "^0.25.0"                                                                                                         |
2026-02-06T00:08:26.7777491Z |                              |     },                                                                                                                                 |
2026-02-06T00:08:26.7777975Z |                              |     {                                                                                                                                  |
2026-02-06T00:08:26.7778541Z |                              |       "dependency_name": "axios",                                                                                                      |
2026-02-06T00:08:26.7779159Z |                              |       "fix_available": false,                                                                                                          |
2026-02-06T00:08:26.7779722Z |                              |       "fix_updates": [],                                                                                                               |
2026-02-06T00:08:26.7780284Z |                              |       "top_level_ancestors": [],                                                                                                       |
2026-02-06T00:08:26.7780928Z |                              |       "explanation": "No patched version available for axios"                                                                          |
2026-02-06T00:08:26.7781524Z |                              |     }                                                                                                                                  |
2026-02-06T00:08:26.7782345Z |                              |   ]                                                                                                                                    |
2026-02-06T00:08:26.7782863Z |                              | }                                                                                                                                      |
2026-02-06T00:08:26.7783493Z +------------------------------+----------------------------------------------------------------------------------------------------------------------------------------+
2026-02-06T00:08:26.9047228Z Failure running container 37fbb9ed07f2f5fe280455a323caf6f33d5dfcf38235b5e2d2df7555ae05a1ae: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-02-06T00:08:27.8493342Z Cleaned up container 37fbb9ed07f2f5fe280455a323caf6f33d5dfcf38235b5e2d2df7555ae05a1ae
2026-02-06T00:08:27.8616217Z   proxy | 2026/02/06 00:08:27 4/1043 calls cached (0%)
2026-02-06T00:08:27.8616786Z 2026/02/06 00:08:27 Posting metrics to remote API endpoint
2026-02-06T00:08:27.8952438Z   proxy | 2026/02/06 00:08:27 Successfully posted metrics data via api client
2026-02-06T00:08:28.4641903Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/aws-samples/jp-prototyping-blog/network/updates/1235727543 (write access to the repository is required to view the log)
2026-02-06T00:08:28.4650897Z 🤖 ~ finished: error reported to Dependabot ~
2026-02-06T00:08:28.4733299Z Post job cleanup.
2026-02-06T00:08:28.6347631Z Cleaning up orphan processes
```

**Fixture:** `1d9ddff8097099a5127b...` conf=0.4
**Evidence:** ['<timestamp>.4944652Z proxy | <timestamp> [774] GET https://registry.npmjs.org:443/postcss-attribute-case-insensitive']
```
2026-02-06T02:20:59.9767130Z |                              |       "requirement": "^0.25.0"                                                                                                         |
2026-02-06T02:20:59.9767694Z |                              |     },                                                                                                                                 |
2026-02-06T02:20:59.9768226Z |                              |     {                                                                                                                                  |
2026-02-06T02:20:59.9768806Z |                              |       "dependency_name": "axios",                                                                                                      |
2026-02-06T02:20:59.9769433Z |                              |       "fix_available": false,                                                                                                          |
2026-02-06T02:20:59.9770030Z |                              |       "fix_updates": [],                                                                                                               |
2026-02-06T02:20:59.9770630Z |                              |       "top_level_ancestors": [],                                                                                                       |
2026-02-06T02:20:59.9771462Z |                              |       "explanation": "No patched version available for axios"                                                                          |
2026-02-06T02:20:59.9772383Z |                              |     }                                                                                                                                  |
2026-02-06T02:20:59.9772914Z |                              |   ]                                                                                                                                    |
2026-02-06T02:20:59.9773438Z |                              | }                                                                                                                                      |
2026-02-06T02:20:59.9774096Z +------------------------------+----------------------------------------------------------------------------------------------------------------------------------------+
2026-02-06T02:21:00.1084935Z Failure running container 505a8af8e30ea6462597eda8cdded3f961f1028037743266a409089175363d2f: Error: Command failed with exit code 1: /bin/sh -c $DEPENDABOT_HOME/dependabot-updater/bin/run update_files
2026-02-06T02:21:01.0053330Z Cleaned up container 505a8af8e30ea6462597eda8cdded3f961f1028037743266a409089175363d2f
2026-02-06T02:21:01.0179746Z   proxy | 2026/02/06 02:21:01 Posting metrics to remote API endpoint
2026-02-06T02:21:01.0184407Z   proxy | 2026/02/06 02:21:01 3/1043 calls cached (0%)
2026-02-06T02:21:01.0555723Z   proxy | 2026/02/06 02:21:01 Successfully posted metrics data via api client
2026-02-06T02:21:01.6815141Z ##[error]Dependabot encountered an error performing the update

Error: The updater encountered one or more errors.

For more information see: https://github.com/aws-samples/jp-prototyping-blog/network/updates/1235851445 (write access to the repository is required to view the log)
2026-02-06T02:21:01.6826680Z 🤖 ~ finished: error reported to Dependabot ~
2026-02-06T02:21:01.6913576Z Post job cleanup.
2026-02-06T02:21:01.8553263Z Cleaning up orphan processes
```

**Fixture:** `1beeb1d11f4f63e7551f...` conf=0.46
**Evidence:** ['<timestamp>.0726429Z echo "Warning: File not found: $img"', '<timestamp>.0703686Z # Get modified image files in PR (case-insensitive)']
```
2026-04-21T09:54:29.2249962Z   [检查] 文件 'testcases/vela_fs_test/fs/stress/write_speed.c'...
2026-04-21T09:54:29.2250513Z   [检查] 文件 'testcases/vela_fs_test/fs_test.mk'...
2026-04-21T09:54:29.2251070Z   [检查] 文件 'testcases/vela_fs_test/rdonly_fs/include/md5.h'...
2026-04-21T09:54:29.2251681Z   [检查] 文件 'testcases/vela_fs_test/rdonly_fs/lib/md5.c'...
2026-04-21T09:54:29.2252306Z   [检查] 文件 'testcases/vela_fs_test/rdonly_fs/md5_test.c'...
2026-04-21T09:54:29.2252641Z 
2026-04-21T09:54:29.2252904Z --- 检查未通过：代码文件中包含中文字符。 ---
2026-04-21T09:54:29.3173209Z ❌ Chinese character check failed in source files
2026-04-21T09:54:29.3186144Z ##[error]Process completed with exit code 1.
2026-04-21T09:54:29.3294385Z Post job cleanup.
2026-04-21T09:54:29.4260790Z [command]/usr/bin/git version
2026-04-21T09:54:29.4298806Z git version 2.53.0
2026-04-21T09:54:29.4350463Z Temporarily overriding HOME='/home/runner/work/_temp/d20959c4-6b9d-4231-b792-bc50bf313cf3' before making global git config changes
2026-04-21T09:54:29.4351314Z Adding repository directory to the temporary git global config as a safe directory
2026-04-21T09:54:29.4356041Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/tests/tests/tests
2026-04-21T09:54:29.4391232Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-04-21T09:54:29.4423613Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-04-21T09:54:29.4648325Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-04-21T09:54:29.4669118Z http.https://github.com/.extraheader
2026-04-21T09:54:29.4681617Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-04-21T09:54:29.4712359Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-04-21T09:54:29.4928702Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-04-21T09:54:29.4958838Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-04-21T09:54:29.5296619Z Cleaning up orphan processes
2026-04-21T09:54:29.5665734Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: actions/checkout@v4. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Node.js 20 will be removed from the runner on September 16th, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

### `test-timeout` — 7 matches, 7 low-conf (100%)

**Fixture:** `b237c6dee2f1b09d3c22...` conf=0.33
**Evidence:** ['<timestamp>.6806840Z ✗ testPostErrorMessage, Asynchronous wait failed: Exceeded timeout of 60 seconds, with unfulfilled expectations: &quot;Evaluate JS&quot;.']
```
2026-01-30T04:56:37.9073650Z 
2026-01-30T04:56:37.9073650Z 
2026-01-30T04:56:37.9073810Z Executed 24 tests, with 0 failures (0 unexpected) in 0.094 (0.108) seconds
2026-01-30T04:57:04.3572840Z 2026-01-30 04:57:04.354 xcodebuild[8301:41177] [MT] IDETestOperationsObserverDebug: 195.745 elapsed -- Testing started completed.
2026-01-30T04:57:04.3665990Z 2026-01-30 04:57:04.354 xcodebuild[8301:41177] [MT] IDETestOperationsObserverDebug: 0.000 sec, +0.000 sec -- start
2026-01-30T04:57:04.3769210Z 2026-01-30 04:57:04.354 xcodebuild[8301:41177] [MT] IDETestOperationsObserverDebug: 195.745 sec, +195.745 sec -- end
2026-01-30T04:57:04.5997480Z ** TEST FAILED **
2026-01-30T04:57:04.6097520Z 
2026-01-30T04:57:11.3272050Z ##[error]Process completed with exit code 65.
2026-01-30T04:57:11.3740200Z Post job cleanup.
2026-01-30T04:57:11.8271180Z [command]/opt/homebrew/bin/git version
2026-01-30T04:57:11.8576410Z git version 2.52.0
2026-01-30T04:57:11.9037250Z Copying '/Users/runner/.gitconfig' to '/Users/runner/work/_temp/e7a15afa-cf08-4185-89aa-e263ec27d1b0/.gitconfig'
2026-01-30T04:57:11.9157230Z Temporarily overriding HOME='/Users/runner/work/_temp/e7a15afa-cf08-4185-89aa-e263ec27d1b0' before making global git config changes
2026-01-30T04:57:11.9226060Z Adding repository directory to the temporary git global config as a safe directory
2026-01-30T04:57:11.9327950Z [command]/opt/homebrew/bin/git config --global --add safe.directory /Users/runner/work/swift/swift
2026-01-30T04:57:11.9646800Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-01-30T04:57:11.9950490Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-01-30T04:57:12.0808010Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-01-30T04:57:12.0915320Z http.https://github.com/.extraheader
2026-01-30T04:57:12.1204280Z [command]/opt/homebrew/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-01-30T04:57:12.1491710Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-01-30T04:57:12.2407090Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-01-30T04:57:12.2711450Z [command]/opt/homebrew/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-01-30T04:57:12.4104160Z Cleaning up orphan processes
```

**Fixture:** `8fc95c79748a3db2db5e...` conf=0.33
**Evidence:** ['<timestamp>.6805800Z ✗ testPostErrorMessage, Asynchronous wait failed: Exceeded timeout of 60 seconds, with unfulfilled expectations: &quot;Evaluate JS&quot;.']
```
2026-03-05T10:48:53.9131270Z 
2026-03-05T10:48:53.9131270Z 
2026-03-05T10:48:53.9131420Z Executed 25 tests, with 0 failures (0 unexpected) in 0.113 (0.135) seconds
2026-03-05T10:49:48.9561100Z 2026-03-05 10:49:48.946 xcodebuild[15604:69279] [MT] IDETestOperationsObserverDebug: 225.292 elapsed -- Testing started completed.
2026-03-05T10:49:48.9659120Z 2026-03-05 10:49:48.946 xcodebuild[15604:69279] [MT] IDETestOperationsObserverDebug: 0.000 sec, +0.000 sec -- start
2026-03-05T10:49:48.9768260Z 2026-03-05 10:49:48.946 xcodebuild[15604:69279] [MT] IDETestOperationsObserverDebug: 225.292 sec, +225.292 sec -- end
2026-03-05T10:49:49.5912540Z ** TEST FAILED **
2026-03-05T10:49:49.6014980Z 
2026-03-05T10:49:57.7790620Z ##[error]Process completed with exit code 65.
2026-03-05T10:49:57.8243670Z Post job cleanup.
2026-03-05T10:49:58.1778470Z [command]/opt/homebrew/bin/git version
2026-03-05T10:49:58.2257440Z git version 2.53.0
2026-03-05T10:49:58.2301650Z Copying '/Users/runner/.gitconfig' to '/Users/runner/work/_temp/c0f60c23-24cb-4d0b-a47c-eda26656eabb/.gitconfig'
2026-03-05T10:49:58.2316700Z Temporarily overriding HOME='/Users/runner/work/_temp/c0f60c23-24cb-4d0b-a47c-eda26656eabb' before making global git config changes
2026-03-05T10:49:58.2367740Z Adding repository directory to the temporary git global config as a safe directory
2026-03-05T10:49:58.2368500Z [command]/opt/homebrew/bin/git config --global --add safe.directory /Users/runner/work/swift/swift
2026-03-05T10:49:58.3151300Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-05T10:49:58.3826060Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-05T10:49:58.5804800Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-05T10:49:58.5835270Z http.https://github.com/.extraheader
2026-03-05T10:49:58.6097150Z [command]/opt/homebrew/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-05T10:49:58.6166890Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-05T10:49:58.7555770Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-05T10:49:58.7621760Z [command]/opt/homebrew/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-05T10:49:58.8565640Z Cleaning up orphan processes
```

**Fixture:** `bf4b64b285c165506be1...` conf=0.33
**Evidence:** ['<timestamp>.8323100Z ✗ testPostErrorMessage, Asynchronous wait failed: Exceeded timeout of 60 seconds, with unfulfilled expectations: &quot;Evaluate JS&quot;.']
```
2026-02-02T18:17:34.0053360Z 
2026-02-02T18:17:34.0053360Z 
2026-02-02T18:17:34.0053690Z Executed 24 tests, with 0 failures (0 unexpected) in 0.226 (0.241) seconds
2026-02-02T18:18:26.8496010Z 2026-02-02 18:18:26.847 xcodebuild[1482:8278] [MT] IDETestOperationsObserverDebug: 1047.974 elapsed -- Testing started completed.
2026-02-02T18:18:26.8588800Z 2026-02-02 18:18:26.847 xcodebuild[1482:8278] [MT] IDETestOperationsObserverDebug: 0.000 sec, +0.000 sec -- start
2026-02-02T18:18:26.8709320Z 2026-02-02 18:18:26.847 xcodebuild[1482:8278] [MT] IDETestOperationsObserverDebug: 1047.974 sec, +1047.974 sec -- end
2026-02-02T18:18:27.3927520Z ** TEST FAILED **
2026-02-02T18:18:27.4138420Z 
2026-02-02T18:18:36.7983620Z ##[error]Process completed with exit code 65.
2026-02-02T18:18:36.9353300Z Post job cleanup.
2026-02-02T18:18:37.6379280Z [command]/opt/homebrew/bin/git version
2026-02-02T18:18:37.7087740Z git version 2.52.0
2026-02-02T18:18:37.8466550Z Copying '/Users/runner/.gitconfig' to '/Users/runner/work/_temp/1d550388-1a57-4b51-806c-7154732b390c/.gitconfig'
2026-02-02T18:18:37.8626740Z Temporarily overriding HOME='/Users/runner/work/_temp/1d550388-1a57-4b51-806c-7154732b390c' before making global git config changes
2026-02-02T18:18:37.8800980Z Adding repository directory to the temporary git global config as a safe directory
2026-02-02T18:18:37.8904460Z [command]/opt/homebrew/bin/git config --global --add safe.directory /Users/runner/work/swift/swift
2026-02-02T18:18:37.9618870Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-02T18:18:38.0492590Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-02T18:18:38.4429010Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-02T18:18:38.4535840Z http.https://github.com/.extraheader
2026-02-02T18:18:38.4843240Z [command]/opt/homebrew/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-02T18:18:38.5183000Z [command]/opt/homebrew/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-02T18:18:38.7639070Z [command]/opt/homebrew/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-02T18:18:38.7743850Z [command]/opt/homebrew/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-02T18:18:39.1096730Z Cleaning up orphan processes
```

### `process-killed-no-logs` — 6 matches, 6 low-conf (100%)

**Fixture:** `32a2bac73415b37bc501...` conf=0.46
**Evidence:** ['<timestamp>.9119790Z `;_q(X,$)}function Z5(Q,X){let Y={...Q};if(X){let $={sandbox:X};if(Y.settings)try{$={...C6(Y.settings),sandbox:X}}catch{}Y.settings=Z0($)}return Y}class B4{options;process;processStdin;processStdout;ready=!1;abortController;exitError;exitListeners=[];processExitHandler;abortHandler;constructor(Q){this.options=Q;this.abortController=Q.abortController||R6(),this.initialize()}getDefaultExecutable(){return _6()?"bun":"node"}spawnLocalProcess(Q){let{command:X,args:Y,cwd:$,env:W,signal:J}=Q,G=W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr?"pipe":"ignore",H=Tq(X,Y,{cwd:$,stdio:["pipe","pipe",G],signal:J,env:W,windowsHide:!0});if(W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr)H.stderr.on("data",(z)=>{let K=z.toString();if(N1(K),this.options.stderr)this.options.stderr(K)});return{stdin:H.stdin,stdout:H.stdout,get killed(){return H.killed},get exitCode(){return H.exitCode},kill:H.kill.bind(H),on:H.on.bind(H),once:H.once.bind(H),off:H.off.bind(H)}}initialize(){try{let{additionalDirectories:Q=[],agent:X,betas:Y,cwd:$,executable:W=this.getDefaultExecutable(),executableArgs:J=[],extraArgs:G={},pathToClaudeCodeExecutable:H,env:B={...process.env},thinkingConfig:z,maxTurns:K,maxBudgetUsd:q,model:U,fallbackModel:V,jsonSchema:L,permissionMode:F,allowDangerouslySkipPermissions:w,permissionPromptToolName:D,continueConversation:j,resume:I,settingSources:T,allowedTools:v=[],disallowedTools:N0=[],tools:O0,mcpServers:c0,strictMcpConfig:t1,canUseTool:q1,includePartialMessages:a1,plugins:P1,sandbox:s1}=this.options,u=["--output-format","stream-json","--verbose","--input-format","stream-json"];if(z)switch(z.type){case"enabled":if(z.budgetTokens===void 0)u.push("--thinking","adaptive");else u.push("--max-thinking-tokens",z.budgetTokens.toString());break;case"disabled":u.push("--thinking","disabled");break;case"adaptive":u.push("--thinking","adaptive");break}if(this.options.effort)u.push("--effort",this.options.effort);if(K)u.push("--max-turns",K.toString());if(q!==void 0)u.push("--max-budget-usd",q.toString());if(U)u.push("--model",U);if(X)u.push("--agent",X);if(Y&&Y.length>0)u.push("--betas",Y.join(","));if(L)u.push("--json-schema",Z0(L));if(this.options.debugFile)u.push("--debug-file",this.options.debugFile);else if(this.options.debug)u.push("--debug");if(B.DEBUG_CLAUDE_AGENT_SDK)u.push("--debug-to-stderr");if(q1){if(D)throw Error("canUseTool callback cannot be used with permissionPromptToolName. Please use one or the other.");u.push("--permission-prompt-tool","stdio")}else if(D)u.push("--permission-prompt-tool",D);if(j)u.push("--continue");if(I)u.push("--resume",I);if(v.length>0)u.push("--allowedTools",v.join(","));if(N0.length>0)u.push("--disallowedTools",N0.join(","));if(O0!==void 0)if(Array.isArray(O0))if(O0.length===0)u.push("--tools","");else u.push("--tools",O0.join(","));else u.push("--tools","default");if(c0&&Object.keys(c0).length>0)u.push("--mcp-config",Z0({mcpServers:c0}));if(T)u.push("--setting-sources",T.join(","));if(t1)u.push("--strict-mcp-config");if(F)u.push("--permission-mode",F);if(w)u.push("--allow-dangerously-skip-permissions");if(V){if(U&&V===U)throw Error("Fallback model cannot be the same as the main model. Please specify a different model for fallbackModel option.");u.push("--fallback-model",V)}if(a1)u.push("--include-partial-messages");for(let b0 of Q)u.push("--add-dir",b0);if(P1&&P1.length>0)for(let b0 of P1)if(b0.type==="local")u.push("--plugin-dir",b0.path);else throw Error(`Unsupported plugin type: ${b0.type}`);if(this.options.forkSession)u.push("--fork-session");if(this.options.resumeSessionAt)u.push("--resume-session-at",this.options.resumeSessionAt);if(this.options.sessionId)u.push("--session-id",this.options.sessionId);if(this.options.persistSession===!1)u.push("--no-session-persistence");let AQ=Z5(G??{},s1);for(let[b0,Z1]of Object.entries(AQ))if(Z1===null)u.push(`--${b0}`);else u.push(`--${b0}`,Z1);if(!B.CLAUDE_CODE_ENTRYPOINT)B.CLAUDE_CODE_ENTRYPOINT="sdk-ts";if(delete B.NODE_OPTIONS,B.DEBUG_CLAUDE_AGENT_SDK)B.DEBUG="1";else delete B.DEBUG;let e1=yq(H),J4=e1?H:W,Q6=e1?[...J,...u]:[...J,H,...u],A9={command:J4,args:Q6,cwd:$,env:B,signal:this.abortController.signal};if(this.options.spawnClaudeCodeProcess)N1(`Spawning Claude Code (custom): ${J4} ${Q6.join(" ")}`),this.process=this.options.spawnClaudeCodeProcess(A9);else{if(!k1().existsSync(H)){let Z1=e1?`Claude Code native binary not found at ${H}. Please ensure Claude Code is installed via native installer or specify a valid path with options.pathToClaudeCodeExecutable.`:`Claude Code executable not found at ${H}. Is options.pathToClaudeCodeExecutable set?`;throw ReferenceError(Z1)}N1(`Spawning Claude Code: ${J4} ${Q6.join(" ")}`),this.process=this.spawnLocalProcess(A9)}this.processStdin=this.process.stdin,this.processStdout=this.process.stdout;let j9=()=>{if(this.process&&!this.process.killed)this.process.kill("SIGTERM")};this.processExitHandler=j9,this.abortHandler=j9,process.on("exit",this.processExitHandler),this.abortController.signal.addEventListener("abort",this.abortHandler),this.process.on("error",(b0)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else this.exitError=Error(`Failed to spawn Claude Code process: ${b0.message}`),N1(this.exitError.message)}),this.process.on("exit",(b0,Z1)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else{let w6=this.getProcessExitError(b0,Z1);if(w6)this.exitError=w6,N1(w6.message)}}),this.ready=!0}catch(Q){throw this.ready=!1,Q}}getProcessExitError(Q,X){if(Q!==0&&Q!==null)return Error(`Claude Code process exited with code ${Q}`);else if(X)return Error(`Claude Code process terminated by signal ${X}`);return}write(Q){if(this.abortController.signal.aborted)throw new F1("Operation aborted");if(!this.ready||!this.processStdin)throw Error("ProcessTransport is not ready for writing");if(this.process?.killed||this.process?.exitCode!==null)throw Error("Cannot write to terminated process");if(this.exitError)throw Error(`Cannot write to process that exited with error: ${this.exitError.message}`);N1(`[ProcessTransport] Writing to stdin: ${Q.substring(0,100)}`);try{if(!this.processStdin.write(Q))N1("[ProcessTransport] Write buffer full, data queued")}catch(X){throw this.ready=!1,Error(`Failed to write to process stdin: ${X.message}`)}}close(){if(this.processStdin)this.processStdin.end(),this.processStdin=void 0;if(this.abortHandler)this.abortController.signal.removeEventListener("abort",this.abortHandler),this.abortHandler=void 0;for(let{handler:Q}of this.exitListeners)this.process?.off("exit",Q);if(this.exitListeners=[],this.process&&!this.process.killed&&this.process.exitCode===null)this.process.kill("SIGTERM"),setTimeout(()=>{if(this.process&&!this.process.killed)this.process.kill("SIGKILL")},5000).unref();if(this.ready=!1,this.processExitHandler)process.off("exit",this.processExitHandler),this.processExitHandler=void 0}isReady(){return this.ready}async*readMessages(){if(!this.processStdout)throw Error("ProcessTransport output stream not available");let Q=xq({input:this.processStdout});try{for await(let X of Q)if(X.trim())try{yield C6(X)}catch(Y){throw N1(`Non-JSON stdout: ${X}`),Error(`CLI output was not valid JSON. This may indicate an error during startup. Output: ${X.slice(0,200)}${X.length>200?"...":""}`)}await this.waitForExit()}catch(X){throw X}finally{Q.close()}}endInput(){if(this.processStdin)this.processStdin.end()}getInputStream(){return this.processStdin}onExit(Q){if(!this.process)return()=>{};let X=(Y,$)=>{let W=this.getProcessExitError(Y,$);Q(W)};return this.process.on("exit",X),this.exitListeners.push({callback:Q,handler:X}),()=>{if(this.process)this.process.off("exit",X);let Y=this.exitListeners.findIndex(($)=>$.handler===X);if(Y!==-1)this.exitListeners.splice(Y,1)}}async waitForExit(){if(!this.process){if(this.exitError)throw this.exitError;return}if(this.process.exitCode!==null||this.process.killed){if(this.exitError)throw this.exitError;return}return new Promise((Q,X)=>{let Y=(W,J)=>{if(this.abortController.signal.aborted){X(new F1("Operation aborted"));return}let G=this.getProcessExitError(W,J);if(G)X(G);else Q()};this.process.once("exit",Y);let $=(W)=>{this.process.off("exit",Y),X(W)};this.process.once("error",$),this.process.once("exit",()=>{this.process.off("error",$)})})}}function yq(Q){return![".js",".mjs",".tsx",".ts",".jsx"].some((Y)=>Q.endsWith(Y))}class z4{returned;queue=[];readResolve;readReject;isDone=!1;hasError;started=!1;constructor(Q){this.returned=Q}[Symbol.asyncIterator](){if(this.started)throw Error("Stream can only be iterated once");return this.started=!0,this}next(){if(this.queue.length>0)return Promise.resolve({done:!1,value:this.queue.shift()});if(this.isDone)return Promise.resolve({done:!0,value:void 0});if(this.hasError)return Promise.reject(this.hasError);return new Promise((Q,X)=>{this.readResolve=Q,this.readReject=X})}enqueue(Q){if(this.readResolve){let X=this.readResolve;this.readResolve=void 0,this.readReject=void 0,X({done:!1,value:Q})}else this.queue.push(Q)}done(){if(this.isDone=!0,this.readResolve){let Q=this.readResolve;this.readResolve=void 0,this.readReject=void 0,Q({done:!0,value:void 0})}}error(Q){if(this.hasError=Q,this.readReject){let X=this.readReject;this.readResolve=void 0,this.readReject=void 0,X(Q)}}return(){if(this.isDone=!0,this.returned)this.returned();return Promise.resolve({done:!0,value:void 0})}}class PQ{sendMcpMessage;isClosed=!1;constructor(Q){this.sendMcpMessage=Q}onclose;onerror;onmessage;async start(){}async send(Q){if(this.isClosed)throw Error("Transport is closed");this.sendMcpMessage(Q)}async close(){if(this.isClosed)return;this.isClosed=!0,this.onclose?.()}}import{randomUUID as gq}from"crypto";class K4{transport;isSingleUserTurn;canUseTool;hooks;abortController;jsonSchema;initConfig;pendingControlResponses=new Map;cleanupPerformed=!1;sdkMessages;inputStream=new z4;initialization;cancelControllers=new Map;hookCallbacks=new Map;nextCallbackId=0;sdkMcpTransports=new Map;sdkMcpServerInstances=new Map;pendingMcpResponses=new Map;firstResultReceivedResolve;firstResultReceived=!1;hasBidirectionalNeeds(){return this.sdkMcpTransports.size>0||this.hooks!==void 0&&Object.keys(this.hooks).length>0||this.canUseTool!==void 0}constructor(Q,X,Y,$,W,J=new Map,G,H){this.transport=Q;this.isSingleUserTurn=X;this.canUseTool=Y;this.hooks=$;this.abortController=W;this.jsonSchema=G;this.initConfig=H;for(let[B,z]of J)this.connectSdkMcpServer(B,z);this.sdkMessages=this.readSdkMessages(),this.readMessages(),this.initialization=this.initialize(),this.initialization.catch(()=>{})}setError(Q){this.inputStream.error(Q)}async stopTask(Q){await this.request({subtype:"stop_task",task_id:Q})}close(){this.cleanup()}cleanup(Q){if(this.cleanupPerformed)return;this.cleanupPerformed=!0;try{this.transport.close();let X=Error("Query closed before response received");for(let{reject:Y}of this.pendingControlResponses.values())Y(X);this.pendingControlResponses.clear();for(let{reject:Y}of this.pendingMcpResponses.values())Y(X);this.pendingMcpResponses.clear(),this.cancelControllers.clear(),this.hookCallbacks.clear();for(let Y of this.sdkMcpTransports.values())try{Y.close()}catch{}if(this.sdkMcpTransports.clear(),Q)this.inputStream.error(Q);else this.inputStream.done()}catch(X){}}next(...[Q]){return this.sdkMessages.next(...[Q])}return(Q){return this.sdkMessages.return(Q)}throw(Q){return this.sdkMessages.throw(Q)}[Symbol.asyncIterator](){return this.sdkMessages}[Symbol.asyncDispose](){return this.sdkMessages[Symbol.asyncDispose]()}async readMessages(){try{for await(let Q of this.transport.readMessages()){if(Q.type==="control_response"){let X=this.pendingControlResponses.get(Q.response.request_id);if(X)X.handler(Q.response);continue}else if(Q.type==="control_request"){this.handleControlRequest(Q);continue}else if(Q.type==="control_cancel_request"){this.handleControlCancelRequest(Q);continue}else if(Q.type==="keep_alive")continue;if(Q.type==="streamlined_text"||Q.type==="streamlined_tool_use_summary")continue;if(Q.type==="result"){if(this.firstResultReceived=!0,this.firstResultReceivedResolve)this.firstResultReceivedResolve();if(this.isSingleUserTurn)L1("[Query.readMessages] First result received for single-turn query, closing stdin"),this.transport.endInput()}this.inputStream.enqueue(Q)}if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.done(),this.cleanup()}catch(Q){if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.error(Q),this.cleanup(Q)}}async handleControlRequest(Q){let X=new AbortController;this.cancelControllers.set(Q.request_id,X);try{let Y=await this.processControlRequest(Q,X.signal),$={type:"control_response",response:{subtype:"success",request_id:Q.request_id,response:Y}};await Promise.resolve(this.transport.write(Z0($)+`']
```
2026-02-27T18:22:13.2128410Z go: downloading github.com/prometheus/client_golang v1.19.1
2026-02-27T18:22:13.2135525Z go: downloading golang.org/x/net v0.36.0
2026-02-27T18:22:13.2735088Z go: downloading github.com/beorn7/perks v1.0.1
2026-02-27T18:22:13.2739484Z go: downloading github.com/cespare/xxhash/v2 v2.2.0
2026-02-27T18:22:13.2745097Z go: downloading github.com/prometheus/client_model v0.5.0
2026-02-27T18:22:13.2960535Z go: downloading github.com/prometheus/common v0.48.0
2026-02-27T18:22:13.3014448Z go: downloading github.com/prometheus/procfs v0.12.0
2026-02-27T18:22:13.3661403Z go: downloading google.golang.org/protobuf v1.33.0
2026-02-27T18:22:13.5596183Z go: downloading golang.org/x/sys v0.30.0
2026-02-27T18:22:13.5601657Z go: downloading golang.org/x/text v0.22.0
2026-02-27T18:22:29.3384652Z Post job cleanup.
2026-02-27T18:22:29.4354958Z [command]/usr/bin/git version
2026-02-27T18:22:29.4392960Z git version 2.53.0
2026-02-27T18:22:29.4437697Z Temporarily overriding HOME='/home/runner/work/_temp/4e3e1c18-5e6c-4ffa-9a45-a2728ac1191a' before making global git config changes
2026-02-27T18:22:29.4439602Z Adding repository directory to the temporary git global config as a safe directory
2026-02-27T18:22:29.4444892Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/aiquila/aiquila
2026-02-27T18:22:29.4484616Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-27T18:22:29.4520385Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-27T18:22:29.4768685Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-27T18:22:29.4794026Z http.https://github.com/.extraheader
2026-02-27T18:22:29.4807887Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-27T18:22:29.4840679Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-27T18:22:29.5079018Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-27T18:22:29.5113447Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-27T18:22:29.5465777Z Cleaning up orphan processes
```

**Fixture:** `2e44b09d162b468b1c6b...` conf=0.46
**Evidence:** ['<timestamp>.2331431Z `;_q(X,$)}function Z5(Q,X){let Y={...Q};if(X){let $={sandbox:X};if(Y.settings)try{$={...C6(Y.settings),sandbox:X}}catch{}Y.settings=Z0($)}return Y}class B4{options;process;processStdin;processStdout;ready=!1;abortController;exitError;exitListeners=[];processExitHandler;abortHandler;constructor(Q){this.options=Q;this.abortController=Q.abortController||R6(),this.initialize()}getDefaultExecutable(){return _6()?"bun":"node"}spawnLocalProcess(Q){let{command:X,args:Y,cwd:$,env:W,signal:J}=Q,G=W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr?"pipe":"ignore",H=Tq(X,Y,{cwd:$,stdio:["pipe","pipe",G],signal:J,env:W,windowsHide:!0});if(W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr)H.stderr.on("data",(z)=>{let K=z.toString();if(N1(K),this.options.stderr)this.options.stderr(K)});return{stdin:H.stdin,stdout:H.stdout,get killed(){return H.killed},get exitCode(){return H.exitCode},kill:H.kill.bind(H),on:H.on.bind(H),once:H.once.bind(H),off:H.off.bind(H)}}initialize(){try{let{additionalDirectories:Q=[],agent:X,betas:Y,cwd:$,executable:W=this.getDefaultExecutable(),executableArgs:J=[],extraArgs:G={},pathToClaudeCodeExecutable:H,env:B={...process.env},thinkingConfig:z,maxTurns:K,maxBudgetUsd:q,model:U,fallbackModel:V,jsonSchema:L,permissionMode:F,allowDangerouslySkipPermissions:w,permissionPromptToolName:D,continueConversation:j,resume:I,settingSources:T,allowedTools:v=[],disallowedTools:N0=[],tools:O0,mcpServers:c0,strictMcpConfig:t1,canUseTool:q1,includePartialMessages:a1,plugins:P1,sandbox:s1}=this.options,u=["--output-format","stream-json","--verbose","--input-format","stream-json"];if(z)switch(z.type){case"enabled":if(z.budgetTokens===void 0)u.push("--thinking","adaptive");else u.push("--max-thinking-tokens",z.budgetTokens.toString());break;case"disabled":u.push("--thinking","disabled");break;case"adaptive":u.push("--thinking","adaptive");break}if(this.options.effort)u.push("--effort",this.options.effort);if(K)u.push("--max-turns",K.toString());if(q!==void 0)u.push("--max-budget-usd",q.toString());if(U)u.push("--model",U);if(X)u.push("--agent",X);if(Y&&Y.length>0)u.push("--betas",Y.join(","));if(L)u.push("--json-schema",Z0(L));if(this.options.debugFile)u.push("--debug-file",this.options.debugFile);else if(this.options.debug)u.push("--debug");if(B.DEBUG_CLAUDE_AGENT_SDK)u.push("--debug-to-stderr");if(q1){if(D)throw Error("canUseTool callback cannot be used with permissionPromptToolName. Please use one or the other.");u.push("--permission-prompt-tool","stdio")}else if(D)u.push("--permission-prompt-tool",D);if(j)u.push("--continue");if(I)u.push("--resume",I);if(v.length>0)u.push("--allowedTools",v.join(","));if(N0.length>0)u.push("--disallowedTools",N0.join(","));if(O0!==void 0)if(Array.isArray(O0))if(O0.length===0)u.push("--tools","");else u.push("--tools",O0.join(","));else u.push("--tools","default");if(c0&&Object.keys(c0).length>0)u.push("--mcp-config",Z0({mcpServers:c0}));if(T)u.push("--setting-sources",T.join(","));if(t1)u.push("--strict-mcp-config");if(F)u.push("--permission-mode",F);if(w)u.push("--allow-dangerously-skip-permissions");if(V){if(U&&V===U)throw Error("Fallback model cannot be the same as the main model. Please specify a different model for fallbackModel option.");u.push("--fallback-model",V)}if(a1)u.push("--include-partial-messages");for(let b0 of Q)u.push("--add-dir",b0);if(P1&&P1.length>0)for(let b0 of P1)if(b0.type==="local")u.push("--plugin-dir",b0.path);else throw Error(`Unsupported plugin type: ${b0.type}`);if(this.options.forkSession)u.push("--fork-session");if(this.options.resumeSessionAt)u.push("--resume-session-at",this.options.resumeSessionAt);if(this.options.sessionId)u.push("--session-id",this.options.sessionId);if(this.options.persistSession===!1)u.push("--no-session-persistence");let AQ=Z5(G??{},s1);for(let[b0,Z1]of Object.entries(AQ))if(Z1===null)u.push(`--${b0}`);else u.push(`--${b0}`,Z1);if(!B.CLAUDE_CODE_ENTRYPOINT)B.CLAUDE_CODE_ENTRYPOINT="sdk-ts";if(delete B.NODE_OPTIONS,B.DEBUG_CLAUDE_AGENT_SDK)B.DEBUG="1";else delete B.DEBUG;let e1=yq(H),J4=e1?H:W,Q6=e1?[...J,...u]:[...J,H,...u],A9={command:J4,args:Q6,cwd:$,env:B,signal:this.abortController.signal};if(this.options.spawnClaudeCodeProcess)N1(`Spawning Claude Code (custom): ${J4} ${Q6.join(" ")}`),this.process=this.options.spawnClaudeCodeProcess(A9);else{if(!k1().existsSync(H)){let Z1=e1?`Claude Code native binary not found at ${H}. Please ensure Claude Code is installed via native installer or specify a valid path with options.pathToClaudeCodeExecutable.`:`Claude Code executable not found at ${H}. Is options.pathToClaudeCodeExecutable set?`;throw ReferenceError(Z1)}N1(`Spawning Claude Code: ${J4} ${Q6.join(" ")}`),this.process=this.spawnLocalProcess(A9)}this.processStdin=this.process.stdin,this.processStdout=this.process.stdout;let j9=()=>{if(this.process&&!this.process.killed)this.process.kill("SIGTERM")};this.processExitHandler=j9,this.abortHandler=j9,process.on("exit",this.processExitHandler),this.abortController.signal.addEventListener("abort",this.abortHandler),this.process.on("error",(b0)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else this.exitError=Error(`Failed to spawn Claude Code process: ${b0.message}`),N1(this.exitError.message)}),this.process.on("exit",(b0,Z1)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else{let w6=this.getProcessExitError(b0,Z1);if(w6)this.exitError=w6,N1(w6.message)}}),this.ready=!0}catch(Q){throw this.ready=!1,Q}}getProcessExitError(Q,X){if(Q!==0&&Q!==null)return Error(`Claude Code process exited with code ${Q}`);else if(X)return Error(`Claude Code process terminated by signal ${X}`);return}write(Q){if(this.abortController.signal.aborted)throw new F1("Operation aborted");if(!this.ready||!this.processStdin)throw Error("ProcessTransport is not ready for writing");if(this.process?.killed||this.process?.exitCode!==null)throw Error("Cannot write to terminated process");if(this.exitError)throw Error(`Cannot write to process that exited with error: ${this.exitError.message}`);N1(`[ProcessTransport] Writing to stdin: ${Q.substring(0,100)}`);try{if(!this.processStdin.write(Q))N1("[ProcessTransport] Write buffer full, data queued")}catch(X){throw this.ready=!1,Error(`Failed to write to process stdin: ${X.message}`)}}close(){if(this.processStdin)this.processStdin.end(),this.processStdin=void 0;if(this.abortHandler)this.abortController.signal.removeEventListener("abort",this.abortHandler),this.abortHandler=void 0;for(let{handler:Q}of this.exitListeners)this.process?.off("exit",Q);if(this.exitListeners=[],this.process&&!this.process.killed&&this.process.exitCode===null)this.process.kill("SIGTERM"),setTimeout(()=>{if(this.process&&!this.process.killed)this.process.kill("SIGKILL")},5000).unref();if(this.ready=!1,this.processExitHandler)process.off("exit",this.processExitHandler),this.processExitHandler=void 0}isReady(){return this.ready}async*readMessages(){if(!this.processStdout)throw Error("ProcessTransport output stream not available");let Q=xq({input:this.processStdout});try{for await(let X of Q)if(X.trim())try{yield C6(X)}catch(Y){throw N1(`Non-JSON stdout: ${X}`),Error(`CLI output was not valid JSON. This may indicate an error during startup. Output: ${X.slice(0,200)}${X.length>200?"...":""}`)}await this.waitForExit()}catch(X){throw X}finally{Q.close()}}endInput(){if(this.processStdin)this.processStdin.end()}getInputStream(){return this.processStdin}onExit(Q){if(!this.process)return()=>{};let X=(Y,$)=>{let W=this.getProcessExitError(Y,$);Q(W)};return this.process.on("exit",X),this.exitListeners.push({callback:Q,handler:X}),()=>{if(this.process)this.process.off("exit",X);let Y=this.exitListeners.findIndex(($)=>$.handler===X);if(Y!==-1)this.exitListeners.splice(Y,1)}}async waitForExit(){if(!this.process){if(this.exitError)throw this.exitError;return}if(this.process.exitCode!==null||this.process.killed){if(this.exitError)throw this.exitError;return}return new Promise((Q,X)=>{let Y=(W,J)=>{if(this.abortController.signal.aborted){X(new F1("Operation aborted"));return}let G=this.getProcessExitError(W,J);if(G)X(G);else Q()};this.process.once("exit",Y);let $=(W)=>{this.process.off("exit",Y),X(W)};this.process.once("error",$),this.process.once("exit",()=>{this.process.off("error",$)})})}}function yq(Q){return![".js",".mjs",".tsx",".ts",".jsx"].some((Y)=>Q.endsWith(Y))}class z4{returned;queue=[];readResolve;readReject;isDone=!1;hasError;started=!1;constructor(Q){this.returned=Q}[Symbol.asyncIterator](){if(this.started)throw Error("Stream can only be iterated once");return this.started=!0,this}next(){if(this.queue.length>0)return Promise.resolve({done:!1,value:this.queue.shift()});if(this.isDone)return Promise.resolve({done:!0,value:void 0});if(this.hasError)return Promise.reject(this.hasError);return new Promise((Q,X)=>{this.readResolve=Q,this.readReject=X})}enqueue(Q){if(this.readResolve){let X=this.readResolve;this.readResolve=void 0,this.readReject=void 0,X({done:!1,value:Q})}else this.queue.push(Q)}done(){if(this.isDone=!0,this.readResolve){let Q=this.readResolve;this.readResolve=void 0,this.readReject=void 0,Q({done:!0,value:void 0})}}error(Q){if(this.hasError=Q,this.readReject){let X=this.readReject;this.readResolve=void 0,this.readReject=void 0,X(Q)}}return(){if(this.isDone=!0,this.returned)this.returned();return Promise.resolve({done:!0,value:void 0})}}class PQ{sendMcpMessage;isClosed=!1;constructor(Q){this.sendMcpMessage=Q}onclose;onerror;onmessage;async start(){}async send(Q){if(this.isClosed)throw Error("Transport is closed");this.sendMcpMessage(Q)}async close(){if(this.isClosed)return;this.isClosed=!0,this.onclose?.()}}import{randomUUID as gq}from"crypto";class K4{transport;isSingleUserTurn;canUseTool;hooks;abortController;jsonSchema;initConfig;pendingControlResponses=new Map;cleanupPerformed=!1;sdkMessages;inputStream=new z4;initialization;cancelControllers=new Map;hookCallbacks=new Map;nextCallbackId=0;sdkMcpTransports=new Map;sdkMcpServerInstances=new Map;pendingMcpResponses=new Map;firstResultReceivedResolve;firstResultReceived=!1;hasBidirectionalNeeds(){return this.sdkMcpTransports.size>0||this.hooks!==void 0&&Object.keys(this.hooks).length>0||this.canUseTool!==void 0}constructor(Q,X,Y,$,W,J=new Map,G,H){this.transport=Q;this.isSingleUserTurn=X;this.canUseTool=Y;this.hooks=$;this.abortController=W;this.jsonSchema=G;this.initConfig=H;for(let[B,z]of J)this.connectSdkMcpServer(B,z);this.sdkMessages=this.readSdkMessages(),this.readMessages(),this.initialization=this.initialize(),this.initialization.catch(()=>{})}setError(Q){this.inputStream.error(Q)}async stopTask(Q){await this.request({subtype:"stop_task",task_id:Q})}close(){this.cleanup()}cleanup(Q){if(this.cleanupPerformed)return;this.cleanupPerformed=!0;try{this.transport.close();let X=Error("Query closed before response received");for(let{reject:Y}of this.pendingControlResponses.values())Y(X);this.pendingControlResponses.clear();for(let{reject:Y}of this.pendingMcpResponses.values())Y(X);this.pendingMcpResponses.clear(),this.cancelControllers.clear(),this.hookCallbacks.clear();for(let Y of this.sdkMcpTransports.values())try{Y.close()}catch{}if(this.sdkMcpTransports.clear(),Q)this.inputStream.error(Q);else this.inputStream.done()}catch(X){}}next(...[Q]){return this.sdkMessages.next(...[Q])}return(Q){return this.sdkMessages.return(Q)}throw(Q){return this.sdkMessages.throw(Q)}[Symbol.asyncIterator](){return this.sdkMessages}[Symbol.asyncDispose](){return this.sdkMessages[Symbol.asyncDispose]()}async readMessages(){try{for await(let Q of this.transport.readMessages()){if(Q.type==="control_response"){let X=this.pendingControlResponses.get(Q.response.request_id);if(X)X.handler(Q.response);continue}else if(Q.type==="control_request"){this.handleControlRequest(Q);continue}else if(Q.type==="control_cancel_request"){this.handleControlCancelRequest(Q);continue}else if(Q.type==="keep_alive")continue;if(Q.type==="streamlined_text"||Q.type==="streamlined_tool_use_summary")continue;if(Q.type==="result"){if(this.firstResultReceived=!0,this.firstResultReceivedResolve)this.firstResultReceivedResolve();if(this.isSingleUserTurn)L1("[Query.readMessages] First result received for single-turn query, closing stdin"),this.transport.endInput()}this.inputStream.enqueue(Q)}if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.done(),this.cleanup()}catch(Q){if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.error(Q),this.cleanup(Q)}}async handleControlRequest(Q){let X=new AbortController;this.cancelControllers.set(Q.request_id,X);try{let Y=await this.processControlRequest(Q,X.signal),$={type:"control_response",response:{subtype:"success",request_id:Q.request_id,response:Y}};await Promise.resolve(this.transport.write(Z0($)+`']
```
2026-02-27T17:46:22.6254833Z go: downloading golang.org/x/net v0.36.0
2026-02-27T17:46:22.6256118Z go: downloading github.com/prometheus/client_golang v1.19.1
2026-02-27T17:46:22.6776130Z go: downloading github.com/beorn7/perks v1.0.1
2026-02-27T17:46:22.6777349Z go: downloading github.com/cespare/xxhash/v2 v2.2.0
2026-02-27T17:46:22.6791381Z go: downloading github.com/prometheus/client_model v0.5.0
2026-02-27T17:46:22.6885082Z go: downloading github.com/prometheus/common v0.48.0
2026-02-27T17:46:22.6899924Z go: downloading github.com/prometheus/procfs v0.12.0
2026-02-27T17:46:22.7318145Z go: downloading google.golang.org/protobuf v1.33.0
2026-02-27T17:46:22.9487979Z go: downloading golang.org/x/sys v0.30.0
2026-02-27T17:46:22.9491012Z go: downloading golang.org/x/text v0.22.0
2026-02-27T17:46:39.4958011Z Post job cleanup.
2026-02-27T17:46:39.5919166Z [command]/usr/bin/git version
2026-02-27T17:46:39.5956420Z git version 2.53.0
2026-02-27T17:46:39.6001673Z Temporarily overriding HOME='/home/runner/work/_temp/da44f7ff-83c6-412e-9e67-0b8652085053' before making global git config changes
2026-02-27T17:46:39.6003638Z Adding repository directory to the temporary git global config as a safe directory
2026-02-27T17:46:39.6008854Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/aiquila/aiquila
2026-02-27T17:46:39.6047158Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-27T17:46:39.6081431Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-27T17:46:39.6376766Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-27T17:46:39.6401029Z http.https://github.com/.extraheader
2026-02-27T17:46:39.6414872Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-27T17:46:39.6448133Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-27T17:46:39.6684482Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-27T17:46:39.6717813Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-27T17:46:39.7073248Z Cleaning up orphan processes
```

**Fixture:** `8a4ff1fb8bcba3fe7c2e...` conf=0.46
**Evidence:** ['<timestamp>.7205370Z `;_q(X,$)}function Z5(Q,X){let Y={...Q};if(X){let $={sandbox:X};if(Y.settings)try{$={...C6(Y.settings),sandbox:X}}catch{}Y.settings=Z0($)}return Y}class B4{options;process;processStdin;processStdout;ready=!1;abortController;exitError;exitListeners=[];processExitHandler;abortHandler;constructor(Q){this.options=Q;this.abortController=Q.abortController||R6(),this.initialize()}getDefaultExecutable(){return _6()?"bun":"node"}spawnLocalProcess(Q){let{command:X,args:Y,cwd:$,env:W,signal:J}=Q,G=W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr?"pipe":"ignore",H=Tq(X,Y,{cwd:$,stdio:["pipe","pipe",G],signal:J,env:W,windowsHide:!0});if(W.DEBUG_CLAUDE_AGENT_SDK||this.options.stderr)H.stderr.on("data",(z)=>{let K=z.toString();if(N1(K),this.options.stderr)this.options.stderr(K)});return{stdin:H.stdin,stdout:H.stdout,get killed(){return H.killed},get exitCode(){return H.exitCode},kill:H.kill.bind(H),on:H.on.bind(H),once:H.once.bind(H),off:H.off.bind(H)}}initialize(){try{let{additionalDirectories:Q=[],agent:X,betas:Y,cwd:$,executable:W=this.getDefaultExecutable(),executableArgs:J=[],extraArgs:G={},pathToClaudeCodeExecutable:H,env:B={...process.env},thinkingConfig:z,maxTurns:K,maxBudgetUsd:q,model:U,fallbackModel:V,jsonSchema:L,permissionMode:F,allowDangerouslySkipPermissions:w,permissionPromptToolName:D,continueConversation:j,resume:I,settingSources:T,allowedTools:v=[],disallowedTools:N0=[],tools:O0,mcpServers:c0,strictMcpConfig:t1,canUseTool:q1,includePartialMessages:a1,plugins:P1,sandbox:s1}=this.options,u=["--output-format","stream-json","--verbose","--input-format","stream-json"];if(z)switch(z.type){case"enabled":if(z.budgetTokens===void 0)u.push("--thinking","adaptive");else u.push("--max-thinking-tokens",z.budgetTokens.toString());break;case"disabled":u.push("--thinking","disabled");break;case"adaptive":u.push("--thinking","adaptive");break}if(this.options.effort)u.push("--effort",this.options.effort);if(K)u.push("--max-turns",K.toString());if(q!==void 0)u.push("--max-budget-usd",q.toString());if(U)u.push("--model",U);if(X)u.push("--agent",X);if(Y&&Y.length>0)u.push("--betas",Y.join(","));if(L)u.push("--json-schema",Z0(L));if(this.options.debugFile)u.push("--debug-file",this.options.debugFile);else if(this.options.debug)u.push("--debug");if(B.DEBUG_CLAUDE_AGENT_SDK)u.push("--debug-to-stderr");if(q1){if(D)throw Error("canUseTool callback cannot be used with permissionPromptToolName. Please use one or the other.");u.push("--permission-prompt-tool","stdio")}else if(D)u.push("--permission-prompt-tool",D);if(j)u.push("--continue");if(I)u.push("--resume",I);if(v.length>0)u.push("--allowedTools",v.join(","));if(N0.length>0)u.push("--disallowedTools",N0.join(","));if(O0!==void 0)if(Array.isArray(O0))if(O0.length===0)u.push("--tools","");else u.push("--tools",O0.join(","));else u.push("--tools","default");if(c0&&Object.keys(c0).length>0)u.push("--mcp-config",Z0({mcpServers:c0}));if(T)u.push("--setting-sources",T.join(","));if(t1)u.push("--strict-mcp-config");if(F)u.push("--permission-mode",F);if(w)u.push("--allow-dangerously-skip-permissions");if(V){if(U&&V===U)throw Error("Fallback model cannot be the same as the main model. Please specify a different model for fallbackModel option.");u.push("--fallback-model",V)}if(a1)u.push("--include-partial-messages");for(let b0 of Q)u.push("--add-dir",b0);if(P1&&P1.length>0)for(let b0 of P1)if(b0.type==="local")u.push("--plugin-dir",b0.path);else throw Error(`Unsupported plugin type: ${b0.type}`);if(this.options.forkSession)u.push("--fork-session");if(this.options.resumeSessionAt)u.push("--resume-session-at",this.options.resumeSessionAt);if(this.options.sessionId)u.push("--session-id",this.options.sessionId);if(this.options.persistSession===!1)u.push("--no-session-persistence");let AQ=Z5(G??{},s1);for(let[b0,Z1]of Object.entries(AQ))if(Z1===null)u.push(`--${b0}`);else u.push(`--${b0}`,Z1);if(!B.CLAUDE_CODE_ENTRYPOINT)B.CLAUDE_CODE_ENTRYPOINT="sdk-ts";if(delete B.NODE_OPTIONS,B.DEBUG_CLAUDE_AGENT_SDK)B.DEBUG="1";else delete B.DEBUG;let e1=yq(H),J4=e1?H:W,Q6=e1?[...J,...u]:[...J,H,...u],A9={command:J4,args:Q6,cwd:$,env:B,signal:this.abortController.signal};if(this.options.spawnClaudeCodeProcess)N1(`Spawning Claude Code (custom): ${J4} ${Q6.join(" ")}`),this.process=this.options.spawnClaudeCodeProcess(A9);else{if(!k1().existsSync(H)){let Z1=e1?`Claude Code native binary not found at ${H}. Please ensure Claude Code is installed via native installer or specify a valid path with options.pathToClaudeCodeExecutable.`:`Claude Code executable not found at ${H}. Is options.pathToClaudeCodeExecutable set?`;throw ReferenceError(Z1)}N1(`Spawning Claude Code: ${J4} ${Q6.join(" ")}`),this.process=this.spawnLocalProcess(A9)}this.processStdin=this.process.stdin,this.processStdout=this.process.stdout;let j9=()=>{if(this.process&&!this.process.killed)this.process.kill("SIGTERM")};this.processExitHandler=j9,this.abortHandler=j9,process.on("exit",this.processExitHandler),this.abortController.signal.addEventListener("abort",this.abortHandler),this.process.on("error",(b0)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else this.exitError=Error(`Failed to spawn Claude Code process: ${b0.message}`),N1(this.exitError.message)}),this.process.on("exit",(b0,Z1)=>{if(this.ready=!1,this.abortController.signal.aborted)this.exitError=new F1("Claude Code process aborted by user");else{let w6=this.getProcessExitError(b0,Z1);if(w6)this.exitError=w6,N1(w6.message)}}),this.ready=!0}catch(Q){throw this.ready=!1,Q}}getProcessExitError(Q,X){if(Q!==0&&Q!==null)return Error(`Claude Code process exited with code ${Q}`);else if(X)return Error(`Claude Code process terminated by signal ${X}`);return}write(Q){if(this.abortController.signal.aborted)throw new F1("Operation aborted");if(!this.ready||!this.processStdin)throw Error("ProcessTransport is not ready for writing");if(this.process?.killed||this.process?.exitCode!==null)throw Error("Cannot write to terminated process");if(this.exitError)throw Error(`Cannot write to process that exited with error: ${this.exitError.message}`);N1(`[ProcessTransport] Writing to stdin: ${Q.substring(0,100)}`);try{if(!this.processStdin.write(Q))N1("[ProcessTransport] Write buffer full, data queued")}catch(X){throw this.ready=!1,Error(`Failed to write to process stdin: ${X.message}`)}}close(){if(this.processStdin)this.processStdin.end(),this.processStdin=void 0;if(this.abortHandler)this.abortController.signal.removeEventListener("abort",this.abortHandler),this.abortHandler=void 0;for(let{handler:Q}of this.exitListeners)this.process?.off("exit",Q);if(this.exitListeners=[],this.process&&!this.process.killed&&this.process.exitCode===null)this.process.kill("SIGTERM"),setTimeout(()=>{if(this.process&&!this.process.killed)this.process.kill("SIGKILL")},5000).unref();if(this.ready=!1,this.processExitHandler)process.off("exit",this.processExitHandler),this.processExitHandler=void 0}isReady(){return this.ready}async*readMessages(){if(!this.processStdout)throw Error("ProcessTransport output stream not available");let Q=xq({input:this.processStdout});try{for await(let X of Q)if(X.trim())try{yield C6(X)}catch(Y){throw N1(`Non-JSON stdout: ${X}`),Error(`CLI output was not valid JSON. This may indicate an error during startup. Output: ${X.slice(0,200)}${X.length>200?"...":""}`)}await this.waitForExit()}catch(X){throw X}finally{Q.close()}}endInput(){if(this.processStdin)this.processStdin.end()}getInputStream(){return this.processStdin}onExit(Q){if(!this.process)return()=>{};let X=(Y,$)=>{let W=this.getProcessExitError(Y,$);Q(W)};return this.process.on("exit",X),this.exitListeners.push({callback:Q,handler:X}),()=>{if(this.process)this.process.off("exit",X);let Y=this.exitListeners.findIndex(($)=>$.handler===X);if(Y!==-1)this.exitListeners.splice(Y,1)}}async waitForExit(){if(!this.process){if(this.exitError)throw this.exitError;return}if(this.process.exitCode!==null||this.process.killed){if(this.exitError)throw this.exitError;return}return new Promise((Q,X)=>{let Y=(W,J)=>{if(this.abortController.signal.aborted){X(new F1("Operation aborted"));return}let G=this.getProcessExitError(W,J);if(G)X(G);else Q()};this.process.once("exit",Y);let $=(W)=>{this.process.off("exit",Y),X(W)};this.process.once("error",$),this.process.once("exit",()=>{this.process.off("error",$)})})}}function yq(Q){return![".js",".mjs",".tsx",".ts",".jsx"].some((Y)=>Q.endsWith(Y))}class z4{returned;queue=[];readResolve;readReject;isDone=!1;hasError;started=!1;constructor(Q){this.returned=Q}[Symbol.asyncIterator](){if(this.started)throw Error("Stream can only be iterated once");return this.started=!0,this}next(){if(this.queue.length>0)return Promise.resolve({done:!1,value:this.queue.shift()});if(this.isDone)return Promise.resolve({done:!0,value:void 0});if(this.hasError)return Promise.reject(this.hasError);return new Promise((Q,X)=>{this.readResolve=Q,this.readReject=X})}enqueue(Q){if(this.readResolve){let X=this.readResolve;this.readResolve=void 0,this.readReject=void 0,X({done:!1,value:Q})}else this.queue.push(Q)}done(){if(this.isDone=!0,this.readResolve){let Q=this.readResolve;this.readResolve=void 0,this.readReject=void 0,Q({done:!0,value:void 0})}}error(Q){if(this.hasError=Q,this.readReject){let X=this.readReject;this.readResolve=void 0,this.readReject=void 0,X(Q)}}return(){if(this.isDone=!0,this.returned)this.returned();return Promise.resolve({done:!0,value:void 0})}}class PQ{sendMcpMessage;isClosed=!1;constructor(Q){this.sendMcpMessage=Q}onclose;onerror;onmessage;async start(){}async send(Q){if(this.isClosed)throw Error("Transport is closed");this.sendMcpMessage(Q)}async close(){if(this.isClosed)return;this.isClosed=!0,this.onclose?.()}}import{randomUUID as gq}from"crypto";class K4{transport;isSingleUserTurn;canUseTool;hooks;abortController;jsonSchema;initConfig;pendingControlResponses=new Map;cleanupPerformed=!1;sdkMessages;inputStream=new z4;initialization;cancelControllers=new Map;hookCallbacks=new Map;nextCallbackId=0;sdkMcpTransports=new Map;sdkMcpServerInstances=new Map;pendingMcpResponses=new Map;firstResultReceivedResolve;firstResultReceived=!1;hasBidirectionalNeeds(){return this.sdkMcpTransports.size>0||this.hooks!==void 0&&Object.keys(this.hooks).length>0||this.canUseTool!==void 0}constructor(Q,X,Y,$,W,J=new Map,G,H){this.transport=Q;this.isSingleUserTurn=X;this.canUseTool=Y;this.hooks=$;this.abortController=W;this.jsonSchema=G;this.initConfig=H;for(let[B,z]of J)this.connectSdkMcpServer(B,z);this.sdkMessages=this.readSdkMessages(),this.readMessages(),this.initialization=this.initialize(),this.initialization.catch(()=>{})}setError(Q){this.inputStream.error(Q)}async stopTask(Q){await this.request({subtype:"stop_task",task_id:Q})}close(){this.cleanup()}cleanup(Q){if(this.cleanupPerformed)return;this.cleanupPerformed=!0;try{this.transport.close();let X=Error("Query closed before response received");for(let{reject:Y}of this.pendingControlResponses.values())Y(X);this.pendingControlResponses.clear();for(let{reject:Y}of this.pendingMcpResponses.values())Y(X);this.pendingMcpResponses.clear(),this.cancelControllers.clear(),this.hookCallbacks.clear();for(let Y of this.sdkMcpTransports.values())try{Y.close()}catch{}if(this.sdkMcpTransports.clear(),Q)this.inputStream.error(Q);else this.inputStream.done()}catch(X){}}next(...[Q]){return this.sdkMessages.next(...[Q])}return(Q){return this.sdkMessages.return(Q)}throw(Q){return this.sdkMessages.throw(Q)}[Symbol.asyncIterator](){return this.sdkMessages}[Symbol.asyncDispose](){return this.sdkMessages[Symbol.asyncDispose]()}async readMessages(){try{for await(let Q of this.transport.readMessages()){if(Q.type==="control_response"){let X=this.pendingControlResponses.get(Q.response.request_id);if(X)X.handler(Q.response);continue}else if(Q.type==="control_request"){this.handleControlRequest(Q);continue}else if(Q.type==="control_cancel_request"){this.handleControlCancelRequest(Q);continue}else if(Q.type==="keep_alive")continue;if(Q.type==="streamlined_text"||Q.type==="streamlined_tool_use_summary")continue;if(Q.type==="result"){if(this.firstResultReceived=!0,this.firstResultReceivedResolve)this.firstResultReceivedResolve();if(this.isSingleUserTurn)L1("[Query.readMessages] First result received for single-turn query, closing stdin"),this.transport.endInput()}this.inputStream.enqueue(Q)}if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.done(),this.cleanup()}catch(Q){if(this.firstResultReceivedResolve)this.firstResultReceivedResolve();this.inputStream.error(Q),this.cleanup(Q)}}async handleControlRequest(Q){let X=new AbortController;this.cancelControllers.set(Q.request_id,X);try{let Y=await this.processControlRequest(Q,X.signal),$={type:"control_response",response:{subtype:"success",request_id:Q.request_id,response:Y}};await Promise.resolve(this.transport.write(Z0($)+`']
```
2026-02-27T18:17:04.9930627Z go: downloading golang.org/x/net v0.36.0
2026-02-27T18:17:05.0159574Z go: downloading github.com/kr/fs v0.1.0
2026-02-27T18:17:05.0779172Z go: downloading github.com/beorn7/perks v1.0.1
2026-02-27T18:17:05.0788219Z go: downloading github.com/cespare/xxhash/v2 v2.2.0
2026-02-27T18:17:05.0794821Z go: downloading github.com/prometheus/client_model v0.5.0
2026-02-27T18:17:05.1141499Z go: downloading github.com/prometheus/common v0.48.0
2026-02-27T18:17:05.1162067Z go: downloading github.com/prometheus/procfs v0.12.0
2026-02-27T18:17:05.1710825Z go: downloading google.golang.org/protobuf v1.33.0
2026-02-27T18:17:05.3871932Z go: downloading golang.org/x/sys v0.30.0
2026-02-27T18:17:05.3872812Z go: downloading golang.org/x/text v0.22.0
2026-02-27T18:17:21.7206377Z Post job cleanup.
2026-02-27T18:17:21.8155306Z [command]/usr/bin/git version
2026-02-27T18:17:21.8194209Z git version 2.53.0
2026-02-27T18:17:21.8239438Z Temporarily overriding HOME='/home/runner/work/_temp/f6f51270-93cd-4edd-b8a3-8d7d27811be6' before making global git config changes
2026-02-27T18:17:21.8241530Z Adding repository directory to the temporary git global config as a safe directory
2026-02-27T18:17:21.8247212Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/aiquila/aiquila
2026-02-27T18:17:21.8286303Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-02-27T18:17:21.8321704Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-02-27T18:17:21.8564284Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-02-27T18:17:21.8585431Z http.https://github.com/.extraheader
2026-02-27T18:17:21.8598296Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-02-27T18:17:21.8629732Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-02-27T18:17:21.8866340Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-02-27T18:17:21.8897808Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-02-27T18:17:21.9259506Z Cleaning up orphan processes
```

### `terraform-state-lock` — 5 matches, 5 low-conf (100%)

**Fixture:** `ccdcbe4bfee227fffcd6...` conf=0.42
**Evidence:** ['<timestamp>.7341208Z Acquiring state lock. This may take a few moments...', '<timestamp>.9058506Z │ Error: Error acquiring the state lock']
```
2026-03-23T03:35:59.9065646Z [31m│[0m [0m
2026-03-23T03:35:59.9066105Z [31m│[0m [0mTerraform acquires a state lock to protect the state from being written
2026-03-23T03:35:59.9066998Z [31m│[0m [0mby multiple users at the same time. Please resolve the issue above and try
2026-03-23T03:35:59.9067675Z [31m│[0m [0magain. For most commands, you can disable locking with the "-lock=false"
2026-03-23T03:35:59.9068216Z [31m│[0m [0mflag, but this is not recommended.
2026-03-23T03:35:59.9068573Z [31m╵[0m[0m
2026-03-23T03:35:59.9120683Z ##[error]Terraform exited with code 1.
2026-03-23T03:35:59.9143454Z ##[error]Process completed with exit code 1.
2026-03-23T03:35:59.9248810Z Post job cleanup.
2026-03-23T03:36:00.0353018Z Post job cleanup.
2026-03-23T03:36:00.1142914Z [command]/usr/bin/git version
2026-03-23T03:36:00.1253851Z git version 2.53.0
2026-03-23T03:36:00.1263797Z Temporarily overriding HOME='/home/runner/work/_temp/3727bcdf-9d5b-4b54-96a7-55f41537b9b5' before making global git config changes
2026-03-23T03:36:00.1264944Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T03:36:00.1265776Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/core/core
2026-03-23T03:36:00.1267740Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-23T03:36:00.1297086Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-23T03:36:00.1500566Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T03:36:00.1522872Z http.https://github.com/.extraheader
2026-03-23T03:36:00.1533005Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-23T03:36:00.1563346Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-23T03:36:00.1763295Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T03:36:00.1792761Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-23T03:36:00.2093975Z Cleaning up orphan processes
2026-03-23T03:36:00.2295276Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: aws-actions/configure-aws-credentials@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `e41f1e6ff3b9b9daee26...` conf=0.42
**Evidence:** ['<timestamp>.7825613Z Acquiring state lock. This may take a few moments...', '<timestamp>.7842603Z │ Error: Error acquiring the state lock']
```
2026-03-23T12:51:24.7849127Z [31m│[0m [0m
2026-03-23T12:51:24.7849475Z [31m│[0m [0mTerraform acquires a state lock to protect the state from being written
2026-03-23T12:51:24.7850029Z [31m│[0m [0mby multiple users at the same time. Please resolve the issue above and try
2026-03-23T12:51:24.7850634Z [31m│[0m [0magain. For most commands, you can disable locking with the "-lock=false"
2026-03-23T12:51:24.7851059Z [31m│[0m [0mflag, but this is not recommended.
2026-03-23T12:51:24.7851339Z [31m╵[0m[0m
2026-03-23T12:51:24.7908702Z ##[error]Terraform exited with code 1.
2026-03-23T12:51:24.7927468Z ##[error]Process completed with exit code 1.
2026-03-23T12:51:24.7998713Z Post job cleanup.
2026-03-23T12:51:24.9032416Z Post job cleanup.
2026-03-23T12:51:24.9761215Z [command]/usr/bin/git version
2026-03-23T12:51:24.9793443Z git version 2.53.0
2026-03-23T12:51:24.9826019Z Temporarily overriding HOME='/home/runner/work/_temp/2256a413-9acc-4830-b38e-bf437272c2f3' before making global git config changes
2026-03-23T12:51:24.9827370Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T12:51:24.9830613Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/core/core
2026-03-23T12:51:24.9862976Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-23T12:51:24.9891829Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-23T12:51:25.0071486Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T12:51:25.0091402Z http.https://github.com/.extraheader
2026-03-23T12:51:25.0099960Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-23T12:51:25.0127536Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-23T12:51:25.0304541Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T12:51:25.0331862Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-23T12:51:25.0608079Z Cleaning up orphan processes
2026-03-23T12:51:25.0834022Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: aws-actions/configure-aws-credentials@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```

**Fixture:** `79a5e15790ea9eacd9fd...` conf=0.42
**Evidence:** ['<timestamp>.9656351Z Acquiring state lock. This may take a few moments...', '<timestamp>.0687411Z │ Error: Error acquiring the state lock']
```
2026-03-23T13:07:38.0694112Z [31m│[0m [0m
2026-03-23T13:07:38.0694499Z [31m│[0m [0mTerraform acquires a state lock to protect the state from being written
2026-03-23T13:07:38.0695144Z [31m│[0m [0mby multiple users at the same time. Please resolve the issue above and try
2026-03-23T13:07:38.0695756Z [31m│[0m [0magain. For most commands, you can disable locking with the "-lock=false"
2026-03-23T13:07:38.0696229Z [31m│[0m [0mflag, but this is not recommended.
2026-03-23T13:07:38.0696531Z [31m╵[0m[0m
2026-03-23T13:07:38.0759234Z ##[error]Terraform exited with code 1.
2026-03-23T13:07:38.0790504Z ##[error]Process completed with exit code 1.
2026-03-23T13:07:38.0892506Z Post job cleanup.
2026-03-23T13:07:38.2032346Z Post job cleanup.
2026-03-23T13:07:38.2830948Z [command]/usr/bin/git version
2026-03-23T13:07:38.2900203Z git version 2.53.0
2026-03-23T13:07:38.2937951Z Temporarily overriding HOME='/home/runner/work/_temp/995b8605-6bad-4373-8202-15706d79e188' before making global git config changes
2026-03-23T13:07:38.2939308Z Adding repository directory to the temporary git global config as a safe directory
2026-03-23T13:07:38.2943661Z [command]/usr/bin/git config --global --add safe.directory /home/runner/work/core/core
2026-03-23T13:07:38.2981324Z [command]/usr/bin/git config --local --name-only --get-regexp core\.sshCommand
2026-03-23T13:07:38.3017689Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || :"
2026-03-23T13:07:38.3262044Z [command]/usr/bin/git config --local --name-only --get-regexp http\.https\:\/\/github\.com\/\.extraheader
2026-03-23T13:07:38.3285951Z http.https://github.com/.extraheader
2026-03-23T13:07:38.3296982Z [command]/usr/bin/git config --local --unset-all http.https://github.com/.extraheader
2026-03-23T13:07:38.3330795Z [command]/usr/bin/git submodule foreach --recursive sh -c "git config --local --name-only --get-regexp 'http\.https\:\/\/github\.com\/\.extraheader' && git config --local --unset-all 'http.https://github.com/.extraheader' || :"
2026-03-23T13:07:38.3566778Z [command]/usr/bin/git config --local --name-only --get-regexp ^includeIf\.gitdir:
2026-03-23T13:07:38.3598280Z [command]/usr/bin/git submodule foreach --recursive git config --local --show-origin --name-only --get-regexp remote.origin.url
2026-03-23T13:07:38.3952203Z Cleaning up orphan processes
2026-03-23T13:07:38.4352098Z ##[warning]Node.js 20 actions are deprecated. The following actions are running on Node.js 20 and may not work as expected: aws-actions/configure-aws-credentials@v5. Actions will be forced to run with Node.js 24 by default starting June 2nd, 2026. Please check if updated versions of these actions are available that support Node.js 24. To opt into Node.js 24 now, set the FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true environment variable on the runner or in your workflow file. Once Node.js 24 becomes the default, you can temporarily opt out by setting ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION=true. For more information see: https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/
```
