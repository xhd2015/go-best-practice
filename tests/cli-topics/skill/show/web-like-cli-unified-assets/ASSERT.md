---
label: e2e
explanation: subprocess CLI against session-built go-best-practice binary
---

## Expected

- Exit code 0.
- Frontmatter name line `go-best-practice/cli/web-like-cli/unified-assets`.
- Recipe markers: the library layout `images/<id>`, md5 dedup, magic-byte
  sniffing (`meta.json` commit marker), and the delete guard's single
  container registry.

## Side Effects

- None (read-only show).

## Errors

- No error from Run.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run failed: %v\nstderr:\n%s", err, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "go-best-practice/cli/web-like-cli/unified-assets") {
		t.Fatalf("stdout missing nested name go-best-practice/cli/web-like-cli/unified-assets:\n%s", resp.Stdout)
	}
	for _, want := range []string{"images/<id>", "meta.json", "md5", "delete guard"} {
		if !strings.Contains(resp.Stdout, want) {
			t.Fatalf("stdout missing cli/web-like-cli/unified-assets marker %q:\n%s", want, resp.Stdout)
		}
	}
}
```
