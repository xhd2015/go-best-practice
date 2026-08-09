---
name: go-best-practice/flags-parsing/subcommand
description: >-
  Sub-command dispatcher patterns with StopOnFirstArg; every level needs --help.
---

# flags — sub-command dispatcher pattern

Use `StopOnFirstArg()` so top-level flags stop parsing at the first
positional argument. The remainder can then be dispatched to a
sub-command handler, which runs its own `lessflags.Parse` over its own
option set.

## Rule: every command level supports `-h` / `--help`

Users explore CLIs by walking the tree with help. **Each level** of the
command hierarchy must answer `-h` / `--help` with **that level's** usage
(not only the root binary).

| Level | Example | What `--help` shows |
|-------|---------|---------------------|
| Root / binary | `mytool --help` | global commands + global flags |
| Sub-command | `mytool install --help` | install usage + install flags |
| Nested sub-command | `mytool skill --help` | skill actions (`--show`, `--install`, …) |
| Nested action that has its own flags | `mytool skill --install --help` | install-target flags only |

**Do not** leave a dispatch node without help. If the user runs
`mytool skill --help` and the parser requires an action flag first, they
get a dead end. Handle `-h`/`--help` **before** (or as an alternative to)
requiring sub-actions.

Checklist when adding a command:

1. Top-level: `Help("-h,--help", topHelp)` or manual `-h`/`--help` when no
   top-level flags.
2. Every `case "cmd":` handler: its own `Help("-h,--help", cmdHelp)` (or
   equivalent for flag-style actions via `skillcmd`).
3. Top-level help text should say: `Run mytool <command> --help for
   command-specific options.`
4. Empty args at a level that only dispatches → print that level's help
   (friendly default).

```go
package main

import (
    "fmt"
    "os"

    "github.com/xhd2015/less-flags"
)

const topHelp = `
Usage: mytool [global-options] <command> [ARGS]

Commands:
  install [<dir>]    install something to <dir>
  run <script>       run a script

Global options:
  --debug        enable debug output
  -h, --help     show this help

Run mytool <command> --help for command-specific options.
`

func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

func run(args []string) error {
    var debug bool
    args, err := lessflags.Bool("--debug", &debug).
        Help("-h,--help", topHelp).
        StopOnFirstArg().
        Parse(args)
    if err != nil {
        return err
    }
    if len(args) == 0 {
        fmt.Print(topHelp)
        return nil
    }

    switch args[0] {
    case "install":
        return handleInstall(args[1:])
    case "run":
        return handleRun(args[1:])
    default:
        return fmt.Errorf("unknown command: %s", args[0])
    }
}

func handleInstall(args []string) error {
    var force bool
    args, err := lessflags.Bool("--force", &force).
        Help("-h,--help", `
Usage: install [--force] <dir>
`).Parse(args)
    if err != nil {
        return err
    }
    if len(args) == 0 {
        return fmt.Errorf("install requires <dir>")
    }
    fmt.Printf("installing to %s (force=%v)\n", args[0], force)
    return nil
}

const runHelp = `
Usage: mytool run [--verbose] <script>

Options:
  --verbose     verbose run output
  -h, --help    show this help
`

func handleRun(args []string) error {
    var verbose bool
    args, err := lessflags.Bool("--verbose", &verbose).
        Help("-h,--help", runHelp).
        Parse(args)
    if err != nil {
        return err
    }
    if len(args) == 0 {
        return fmt.Errorf("run requires a script (try --help)")
    }
    fmt.Printf("running %s (verbose=%v)\n", args[0], verbose)
    return nil
}
```

## No toplevel flags

When the toplevel itself has no flags, skip `lessflags.Parse` entirely.
Just check the first arg, manually handle `-h`/`--help`, and
dispatch. Each sub-command parses its own lessflags.

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/xhd2015/less-flags"
)

const topHelp = `
Usage: mytool <command> [OPTIONS]

Commands:
  install     install deny rules
  clean       remove installed rules

Run mytool <command> --help for command-specific options.
`

func main() {
    if err := run(os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", "mytool", err)
        os.Exit(1)
    }
}

func run(args []string) error {
    if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
        fmt.Print(strings.TrimPrefix(topHelp, "\n"))
        return nil
    }

    cmd := args[0]
    cmdArgs := args[1:]
    switch cmd {
    case "install":
        return handleInstall(cmdArgs)
    case "clean":
        return handleClean(cmdArgs)
    default:
        return fmt.Errorf("unknown command: %s", cmd)
    }
}

const installHelp = `
Usage: mytool install [--dir DIR] [--dry-run]

Options:
  --dir <dir>   project directory (default: current directory)
  --dry-run     show what would be changed without writing
  -h,--help     show help
`

func handleInstall(args []string) error {
    var dirFlag *string
    var dryRun bool
    _, err := lessflags.String("--dir", &dirFlag).
        Bool("--dry-run", &dryRun).
        Help("-h,--help", installHelp).
        Parse(args)
    if err != nil {
        return err
    }

    dir := "."
    if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
        dir = strings.TrimSpace(*dirFlag)
    }

    if dryRun {
        fmt.Printf("Would install to %s\n", dir)
        return nil
    }

    fmt.Printf("Installed to %s\n", dir)
    return nil
}

func handleClean(args []string) error {
    var dirFlag *string
    _, err := lessflags.String("--dir", &dirFlag).
        Help("-h,--help", `
Usage: mytool clean [--dir DIR]

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).Parse(args)
    if err != nil {
        return err
    }

    dir := "."
    if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
        dir = strings.TrimSpace(*dirFlag)
    }

    fmt.Printf("Cleaned %s\n", dir)
    return nil
}
```

## Notes

- **Every level needs `--help`** — root, each word sub-command, and any
  nested flag surface (e.g. `skill --help` and `skill --install --help`).
  Prefer `lessflags.Help("-h,--help", helpText)` inside each handler; for
  skill flag CLIs use `skillcmd` (`SingleSkill.Help` / `Registry.Help`).
- Without `StopOnFirstArg()`, `flags` would try to interpret
  sub-command flags (e.g. `--force` after `install`) against the
  top-level spec and fail with `unrecognized flag`.
- When the toplevel has **no** flags, skip `lessflags.Parse` entirely
  at the toplevel. Just check `args[0]` for `-h`/`--help` and
  dispatch raw args to sub-commands.
- Each handler reuses `lessflags.Help(...)` for its own `--help` so
  help text stays next to that command's flags.
- Point users deeper: top help mentions `mytool <command> --help`.
- See the `flags-parsing/types` sub-topic for the full list of
  supported target types.
- For an opaque trailing command line (`myapp --exec <cmd> [args...]`),
  use `Cut` instead of word dispatch — see `flags-parsing/cut`.
- To forward a filtered flag set to a child process, use
  `CollectParsedFlags` — see `flags-parsing/collect`.
- Skill CLIs: see `cli/skill-cli` for per-level help on `skill` /
  `skills` / `--install`.
