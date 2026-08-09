## Expected

- Non-zero exit.
- Error indicates invalid/missing action (word `show` is not an action flag).

## Side Effects

- None.

## Exit Code

- 1

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
		t.Fatalf("Run failed: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for legacy skill show, got 0\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	// Must not successfully print the full skill body as the old `skill show` did.
	if strings.Contains(resp.Stdout, "# Go Best Practice Skill") && resp.ExitCode == 0 {
		t.Fatalf("legacy skill show must not succeed with full body")
	}
	combined := resp.Stdout + resp.Stderr
	if combined == "" {
		t.Fatalf("expected error output for legacy skill show")
	}
}
```
