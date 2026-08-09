# Scenario

**Bug**: `skill` without an action flag must explain expected flags

```
# missing action under skill
user -> go-best-practice skill -> error mentioning --show / --install / --list
```

## Preconditions

- `go-best-practice` binary is built and on `req.Binary`.

## Steps

1. Set `req.Args = ["skill"]` (no --show/--install/--list).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill"}
	return nil
}
```
