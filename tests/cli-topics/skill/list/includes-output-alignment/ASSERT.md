---
label: e2e
explanation: subprocess CLI against session-built go-best-practice binary
---

## Expected

- Exit code 0.
- stdout lists skill name `go-best-practice`.
- stdout contains topic path line `cli/output-alignment`.
- stdout still lists prior topics: `cli`, `cli/color`, `flags-parsing`.
- Trailing newline after last content line (fmt.Println inventory).

## Side Effects

- None (read-only list).

## Errors

- No error from Run (build may fail until product exists → RED).

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
	// Line-aware inventory: skill name + exact topic paths (not substrings of nested paths).
	lines := map[string]bool{}
	for _, line := range strings.Split(resp.Stdout, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			lines[s] = true
		}
	}
	for _, want := range []string{
		"go-best-practice",
		"cli",
		"cli/color",
		"cli/output-alignment",
		"flags-parsing",
	} {
		if !lines[want] {
			t.Fatalf("stdout missing inventory line %q:\n%s", want, resp.Stdout)
		}
	}
}
```
