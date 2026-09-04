---
name: go-best-practice/cli/output/staged-markers
description: >-
  Stage-style CLI progress: flush-left [n/total] markers (exactly one per
  stage) as the spine; kind-aligned detail under that marker (stderr);
  blank stderr line then flush-left stdout summary after all stages.
---

# staged-markers — `[n/total]` stage spine

When a Go CLI runs a **fixed multi-stage pipeline** (launch → work →
validate → ship → …), print **stage markers** so humans can skim
progress and tools can grep the outline. Markers are the spine;
optional detail logs nest under the current stage’s **kind** column.

Use this for a **known stage count** decided up front (possibly from
flags). For unbounded item streams, prefer `cli/output/streaming` (line-at-a-time
results), not a fake stage counter per item.

## Policy

| Concern | Default | Notes |
| ------- | ------- | ----- |
| Stream | **stderr** | Primary results stay on **stdout** |
| Marker | `[n/total] kind msg` | Flush left; never indent markers |
| `n` | 1-based | Stable for the run’s stage plan |
| `total` | Fixed before first marker | May depend on flags (e.g. 2 / 4 / 5) |
| `kind` | Short lowercase name | Pad with `%-12s` so msgs align |
| Markers per stage | **Exactly one** `[n/total]` | No start-then-`ok` second marker for the same `n` |
| Detail | Kind-aligned under open stage | `notice:`, `ok`, `would:`, evidence `skip:` / `skipped (…)` |
| Verbose | Opt-in | Default = spine; `-v` adds kind-aligned detail |
| Dry-run | Same probes as live | Gate mutations; kind-aligned `would:`; never stamp `skip (dry-run)` (see `cli/dry-run`) |
| Post-stage summary | Blank stderr line, then stdout | After the last stage output, print one blank line on **stderr**, then the flush-left product / summary on **stdout** (not kind-aligned). Keeps piped stdout clean while separating bands on a TTY |
| Fatal errors | Flush-left `Error:` | Non-zero exit; keep prior markers. **No** blank before mid-pipeline `Error:` (it aborts the open stage) |

## Marker format

```go
fmt.Fprintf(w, "[%d/%d] %-12s %s\n", n, total, kind, msg)
```

Typical `msg` fragments: stage intent (`wait result.json`,
`commit+push+MR`), or a short final status when there is no detail
(`ok`, `skipped (reason)`). Do **not** use `skip (dry-run)` — the user
already passed `--dry-run`.

```text
[1/5] launch       headless create-mr auto-merge
[2/5] agent        wait result.json
      ok
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

## Kind-aligned detail (under the open stage)

Detail lines must **not** reuse `[n/total]`. Indent so text starts at
the same column as `kind` (`launch`, `validate`, …):

```go
func progress(w io.Writer, n, total int, kind, msg string) {
    prefix := fmt.Sprintf("[%d/%d] ", n, total)
    fmt.Fprintf(w, "%s%-12s %s\n", prefix, kind, msg)
}

func stageDetail(w io.Writer, n, total int, format string, args ...any) {
    prefix := fmt.Sprintf("[%d/%d] ", n, total)
    indent := strings.Repeat(" ", len(prefix))
    fmt.Fprintf(w, indent+format+"\n", args...)
}
// stageDetail(w, n, total, "ok")
// stageDetail(w, n, total, "would: git push …")
// stageDetail(w, n, total, "notice: …")
```

```text
[4/5] ship         commit+push+MR
      notice: ship: git commit -m "…"
      notice: ship: git push origin HEAD:…
      ok
