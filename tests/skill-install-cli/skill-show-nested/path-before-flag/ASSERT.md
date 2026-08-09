## Expected

- Exit code 0.
- stdout contains skill-cli topic markers (e.g. `skill-cli` and flag action documentation).
- stdout frontmatter `name` contains `go-best-practice/cli/skill-cli` when present.

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
		t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "skill-cli") {
		t.Fatalf("stdout missing skill-cli marker:\n%s", resp.Stdout)
	}
	// Nested frontmatter name convention after Shape 3 migration
	if !strings.Contains(resp.Stdout, "go-best-practice/cli/skill-cli") {
		t.Fatalf("stdout missing nested name go-best-practice/cli/skill-cli:\n%s", resp.Stdout)
	}
}
```
