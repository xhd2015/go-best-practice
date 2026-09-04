---
name: go-best-practice/cli/inline-tui-mouse
description: >-
  Mouse hit-testing for inline (non-alt-screen) terminal UIs: view-local
  hitmaps, CSI 6n origin measurement on a single stdin path, dual-origin
  fallback, and anti-patterns (sleep probes, parallel /dev/tty reads).
---

# inline-tui-mouse — mouse coordinates for inline TUIs

Use this recipe when a Go TUI paints **inline** (no alt-screen), enables
mouse cell reporting, and must map clicks to widgets (buttons, rows,
chips). Primary reference stack: **Bubble Tea** (`charmbracelet/bubbletea`);
the coordinate and CSI 6n rules apply to any library with the same paint
model.

## Problem

| Space | Coordinates |
| ----- | ----------- |
| **Mouse** | Viewport-absolute: `(0,0)` = top-left of the terminal (Bubble Tea: 0-based after SGR −1) |
| **Hitmap** | View-local: `Y = 0` is the first line of the UI frame |

Without a correct **origin** (absolute row of UI line 0):

```text
localY = absY - originY0
```

clicks land on the wrong row (e.g. “gen-commit-msg Run” fires “tag-next”).
**Mid-pane** start (cursor not at top or bottom) is common for bare CLIs;
top/bottom-only heuristics are not enough as the primary strategy.

## Coordinate model

```text
absX, absY     mouse cell (viewport-absolute, 0-based)
viewLines      number of lines the renderer paints for this frame
originY0       absolute row of the first UI line (0-based)
localY         absY - originY0
```

**Hit region** (half-open):

```text
y0 ≤ localY < y1
if x1 > x0:  x0 ≤ absX < x1   else whole line
```

Widgets that run an action store a non-empty `RunStage` (or equivalent id).
Left/right splits (toggle vs Run chip) use different `x0..x1` on the same `y`.

### viewLines must match paint

`viewLines` must equal the number of lines Bubble Tea’s renderer will
count for the `View()` string.

**Do not** end `View()` with a bare trailing `"\n"` if you also set
`viewLines = len(lines)`: `strings.Split` then yields an extra empty
segment → `linesRendered = viewLines+1` → bottom-origin off-by-one
(clicks hit the row *above* the pointer).

```go
// Good: join without trailing newline
return strings.Join(lines, "\n")

// Bad when viewLines == len(lines):
// return strings.Join(lines, "\n") + "\n"
```

## Anti-patterns (do not use)

| Approach | Why it fails |
| -------- | ------------ |
| `time.Sleep` / `tea.Tick(N)` then probe | Orders on wall clock, not paint |
| Open `/dev/tty` and **read** CSI 6n while tea owns mouse | Second input consumer **steals mouse SGR** (`ESC [ < … M`) |
| `os.Stdout.Write(CSI 6n)` in a Cmd after `View` | Bubble Tea flush is often **async** (ticker); probe can run **before** the frame is on screen → **stale cursor** |
| Clamp `row1 < viewLines` → `originY0 = 0` and treat as success | Marks impossible CPR as “top-anchored”; mid-pane mis-hits |
| Dual-origin (top/bottom) **only** as primary | Fails when UI starts mid-viewport |
| `tea.WithInput(plain io.Reader)` without `term.File` | Tea skips **MakeRaw** → no reliable mouse |

## Sound approach (event-driven CSI 6n)

### Pipeline

```text
1. Build hitmap with the same layout rules as paint (pure).
2. When origin unknown and height/viewLines ready:
   append CSI 6n once to the View string (same buffer as the frame).
3. Terminal replies CPR on stdin: ESC [ <row> ; <col> R  (1-based).
4. Filter peels CPR from the same stdin as keys/mouse; never a parallel reader.
5. originY0 = row1 - viewLines  only if row1 >= viewLines; else fail → dual-origin.
6. Mouse: Resolve(abs, originY or dual-origin fallback). Clicks never block on CPR.
7. Resize / viewLines change: bump layoutGen, clear origin, re-emit CSI 6n on next View.
```

**No sleep for correctness.** Ordering is: query rides with the paint buffer;
reply is async on stdin; until known, dual-origin is a soft fallback only.

