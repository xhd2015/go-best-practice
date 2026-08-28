---
name: go-best-practice/cli/streaming
description: >-
  Stream CLI output as work proceeds; avoid buffering all output until
  the end. When buffering is justified, flush, NDJSON vs full JSON.
---

# streaming — stream CLI output as you go

When a Go CLI produces results over time (scan, list, fetch, process),
**write each ready unit of output when it is ready**. Do not accumulate
everything in a `[]string` / `strings.Builder` only to dump at the end.

Streaming is the **default design**. Buffering is the special case you
justify (tables, sort/aggregate, atomic summary, full JSON document,
atomic file replace).

There is **no** recommended `--stream` flag. Prefer streaming always;
do not keep buffer-by-default and ask users to opt in.

## Why stream

| Concern | Buffer until end | Stream as you go |
| ------- | ---------------- | ---------------- |
| Perceived latency | Looks hung | Immediate feedback |
| Large output | High memory | Bounded memory |
| Crash / Ctrl-C | Often nothing useful | Partial but real output |
| Pipelines | Consumer waits for EOF | Consumer can start early |
| Progress | Extra channel or spinner only | Output itself shows progress |

## Policy

### Default: stream

- Print each **logical unit** as soon as it is complete (line, record,
  row, event).
- Primary results → **stdout**; progress / diagnostics → **stderr**.
- On error after partial stdout: still exit non-zero. Do **not** try to
  rewind stdout. Document that partial output may exist.

### When buffering is OK

| Situation | Why buffer | Pattern |
| --------- | ---------- | ------- |
| Column-aligned tables needing max widths | Need all cells first | Collect → measure → print (flush promptly after build) |
| Sorted / top-N / aggregate-only | Need full set | Collect → sort/agg → print |
| Atomic “all or nothing” human summary | Partial success would mislead | Compute → single final write (or print only on success) |
| Single JSON **document** (array/object) | Closing `]` / `}` | Prefer NDJSON for long lists; else buffer document |
| Atomic file replace | Avoid half-written files | Write temp → rename |

If none of those apply, **do not** hold output until the end.

## Anti-pattern → preferred

**Anti-pattern** — hold everything, then dump:

```go
var lines []string
for _, item := range items {
    lines = append(lines, format(item))
}
fmt.Print(strings.Join(lines, "\n"))
```

**Preferred** — stream each unit:

```go
for _, item := range items {
    fmt.Fprintln(os.Stdout, format(item))
}
```

Same idea with incremental work (scan disk, page an API, process files):
emit a line as soon as that item is ready, not after the whole job.

## Streaming recipes

### Line-at-a-time

```go
package main

import (
    "fmt"
    "os"
)

func listPaths(paths []string) {
    for _, p := range paths {
        fmt.Fprintln(os.Stdout, p)
    }
}
```

Use `fmt.Fprintln` / `fmt.Fprintf` to the real writer (`os.Stdout` or an
injected `io.Writer` in tests). Avoid building a giant string first.

### High volume: `bufio.Writer` + flush

Streaming ≠ one syscall per byte. For many small writes, buffer for
performance **and** still flush so the user (or pipe) sees data during
long work:

```go
import (
    "bufio"
    "fmt"
    "os"
)

func listPaths(paths []string) error {
    w := bufio.NewWriter(os.Stdout)
    defer w.Flush()

    for _, p := range paths {
        if _, err := fmt.Fprintln(w, p); err != nil {
            return err
        }
        // Optional: flush every N lines, or before a long blocking step
        // so interactive users see progress while work continues.
    }
    return w.Flush()
}
```

**Flush discipline:**

- `defer w.Flush()` so exit paths still push remaining bytes.
- Flush before a long wait (network, subprocess) when the user should
  already see prior lines.
- On error, flush what you have, then write the error to **stderr**.

### Progress on stderr, results on stdout

```go
fmt.Fprintln(os.Stderr, "scanning…")
for _, p := range matches {
    fmt.Fprintln(os.Stdout, p) // safe to pipe; progress stays on stderr
}
fmt.Fprintf(os.Stderr, "done (%d matches)\n", len(matches))
```

