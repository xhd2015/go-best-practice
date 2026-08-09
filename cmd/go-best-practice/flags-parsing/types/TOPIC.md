---
name: go-best-practice/flags-parsing/types
description: >-
  Supported flag target types for less-flags.
---

# flags — supported target types

Every flag method takes a pointer target. Both `*T` and `**T` are
supported (the `**T` form lets the caller distinguish "unset" from
"zero value").

| Method                              | `*T`              | `**T`               |
| ----------------------------------- | ----------------- | ------------------- |
| `Bool(names, target)`               | `*bool`           | `**bool`            |
| `String(names, target)`             | `*string`         | `**string`          |
| `Int(names, target)`                | `*int` / `*int64` | `**int` / `**int64` |
| `Duration(names, target)`           | `*time.Duration`  | `**time.Duration`   |
| `StringSlice(names, target)`        | `*[]string`       | `**[]string`        |
| `Cut(names, target)`                | `*[]string`       | `**[]string`        |

`StringSlice` appends one value per occurrence and keeps parsing.
`Cut` assigns **all tokens after the marker** once and **stops** parsing
(body not re-parsed; equals form rejected). Recipe: `flags-parsing/cut`.

## Detecting "unset" with `**T`

```go
var timeout *time.Duration
remain, err := lessflags.Duration("--timeout", &timeout).Parse(os.Args[1:])
if err != nil { /* ... */ }
if timeout == nil {
    // --timeout was not provided
} else {
    fmt.Println("timeout =", *timeout)
}
_ = remain
```

## Repeatable slice flags

`StringSlice` appends each occurrence:

```bash
myapp --file a.txt --file b.txt --file c.txt
# files == []string{"a.txt", "b.txt", "c.txt"}
```

## Boolean flags

Bool flags may be passed bare (`--verbose`) or with an explicit value
(`--verbose=true`, `--verbose=false`).

## Cut targets

Public API: `Cut(names, target *[]string)`. Prefer `Cut` when the rest of
the line is an external command; prefer `StringSlice` for repeatable
single values. See `flags-parsing/cut`.

## CollectParsedFlags (not a target type)

`CollectParsedFlags(dst *Flags)` records occurrences for
`Reconstruct` / `Remove`. See `flags-parsing/collect`.