### Origin state machine

```text
Unknown  → need to emit CSI 6n on next View
Pending  → query already in a frame; waiting for CPR
Known    → originY set; use known map
Failed   → no usable CPR for this layout; dual-origin only

layoutGen++ on resize / invalidate
pendingGen / pendingViewLines captured when emitting
Ignore CPR if phase != Pending or pendingGen != layoutGen
```

### Dual-origin fallback (not primary)

When `originY == nil`:

1. **Top:** `localY = absY`
2. **Bottom:** `localY = absY - max(0, height - viewLines)`

Prefer a hit with non-empty `RunStage` when candidates disagree.
Good for true top- or bottom-anchored UIs; **insufficient alone for mid-pane**.

## Minimal code framework

Portable sketches — adapt names to your package. Bubble Tea notes called out.

### Hitmap + resolve (pure)

```go
type Hit struct {
    Y0, Y1, X0, X1 int
    Focus          int    // keyboard focus index; -1 if unused
    RunStage       string // non-empty => click runs this action
}

func HitTest(hits []Hit, x, localY int) (Hit, bool) {
    for _, h := range hits {
        if localY < h.Y0 || localY >= h.Y1 {
            continue
        }
        if h.X1 > h.X0 && (x < h.X0 || x >= h.X1) {
            continue
        }
        return h, true
    }
    return Hit{}, false
}

// ResolveMouse maps absolute mouse → hit.
// originY nil => dual-origin; non-nil => localY = absY - *originY.
func ResolveMouse(absX, absY, height, viewLines int, originY *int, hits []Hit) (h Hit, kind string, ok bool) {
    if originY != nil {
        localY := absY - *originY
        h, ok = HitTest(hits, absX, localY)
        return h, "known", ok
    }
    // Dual-origin: top then bottom; prefer RunStage when they disagree.
    type cand struct {
        h      Hit
        localY int
        kind   string
    }
    var cands []cand
    if hit, yes := HitTest(hits, absX, absY); yes {
        cands = append(cands, cand{hit, absY, "top"})
    }
    origin := 0
    if height > 0 && viewLines > 0 {
        origin = height - viewLines
        if origin < 0 {
            origin = 0
        }
    }
    botY := absY - origin
    if hit, yes := HitTest(hits, absX, botY); yes {
        if len(cands) == 0 || cands[0].localY != botY {
            cands = append(cands, cand{hit, botY, "bottom"})
        }
    }
    if len(cands) == 0 {
        return Hit{}, "", false
    }
    best := cands[0]
    for _, c := range cands[1:] {
        if c.h.RunStage != "" && best.h.RunStage == "" {
            best = c
        }
    }
    return best.h, best.kind, true
}
```

### CSI 6n / CPR parse

```go
const CSI6n = "\x1b[6n"

// ParseCPR: first complete ESC [ row ; col R (1-based). Noise ignored.
func ParseCPR(buf []byte) (row1, col1 int, ok bool)

// OriginFromCPR for live TUI: require row1 >= viewLines (cursor on last
// line of a viewLines-tall frame). Do not treat row1 < viewLines as origin 0.
func OriginFromCPR(row1, viewLines int) (originY0 int, ok bool) {
    if viewLines <= 0 || row1 < 1 || row1 < viewLines {
        return 0, false
    }
    return row1 - viewLines, true
}

// DemuxCPRBytes peels complete CPRs from a stream; forwards mouse SGR
// (ESC [ < … M/m) and all other bytes. hold/rest for incomplete CSI.
func DemuxCPRBytes(hold, data []byte) (events []struct{ Row1, Col1 int }, forward, rest []byte)
```

**Live path:** if `row1 < viewLines`, mark origin **Failed** (dual-origin). Do
not clamp to success-at-zero.

### Stdin filter (Bubble Tea)

Must implement **`term.File`** so `MakeRaw` still runs on the real TTY:

