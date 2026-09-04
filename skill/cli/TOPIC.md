---
name: go-best-practice/cli
description: >-
  CLI UX (layout, output/*, dry-run, config, inline TUI mouse) and skill
  CLI packaging shapes. Load a child with:
  go-best-practice skill --show cli/<topic>
---

# cli — CLI UX and skill CLI packaging

Recipes for building Go CLIs: how to lay out packages, how output looks
and streams, interactive terminal UIs, persisted preferences, and how to
ship skill binaries that embed `SKILL.md` / nested `TOPIC.md` trees.

This is a **category index**. `project-layout` covers source layout;
`output/*` covers emit/render; `dry-run` and `config` are behavior /
prefs; `inline-tui-mouse` is mouse hit-testing for inline TUIs;
`skill-cli` is how to package skill CLIs. Flag parsing lives separately
under `flags-parsing`.

## Topics

- `project-layout` — thin main / `cmd`, `run` package, assets outside
  `cmd` (matches `kool create go-cli`)
- `output` — emit and render CLI text (category)
  - `streaming` — stream units as ready; stdout vs stderr; NDJSON
  - `color` — `--color` / `--no-color`, TTY auto, `NO_COLOR`
  - `staged-markers` — `[n/total]` spine (one per stage); kind-aligned detail
  - `alignment` — column measure/pad/truncate (rune width, ANSI-safe)
- `dry-run` — one pipeline with side-effect gates; avoid a separate
  dry-run function that duplicates logic
- `config` — persist flag preferences in tool-home `config.json`:
  `--set-config` / `--show-config` / `--no-config`, precedence, gray
  `notice:` when a value comes from config
- `skill-cli` — skill CLI shapes: single-skill, multi-skill host,
  topic discovery
- `inline-tui-mouse` — mouse hit-testing for inline (non-alt-screen)
  TUIs: view-local hitmaps, CSI 6n origin on one stdin path, dual-origin
  fallback, anti-patterns

## Retrieve

```bash
go-best-practice skill --show cli
go-best-practice skill --show cli/project-layout
go-best-practice skill --show cli/output
go-best-practice skill --show cli/output/streaming
go-best-practice skill --show cli/output/color
go-best-practice skill --show cli/output/staged-markers
go-best-practice skill --show cli/output/alignment
go-best-practice skill --show cli/dry-run
go-best-practice skill --show cli/config
go-best-practice skill --show cli/skill-cli
go-best-practice skill --show cli/inline-tui-mouse
go-best-practice skill cli/output/color --show
```

## See also

- `kool-create` — scaffolds including `go-cli`
- `flags-parsing` — less-flags and sub-command `--help` at every level
- `cmd-exec` — running external commands (inherit stdout/stderr)