```

| Concern | Rule |
| ------- | ---- |
| Align to | Start of `kind` for the **open** stage’s `n` / `total` |
| Nest depth | One level under the current stage |
| Detail prefix | `notice:` (verbose micro-steps); `would:` (planned mutate); `ok` (stage finished after intent msg) |
| Non-fatal `warning:` | Same kind-column indent |
| Evidence no-op | `skip:` / `skipped (…)` with a real reason — not “because dry-run” |
| Fatal `Error:` | Flush left (not nested) |

When `total >= 10`, unpadded `%d` can shift the kind column by one
(`[9/12]` vs `[10/12]`). Accept that by default; optionally pad both
numbers to `digits(total)` (prefer space-pad over zero-pad) if a fixed
column matters for that tool.

Grep the spine with `^\[[0-9]+/` — markers stay flush left. A correct
run has **exactly `total`** matching lines.

## Post-stage summary

When the pipeline finishes and there is a **summary / product** line
(e.g. `seeded path@ver`, `ok  mr=…`):

1. Finish the last stage’s marker and any kind-aligned detail on stderr.
2. Print **one blank line on stderr** (`fmt.Fprintln(stderr)`).
3. Print the summary flush-left on **stdout** — not under the kind column.

Do not put the summary on stderr as kind-aligned detail. Do not omit the
blank when both streams are shown together on a TTY (the gap is what
distinguishes the summary from the spine). Mid-pipeline `Error:` stays
flush-left on stderr with **no** preceding blank.

## Verbosity and dry-run

| Mode | stderr |
| ---- | ------ |
| Default | One marker per stage (intent or short status as `msg`) |
| Verbose | Same spine + kind-aligned `notice:` / `ok` / `would:` under the open stage |
| Dry-run | Same spine and **real probes** as live; gate mutations; kind-aligned `would:` under mutate stages; **never** `skip (dry-run)` (see `cli/dry-run`) |

Quiet mode must remain scannable as an outline; do not dump git/API
micro-steps without a verbose gate.

## CLI shape examples

**Success (markers + verbose detail on stderr; blank; result on stdout):**

```text
$ mytool sink --create-mr --auto-merge -v
[1/5] launch       headless create-mr auto-merge
[2/5] agent        wait result.json
      notice: ping loop started
      ok
[3/5] validate     result.json
      notice: ship: read/validate /path/result.json
      ok
[4/5] ship         commit+push+MR
      notice: ship: git add -- topics/p.md
      notice: ship: git commit -m "docs(kb): …"
      notice: ship: git push origin HEAD:tester/…
      ok
[5/5] merge        ok

ok  mr=https://example.com/merge_requests/12
```

**Product after stages (same separator; e.g. seed):**

```text
$ kool go modcache seed -v .
[1/2] resolve      tag v0.0.2 (latest)
[2/2] download     github.com/xhd2015/lib@v0.0.2
      notice: go mod download -json github.com/xhd2015/lib@v0.0.2
      ok

seeded github.com/xhd2015/lib@v0.0.2
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

**Error mid-pipeline (flush-left `Error:`, non-zero — no blank before `Error:`):**

```text
$ mytool sink --create-mr
[1/4] launch       headless create-mr
[2/4] agent        wait result.json
Error: propose agent: timeout
```

**Dry-run (probes run; mutations gated; no `skip (dry-run)` stamps):**

```text
$ mytool sink --dry-run --create-mr --auto-merge -v
[1/5] launch       headless create-mr auto-merge
[2/5] agent        wait result.json
      would: wait result.json
[3/5] validate     result.json
[4/5] ship         commit+push+MR
      would: git add -- topics/p.md
      would: git commit -m "docs(kb): …"
      would: git push origin HEAD:tester/…
[5/5] merge        merge MR
      would: merge MR
```

When dry-run also emits a stdout product (`would: seed …`), use the same
blank stderr line before that product.

Color (when enabled via `cli/output/color`): yellow `warning:` and red
`Error:` on stderr; optional gray meta on markers. No ANSI in `--json`
or other machine-readable stdout.

## Anti-patterns

| Avoid | Prefer |
| ----- | ------ |
| Second `[n/total]` for start then `ok` | One marker; append `ok` / `notice:` under kind |
| Renumber `n` for every log line | One `n` per stage; detail under kind |
| Indent `[n/total]` lines | Markers flush left |
| Flush-left `notice:` / `would:` between stages | Kind-column indent |
| Wrap every detail as `[4/5] ship …` | `notice:` / `would:` / `ok` under kind |
| Every stage labeled `skip (dry-run)` | Same probe msgs as live; kind-aligned `would:` for gated mutates |
| Summary stuck to last stage (no blank) | Blank line on stderr, then flush-left stdout product |
| Kind-align the stdout summary under the last stage | Summary stays on stdout, flush left |
| Shrink `total` when skipping later stages | Keep `total`; emit `skipped (…)` with evidence |
| `[i/N]` for every scanned file as “stages” | Stream results (`cli/output/streaming`) |

## Testing notes

| Goal | Approach |
| ---- | -------- |
| Marker shape | Capture stderr; exact `"%d/%d] %-12s"` lines |
| One marker per stage | Assert **exactly `total`** lines matching `^\[[0-9]+/` |
| Kind indent | Assert detail lines start with `strings.Repeat(" ", len(fmt.Sprintf("[%d/%d] ", n, total)))` |
| Quiet vs verbose | Without `-v`, no micro-step `notice:`; with `-v`, notices appear under markers |
| Dry-run | Every stage present; probes/`would:` as live would plan; **no** `skip (dry-run)` |
| Post-stage summary | stderr ends with a blank line after the last stage/detail; product on stdout flush-left |
| Partial error | Prior markers kept; flush-left `Error:` with **no** preceding blank; non-zero exit |

## Out of scope

- Spinner / progress-bar library recipes
- Interactive TUI stage redraw
- `--stream` flags (see `cli/output/streaming`)

## See also

- `cli/output/streaming` — progressive stdout; progress/diagnostics on stderr
- `cli/output/alignment` — measure/pad when building custom columns
- `cli/output/color` — ANSI gating; never color machine-readable output
- `cli/dry-run` — one pipeline; probes run; gate mutations; no
  `skip (dry-run)` stamps

Reveal with:

```bash
go-best-practice skill --show cli/output/staged-markers
```
