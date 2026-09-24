# Scenario

**Feature**: skill cli/web-like-cli/unified-assets --show prints the unified asset library topic

```
user -> go-best-practice skill --show cli/web-like-cli/unified-assets
  -> nested topic body from the embed Tree
```

## Steps

1. Set Args to the nested storage topic under cli/web-like-cli.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"skill", "--show", "cli/web-like-cli/unified-assets"}
	return nil
}
```
