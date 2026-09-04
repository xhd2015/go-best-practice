# CLI topics surface — go-best-practice extract (P1)

Doc-style e2e tests for the standalone `go-best-practice` CLI after extract
from the skills monorepo. Covers topic inventory (including new
`cli/output/alignment`), nested show (both flag orders), root skill show, and
`vet` help.

**Layer:** L3 e2e — build `./cmd/go-best-practice` once per session, subprocess
CLI. Every leaf is labeled `e2e`.

# DSN (Domain Specific Notion)

### Participants

- **User / harness** — runs the built `go-best-practice` binary with skill or
  vet argv (no install, no monorepo binary).
- **CLI router (`handle`)** — dispatches top-level `skill` and `vet` (and
  related commands after port).
- **skillcmd SingleSkill** — serves embedded root `SKILL.md` and nested
  `path/TOPIC.md` tree for `--list` / `--show` (both flag orders).
- **Embed FS** — root skill index plus topics under `cli/`, `flags-parsing/`,
  etc., including **new** `cli/output/alignment/TOPIC.md`.
- **vet** — best-practice checker subcommand with less-flags help.

### Behaviors

- **List** — `skill --list` prints skill name then flat topic paths (one per
  line), including `cli/output/alignment` and prior topics (`cli`,
  `cli/output/color`, `flags-parsing`, …).
- **Show nested** — `skill --show cli/output/alignment` and
  `skill cli/output/alignment --show` print the nested topic body with
  frontmatter name `go-best-practice/cli/output/alignment` and alignment
  guidance (pad / width / truncate or similar).
- **Show root** — `skill --show` prints root `SKILL.md` naming
  `go-best-practice` and indexing `cli/output/alignment`.
- **Vet help** — `vet -h` / `vet --help` exits 0 and prints usage mentioning
  vet / best-practice violations.

### Pipeline sketch

```
req.Args (e.g. skill --list | skill --show … | vet -h)
  -> session-cached go build ./cmd/go-best-practice
  -> exec.Command(binary, args...)  // cmd.Env GOWORK=off; no process chdir/setenv
  -> Response{Stdout, Stderr, ExitCode}
```

## Decision Tree

```text
cli-topics/
├── skill/
│   ├── list/
│   │   └── includes-alignment
│   └── show/
│       ├── root/
│       │   └── mentions-alignment
│       └── nested-alignment/
│           ├── flag-before-path
│           └── path-before-flag
└── vet/
    └── help/
        └── short-flag
```

## Test Index

| Leaf | Args | Expected markers (subset) |
|------|------|---------------------------|
| `skill/list/includes-alignment` | `skill --list` | `go-best-practice`, `cli/output/alignment`, `cli`, `cli/output/color`, `flags-parsing` |
| `skill/show/root/mentions-alignment` | `skill --show` | `go-best-practice`, `cli/output/alignment` in Topics/index |
| `skill/show/nested-alignment/flag-before-path` | `skill --show cli/output/alignment` | name `go-best-practice/cli/output/alignment`; pad/width/truncate guidance |
| `skill/show/nested-alignment/path-before-flag` | `skill cli/output/alignment --show` | same nested markers as flag-before-path |
| `vet/help/short-flag` | `vet -h` | exit 0; usage mentions `vet` and best-practice / violations |

## How to Run

```sh
doctest vet ./tests/cli-topics
doctest test -v ./tests/cli-topics
doctest test -v --label e2e ./tests/cli-topics
```

Classic TDD: expect **RED** until `cmd/go-best-practice` exists in this module
and embeds `cli/output/alignment` (do not call the skills monorepo binary).

## Version

0.0.2

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// Request drives one CLI invocation against the session-built binary.
// Leaves set Args only; Binary is filled by root Setup.
type Request struct {
	Binary string
	Args   []string
}

// Response captures subprocess stdout, stderr, and exit status.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.Binary == "" {
		return nil, fmt.Errorf("CLI binary path is required")
	}
	if len(req.Args) == 0 {
		return nil, fmt.Errorf("CLI args are required")
	}

	cmd := exec.Command(req.Binary, req.Args...)
	// Injectable child env only — no t.Setenv / os.Setenv / process chdir.
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cmd.ProcessState != nil {
		resp.ExitCode = cmd.ProcessState.ExitCode()
	}
	if runErr != nil {
		if _, ok := runErr.(*exec.ExitError); !ok {
			return resp, fmt.Errorf("run %s: %w", filepath.Base(req.Binary), runErr)
		}
	}
	return resp, nil
}
```
