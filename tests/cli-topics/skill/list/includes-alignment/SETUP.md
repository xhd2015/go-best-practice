# Scenario

**Feature**: skill --list includes cli/output/alignment and prior topics

```
# inventory must list new nested topic and existing ones
user -> go-best-practice skill --list
  -> stdout lines: go-best-practice, cli, cli/output/color, cli/output/alignment, flags-parsing, …
```

## Preconditions

- Embed tree includes `cli/output/alignment/TOPIC.md` plus prior topics.

## Steps

1. Set Args to `skill --list`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "--list"}
	return nil
}
```