Typical session:

```text
$ mytool find --name '*.go' | wc -l
# stderr (TTY):
scanning…
done (3 matches)
# stdout (pipe to wc): only paths
```

### Partial output then error

```text
$ mytool find ./ok ./boom
./ok/a.go
Error: open ./boom: permission denied
```

Stdout already has `./ok/a.go`. Exit non-zero. Do **not** buffer all
paths just so a late error can suppress them—unless the command’s
contract is explicitly “all or nothing” (then document that and buffer
by design).

## Tables and sorted lists

Two-pass is legitimate when alignment or order requires the full set:

```go
// OK: need widths / sort
rows := collectRows(items)
sort.Slice(rows, ...)
printTable(os.Stdout, rows) // still write promptly once ready
```

Alternatives when you want progressive output:

- Fixed column widths (stream rows with padding).
- Unaligned stream (`key=value` or one field per line).
- NDJSON records (machine-readable, streamable).

## Machine-readable: prefer NDJSON for long lists

For unbounded or long-running lists, emit **one JSON object per line**
(NDJSON / JSON Lines):

```text
{"id":"a","status":"ok"}
{"id":"b","status":"ok"}
```

```go
enc := json.NewEncoder(os.Stdout)
for _, item := range items {
    if err := enc.Encode(item); err != nil { // Encode adds '\n'
        return err
    }
}
```

**Soft rule:** full JSON arrays/objects are fine for **small, fixed**
payloads. For unbounded lists, prefer NDJSON so you never hold the
entire document in memory or delay the first record until the last.

Do **not** put ANSI color in JSON / NDJSON (see `cli/color`).

## CLI shape examples

**Success (streamed rows on stdout):**

```text
$ mytool find --name '*.go'
./cmd/root.go
./internal/scan.go
./internal/scan_test.go

3 paths
```

Rows appear as discovered; a final count may be the last stdout line or
gray meta on stderr.

**Warning (stderr, exit 0):**

```text
$ mytool find ./missing ./present
warning: skip ./missing: no such file or directory
./present/a.go
```

**Error after partial output (stderr, non-zero):**

```text
$ mytool find ./ok ./boom
./ok/a.go
Error: open ./boom: permission denied
```

**Help (machine vs human):**

```text
$ mytool find --help
Usage: mytool find [OPTIONS] [PATH]...
  --json    emit NDJSON records on stdout (no ANSI)
```

Color (when enabled via `cli/color`): yellow `warning:` and red `Error:`
on stderr; green success tokens and gray meta on stdout for human mode.
No ANSI in `--json` / NDJSON.

## Testing notes

| Goal | Approach |
| ---- | -------- |
| Stream order | Inject `io.Writer`; assert lines appear in discovery order |
| No end-only dump | Prefer unit tests that fail if implementation only writes once at the end (e.g. spy writer counting Write calls during a fake slow iterator) |
| Stderr vs stdout | Capture both; progress must not pollute piped stdout |
| NDJSON | Decode line-by-line; each line a complete object |
| Partial error | Simulate failure mid-loop; expect prior lines on stdout and non-zero exit |

Harness pipes are fine: streaming behavior does not depend on TTY.

## Out of scope

- `--stream` / `--buffer` flags as the primary UX
- Full interactive TUI redraw (bubbletea, etc.)
- Spinner / progress-bar library recipes
- Stage-style `[n/total]` spines (see `cli/staged-markers`)
- Holding all human output only because stdout is a pipe (pipelines
  benefit from streaming too)

## See also

- `cli/staged-markers` — fixed multi-stage `[n/total]` progress on
  stderr; kind-aligned detail between stages
- `cli/color` — when to emit ANSI; never color machine-readable output
- `cmd-exec` — external commands inherit stdout/stderr (live stream by
  default); prefer that over capturing then reprinting unless you need
  the bytes
- `flags-parsing` — flag and help text patterns

Reveal with:

```bash
go-best-practice skill --show cli/streaming
```
