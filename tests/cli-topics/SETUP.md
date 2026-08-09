# Scenario

**Feature**: standalone go-best-practice CLI serves embedded topics and vet

```
# build once per doctest session from this module only
harness -> go build ./cmd/go-best-practice -> session cache binary

# skill list / show and vet help via subprocess
user -> go-best-practice skill --list|--show … -> skillcmd SingleSkill + embed FS
user -> go-best-practice vet -h -> vet usage (exit 0)
```

## Preconditions

- Module root is `github.com/xhd2015/go-best-practice` (walk up from
  `d.DOCTEST_ROOT` until `go.mod`).
- Product package `./cmd/go-best-practice` will exist after implementer port
  (build is expected to fail until then → RED).
- Session cache: `$TMPDIR/cli-topics-doctest-<d.DOCTEST_SESSION_ID>/` holds the
  binary, a `.ready` sentinel, and a `.lock` file (flock) so parallel leaves
  share one build.
- Harness never calls the skills monorepo binary.

## Steps

1. Resolve module root from `d.DOCTEST_ROOT`.
2. Build `./cmd/go-best-practice` once per session under file lock.
3. Set `req.Binary` to the cached path; initialize `req.Args` if nil.
4. Leaves append concrete argv; `Run` execs the binary.

## Context

- Parallel-safe: child `cmd.Env` only (`GOWORK=off`); no `t.Setenv`,
  `os.Setenv`, `os.Chdir`, or hijacking `os.Stdout`.
- Install path is documented by product behavior only:
  `go install github.com/xhd2015/go-best-practice/cmd/go-best-practice@latest`
  (not exercised in this tree — P2).

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func moduleRoot(d *session.Doctest) (string, error) {
	dir := d.DOCTEST_ROOT
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot find go.mod from %s", d.DOCTEST_ROOT)
		}
		dir = parent
	}
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "cli-topics-doctest-"+d.DOCTEST_SESSION_ID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

// buildGoBestPracticeOnce builds ./cmd/go-best-practice once per doctest session.
func buildGoBestPracticeOnce(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cacheDir := sessionCacheDir(d)
	bin := filepath.Join(cacheDir, "go-best-practice")
	ready := filepath.Join(cacheDir, "go-best-practice.ready")
	lock := filepath.Join(cacheDir, "go-best-practice.lock")

	err := withFileLock(t, lock, func() error {
		if _, err := os.Stat(ready); err == nil {
			if _, err := os.Stat(bin); err == nil {
				return nil
			}
		}
		root, err := moduleRoot(d)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/go-best-practice")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("build go-best-practice: %w\n%s", err, out)
		}
		if err := os.Chmod(bin, 0o755); err != nil {
			return err
		}
		return os.WriteFile(ready, []byte("ok"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return bin, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Args == nil {
		req.Args = []string{}
	}
	bin, err := buildGoBestPracticeOnce(t, d)
	if err != nil {
		return err
	}
	req.Binary = bin
	return nil
}
```
