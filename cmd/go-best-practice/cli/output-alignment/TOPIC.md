---
name: go-best-practice/cli/output-alignment
description: >-
  Align CLI columns and fields: hand-rolled measure + pad (primary),
  optional tabwriter, rune-count width, ANSI-safe measure, truncate.
---

# output-alignment — columns, pad, width, and truncate

When a Go CLI prints tables or side-by-side fields, **measure visible
width**, then **pad** (or **truncate**) so columns line up. Prefer a
small hand-rolled measure + pad core over reaching for layout libraries
first. `text/tabwriter` is optional for quick scripts; production CLIs
usually own width math so ANSI color, truncation, and fixed columns stay
predictable.

## Policy

| Concern | Default | Notes |
| ------- | ------- | ----- |
| Width unit | **rune count** | Not `len(string)` (bytes); not terminal cell width for CJK unless you need it |
| Color / ANSI | **strip for measure** | Pad/truncate on visible width; re-emit original escapes |
| Layout engine | **hand-rolled pad** | `tabwriter` optional for ad-hoc dumps |
| Column width | **dynamic max** or **fixed** | Dynamic = max over rows; fixed = clamp / truncate |
| Overlong cells | **truncate** with ellipsis | Prefer middle or end truncate; keep pad on remainder |

## Measure visible width

Default width is **rune count** after stripping CSI/ANSI sequences:

```go
import (
    "regexp"
    "unicode/utf8"
)

// crude but practical: strip CSI ... m and similar ESC sequences
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func visibleWidth(s string) int {
    plain := ansiRE.ReplaceAllString(s, "")
    return utf8.RuneCountInString(plain)
}
```

Do **not** use `len(s)` for column width — multi-byte UTF-8 would
mis-align. For CJK double-width glyphs, upgrade measure later; rune
count is the documented default for this recipe.

## Pad (primary)

Pad on the **right** (or left for numbers) to a target width using
spaces based on `visibleWidth`:

```go
func padRight(s string, width int) string {
    w := visibleWidth(s)
    if w >= width {
        return s
    }
    return s + strings.Repeat(" ", width-w)
}

func padLeft(s string, width int) string {
    w := visibleWidth(s)
    if w >= width {
        return s
    }
    return strings.Repeat(" ", width-w) + s
}
```

### Dynamic column widths

Collect cells → measure max visible width per column → pad each cell:

```go
// rows is [][]string; colWidths[j] = max visible width of column j
for _, row := range rows {
    for j, cell := range row {
        if w := visibleWidth(cell); w > colWidths[j] {
            colWidths[j] = w
        }
    }
}
for _, row := range rows {
    for j, cell := range row {
        fmt.Fprint(out, padRight(cell, colWidths[j]))
        if j+1 < len(row) {
            fmt.Fprint(out, "  ") // column gap
        }
    }
    fmt.Fprintln(out)
}
```

### Fixed column widths

When the terminal width is known or a schema defines columns, clamp:

```go
func fit(s string, width int) string {
    if visibleWidth(s) <= width {
        return padRight(s, width)
    }
    return truncate(s, width) // see below
}
```

## Truncate

When a cell exceeds the column width, **truncate** and append an
ellipsis (or hard cut for machine output). Measure and cut in runes
after ANSI strip, then re-apply color if you stripped it for measure:

```go
func truncate(s string, width int) string {
    if width <= 0 {
        return ""
    }
    plain := ansiRE.ReplaceAllString(s, "")
    runes := []rune(plain)
    if len(runes) <= width {
        return s // original (may include ANSI)
    }
    if width == 1 {
        return "…"
    }
    return string(runes[:width-1]) + "…"
}
```

For colored cells, prefer: strip → truncate runes → re-wrap with the
same open/close sequences, or truncate plain text and color only after
fit. Never pad using byte length of ANSI-laden strings.

## tabwriter (optional)

`text/tabwriter` is fine for throwaway debug tables. It does **not**
know about ANSI or rune width:

```go
w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
fmt.Fprintln(w, "NAME\tSTATUS\tMSG")
fmt.Fprintln(w, "job-a\tok\tdone")
w.Flush()
```

Use tabwriter when cells are plain ASCII and you do not need truncate
or color-safe measure. Otherwise stay with hand-rolled pad + width.

## ANSI-safe alignment

1. **Measure** only the visible portion (strip CSI).
2. **Pad / truncate** using that width.
3. **Print** the original string (with escapes) plus pure space padding.

```go
func padRightANSI(s string, width int) string {
    w := visibleWidth(s)
    if w >= width {
        return s
    }
    return s + strings.Repeat(" ", width-w)
}
```

See also `cli/color` for when to emit ANSI at all (`--color` /
`--no-color` / TTY / `NO_COLOR`).

## Anti-patterns

| Anti-pattern | Prefer |
| ------------ | ------ |
| `fmt.Sprintf("%-20s", s)` with colored or multi-byte text | `padRight` on `visibleWidth` |
| `len(s)` as column width | `utf8.RuneCountInString` on stripped text |
| Buffer entire table only to align when streaming would do | Stream rows when widths are **fixed**; buffer when **dynamic** max is required |
| tabwriter + ANSI without stripping | Hand-rolled measure + pad |

## See also

- `cli/streaming` — when to buffer for tables vs stream fixed-width rows
- `cli/color` — three-mode color policy; strip for measure, emit when allowed
- `flags-parsing` — flag shapes for `--width` / output options if exposed
