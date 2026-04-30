# Ansible YAML task or variable file syntax error

**Playbook ID:** `ansible-yaml-syntax-error`
**Category:** build
**Severity:** high
**Tags:** `ansible`, `yaml`, `infrastructure`, `configuration`, `iac`, `playbook`, `roles`

## What this failure means

Ansible failed to parse a YAML file — a task file, variable file, role
default, or handler — while loading the playbook. The CI pipeline cannot
continue until the YAML syntax is corrected.

## Common log signals

```text
Syntax Error while loading YAML script
ERROR! Syntax Error while loading YAML
```

## Diagnosis

Ansible's YAML loader emits this error when it encounters a syntax violation
in a task file, variables file, handler, or role default. The error message
identifies the exact file and, in newer Ansible versions, marks the offending
line with a caret (`^`):

```
Syntax Error while loading YAML script, /path/to/roles/<role>/tasks/main.yml

The offending line appears to be:
    key: value
     misindented: true
         ^ here
```

Common causes:

- **Incorrect indentation**: YAML uses spaces exclusively — tabs are not
  permitted and will cause a parse failure.
- **Incorrect alignment under a mapping key**: a value block that is not
  indented consistently relative to its parent key.
- **Missing or extra colons or dashes**: a bare colon inside an unquoted
  string, or a list item dash not followed by a space.
- **Unclosed quotes**: a single or double quote opened but never closed.
- **Duplicate keys**: two keys at the same level in a mapping block.
- **Merge conflict markers** left in the file: lines beginning with
  `<<<<<<<`, `=======`, or `>>>>>>>`.
- **Jinja2 template expressions in bare YAML**: a `{{ }}` block at the
  start of a value must be quoted.

## Fix steps

1. Identify the file and line from the error message.

2. Run Ansible's built-in syntax check on the playbook:

   ```bash
   ansible-playbook --syntax-check playbook.yml
   ```

3. Validate the offending file with the Python YAML parser:

   ```bash
   python3 -c "import yaml; yaml.safe_load(open('<file>').read()); print('OK')"
   ```

4. Use `yamllint` for detailed whitespace and structure feedback:

   ```bash
   yamllint <file>
   ```

5. Scan for tabs:

   ```bash
   grep -Pn "\t" <file>
   ```

6. Look for merge conflict markers:

   ```bash
   grep -n "^<<<\|^===\|^>>>" <file>
   ```

7. If the file uses Jinja2 templates, quote any value that starts with
   `{{`:

   ```yaml
   # BAD
   dest: {{ app_dir }}/app.conf

   # GOOD
   dest: "{{ app_dir }}/app.conf"
   ```

8. After fixing, re-run:

   ```bash
   ansible-playbook --syntax-check playbook.yml
   ```

## Validation

- `ansible-playbook --syntax-check playbook.yml` exits 0 with no errors.
- `yamllint roles/<role>/tasks/main.yml` reports no errors.
- The CI pipeline completes the Ansible run without YAML parse failures.

## Likely files to inspect

- `*.yml`
- `roles/*/tasks/main.yml`
- `roles/*/tasks/*.yml`
- `roles/*/vars/main.yml`
- `roles/*/defaults/main.yml`
- `roles/*/handlers/main.yml`
- `group_vars/*.yml`
- `host_vars/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ansible-yaml-syntax-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Ansible YAML task or variable file syntax error
- Build: ansible yaml task or variable file syntax error
- Syntax Error while loading YAML script
- GitHub Actions ansible yaml task or variable file syntax error
- faultline explain ansible-yaml-syntax-error


---

*Generated from [playbooks/bundled/log/build/ansible-yaml-syntax-error.yaml](../../../playbooks/bundled/log/build/ansible-yaml-syntax-error.yaml). Do not edit directly — run `make docs-generate`.*
