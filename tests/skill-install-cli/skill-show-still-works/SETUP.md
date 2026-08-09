# Scenario

**Feature**: `skill --show` continues to work after install flag migration

```
# regression: skill --show prints embedded SKILL.md
user -> go-best-practice skill --show -> stdout contains go-best-practice
```

## Preconditions

- `go-best-practice` binary is built and on `req.Binary`.

## Steps

1. Set `req.Args = ["skill", "--show"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "--show"}
	return nil
}
```
