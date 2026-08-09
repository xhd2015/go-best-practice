---
label: e2e
explanation: subprocess CLI against session-built go-best-practice binary
---

## Expected

- Exit code 0.
- stdout frontmatter / body contains nested name
  `go-best-practice/cli/output-alignment`.
- stdout includes alignment guidance keywords (at least one of: pad, width,
  truncate, column).

## Side Effects

- None.

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
	if !strings.Contains(resp.Stdout, "go-best-practice/cli/output-alignment") {
		t.Fatalf("stdout missing nested name go-best-practice/cli/output-alignment:\n%s", resp.Stdout)
	}
	// Prefer body keywords stronger than "align" (also appears in the path name).
	hasGuidance := strings.Contains(resp.Stdout, "pad") ||
		strings.Contains(resp.Stdout, "width") ||
		strings.Contains(resp.Stdout, "truncate") ||
		strings.Contains(resp.Stdout, "column")
	if !hasGuidance {
		t.Fatalf("stdout missing alignment guidance (pad/width/truncate/column):\n%s", resp.Stdout)
	}
}
```
