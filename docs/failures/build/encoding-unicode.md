# Unicode or character encoding error

**Playbook ID:** `encoding-unicode`
**Category:** build
**Severity:** medium
**Tags:** `unicode`, `encoding`, `utf-8`, `ascii`, `charset`, `locale`

## What this failure means

A CI step failed because a file or data stream contains characters that
cannot be interpreted in the assumed encoding. This commonly manifests as a
Python `UnicodeDecodeError`, Java `MalformedInputException`, or a locale
warning turning a build tool error non-fatal.

## Common log signals

```text
UnicodeDecodeError
UnicodeEncodeError
codec can't decode
codec can't encode
invalid byte sequence
invalid multibyte character
invalid character in identifier
character encoding
```

## Diagnosis

Encoding failures occur when:
1. The CI runner's locale is `C` or `POSIX` rather than `en_US.UTF-8`,
   causing tools to default to ASCII
2. A source file, test fixture, or input data contains non-ASCII characters
3. A file was saved with Latin-1 or Windows-1252 encoding rather than UTF-8
4. A string conversion assumes ASCII but encounters an emoji, accented
   character, or non-Latin script

Check the runner's locale:

```bash
locale
echo $LANG $LC_ALL $LC_CTYPE
```

A minimal CI runner may show `LANG=C` or no locale set at all.

Find non-UTF-8 files:

```bash
find . -name "*.py" -exec file {} \; | grep -v "ASCII\|UTF-8"
```

## Fix steps

1. Set UTF-8 locale in CI:

   ```yaml
   # GitHub Actions / most CI
   env:
     LANG: en_US.UTF-8
     LC_ALL: en_US.UTF-8
     PYTHONIOENCODING: utf-8
   ```

2. In Python, open files with explicit encoding:

   ```python
   # GOOD
   with open("file.txt", encoding="utf-8") as f:
       data = f.read()

   # Python 3.15+ will warn on implicit encoding in open()
   # use PYTHONWARNDEFAULTENCODING=1 to audit existing code
   ```

3. Convert non-UTF-8 files to UTF-8:

   ```bash
   # Detect encoding
   file -i suspicious-file.txt
   chardet suspicious-file.txt     # requires chardet package

   # Convert Latin-1 to UTF-8
   iconv -f latin1 -t utf-8 input.txt > output.txt
   ```

4. In Java, specify charset explicitly:

   ```java
   Files.readString(path, StandardCharsets.UTF_8);
   new InputStreamReader(stream, StandardCharsets.UTF_8);
   ```

5. Ensure the database connection and ORM use UTF-8:

   ```
   # MySQL
   jdbc:mysql://host/db?characterEncoding=UTF-8

   # PostgreSQL: set client_encoding = 'UTF8' in the connection
   ```

## Validation

- Re-run the failing step after setting the locale.
- Confirm `locale` shows `LANG=en_US.UTF-8`.
- Run the failing test or build and confirm no codec errors.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain encoding-unicode
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Unicode or character encoding error
- Build: unicode or character encoding error
- invalid character in identifier
- GitHub Actions unicode or character encoding error
- faultline explain encoding-unicode


---

*Generated from [playbooks/bundled/log/runtime/encoding-unicode.yaml](../../../playbooks/bundled/log/runtime/encoding-unicode.yaml). Do not edit directly — run `make docs-generate`.*
