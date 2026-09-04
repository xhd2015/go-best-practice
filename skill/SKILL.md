---
name: go-best-practice
description: >-
  Index of Go best-practice recipes (kool create, external commands,
  CLI UX, less-flags, go:embed assets/version, RFC3339 time strings).
  Use when scaffolding a Go project or choosing a CLI, flag, dry-run,
  or embed pattern. Load a topic: go-best-practice skill --show <topic-path>
---

# Go Best Practice Skill

This skill is an **index**. Load a recipe with
`go-best-practice skill --show <topic>` (or
`go-best-practice skill <topic> --show`). Nested topics use a
slash-separated path (e.g. `flags-parsing/types`, `cli/dry-run`).

## Topics

- `kool-create` — scaffold new projects with `kool create` (go-cli,
  react, go-react, frontend, server, electron, …)
- `cmd-exec` — running external commands with
  `github.com/xhd2015/xgo/support/cmd` (Debug mode, output capture,
  env vars, directory, I/O redirect)
- `cli` — CLI UX, project layout, and skill CLI packaging
  - `project-layout` — thin main / `cmd`, `run` package, assets outside
    `cmd` (matches `kool create go-cli`)
  - `color` — terminal ANSI color: `--color` / `--no-color`, TTY
    auto, and the `NO_COLOR` env convention
  - `streaming` — stream CLI output as work proceeds; avoid
    buffering all output until the end (when to buffer, flush,
    NDJSON vs full JSON)
  - `staged-markers` — stage-style progress: flush-left `[n/total]`
    spine, kind-aligned indented detail between stages (stderr)
  - `output-alignment` — column pad/width/truncate; hand-rolled
    measure + pad (primary), rune-count width, ANSI-safe measure
  - `dry-run` — one pipeline with side-effect gates; avoid a
    separate dry-run function that duplicates logic
  - `config` — persist flag preferences in tool-home `config.json`:
    `--set-config` / `--show-config` / `--no-config`, precedence,
    gray `notice:` when a value comes from config
  - `skill-cli` — skill CLI shapes: single-skill, multi-skill host,
    topic discovery
  - `inline-tui-mouse` — mouse hit-testing for inline (non-alt-screen)
    TUIs: CSI 6n origin on one stdin path, dual-origin fallback,
    anti-patterns (sleep probes, parallel `/dev/tty` reads)
- `flags-parsing` — CLI flag parsing with
  `github.com/xhd2015/less-flags`
  - `types` — supported target types (`*bool`, `*string`, `*int`,
    `*time.Duration`, `*[]string`, `Cut`, and `**T` variants)
  - `subcommand` — sub-command dispatcher patterns (with `StopOnFirstArg` and no-toplevel-flags variants)
  - `cut` — cut flags: consume all remaining tokens after a marker
  - `collect` — `CollectParsedFlags` / `Flags.Reconstruct` / `Remove`
- `go-embed-assets` — ship generated UI/extension assets with
  `//go:embed`: placeholders so bare `go install` compiles, fat local
  bundle, and hydrate from version-pinned GitHub release archives
- `go-embed-version` — tracked `VERSION.txt` + `//go:embed`; stamp
  during build/release and always restore original bytes
- `time-string` — persist instants as RFC3339 strings with **local
  offset** (e.g. `+08:00`), not always `Z`

## Retrieve

```bash
# list paths
go-best-practice topics
go-best-practice skill --list

# this index
go-best-practice skill --show

# category / nested (slash path; both flag orders)
go-best-practice skill --show cli
go-best-practice skill --show cli/project-layout
go-best-practice skill --show cli/dry-run
go-best-practice skill --show cli/staged-markers
go-best-practice skill --show cli/output-alignment
go-best-practice skill --show cli/config
go-best-practice skill --show go-embed-version
go-best-practice skill --show time-string
go-best-practice skill flags-parsing/types --show

# YAML frontmatter only
go-best-practice skill --show --header
```

Every path under **Topics** loads the same way; use `skill --list` for a
flat inventory.
