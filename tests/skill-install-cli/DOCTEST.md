# Skill Install CLI — go-best-practice (coverage backfill)

Doc-style e2e tests for `skill --install` / `skill --show` / top-level `install`
on the standalone `go-best-practice` module CLI. Ported from the skills monorepo
`tests/skill-install-cli/go-best-practice/*` leaves. Covers dry-run (local and
global), real install with nested `cli/skill-cli/TOPIC.md`, show root and nested
(both flag orders), bare-path / legacy / no-action errors, and the top-level
install alias.

**Layer:** subprocess CLI against session-built `./cmd/go-best-practice` with
isolated `cmd.Dir` / `HOME` for install leaves. Only the full-integration real
install leaf is labeled `e2e` (L3 ≤10%); short dispatch / dry-run / show /
error leaves are unlabeled for fast discovery.

# DSN (Domain Specific Notion)

### Participants

- **User / harness** — runs the session-built `go-best-practice` binary with
  skill or top-level install argv (this module only; no skills monorepo binary).
- **CLI router (`handle`)** — dispatches top-level `skill`, `install` alias, and
  related commands.
- **skillcmd SingleSkill** — classifies `--show` / `--install` / `--list` action
  flags and remaining path/install flags; rejects bare paths and word forms.
- **Install resolver** — targets local `.agents/skills/go-best-practice`, global
  `~/.agents/skills/go-best-practice` (under isolated `HOME`), or explicit dir;
  dry-run prints `[dry-run]` without writing.
- **Embed FS** — root `SKILL.md` plus nested `path/TOPIC.md` tree (including
  `cli/skill-cli` and newer topics such as `cli/output-alignment`).

### Behaviors

- **Install dry-run (local)** — `skill --install --dry-run` prints `[dry-run]` and
  mentions `.agents/skills/go-best-practice`; creates no files under work dir.
- **Install dry-run (global)** — `skill --install --global --dry-run` resolves the
  absolute path under isolated `HOME`.
- **Install real** — `skill --install` writes `SKILL.md` and nested
  `cli/skill-cli/TOPIC.md` (not nested `SKILL.md`, not legacy `topics/*.md`).
- **Show root / nested** — `skill --show` prints skill content; nested
  `cli/skill-cli` works with flag-before-path and path-before-flag orders.
- **Top-level install alias** — `install --dry-run` matches `skill --install
  --dry-run` default target.
- **Errors** — bare `skill <path>`, legacy `skill show`, and bare `skill` exit
  non-zero and mention expected action flags.

### Pipeline sketch

```
req.Args (skill --install|…|install --dry-run)
  -> session-cached go build ./cmd/go-best-practice
  -> exec.Command(binary, args...)  // cmd.Dir work dir; cmd.Env HOME + GOWORK=off
  -> Response{Stdout, Stderr, ExitCode, WorkDir, HomeDir}
```

## Decision Tree

```text
skill-install-cli/
├── skill-install/
│   ├── dry-run-default/       # local --dry-run → .agents/skills/go-best-practice
│   ├── global-dry-run/        # --global --dry-run under HOME
│   └── includes-topics/       # real install → SKILL.md + cli/skill-cli/TOPIC.md
├── skill-show-still-works/    # skill --show regression
├── skill-show-nested/
│   ├── flag-before-path/      # skill --show cli/skill-cli
│   └── path-before-flag/      # skill cli/skill-cli --show
├── bare-path-no-action/       # skill cli/skill-cli → error
├── legacy-skill-show/         # skill show → error
├── top-level-install-alias/   # install --dry-run ≡ skill --install --dry-run
└── unknown-subcommand/        # skill alone → error with action hints
```

## Test Index

| Leaf | Args | Expected markers (subset) |
|------|------|---------------------------|
| `skill-install/dry-run-default` | `skill --install --dry-run` | exit 0; `[dry-run]`; `.agents/skills/go-best-practice`; no files |
| `skill-install/global-dry-run` | `skill --install --global --dry-run` | exit 0; `[dry-run]`; absolute `HOME/.agents/skills/go-best-practice` |
| `skill-install/includes-topics` | `skill --install` | exit 0; `Installed skill to:`; `SKILL.md` + `cli/skill-cli/TOPIC.md` |
| `skill-show-still-works` | `skill --show` | exit 0; stdout contains `go-best-practice` |
| `skill-show-nested/flag-before-path` | `skill --show cli/skill-cli` | exit 0; `skill-cli`; name `go-best-practice/cli/skill-cli` |
| `skill-show-nested/path-before-flag` | `skill cli/skill-cli --show` | same nested markers as flag-before-path |
| `bare-path-no-action` | `skill cli/skill-cli` | non-zero; mentions `--show` or `--install` |
| `legacy-skill-show` | `skill show` | non-zero; does not succeed as full body show |
| `top-level-install-alias` | `install --dry-run` | exit 0; `[dry-run]`; `.agents/skills/go-best-practice` |
| `unknown-subcommand` | `skill` | non-zero; mentions `--show` or `--install` |

## How to Run

```sh
doctest vet ./tests/skill-install-cli
doctest test -v ./tests/skill-install-cli
doctest test -v --label e2e ./tests/skill-install-cli   # includes-topics only
```

Coverage backfill: product already implements install/show/alias in this module
— expect **GREEN** after harness points at `./cmd/go-best-practice` (no monorepo
binary). Do not touch sealed `./tests/cli-topics`.


## Version

0.0.2

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

// Request drives one CLI invocation against the session-built binary.
// Leaves set Args and optional scope flags; Binary/SkillName filled by root Setup.
type Request struct {
	Binary        string
	SkillName     string
	Args          []string
	UseGlobalHome bool
	UseWorkDir    bool
}

// Response captures subprocess output, exit status, and isolated scope dirs.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	WorkDir  string
	HomeDir  string
}

// stripEnvPrefix drops KEY= entries from environ so a later KEY=value wins cleanly.
func stripEnvPrefix(environ []string, prefix string) []string {
	out := make([]string, 0, len(environ))
	for _, e := range environ {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
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

	resp := &Response{}
	// Injectable child env only — no t.Setenv / os.Setenv / process chdir.
	env := stripEnvPrefix(os.Environ(), "GOWORK=")
	env = append(env, "GOWORK=off")

	if req.UseWorkDir {
		resp.WorkDir = t.TempDir()
	}
	if req.UseGlobalHome {
		resp.HomeDir = t.TempDir()
		env = stripEnvPrefix(env, "HOME=")
		env = append(env, "HOME="+resp.HomeDir)
	}

	cmd := exec.Command(req.Binary, req.Args...)
	cmd.Env = env
	if resp.WorkDir != "" {
		cmd.Dir = resp.WorkDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
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
