---
name: go-best-practice/cli/project-layout/script/install
description: >-
  Local CLI install scripts: LookPath or ~/.local/bin, refresh existing
  GOPATH/GOBIN/.local copies, EnsureOnPATH, macOS ad-hoc codesign. Use
  github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install.
---

# script/install — PATH-aware local binary install

Bare `go install ./cmd/…` always writes `$GOBIN` / `$GOPATH/bin`. Fresh installs
then miss `~/.local/bin` (common on developer PATH), and older copies earlier on
PATH stay stale.

Prefer a project **`script/…/install`** that stages embeds (see
`go-embed-assets`) then installs with
[`gotool/localbin/install`](https://pkg.go.dev/github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install).

## Rules

1. **Prefer the install script** over plain `go install` / `go build` when the
   binary must embed fat assets or you care which PATH entry is updated.
2. **Primary dest**
   - existing `LookPath(<bin>)` if found, else
   - **`~/.local/bin/<bin>`** (create the directory; not GOPATH first).
3. **Write with `go build -o <dest>`** so any LookPath path can be overwritten.
4. **Binary name** matches `go install`: last element of the package path;
   module-root package (`.`) → module path basename.
5. **Extra copies:** if these paths **already have** the binary and differ from
   primary, refresh them too: `~/.local/bin/<bin>`, `$GOBIN/<bin>`,
   `$GOPATH/bin/<bin>`.
6. **PATH:** when writing under default `~/.local/bin`, call
   `shell/localbin.EnsureOnPATH` (rc marker; best-effort warnings).
7. **macOS:** `codesign --force --sign -` every written path; on failure print
   `warning:` and continue.

## Library API

```go
import localinstall "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install"

res, err := localinstall.Install(localinstall.Options{
    Dir:     moduleRoot,
    Package: "./cmd/mytool", // or "." for root main
    Stdout:  os.Stdout,
    Stderr:  os.Stderr,
})
// res.Primary, res.Extras, res.BinName, res.PATHEnsured
```

Optional `BinName` overrides go-install naming. Inject `LookPath` / `Build` /
`CodeSign` / `EnsurePATH` in tests.

## Anti-patterns

| Anti-pattern | Prefer |
| ------------ | ------ |
| Fresh install → only `$GOPATH/bin` | Primary `~/.local/bin` + EnsureOnPATH |
| `go install` only, assume PATH hits GOBIN | `localbin/install` |
| Leave stale `GOPATH/bin` copy when refreshing PATH hit | ExtraTargets existence refresh |
| Overwrite without codesign (darwin) | Ad-hoc codesign after every write |
| Agents/docs saying `go build` for fat embed | Document `go run ./script/…/install` |

## Cross-links

- Fat embed staging before install: `go-embed-assets` (Layer 3)
- Thin `cmd/` / `script/` layout: parent `cli/project-layout`
- PATH rc helper: `shell/localbin.EnsureOnPATH`
