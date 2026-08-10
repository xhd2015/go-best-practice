// usage: go run ./script/build (go build -o bin/go-best-practice)
//
// Proposed behavior (sketch):
//   1. Parse optional flags if any (default: native go build).
//   2. Run go build -o bin/go-best-practice for ./cmd/go-best-practice.
//   3. Exit non-zero on build failure.
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	fmt.Println("==> Building")
	return cmd.Debug().Run("go", "build", "-o", "bin/go-best-practice", "./cmd/go-best-practice")
}
