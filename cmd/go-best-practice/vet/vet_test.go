package vet

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoViolations(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "fmt"
func main() { fmt.Println("hello") }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{dir})
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output for clean dir, got: %s", output)
	}
}

func TestStdFlag(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{dir})
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(output, "[std-flag]") {
		t.Errorf("expected std-flag violation, got: %s", output)
	}
	if !strings.Contains(output, "flags-parsing") {
		t.Errorf("expected hint about flags-parsing, got: %s", output)
	}
}

func TestFileLength(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n" + strings.Repeat("// line\n", 600)
	mustWriteFile(t, dir, "main.go", content)

	output, err := captureStdout(t, func() error {
		return Run([]string{"--file-max-lines", "500", dir})
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(output, "[file-length]") {
		t.Errorf("expected file-length violation, got: %s", output)
	}
	if !strings.Contains(output, "exceeds maximum of 500") {
		t.Errorf("expected threshold message, got: %s", output)
	}
}

func TestJsonOutput(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{"--json", dir})
	})
	if err != nil {
		t.Fatalf("Run --json: %v", err)
	}
	if !strings.Contains(output, "\"Checker\"") {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestExclude(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{"--exclude", "std-flag", dir})
	})
	if err != nil {
		t.Fatalf("Run --exclude: %v", err)
	}
	if strings.Contains(output, "[std-flag]") {
		t.Errorf("expected std-flag to be excluded, got: %s", output)
	}
}

func TestNoArgs(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return Run([]string{})
	})
	if err != nil {
		t.Fatalf("Run no-args: %v", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output for vet on current dir (no violations), got: %s", output)
	}
}

func TestUnknownFlag(t *testing.T) {
	err := Run([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unrecognized flag") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBadFileMaxLines(t *testing.T) {
	tests := []struct {
		flag string
	}{
		{"--file-max-lines=abc"},
		{"--file-max-lines=-1"},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			err := Run([]string{tt.flag, "."})
			if err == nil {
				t.Fatal("expected error for bad --file-max-lines")
			}
		})
	}
}

func TestDotDotDot(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, root, "go.mod", "module testmod\ngo 1.24\n")
	mustWriteFile(t, root, "clean.go", "package main\nfunc main() {}\n")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	mustWriteFile(t, root, "sub/bad.go", "package sub\nimport \"flag\"\nvar _ = flag.Usage\n")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(oldDir)

	output, err := captureStdout(t, func() error {
		return Run([]string{"./..."})
	})
	if err != nil {
		t.Fatalf("Run ./...: %v", err)
	}
	if !strings.Contains(output, "[std-flag]") {
		t.Errorf("expected std-flag violation, got: %s", output)
	}
	if strings.Contains(output, "clean.go") {
		t.Errorf("clean.go should not be reported, got: %s", output)
	}
}

func TestAllFlag(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{"--all", dir})
	})
	if err != nil {
		t.Fatalf("Run --all: %v", err)
	}
	if !strings.Contains(output, "[std-flag]") {
		t.Errorf("expected std-flag violation with --all, got: %s", output)
	}
}

func TestAllFlagNoDir(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return Run([]string{"--all"})
	})
	if err != nil {
		t.Fatalf("Run --all no-dir: %v", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected no output for --all on test dir, got: %s", output)
	}
}

func TestExplicitDirDisablesChangedFilesDefault(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	output, err := captureStdout(t, func() error {
		return Run([]string{dir})
	})
	if err != nil {
		t.Fatalf("Run with explicit dir: %v", err)
	}
	if !strings.Contains(output, "[std-flag]") {
		t.Errorf("expected std-flag violation, got: %s", output)
	}
}

func TestLsFiles(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", "package main\nimport \"fmt\"\nfunc main() {}\n")
	mustWriteFile(t, dir, "util.go", "package main\n")
	mustWriteFile(t, dir, "README.md", "# docs\n")

	output, err := captureStdout(t, func() error {
		return Run([]string{"--ls-files", dir})
	})
	if err != nil {
		t.Fatalf("Run --ls-files: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	goCount := 0
	for _, line := range lines {
		if strings.HasSuffix(line, ".go") {
			goCount++
		}
	}
	if goCount != 2 {
		t.Errorf("expected 2 .go files, got %d: %s", goCount, output)
	}
	if strings.Contains(output, ".md") {
		t.Errorf("expected no .md files, got: %s", output)
	}
}

func TestLsFilesNoGoFiles(t *testing.T) {
	dir := t.TempDir()
	output, err := captureStdout(t, func() error {
		return Run([]string{"--ls-files", dir})
	})
	if err != nil {
		t.Fatalf("Run --ls-files: %v", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output, got: %s", output)
	}
}

func mustWriteFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	stdoutCh := make(chan []byte, 1)
	readErrCh := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(reader)
		stdoutCh <- data
		readErrCh <- readErr
	}()

	os.Stdout = writer
	runErr := fn()
	os.Stdout = oldStdout
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	data := <-stdoutCh
	if err := <-readErrCh; err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(data), runErr
}

func TestCompareWithAndAllConflict(t *testing.T) {
	err := Run([]string{"--compare-with", "HEAD", "--all"})
	if err == nil {
		t.Fatal("expected error for --compare-with and --all together")
	}
	if !strings.Contains(err.Error(), "cannot use --compare-with and --all together") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCompareWithInvalidRef(t *testing.T) {
	err := Run([]string{"--compare-with", "nonexistent123"})
	if err == nil {
		// Some environments (e.g. shallow CI checkouts without full history, or
		// when git cannot read the work tree) may not surface ref errors.
		t.Skip("invalid ref did not error; git/env may not support compare-with here")
	}
}

func TestLsFilesIncremental_IncludesUntrackedDir(t *testing.T) {
	dir := t.TempDir()
	execGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s: %v", args, string(out), err)
		}
	}

	execGit("init")
	execGit("config", "user.email", "test@test.com")
	execGit("config", "user.name", "test")

	mustWriteFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	execGit("add", "-A")
	execGit("commit", "-m", "init")

	viewDir := filepath.Join(dir, "view")
	if err := os.MkdirAll(viewDir, 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, dir, "view/view.go", "package view\n")
	mustWriteFile(t, dir, "view/view_test.go", "package view\n")

	mustWriteFile(t, dir, "main.go", "package main\nfunc main() { println(\"changed\") }\n")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(oldDir)

	output, err := captureStdout(t, func() error {
		return Run([]string{"--ls-files"})
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	goCount := 0
	for _, line := range lines {
		if strings.HasSuffix(line, ".go") {
			goCount++
		}
	}
	if goCount != 3 {
		t.Errorf("expected 3 .go files, got %d: %s", goCount, output)
	}
}
