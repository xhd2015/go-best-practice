---
label: e2e
explanation: subprocess CLI against session-built go-best-practice binary
---

## Expected

- Exit code 0 (less-flags Help prints and exits 0).
- stdout mentions `vet` (usage).
- stdout mentions best-practice and/or violations.

## Side Effects

- None.

## Errors

- No error from Run (ExitError is not returned for exit 0).

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
	out := resp.Stdout
	if !strings.Contains(out, "vet") {
		t.Fatalf("stdout missing vet usage marker:\n%s", out)
	}
	if !strings.Contains(out, "best-practice") && !strings.Contains(out, "violations") {
		t.Fatalf("stdout missing best-practice / violations guidance:\n%s", out)
	}
}
```