```go
// charmbracelet/x/term.File:
//   io.ReadWriteCloser + Fd() uintptr

type CPRFilter struct {
    F             *os.File // usually os.Stdin
    Ch            chan<- CPRMsg
    hold, pending []byte
}

func (f *CPRFilter) Fd() uintptr                  { return f.F.Fd() }
func (f *CPRFilter) Write(p []byte) (int, error)  { return f.F.Write(p) }
func (f *CPRFilter) Close() error                 { return nil } // never close process stdin
func (f *CPRFilter) Read(p []byte) (int, error) {
    // demux F.Read → peel CPR into Ch; return forward bytes (and pending)
}
```

```go
ch := make(chan CPRMsg, 8)
in := NewCPRFilter(os.Stdin, ch)
p := tea.NewProgram(m,
    tea.WithInput(in), // custom input — filter MUST expose Fd()
    tea.WithOutput(os.Stdout),
    tea.WithMouseCellMotion(),
)
// Init: return waitCPR(ch) and re-arm after each CPR
```

A plain `io.Reader` wrapper **without** `Fd()` becomes `customInput` and tea
**does not** enter raw mode → mouse dies.

### View: emit CSI 6n in the same paint buffer

```go
func (m *Model) View() string {
    body := m.render() // sets m.viewLines = len(lines); no trailing "\n"
    if m.originPhase == OriginKnown && m.knownViewLines != m.viewLines {
        m.invalidateOrigin() // layoutGen++; phase = Unknown; originY = nil
    }
    if m.originPhase == OriginUnknown && m.height > 0 && m.viewLines > 0 {
        body += CSI6n // same string → same renderer buffer → same flush
        m.originPhase = OriginPending
        m.pendingGen = m.layoutGen
        m.pendingViewLines = m.viewLines
    }
    return body
}
```

Do **not** rely on a separate `os.Stdout.Write(CSI6n)` Cmd after View: the
standard renderer may still be holding the frame in a buffer/ticker.

### On CPR message

```go
// if phase != Pending || pendingGen != layoutGen → ignore (re-arm waitCPR)
// if row1 < pendingViewLines → phase = Failed; originY = nil
// else originY = row1 - pendingViewLines; phase = Known; knownViewLines = pendingViewLines
```

### On mouse

```go
h, _, ok := ResolveMouse(msg.X, msg.Y, height, viewLines, originY, hitmap)
// if !ok && originY != nil → optional second try with originY=nil (dual fallback)
// if RunStage != "" → run action; else focus semantics
```

## Checklist

- [ ] Hitmap Y/X built with the same width/layout as paint  
- [ ] `viewLines` matches painted line count (no extra trailing `\n`)  
- [ ] CSI 6n only via **View string** (or equivalent atomic paint buffer)  
- [ ] One stdin path; CPR demuxed; mouse SGR forwarded  
- [ ] Input wrapper is `term.File` (`Fd` on real `*os.File`)  
- [ ] Reject `row1 < viewLines`; dual-origin only as fallback  
- [ ] Invalidate origin on resize / viewLines change  
- [ ] No sleep/tick as probe ordering; no parallel `/dev/tty` CPR read  

## Tests to write in the app

| Case | Expected |
| ---- | -------- |
| Demux: CPR then mouse SGR | CPR peeled; mouse bytes unchanged |
| Demux: incomplete CSI across reads | held until complete |
| `OriginFromCPR(26, 20)` | origin 6, ok |
| Live rule: `row1=9`, `viewLines=20` | fail / reject (not origin 0) |
| Dual-origin top/bottom tables | distinct stages (e.g. gen-commit-msg ≠ tag-next) |
| Known mid-pane origin | click maps to correct `RunStage` |
| Filter `Fd()` | documented; smoke raw mouse on real TTY |

Prefer pure unit tests for demux/resolve; optional env-gated file log for
human calibration (`WRK_TUI_MOUSE_DEBUG`-style) while debugging.

## Optional: alt-screen

With alt-screen, UI is typically top-left and `localY ≈ absY` after home.
Simpler coordinates; different UX (leaves primary screen). This topic is for
**inline** TUIs that must stay on the primary screen.

## Retrieve

```bash
go-best-practice skill --show cli/inline-tui-mouse
go-best-practice skill cli/inline-tui-mouse --show
```

## See also

- `cli/output/color` — TTY / `NO_COLOR` for status paints  
- `cli/output/streaming` — progressive stdout (orthogonal to mouse)  
