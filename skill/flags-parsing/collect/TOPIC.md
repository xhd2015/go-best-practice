---
name: go-best-practice/flags-parsing/collect
description: >-
  CollectParsedFlags: record CLI flag occurrences for Reconstruct and Remove.
---

# flags — CollectParsedFlags

`CollectParsedFlags` records each successfully parsed flag occurrence into a
`Flags` value (CLI order, name form as written). Use it to rebuild argv for a
child process, drop parent-only flags, or inspect what the user typed.

```bash
go get github.com/xhd2015/less-flags@latest
```

## Types

```go
// One successfully parsed flag occurrence as written on the CLI.
// Value is empty for bare bools; Reconstruct emits the name only in that case.
type Flag struct {
    Flag  string
    Value string
}

type Flags []Flag

func (f Flags) Reconstruct() []string // argv tokens in CLI order
func (f Flags) String() string        // space-joined Reconstruct (debug)
func (f Flags) Remove(names string) Flags // drop aliases; original not mutated
```

Chain before `Parse`:

```go
var recorded lessflags.Flags
remain, err := lessflags.String("--session-id", &id).
    Bool("--open", &open).
    CollectParsedFlags(&recorded).
    Parse(os.Args[1:])
```

Without `CollectParsedFlags`, parse behavior is unchanged (collection is a
no-op when not chained).

## Reconstruct rules

| Case | Recorded | Reconstruct tokens |
|------|----------|--------------------|
| `--session-id s1` | `{Flag: "--session-id", Value: "s1"}` | `--session-id`, `s1` |
| `--session-id=s1` | same Value `"s1"` | still **space form** `--session-id s1` |
| bare `--open` | `{Flag: "--open", Value: ""}` | `--open` only |
| `--files a --files b` | two entries | `--files a --files b` |
| CLI order | preserved | not registration order |

## Remove

`Remove(names)` returns a **new** `Flags` without entries whose name matches
any comma-separated alias. The original is not mutated.

```go
// Drop whichever form appeared: -s or --session.
child := recorded.Remove("-s,--session")
```

## Example: parent flags → child argv

Common pattern: parent CLI accepts extra flags (e.g. open a new terminal) that
must not be forwarded to the child.

```go
package main

import (
    "fmt"
    "os"

    "github.com/xhd2015/less-flags"
)

func main() {
    var open, newTerminal, auto bool
    var sessionID string
    var recorded lessflags.Flags

    _, err := lessflags.Bool("--auto-send-or-resume", &auto).
        Bool("--new-terminal", &newTerminal).
        String("--session-id", &sessionID).
        Bool("--open", &open).
        CollectParsedFlags(&recorded).
        Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    // Parent-only: --new-terminal
    childArgs := recorded.Remove("--new-terminal").Reconstruct()
    // e.g. input:
    //   --auto-send-or-resume --new-terminal --session-id sess-1 --open
    // childArgs:
    //   ["--auto-send-or-resume", "--session-id", "sess-1", "--open"]

    if newTerminal {
        fmt.Println("spawn new terminal with:", childArgs)
    } else {
        fmt.Println("run in-process with:", childArgs)
    }
}
```

## Collect + Cut (v1)

If a `Cut` flag matches, collection records the **marker only** (`Value`
empty). The cut body lives in the `Cut` target, not in `Flags.Reconstruct()`.
Read the cut target when you need the body.

See `flags-parsing/cut`.

## Related

- `flags-parsing` — overview and basic parse
- `flags-parsing/cut` — consume remaining args after a marker
- `flags-parsing/types` — supported target types

Reveal with:

```bash
go-best-practice skill --show flags-parsing/collect
```
