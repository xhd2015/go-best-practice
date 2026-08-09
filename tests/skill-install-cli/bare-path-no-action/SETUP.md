# Scenario

**Feature**: bare topic path under skill without action is rejected

```
# missing action flag
user -> go-best-practice skill cli/skill-cli -> error (no bare path action)
```

## Steps

1. Set Args `skill cli/skill-cli` without --show/--install.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "cli/skill-cli"}
	return nil
}
```
