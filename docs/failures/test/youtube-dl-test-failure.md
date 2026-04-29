# youtube-dl / yt-dlp test failure due to unavailable video

**Playbook ID:** `youtube-dl-test-failure`
**Category:** test
**Severity:** low
**Tags:** `youtube-dl`, `yt-dlp`, `video`, `external-service`, `network`, `flaky`

## What this failure means

A youtube-dl or yt-dlp test case failed because the target video or page was
unavailable at the time the test ran. This is an external-service failure, not
a code defect. Common causes include videos deleted or made private, HTTP 403/404
responses from video platforms, URLs that are geo-restricted, or the platform
changing its embed format so the extractor can no longer parse it.

## Common log signals

```text
yt-dl.org
yt-dlp.org
unable to download webpage: timed out
unable to download video data: timed out
unable to download video data: HTTP Error 403
unable to download video data: HTTP Error 404
unable to download video data: HTTP Error 401
unable to download video data: HTTP Error 429
```

## Diagnosis

The test suite calls a live video URL during CI. The video host returned an error
(HTTP 4xx/5xx, redirect loop, timeout, or "unsupported URL") causing youtube-dl
to raise an exception and the test to fail.

Typical error messages include:

- `Unable to extract <field>; please report this issue on https://yt-dl.org/bug`
- `unable to download video data: HTTP Error 403: Forbidden`
- `unable to download video data: HTTP Error 404: Not Found`
- `unable to download video data: HTTP Error 401: Unauthorized`
- `unable to download video data: HTTP Error 429: Too Many Requests`
- `Unsupported URL: <url>; please report this issue on https://yt-dl.org/bug`
- `Unable to download webpage: timed out`
- `YouTube said: This video is no longer available`
- `No video formats found; please report this issue on https://yt-dl.org/bug`

## Fix steps

1. Find the failing test case name (e.g. `test_YouTube_1`) in the output.
2. Check whether the test video is still accessible from the command line:
   ```
   youtube-dl -v <url>
   ```
3. If the video is gone or geo-blocked, replace the test URL with a stable
   alternative or mark the test as an expected failure / skip for that URL.
4. If the extractor logic changed upstream, update to the latest youtube-dl
   release:
   ```
   pip install --upgrade youtube-dl
   # or
   pip install --upgrade yt-dlp
   ```
5. For persistent failures against a specific platform, consider mocking the
   HTTP layer in CI so tests do not depend on live network access.

## Validation

- Re-run the specific failing test: `python -m pytest test/test_download.py -k test_YouTube_1 -v`
- Confirm the test passes or is correctly skipped.
- No other test_download tests regress.

## Likely files to inspect

- `test/test_download.py`
- `test/test_networking.py`
- `test/test_utils.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain youtube-dl-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- youtube-dl / yt-dlp test failure due to unavailable video
- Test: youtube-dl / yt-dlp test failure due to unavailable video
- unable to download video data: HTTP Error 403
- faultline explain youtube-dl-test-failure


---

*Generated from [playbooks/bundled/log/test/youtube-dl-test-failure.yaml](../../playbooks/bundled/log/test/youtube-dl-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
