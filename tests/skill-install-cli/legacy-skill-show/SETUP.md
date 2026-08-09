# Scenario

**Feature**: legacy word form `skill show` is rejected after flag migration

```
# word subcommand no longer valid
user -> go-best-practice skill show -> error
```

## Steps

1. Set Args `skill show` (legacy).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "show"}
	return nil
}
```
