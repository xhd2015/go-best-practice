---
name: go-best-practice/cli/project-layout
description: >-
  Organize Go CLI source: thin cmd or root main, run package for logic,
  assets and libraries outside cmd. Matches kool create go-cli.
---

# project-layout — organize Go CLI source code

Keep **`package main` thin**. Put flags, dispatch, domain logic, and embed
trees in importable packages. Do **not** grow a fat tree under `cmd/<bin>/`.

This recipe matches `kool create go-cli` and scales to multi-command tools
like kool itself.

## Rules

1. **Entry only in main** — parse nothing heavy; call `run.Main(args)` (or
   equivalent); print errors to stderr; `os.Exit(1)` on error.
2. **Logic in non-main packages** — testable with `go test` without building
   a binary.
3. **Assets out of `cmd/`** — `SKILL.md`, `TOPIC.md`, static files live in a
   dedicated package directory (e.g. `skill/`) with `//go:embed` there.
4. **`cmd/<name>` is optional** — use it when you need a **named install path**
   (`go install module/cmd/name@latest`) or multiple binaries. Still keep each
   `cmd/<name>/main.go` a few lines.

## Anti-patterns

| Anti-pattern | Prefer |
| ------------ | ------ |
| All code under `cmd/foo/` (handlers, vet, embeds, tests) | Thin `cmd/foo/main.go` + `run/`, `vet/`, `skill/` at module root |
| Unexported logic only in `main` (untestable) | `run.Main` / `Run(config)` in a library package |
| Nesting libraries as `cmd/foo/internalbar` when they are product core | Top-level or `internal/` packages |
| Embedding `../skill` from another package | Embed **inside** the package directory that owns the files |

`//go:embed` patterns cannot use `..`. Put the embed directive next to the
files (e.g. `skill/embed.go` next to `skill/SKILL.md`). For a product
version string, put tracked `VERSION.txt` next to the `//go:embed` (see
`go-embed-version`); do not embed `../VERSION.txt`.

## Shape A — small single binary (kool `go-cli` template)

Scaffold:

```bash
kool create go-cli mytool
```

Layout:

```text
mytool/
├── main.go          # package main — thin
├── run/
│   └── run.go       # package run — Main(args), flags, Run(config)
├── go.mod
└── .gitignore
```

```go
// main.go
package main

import (
	"fmt"
	"os"

	"example.com/mytool/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

```go
// run/run.go
package run

func Main(args []string) error {
	// less-flags parse → Run(config)
	return Run(config)
}

func Run(config Config) error {
	// real work
	return nil
}
```

Install (root main):

```bash
go install example.com/mytool@latest
```

## Shape B — named `cmd/<bin>` install path (still thin)

When the module should install as `go install module/cmd/mytool@latest`
(or hosts several binaries), keep **only** the entry under `cmd/`:

```text
module/
├── cmd/mytool/
│   └── main.go      # thin → run.Main
├── run/
│   └── run.go
├── vet/             # domain packages at module root
├── skill/           # embeds: SKILL.md + topics (if skill CLI)
│   ├── embed.go
│   ├── SKILL.md
│   └── cli/...
└── go.mod
```

```go
// cmd/mytool/main.go
package main

import (
	"fmt"
	"os"

	"example.com/module/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

This repo (`go-best-practice`) uses **Shape B**.

## Shape C — multi-command host (kool-scale)

```text
kool/
├── main.go          # switch on args[0] → tools/*
├── tools/           # one package (or tree) per command family
├── pkgs/            # shared helpers
└── cmd/             # optional *extra* binaries only
```

Root `main` dispatches; work never piles into a single fat `cmd` package.

## Skill CLIs

Shape 3 skill trees (`SKILL.md` + nested `TOPIC.md`) should live in an embed
package (e.g. `skill/`), not under `cmd/<bin>/`. Wiring stays in `run` via
`skillcmd.SingleSkill{ RootContent, TreeFS }`. See `cli/skill-cli`.

## Choosing a shape

| Situation | Shape |
| --------- | ----- |
| New small CLI, one binary, simple module path | **A** (`kool create go-cli`) |
| Want `…/cmd/name@latest` or multiple binaries later | **B** |
| Many subcommands, growing tool suite | **C** |

## Testing notes

| Layer | Where |
| ----- | ----- |
| Unit / L1–L2 | `run`, `vet`, helpers — `go test ./...` |
| Binary smoke | `go build -o bin ./cmd/mytool` or root `.` then subprocess (L3 / e2e) |

Prefer testing `run.Main` / `Run` with buffers or pure returns over only
black-box binary tests.

## Out of scope

- Flag parsing details (`flags-parsing`)
- Color / streaming / tables (`cli/output/color`, `cli/output/streaming`, `cli/output/alignment`)
- Skill action flags (`cli/skill-cli`)

## See also

- `kool-create` — `kool create go-cli` and other scaffolds
- `cli/skill-cli` — skill embed + `--show` / `--install` shapes
- `flags-parsing/subcommand` — dispatch + `--help` at every level
- `go-embed-version` — tracked `VERSION.txt` stamp/restore beside the embed

Reveal with:

```bash
go-best-practice skill --show cli/project-layout
```
