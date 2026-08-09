# Scenario

**Feature**: global dry-run resolves install target under HOME

```
# HOME is isolated temp dir for global scope (child env)
user -> go-best-practice skill --install --global --dry-run -> HOME/.agents/skills/go-best-practice
```

## Preconditions

- Child process receives `HOME` set to a temporary directory (not process env).

## Steps

1. Set `req.UseGlobalHome = true`.
2. Set `req.Args = ["skill", "--install", "--global", "--dry-run"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseGlobalHome = true
	req.Args = []string{"skill", "--install", "--global", "--dry-run"}
	return nil
}
```
