# Scenario

**Feature**: top-level `install` remains backward-compatible with `skill --install`

```
# legacy entry point still routes to install handler
user -> go-best-practice install --dry-run -> [dry-run] .agents/skills/go-best-practice
```

## Preconditions

- Command runs with `cmd.Dir` set to an isolated work directory.

## Steps

1. Set `req.UseWorkDir = true`.
2. Set `req.Args = ["install", "--dry-run"]` (no `skill` prefix).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseWorkDir = true
	req.Args = []string{"install", "--dry-run"}
	return nil
}
```
