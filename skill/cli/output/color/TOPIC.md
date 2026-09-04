---
name: go-best-practice/cli/output/color
description: >-
  Terminal ANSI color for CLI output: --color / --no-color, TTY auto,
  and the NO_COLOR environment convention.
---

# color — terminal ANSI color for CLIs

When a Go CLI prints human-facing status, summaries, or highlights,
gate ANSI color with an explicit three-mode policy. Default is **auto**
(TTY detection), overridden by `--color` / `--no-color` and by the
industry `NO_COLOR` convention.

Do **not** implement `FORCE_COLOR`, `CLICOLOR`, or `CLICOLOR_FORCE` in
this recipe. Force-on is only via `--color`.

## Policy (three modes)

| Mode | How selected | Color on? |
| ---- | ------------ | --------- |
| **Auto** (default) | neither flag | TTY → on, unless `NO_COLOR` non-empty → off; non-TTY → off |
| **Always** | `--color` | always on (ignores TTY and `NO_COLOR`) |
| **Never** | `--no-color` | always off (ignores TTY and `NO_COLOR`) |

**Conflict:** `--color` and `--no-color` together must fail with:

```text
--color and --no-color cannot be specified together
```

### Resolve order

```text
if --color && --no-color → error (message above)
if --color               → Always → true
if --no-color            → Never  → false
// Auto only:
if NO_COLOR != ""        → false   // any non-empty value
else                     → IsTerminal(stdout)
```

CLI flags always win over env and TTY. `NO_COLOR` applies only in **auto**.

## NO_COLOR (no-color.org)

Canonical disable-color env: [no-color.org](https://no-color.org/).

| Value | Effect in auto |
| ----- | -------------- |
| unset | no effect |
| empty (`NO_COLOR=`) | no effect (not considered set) |
| any non-empty (`1`, `true`, `yes`, …) | disable color |

Value content is ignored; presence of a non-empty string is enough:

```go
if os.Getenv("NO_COLOR") != "" {
    // disable color in auto mode
}
```

Document `NO_COLOR=1` as the common user-facing example.

## ColorMode model

```go
type ColorMode int

const (
    ColorAuto ColorMode = iota // default
    ColorAlways                // --color
    ColorNever                 // --no-color
)

// ResolveColor returns whether ANSI escapes should be emitted.
// stdoutIsTTY should come from term.IsTerminal on the real stdout fd.
// noColorEnv is os.Getenv("NO_COLOR") (empty string if unset).
func ResolveColor(mode ColorMode, stdoutIsTTY bool, noColorEnv string) bool {
    switch mode {
    case ColorAlways:
        return true
    case ColorNever:
        return false
    default: // ColorAuto
        if noColorEnv != "" {
            return false
        }
        return stdoutIsTTY
    }
}
```

## CLI flags (less-flags)

Parse both bool flags, reject mutual use, map to `ColorMode`:

```go
package main

import (
    "fmt"
    "os"

    lessflags "github.com/xhd2015/less-flags"
    "golang.org/x/term"
)

func parseColorMode(args []string) (ColorMode, []string, error) {
    var colorFlag, noColorFlag bool
    remain, err := lessflags.
        Bool("--color", &colorFlag).
        Bool("--no-color", &noColorFlag).
        Parse(args)
    if err != nil {
        return 0, nil, err
    }
    // Always detect conflict from less-flags results — do not hand-scan argv.
    if colorFlag && noColorFlag {
        return 0, nil, fmt.Errorf("--color and --no-color cannot be specified together")
    }

    mode := ColorAuto
    if colorFlag {
        mode = ColorAlways
    }
    if noColorFlag {
        mode = ColorNever
    }
    return mode, remain, nil
}

func useColor(mode ColorMode) bool {
    return ResolveColor(mode, term.IsTerminal(int(os.Stdout.Fd())), os.Getenv("NO_COLOR"))
}
```

Parse with less-flags only; reject both flags from the bool results after
`Parse`. Do not manually scan `args` for `"--color"` / `"--no-color"`
(misses equals forms and duplicates the parser). Help text should list:

```text
  --color      force ANSI color on (even when stdout is not a TTY)
  --no-color   force ANSI color off
```

## TTY detection

Prefer `golang.org/x/term`:

```go
import "golang.org/x/term"

tty := term.IsTerminal(int(os.Stdout.Fd()))
```

Install:

```bash
go get golang.org/x/term@latest
```

**Alternative** when coloring against a generic `io.Writer` that may be
an `*os.File` (e.g. redirected test pipe):

```go
func writerIsTTY(w io.Writer) bool {
    f, ok := w.(*os.File)
    if !ok {
        return false
    }
    st, err := f.Stat()
    if err != nil {
        return false
    }
    return (st.Mode() & os.ModeCharDevice) != 0
}
```

Use the same writer you actually print to (usually `os.Stdout`).

## Applying color (gate every SGR)

Never emit escapes when color is off. Wrap through one style helper:

```go
const (
    ansiReset = "\x1b[0m"
    ansiRed   = "\x1b[31m"
    ansiGreen = "\x1b[32m"
    ansiGray  = "\x1b[90m"
)

type colorStyle struct{ enabled bool }

func newColorStyle(mode ColorMode) colorStyle {
    return colorStyle{enabled: useColor(mode)}
}

func (c colorStyle) wrap(code, s string) string {
    if !c.enabled {
        return s
    }
    return code + s + ansiReset
}

func (c colorStyle) red(s string) string   { return c.wrap(ansiRed, s) }
func (c colorStyle) green(s string) string { return c.wrap(ansiGreen, s) }
func (c colorStyle) gray(s string) string  { return c.wrap(ansiGray, s) }
```

Typical mapping: pass/success → green, fail/error → red, meta/duration → gray.
Always reset after each colored token so later plain text is not tinted.

## Testing notes

Harness stdout is usually a **pipe** (non-TTY), so **auto is off** unless
you force `--color` / `ColorAlways`.

| Scenario | Setup | Expect |
| -------- | ----- | ------ |
| Force on | `--color` or Always, non-TTY | ANSI present (`\x1b`) |
| Force off | `--no-color` or Never | no ANSI |
| `NO_COLOR` auto | `NO_COLOR=1`, no flags, non-TTY or TTY | no ANSI |
| Force wins | `--color` + `NO_COLOR=1` | ANSI present |
| Auto pipe | no flags, stdout pipe, no `NO_COLOR` | no ANSI |
| Conflict | `--color` + `--no-color` | error: `cannot be specified together` |

Strip parent `NO_COLOR` in the test harness when leaves must own the env
(so ambient developer shell settings do not flake tests). Re-inject
`NO_COLOR=1` only in leaves that assert disable behavior.

Assert specific SGR sequences when useful (e.g. green pass token), not
only “any escape present”.

## Out of scope

- `FORCE_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`
- Full terminfo / 256-color palettes
- Coloring JSON or machine-readable output (keep those plain)

## See also

- `cli/output/streaming` — stream output as work proceeds
- `cli/config` — gray `notice:` when a preference comes from config.json
- `flags-parsing` — bool flags and help text with less-flags
- [no-color.org](https://no-color.org/) — `NO_COLOR` convention

Reveal with:

```bash
go-best-practice skill --show cli/output/color
```
