# Scenario

**Feature**: skill cli/output/alignment --show prints nested topic (path before flag)

```
user -> go-best-practice skill cli/output/alignment --show
  -> same nested markers as flag-before-path
```

## Steps

1. Set Args for path-before-flag order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "cli/output/alignment", "--show"}
	return nil
}
```
