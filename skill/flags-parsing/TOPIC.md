---
name: go-best-practice/flags-parsing
description: >-
  CLI flag parsing with less-flags; nested CLIs need --help at every level.
---

# flags — CLI flag parsing

A fluent flag parser for Go, **mainly used for CLI flag parsing**
(boolean/string/int/duration/string-slice options, with single or
multiple aliases per flag, built-in `--help`, **Cut** for opaque command
tails, and **CollectParsedFlags** to reconstruct/filter argv).

Install:

```bash
go get github.com/xhd2015/less-flags@latest
```

Example:

```go
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/xhd2015/less-flags"
)

const help = `
Usage: myapp [options] [args...]

Options:
  --timeout DURATION  set timeout duration
  --file FILE         add files to process, can be repeated
  -v, --verbose       enable verbose output
  -h, --help          show help
`

func main() {
    var verbose bool
    var timeout time.Duration
    var files []string

    remainArgs, err := lessflags.Duration("--timeout", &timeout).
        StringSlice("--file", &files).
        Bool("-v,--verbose", &verbose).
        Help("-h,--help", help).
        Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Println("verbose:", verbose)
    fmt.Println("timeout:", timeout)
    fmt.Println("files:", files)
    fmt.Println("args:", remainArgs)
}
```

## Help options

- `Help("-h,--help", helpText)` — prints `helpText` and exits 0.
- `HelpFunc("-h,--help", fn)` — calls `fn` and exits 0.
- `HelpNoExit()` — do not exit on help; `Parse` returns `lessflags.ErrHelp`.

### Nested CLIs: every sub-command level needs `--help`

A single top-level `Help(...)` is not enough once you dispatch to
sub-commands. Users run `mytool <cmd> --help` to learn **that** command's
flags. Wire `-h`/`--help` in **every** handler (and empty-args → that
level's help). Patterns and a full example live in
`flags-parsing/subcommand`.

## Multiple flag names

Comma-separate aliases in the name string:

```go
lessflags.Bool("-v,--verbose", &verbose)
lessflags.String("-o,--output", &output)
```

## Cut and CollectParsedFlags (overview)

**Cut** — stop parsing at a marker and copy all following tokens into a
slice (not re-parsed). Use for `myapp --exec <command> [args...]`.

```go
var execArgs []string
remain, err := lessflags.Bool("--verbose", &verbose).
    Cut("--exec", &execArgs).
    Parse(os.Args[1:])
```

**CollectParsedFlags** — record each parsed occurrence, then
`Reconstruct()` argv or `Remove(names)` parent-only flags before
forwarding to a child.

```go
var recorded lessflags.Flags
_, err := lessflags.Bool("--open", &open).
    Bool("--new-terminal", &nt).
    CollectParsedFlags(&recorded).
    Parse(os.Args[1:])
childArgs := recorded.Remove("--new-terminal").Reconstruct()
```

Full recipes: `flags-parsing/cut` and `flags-parsing/collect`.

## Sub-topics

- `flags-parsing/types` — the full list of supported target types
  (`*bool`, `*string`, `*int`/`*int64`, `*time.Duration`, `*[]string`,
  `Cut`, and their `**T` variants), plus how to detect "unset".
- `flags-parsing/subcommand` — sub-command dispatcher pattern using
  `StopOnFirstArg`.
- `flags-parsing/cut` — cut flags (opaque remaining command line).
- `flags-parsing/collect` — collect, reconstruct, and remove parsed flags.

Reveal with:

```bash
go-best-practice skill --show flags-parsing/types
go-best-practice skill --show flags-parsing/subcommand
go-best-practice skill --show flags-parsing/cut
go-best-practice skill --show flags-parsing/collect
```
