## Expected

- Command exits with code 0.
- stdout contains `[dry-run]`.
- stdout mentions `.agents/skills/go-best-practice`.

## Side Effects

- No files are created under the work directory.

## Errors

- No error from `Run`.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
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
	if !strings.Contains(resp.Stdout, "[dry-run]") {
		t.Fatalf("stdout missing [dry-run]:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, ".agents/skills/go-best-practice") {
		t.Fatalf("stdout missing default target path:\n%s", resp.Stdout)
	}
	target := filepath.Join(resp.WorkDir, ".agents", "skills", "go-best-practice")
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("dry-run must not create target dir %s", target)
	}
}
```
