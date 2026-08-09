// Command go-best-practice is a thin entrypoint; logic lives in package run.
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/go-best-practice/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
