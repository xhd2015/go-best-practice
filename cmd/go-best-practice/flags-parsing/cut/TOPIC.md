---
name: go-best-practice/flags-parsing/cut
description: >-
  Cut flags: consume all remaining tokens after a marker without re-parsing.
---

# flags — Cut (consume remaining args)

Use `Cut` when a flag marks the start of a **foreign command line** (or any
opaque tail). Everything after the marker is copied into a `*[]string` and
parsing **stops**. Tokens in the body are **not** re-parsed as less-flags
options.

Install:

```bash
go get github.com/xhd2015/less-flags@latest
```

## When to use Cut vs StringSlice

| | `StringSlice` | `Cut` |
|---|---------------|-------|
| Values | One value per occurrence; appends | All tokens after the marker once |
| After the flag | Parsing continues | Parsing stops |
| Body tokens | Must be a single flag value | May look like flags (`--x`, `--`) |
| Equals form | `--file=a` OK | `--exec=ls` rejected |

Typical shape:

```text
myapp [own-flags...] --exec <command> [args...]
```

## Example

```go
package main

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/xhd2015/less-flags"
)

const help = `
Usage: myapp [--verbose] --exec <command> [args...]

Options:
  --verbose     enable verbose output
  --exec ...    run the remaining tokens as a command (required body)
  -h, --help    show help
`

func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run(args []string) error {
    var verbose bool
    var execArgs []string

    remain, err := lessflags.Bool("--verbose", &verbose).
        Cut("--exec", &execArgs).
        Help("-h,--help", help).
        Parse(args)
    if err != nil {
        return err
    }
    _ = remain // positionals before --exec, if any

    if len(execArgs) == 0 {
        return fmt.Errorf("missing --exec <command>")
    }
    if verbose {
        fmt.Println("running:", execArgs)
    }
    cmd := exec.Command(execArgs[0], execArgs[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

```bash
myapp --verbose --exec ls -la /tmp
# verbose=true, execArgs=["ls", "-la", "/tmp"]

myapp --exec --verbose foo
# execArgs=["--verbose", "foo"]  — not parsed as myapp's --verbose

myapp --exec -- ls
# execArgs=["--", "ls"]  — "--" is a literal token inside the cut body
```

## Rules

1. **Body required** — bare `--exec` with no following tokens errors
   (`requires a command`).
2. **No equals form** — `--exec=ls` is rejected.
3. **Not reparsed** — cut body is raw strings.
4. **Positionals before the marker** still land in `remainArgs`.
5. **Stops parse** — nothing after the marker returns in `remainArgs`.

```go
// args: file1 --exec echo hi
var execArgs []string
remain, err := lessflags.Cut("--exec", &execArgs).Parse(
    []string{"file1", "--exec", "echo", "hi"},
)
// remain == ["file1"], execArgs == ["echo", "hi"]
```

## Target type

`Cut(names, target *[]string)` — the public signature takes `*[]string`.
Internally `**[]string` is also accepted (same as other slice targets).

## Related

- `flags-parsing/types` — full type table
- `flags-parsing/collect` — record parsed flags; Cut is recorded as marker only (v1)
- `flags-parsing/subcommand` — word sub-commands with `StopOnFirstArg` (different pattern)

Reveal with:

```bash
go-best-practice skill --show flags-parsing/cut
```
