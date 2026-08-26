---
name: go-best-practice/cli/config
description: >-
  Persist CLI flag preferences in a tool-home config.json: explicit
  set/show, precedence, --no-config, and a gray notice: when a value
  comes from config.
---

# config — persist CLI preferences

Stop making users re-type the same flags. Preferences are **explicit
writes** into a small JSON file under the tool home—not silent
“last used” mutation on every run.

Shape inspired by `wrk` (`--set-config`) and `explain` (flat
`--show-config` + config-sourced `notice:`). Keep the recipe
product-agnostic.

## When to use

- A flag is repeated often (`--agent-runner`, default model, UX mode).
- Users need a one-shot bypass without deleting prefs (`--no-config`).
- Tests need an isolated config home via env.

## Policy

1. **Storage** — `$TOOL_HOME/config.json` (or a command-scoped home).
   Include a `version` field. **Merge-only** writes: change only keys
   implied by flags; preserve unknown top-level keys. Write atomically
   (temp file + rename) with indent + trailing newline.
2. **CLI surface (simple tools)**  
   - `--set-config` + preference flags and/or `--clear-*` (write-only;
     empty stdout on success)  
   - `--show-config` (pretty-print JSON; missing file → `{}`)  
   - `--no-config` (skip reading config for this normal run)  
   Action-heavy CLIs may nest show under `--set-config --show` (wrk);
   prefer a top-level `--show-config` when set-config is not already a
   mode dispatcher.
3. **Precedence** — explicit CLI flag → config (unless `--no-config`) →
   built-in default.
4. **Clear** — `--clear-<name>` under `--set-config` deletes that key.
   Do not treat empty `--flag=` as “clear” if empty also means “unset /
   do not change.”
5. **Corrupt JSON** — fail hard on read (`Error: parse config.json: …`).
   Do not silently fall back to built-ins.
6. **Mutual exclusion** — `--set-config` and `--show-config` conflict;
   neither takes a normal command message; clear flags require
   `--set-config`.
7. **Config-sourced notice** — when a preference value is taken from
   the file (CLI flag empty, not `--no-config`, config key non-empty),
   print one **stderr** line before the primary work:

   ```text
   notice: agent-runner=codex (from config)
   ```

   Color **only** the `notice:` prefix **gray** via `cli/color`
   (`ModeFromFlags`, `EnabledFor` on stderr). No ANSI in
   `--show-config` JSON. Wire `--color` / `--no-color` on the ask path
   that prints the notice.

## Precedence table

| Source | Wins when |
| ------ | --------- |
| CLI flag | non-empty / explicitly set |
| `config.json` | flag empty and not `--no-config` |
| Built-in default | both above empty / skipped |

## CLI examples

```text
$ mytool --set-config --agent-runner codex
$ mytool --show-config
{
  "version": 1,
  "agent_runner": "codex"
}

$ mytool "hello"
notice: agent-runner=codex (from config)
…answer on stdout…

$ mytool --agent-runner grok "hello"    # flag wins; no notice
$ mytool --no-config "hello"            # built-in; no notice

$ mytool --set-config --clear-agent-runner
$ mytool --set-config --show-config
Error: --set-config is mutually exclusive with --show-config

$ mytool --color --no-color "hello"
Error: --color and --no-color cannot be specified together
```

Help snippets:

```text
  --set-config              persist preference flags into config.json
  --show-config             pretty-print config.json (missing → {})
  --no-config               do not read config.json for this run
  --clear-agent-runner      with --set-config: remove agent_runner
  --color / --no-color      force ANSI on/off for human notices
```

## Test isolation

- Point config at a temp home via a dedicated env
  (e.g. `MYTOOL_DEBUG_CONFIG_HOME` or the tool’s existing config-home
  override).
- E2E / doctest leaves should set that env on `cmd.Env`, not mutate
  process-global env in a shared harness when that breaks parallelism.
- Assert the notice with `--no-color` (plain text) or `--color` (gray
  SGR `\x1b[90m` on `notice:` only). Pipes are non-TTY → Auto is off
  unless `--color`.

## Anti-patterns

| Avoid | Prefer |
| ----- | ------ |
| Rewrite full config on every set | Merge-only map write |
| Auto-save last flags on each run | Explicit `--set-config` |
| Share one config across CLIs with different ID spaces | Per-tool / per-command home |
| `warning:` for normal config apply | `notice:` (gray prefix) |
| Notice on stdout next to the answer / JSON | stderr only |
| Silent fallback on corrupt JSON | Fail hard |

## Out of scope

- Full application config schemas, YAML, multi-file layers
- Remote / synced settings
- Auto-detecting “last used” as preference

## See also

- `cli/color` — `--color` / `--no-color`, gray `notice:` prefix
- `flags-parsing` — bool/string flags and help text with less-flags
- `cli/dry-run` — one control flow; do not fork a second “config path”

Reveal with:

```bash
go-best-practice skill --show cli/config
```
