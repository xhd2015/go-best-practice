package vet

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet"
	govetfiles "github.com/xhd2015/dot-pkgs/go-pkgs/govet/files"
	"github.com/xhd2015/gitops/git"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `
Usage: go-best-practice vet [flags] [dirs]

Check codebase for best-practice violations.

Flags:
  --json                 Output violations as JSON
  --all                  Vet all files (default: only changed files)
  --compare-with COMMIT  Compare working tree against COMMIT (cannot use with --all)
  --file-max-lines N     Max lines per file (default 500); 0 disables
  --exclude CHECKER      Exclude a checker (repeatable)
`

func Run(args []string) error {
	var jsonOutput, all, lsFiles bool
	var fileMaxLines int
	var compareWith string
	var excludes []string

	fileMaxLines = 500
	args, err := lessflags.Bool("--json", &jsonOutput).
		Bool("--all", &all).
		Bool("--ls-files", &lsFiles).
		String("--compare-with", &compareWith).
		Int("--file-max-lines", &fileMaxLines).
		StringSlice("--exclude", &excludes).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if compareWith != "" && all {
		return fmt.Errorf("cannot use --compare-with and --all together")
	}
	if fileMaxLines < 0 {
		return fmt.Errorf("--file-max-lines must be non-negative, got %d", fileMaxLines)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	resolvedFiles, err := resolveVetFiles(workDir, all, compareWith, args)
	if err != nil {
		return err
	}

	if lsFiles {
		for _, f := range resolvedFiles {
			fmt.Println(f)
		}
		return nil
	}

	cfg := govet.Config{
		FileMaxLines:    fileMaxLines,
		ExcludeCheckers: excludes,
		Files:           resolvedFiles,
	}

	violations, err := govet.Run(cfg)
	if err != nil {
		return err
	}
	return printViolations(violations, jsonOutput)
}

func resolveVetFiles(workDir string, all bool, compareWith string, args []string) ([]string, error) {
	if len(args) == 0 {
		if !all || compareWith != "" {
			changedFiles, err := getChangedGoFiles(workDir, compareWith)
			if err != nil {
				return nil, err
			}
			if compareWith != "" {
				return govetfiles.ResolveGoFiles(workDir, changedFiles)
			}
			if len(changedFiles) > 0 {
				return govetfiles.ResolveGoFiles(workDir, changedFiles)
			}
		}
		return govetfiles.ResolveGoFiles(workDir, []string{"."})
	}
	return govetfiles.ResolveGoFiles(workDir, args)
}

func getChangedGoFiles(workDir string, compareWith string) ([]string, error) {
	// Outside a git work tree there is no changed-file set; caller falls back
	// to vetting all files under workDir.
	root, inRepo := lookupGitRoot(workDir)
	if !inRepo {
		return nil, nil
	}
	var paths []string
	var err error
	if compareWith != "" {
		paths, err = git.GetOnDiskChangedFiles(root, git.CompareWith(compareWith), git.ResolvePathsToFiles())
	} else {
		paths, err = git.GetOnDiskChangedFiles(root, git.ResolvePathsToFiles())
	}
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}
	var goFiles []string
	for _, p := range paths {
		goFiles = append(goFiles, filepath.Join(root, p))
	}
	if len(goFiles) == 0 {
		return nil, nil
	}
	return goFiles, nil
}

// lookupGitRoot returns the repo root and whether dir is inside a git work tree.
// A missing/invalid repo is not an error for callers that fall back to all files.
func lookupGitRoot(dir string) (string, bool) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

func printViolations(violations []govet.Violation, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(violations)
	}

	for _, v := range violations {
		fmt.Printf("%s:%d:%d: [%s] %s\n", v.File, v.Line, v.Col, v.Checker, v.Message)
		if v.Hint != "" {
			fmt.Printf("  → %s\n", v.Hint)
		}
	}
	return nil
}
