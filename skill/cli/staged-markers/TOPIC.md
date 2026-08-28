---
name: go-best-practice/cli/staged-markers
description: >-
  Stage-style CLI progress: flush-left [n/total] markers as the spine,
  kind-aligned indented detail between stages (stderr).
---

# staged-markers — `[n/total]` stage spine

When a Go CLI runs a **fixed multi-stage pipeline** (launch → work →
validate → ship → …), print **stage markers** so humans can skim
progress and tools can grep the outline. Markers are the spine;
optional detail logs nest under the current stage’s **kind** column.

Use this for a **known stage count** decided up front (possibly from
flags). For unbounded item streams, prefer `cli/streaming` (line-at-a-time
results), not a fake stage counter per item.

## Policy

| Concern | Default | Notes |
| ------- | ------- | ----- |
| Stream | **stderr** | Primary results stay on **stdout** |
| Marker | `[n/total] kind msg` | Flush left; never indent markers |
| `n` | 1-based | Stable for the run’s stage plan |
| `total` | Fixed before first marker | May depend on flags (e.g. 2 / 4 / 5) |
| `kind` | Short lowercase name | Pad with `%-12s` so msgs align |
| Same `n` twice | Allowed | Long stages: start intent, then `ok` / `skip` |
| Interleaving | Allowed | Detail / `notice:` between markers |
| Detail indent | **Kind column** | Leading spaces = `len("[%d/%d] ")` for open stage |
| Verbose | Opt-in | Default = spine; `-v` adds kind-aligned detail |
| Fatal errors | Flush-left `Error:` | Non-zero exit; keep prior markers |

## Marker format

```go
fmt.Fprintf(w, "[%d/%d] %-12s %s\n", n, total, kind, msg)
```

Typical `msg` fragments: start intent (`wait result.json`,
`commit+push+MR`), `ok`, `skip (dry-run)`, `skipped (reason)`.

```text
[1/5] launch       headless create-mr auto-merge
[2/5] agent        wait result.json
[2/5] agent        ok
```

### Compute `total` once

```go
func stageTotal(createMR, autoMerge bool) int {
    if !createMR {
        return 2
    }
    if autoMerge {
        return 5
    }
    return 4
}
```

Do not change `total` after the first marker. Skip remaining work with
`skipped (…)` on later stages rather than shrinking the denominator.

## Kind-aligned detail (between stages)

Detail lines must **not** reuse `[n/total]`. Indent so text starts at
the same column as `kind` (`launch`, `validate`, …):

```go
func progress(w io.Writer, n, total int, kind, msg string) {
    prefix := fmt.Sprintf("[%d/%d] ", n, total)
    fmt.Fprintf(w, "%s%-12s %s\n", prefix, kind, msg)
}

func stageNotice(w io.Writer, n, total int, format string, args ...any) {
    prefix := fmt.Sprintf("[%d/%d] ", n, total)
    indent := strings.Repeat(" ", len(prefix))
    fmt.Fprintf(w, indent+"notice: "+format+"\n", args...)
}
```

```text
[4/5] ship         commit+push+MR
      notice: ship: git commit -m "…"
      notice: ship: git push origin HEAD:…
[4/5] ship         ok
```

| Concern | Rule |
| ------- | ---- |
| Align to | Start of `kind` for the **open** stage’s `n` / `total` |
| Nest depth | One level under the current stage |
| Detail prefix | Prefer `notice:` for verbose micro-steps |
| Non-fatal `warning:` | Same kind-column indent |
| Fatal `Error:` | Flush left (not nested) |

When `total >= 10`, unpadded `%d` can shift the kind column by one
(`[9/12]` vs `[10/12]`). Accept that by default; optionally pad both
numbers to `digits(total)` (prefer space-pad over zero-pad) if a fixed
column matters for that tool.

Grep the spine with `^\[[0-9]+/` — markers stay flush left.

## Verbosity and dry-run

| Mode | stderr |
| ---- | ------ |
| Default | Stage markers (start / `ok` / `skip`) |
| Verbose | Same spine + kind-aligned `notice:` (and similar) between markers |
| Dry-run | Still emit every stage; msgs like `skip (dry-run)` (see `cli/dry-run`) |

Quiet mode must remain scannable as an outline; do not dump git/API
micro-steps without a verbose gate.

## CLI shape examples

**Success (markers + verbose detail on stderr; result on stdout):**

```text
$ mytool sink --create-mr --auto-merge -v
[1/5] launch       headless create-mr auto-merge
[2/5] agent        wait result.json
      notice: ping loop started
[2/5] agent        ok
[3/5] validate     result.json
      notice: ship: read/validate /path/result.json
[3/5] validate     ok
[4/5] ship         commit+push+MR
      notice: ship: git add -- topics/p.md
      notice: ship: git commit -m "docs(kb): …"
      notice: ship: git push origin HEAD:tester/…
[4/5] ship         ok
[5/5] merge        ok
ok  mr=https://example.com/merge_requests/12
```

**Warning (partial skip, exit 0):**

```text
$ mytool sink --create-mr
[1/4] launch       headless create-mr
[2/4] agent        ok
[3/4] validate     ok
[4/4] ship         skipped (no new knowledges)
warning: nothing to ship
```

**Error mid-pipeline (flush-left `Error:`, non-zero):**

```text
$ mytool sink --create-mr
[1/4] launch       headless create-mr
[2/4] agent        wait result.json
Error: propose agent: timeout
```

**Dry-run (full stage plan):**

```text
$ mytool sink --dry-run --create-mr --auto-merge
[1/5] launch       skip (dry-run) headless create-mr auto-merge
[2/5] agent        skip (dry-run)
[3/5] validate     skip (dry-run)
[4/5] ship         skip (dry-run)
[5/5] merge        skip (dry-run)
```

Color (when enabled via `cli/color`): yellow `warning:` and red
`Error:` on stderr; optional gray meta on markers. No ANSI in `--json`
or other machine-readable stdout.

## Anti-patterns

| Avoid | Prefer |
| ----- | ------ |
| Renumber `n` for every log line | One `n` per stage; detail under kind |
| Indent `[n/total]` lines | Markers flush left |
| Flush-left `notice:` between stages | Kind-column indent |
| Wrap every detail as `[4/5] ship …` | `notice:` (or plain) under kind |
| Shrink `total` when skipping later stages | Keep `total`; emit `skipped (…)` |
| `[i/N]` for every scanned file as “stages” | Stream results (`cli/streaming`) |

## Testing notes

| Goal | Approach |
| ---- | -------- |
| Marker shape | Capture stderr; exact `"%d/%d] %-12s"` lines for each stage |
| Kind indent | Assert detail lines start with `strings.Repeat(" ", len(fmt.Sprintf("[%d/%d] ", n, total)))` |
| Quiet vs verbose | Without `-v`, no micro-step `notice:`; with `-v`, notices appear between markers |
| Dry-run | Every stage present; msgs contain `skip (dry-run)` |
| Partial error | Prior markers kept; flush-left `Error:`; non-zero exit |

## Out of scope

- Spinner / progress-bar library recipes
- Interactive TUI stage redraw
- `--stream` flags (see `cli/streaming`)

## See also

- `cli/streaming` — progressive stdout; progress/diagnostics on stderr
- `cli/output-alignment` — measure/pad when building custom columns
- `cli/color` — ANSI gating; never color machine-readable output
- `cli/dry-run` — one pipeline; dry-run still walks the same stages

Reveal with:

```bash
go-best-practice skill --show cli/staged-markers
```
