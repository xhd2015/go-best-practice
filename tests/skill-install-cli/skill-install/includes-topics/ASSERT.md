---
label: e2e
explanation: subprocess CLI real install side effects under isolated work dir
---

## Expected

- Command exits with code 0.
- stdout contains `Installed skill to:`.
- `SKILL.md` exists at `.agents/skills/go-best-practice/SKILL.md`.
- `cli/skill-cli/TOPIC.md` exists under the same skill directory (nested topics use
  TOPIC.md, not nested SKILL.md).
- Nested `cli/skill-cli/SKILL.md` and legacy `topics/skill-cli.md` must not exist.

## Side Effects

- Skill directory and nested TOPIC.md files are created on disk under the work dir.

## Errors

- No error from `Run`.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run failed: %v\nstderr:\n%s", err, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Installed skill to:") {
		t.Fatalf("stdout missing install confirmation:\n%s", resp.Stdout)
	}

	skillDir := filepath.Join(resp.WorkDir, ".agents", "skills", "go-best-practice")
	skillMD := filepath.Join(skillDir, "SKILL.md")
	nested := filepath.Join(skillDir, "cli", "skill-cli", "TOPIC.md")
	nestedOldSkill := filepath.Join(skillDir, "cli", "skill-cli", "SKILL.md")
	legacyTopic := filepath.Join(skillDir, "topics", "skill-cli.md")

	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Fatalf("SKILL.md missing at %s: %v", skillMD, statErr)
	}
	if _, statErr := os.Stat(nested); statErr != nil {
		t.Fatalf("cli/skill-cli/TOPIC.md missing at %s: %v", nested, statErr)
	}
	if _, statErr := os.Stat(nestedOldSkill); statErr == nil {
		t.Fatalf("cli/skill-cli/SKILL.md must not be installed at %s", nestedOldSkill)
	}
	if _, statErr := os.Stat(legacyTopic); statErr == nil {
		t.Fatalf("legacy topics/skill-cli.md must not be installed at %s", legacyTopic)
	}
}
```
