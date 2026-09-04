---
name: go-best-practice/cli/output
description: >-
  How CLI text is emitted and rendered: streaming, color, staged-markers,
  and column alignment. Load a child with:
  go-best-practice skill --show cli/output/<topic>
---

# output — emit and render CLI text

Recipes for **what appears on the terminal**: when to print, which stream,
ANSI color, stage spines, and column alignment. Behavior gates (dry-run),
config, layout, and skill packaging stay under `cli/` (not here).

This is a **category index**. Load a child topic for the full recipe.

## Topics

- `streaming` — stream units as they are ready; stdout results vs stderr
  progress; when buffering is OK; NDJSON for long lists
- `color` — `--color` / `--no-color`, TTY auto, `NO_COLOR`; never color
  machine-readable output
- `staged-markers` — fixed multi-stage `[n/total]` spine on stderr (one
  marker per stage); kind-aligned detail under the open stage
- `alignment` — measure/pad/truncate columns (rune width, ANSI-safe)

## Retrieve

```bash
go-best-practice skill --show cli/output
go-best-practice skill --show cli/output/streaming
go-best-practice skill --show cli/output/color
go-best-practice skill --show cli/output/staged-markers
go-best-practice skill --show cli/output/alignment
go-best-practice skill cli/output/alignment --show
```

## See also

- `cli/dry-run` — one pipeline; probes run; gate mutations (`would:` lines)
- `cli` — parent CLI index (layout, config, skill-cli, TUI mouse)
- `cmd-exec` — external commands inherit stdout/stderr
