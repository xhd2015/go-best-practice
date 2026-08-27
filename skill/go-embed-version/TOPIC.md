---
name: go-best-practice/go-embed-version
description: >-
  Track VERSION.txt next to //go:embed; stamp it for a build/release,
  then always restore original bytes so git stays a floor.
---

# go-embed-version — stamp VERSION.txt for go:embed, restore after

Ship a **product version** with `//go:embed` of a plaintext
`VERSION.txt`. Git keeps a **floor** (valid `X.Y.Z`). The release
script **temporarily** writes the real version so the compiler bakes
it in, then **always restores** the original bytes.

This is not `go-embed-assets` (generated UI trees, placeholders,
runtime hydrate). Here the file is tracked and becomes the binary
version string.

## Problem

`//go:embed` is compile-time. Release versions often come from a git
tag or a remote feed, not from committing every bump.

| Approach | Failure |
| -------- | ------- |
| Commit every ship | git history tracks ships; easy to forget, floor and last-ship diverge |
| Leave `VERSION.txt` dirty | next `git status` / commit leaks the stamp |
| `-ldflags -X pkg.Version=…` | exact import path; easy to miss the var; value is invisible in the tree |
| Embed `../VERSION.txt` | **illegal** — `go:embed` cannot use `..` |

## Layout

Put the file next to the embed directive (`cli/project-layout`):

```text
run/
  VERSION.txt    # tracked floor, canonical X.Y.Z
  version.go     # //go:embed VERSION.txt
```

```go
package run

import _ "embed"

//go:embed VERSION.txt
var VERSION string
```

Commit a **valid** `X.Y.Z` (not empty). Bare `go install` / module zip
must compile. Trim at use if a trailing newline is present.

`REVISION.txt` uses the **same wrapper** as a second file when you
need a git SHA besides the semver.

## Stamp then restore

One helper around build/release:

1. Read original bytes (or note missing).
2. `defer` restore — runs on success, error, and panic.
3. Stamp (write the release version).
4. Run `go build` / upload / tests that must see the stamped file.
5. Restore **exact** original bytes (including “no trailing newline”).
   If the file was missing, delete it.

Never `git add` the stamp. Restore failure: yellow `warning: restore …`
on stderr; fail the command if the run had otherwise succeeded.

### Sketch

```go
func With(path string, stamp, run func() error) (err error) {
    orig, ok, err := readOriginal(path) // missing → ok=false, err=nil
    if err != nil {
        return err
    }
    defer func() {
        if restoreErr := restore(path, orig, ok); restoreErr != nil {
            // stderr: warning: restore <path>: …
            if err == nil {
                err = fmt.Errorf("restore %s: %w", path, restoreErr)
            }
        }
    }()
    if stamp != nil {
        if err = stamp(); err != nil {
            return err
        }
    }
    if run != nil {
        err = run()
    }
    return err
}

func restore(path string, orig []byte, existed bool) error {
    if !existed {
        err := os.Remove(path)
        if err != nil && !os.IsNotExist(err) {
            return err
        }
        return nil
    }
    return os.WriteFile(path, orig, 0644)
}
```

Release script:

```go
err := With("run/VERSION.txt", func() error {
    return os.WriteFile("run/VERSION.txt", []byte(next+"\n"), 0644)
}, func() error {
    return buildAndUpload() // go:embed sees next
})
```

`--dry-run` still stamps if the dry-run **builds** (same pipeline as
`cli/dry-run`); skip only the upload. Restore still runs.

## Stamp sources

The wrapper is shared. What you write is product-specific:

| Source | When |
| ------ | ---- |
| Nearest git tag `vX.Y.Z` | tagged CLI / library releases |
| Remote feed’s next patch | git is a floor; last ship lives on the feed (`local <= remote` → patch of remote) |
| Committed bump (`--major` / `--set`) | new floor in git — **not** the temporary stamp |

When local is behind remote, stamp **remote’s next patch**, not
`BumpPatch(local)` — local+1 can still be a downgrade.

## Tests

Unit-test the wrapper in a temp dir (no real upload):

| Case | Expect |
| ---- | ------ |
| Success | run sees stamped bytes; file restored |
| `run` error | error returned; file restored |
| Stamp error | `run` not called; file restored |
| Panic | restore still ran (`defer`) |
| Original had no trailing newline | restored bytes equal original |
| File missing before stamp | file deleted after |

## Anti-patterns

| Anti-pattern | Prefer |
| ------------ | ------ |
| `-ldflags -X import.path.VERSION=…` as the default | embed `VERSION.txt` (visible, no import-path fragility) |
| Restore only on success | `defer` restore always |
| `git add VERSION.txt` in the release script | restore; commit floors with a dedicated bump tool |
| Leave the stamp in the working tree | restore exact bytes |
| `//go:embed ../VERSION.txt` | file beside the directive |
| Empty / missing `VERSION.txt` in git | committed valid `X.Y.Z` |

## See also

- `go-embed-assets` — generated trees + hydrate; uses this version as
  the URL pin
- `cli/project-layout` — embed next to files; no `..`
- `cli/dry-run` — stamp still happens when dry-run builds
- `cmd-exec` — running `go build` / generate from the release script

Reveal with:

```bash
go-best-practice skill --show go-embed-version
```
