# Scenario

**Feature**: vet -h prints usage mentioning vet and best-practice violations

```
user -> go-best-practice vet -h
  -> exit 0, stdout Usage … vet … best-practice violations
```

## Steps

1. Set Args to `vet -h`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"vet", "-h"}
	return nil
}
```
