// usage: go run ./script/install
//
// Builds via ./script/build, then installs with gotool/localbin/install
// (PATH LookPath → go build -o, optional ~/.local/bin mirror, macOS codesign).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	localinstall "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install"
	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := handle(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle() error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	fmt.Println("==> Building")
	if err := cmd.Debug().Dir(root).Run("go", "run", "./script/build"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("==> Installing (localbin)")
	_, err = localinstall.Install(localinstall.Options{
		Dir:     root,
		Package: "./cmd/go-best-practice",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
	if err != nil {
		return err
	}
	fmt.Println("install complete")
	return nil
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}
