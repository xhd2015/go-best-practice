---
label: e2e
explanation: subprocess CLI against session-built go-best-practice binary
---

## Expected

- Exit code 0.
- stdout contains skill name `go-best-practice`.
- stdout mentions `cli/output-alignment` in the Topics index (or equivalent
  topic listing in root SKILL.md).

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
	if !strings.Contains(resp.Stdout, "go-best-practice") {
		t.Fatalf("stdout missing skill name go-best-practice:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "cli/output-alignment") {
		t.Fatalf("stdout missing cli/output-alignment in root index:\n%s", resp.Stdout)
	}
}
```
