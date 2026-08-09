# Scenario

**Feature**: standalone go-best-practice CLI supports skill --install / --show and install alias

```
# build once per doctest session from this module only
harness -> go build ./cmd/go-best-practice -> session cache binary

# skill install / show / top-level install via subprocess
user -> go-best-practice skill --install|--show … -> skillcmd SingleSkill + embed FS
user -> go-best-practice install --dry-run -> skill --install equivalent
```

## Preconditions

- Module root is `github.com/xhd2015/go-best-practice` (walk up from
  `d.DOCTEST_ROOT` until `go.mod`).
- Product package `./cmd/go-best-practice` exists and embeds root `SKILL.md` plus
  nested topics (including `cli/skill-cli/TOPIC.md`).
- Session cache: `$TMPDIR/skill-install-cli-doctest-<d.DOCTEST_SESSION_ID>/`
  holds the binary, a `.ready` sentinel, and a `.lock` file (flock) so parallel
  leaves share one build.
- Harness never calls the skills monorepo binary.
- Install leaves isolate scope via child `cmd.Dir` (work dir) and/or child
  `HOME=` (global); no process `t.Setenv` / `os.Chdir`.

## Steps

1. Resolve module root from `d.DOCTEST_ROOT`.
2. Build `./cmd/go-best-practice` once per session under file lock.
3. Set `req.Binary` to the cached path; set `req.SkillName = "go-best-practice"`;
   initialize `req.Args` if nil.
4. Leaves append concrete argv and optional `UseWorkDir` / `UseGlobalHome`;
   `Run` execs the binary.

## Context

- Parallel-safe: child `cmd.Env` (`GOWORK=off`, optional `HOME=`) and `cmd.Dir`
  only; no `t.Setenv`, `os.Setenv`, `os.Chdir`, or hijacking `os.Stdout`.
- Coverage backfill (P2): product already correct; GREEN expected.

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
	return filepath.Join(os.TempDir(), "skill-install-cli-doctest-"+d.DOCTEST_SESSION_ID)
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
	req.SkillName = "go-best-practice"
	return nil
}
```
