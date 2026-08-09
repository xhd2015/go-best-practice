# Scenario

**Feature**: skill --show cli/output-alignment prints nested topic (flag before path)

```
user -> go-best-practice skill --show cli/output-alignment
  -> go-best-practice/cli/output-alignment body + alignment guidance
```

## Steps

1. Set Args for flag-before-path order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "--show", "cli/output-alignment"}
	return nil
}
```
