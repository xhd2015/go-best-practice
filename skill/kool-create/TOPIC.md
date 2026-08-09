---
name: go-best-practice/kool-create
description: >-
  Scaffold new projects with kool create (go-cli, server, react, …).
---

# kool create — scaffold a new project

Install the `kool` CLI:

```bash
go install github.com/xhd2015/kool@latest
```

Scaffold a new project from an embedded template:

```bash
kool create <TEMPLATE> <project-name>
```

## Available templates

| Template | What you get |
| -------- | ------------ |
| **`go-cli`** | Go CLI: thin root `main.go` + `run/` package (less-flags) |
| `server` | Go-only backend (HTTP server layout) |
| `go-react` | Go backend + React frontend in one repo |
| `frontend` | React-only frontend (`bun install`) |
| `react` | Plain React project |
| `electron` | Electron app (`npm install`) |
| `macos-app-go-daemon` | macOS app + Go daemon (when available in your kool version) |

Run `kool create --help` for the full list on your installed kool.

## Examples

```bash
# recommended Go CLI layout (main + run/)
kool create go-cli mytool

# go-only backend (renames go.mod.template -> go.mod,
# rewrites module path from the git remote when possible)
kool create server my-project

# go backend + react frontend in one repo
kool create go-react my-project

# react-only frontend (runs `bun install` automatically)
kool create frontend my-project

# plain react project
kool create react my-project

# electron app (runs `npm install` automatically)
kool create electron my-project
```

## After scaffolding

- **`go-cli`**: `cd mytool && go run .` — layout is thin `main.go` calling
  `run.Main`; put real logic in `run/`. See `cli/project-layout`.
- `frontend`: `cd my-project && bun watch`
- `electron`: `cd my-project && npm run dev`
- `server`: a `main.go` is created with the module path set to your git
  remote (e.g. `github.com/you/repo/my-project`), or to the default
  template path if no remote is found.

Show full help:

```bash
kool create --help
```

## See also

- `cli/project-layout` — when to use root main vs thin `cmd/<bin>`, where
  to put embeds and domain packages

Reveal with:

```bash
go-best-practice skill --show kool-create
```
