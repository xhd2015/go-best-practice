# Scenario

**Feature**: real install copies nested cli/skill-cli/TOPIC.md alongside root SKILL.md

```
# non-dry-run install writes skill tree (Shape 3 nested paths)
user -> go-best-practice skill --install -> .agents/skills/go-best-practice/SKILL.md + cli/skill-cli/TOPIC.md
```

## Preconditions

- Command runs with `cmd.Dir` set to an isolated work directory with no
  pre-existing install.

## Steps

1. Set `req.UseWorkDir = true`.
2. Set `req.Args = ["skill", "--install"]` (no `--dry-run`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseWorkDir = true
	req.Args = []string{"skill", "--install"}
	return nil
}
```
