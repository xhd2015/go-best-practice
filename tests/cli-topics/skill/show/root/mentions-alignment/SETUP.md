# Scenario

**Feature**: root skill --show names go-best-practice and indexes cli/output/alignment

```
# root index must surface the new nested topic
user -> go-best-practice skill --show
  -> stdout: go-best-practice + Topics mentioning cli/output/alignment
```

## Preconditions

- Root SKILL.md Topics section includes `cli/output/alignment` (or equivalent index path).

## Steps

1. Set Args to `skill --show`.

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
