## Expected

- Non-zero exit code.
- stderr (or stdout) mentions expected action flags such as `--show` and/or `--install`.
- Must not imply legacy word subcommands `skill show` / `skill install` as the only form.

## Side Effects

- None.

## Errors

- Process exits with error status; `Run` returns nil (exit error swallowed).

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
		t.Fatalf("expected non-zero exit, got 0\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "--show") && !strings.Contains(combined, "--install") {
		t.Fatalf("error must mention --show or --install:\n%s", combined)
	}
}
```
