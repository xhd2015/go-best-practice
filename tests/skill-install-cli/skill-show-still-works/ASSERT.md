## Expected

- Command exits with code 0.
- stdout contains `go-best-practice`.
- Root index may also mention newer topics (e.g. `cli/output/alignment`); that
  must not break this assert.

## Side Effects

- None beyond process stdout.

## Errors

- No error from `Run`.

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
		t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "go-best-practice") {
		t.Fatalf("stdout missing skill name:\n%s", resp.Stdout)
	}
}
```
