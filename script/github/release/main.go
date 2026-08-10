// usage: go run ./script/github/release [--dry-run]
//
// Proposed behavior (sketch):
//   1. Parse --dry-run / --help flags.
//   2. Resolve tag and credentials (soft-warn on dry-run; hard-fail live).
//   3. Plan artifact names with the same formula as BuildRelease.
//   4. Dry-run: print plan without building or uploading.
//   5. Live: build multi-platform assets, create/upload GitHub Release.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/github"
	"github.com/xhd2015/kool/pkgs/release"
	"github.com/xhd2015/less-flags"
)

const (
	releaseName = "go-best-practice"
	packagePath = "./cmd/go-best-practice"
)

const help = `
Usage: go run ./script/github/release [--dry-run]

Release go-best-practice to GitHub Releases.

Options:
  --dry-run    print what would be done without actually building or uploading
  -h,--help    show help message
`

func main() {
	if err := handle(); err != nil {
		fmt.Fprintf(os.Stderr, "go-best-practice release: %v\n", err)
		os.Exit(1)
	}
}

func handle() error {
	var dryRun bool
	args, err := lessflags.
		Bool("--dry-run", &dryRun).
		Help("-h,--help", help).
		Parse(os.Args[1:])
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	if dryRun {
		return dryRunRelease()
	}
	return liveRelease()
}

func dryRunRelease() error {
	tag, err := release.GetTag()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		tag = "(unknown)"
	}
	creds, err := release.LoadCredentials(".upload-credentials.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		creds = &release.Credentials{Owner: "xhd2015", Repo: "go-best-practice"}
	}
	fmt.Printf("[dry-run] tag: %s\n", tag)
	for _, spec := range release.DefaultSpecs {
		fmt.Printf("[dry-run] would build: %s-%s-%s-%s\n", releaseName, tag, spec.OS, spec.Arch)
	}
	fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n", creds.Owner, creds.Repo)
	return nil
}

func liveRelease() error {
	result, err := release.BuildRelease(releaseName, nil, release.DefaultSpecs, release.WithPackagePath(packagePath))
	if err != nil {
		return err
	}
	creds, err := release.LoadCredentials(".upload-credentials.json")
	if err != nil {
		return err
	}
	client := github.NewReleaseClient(creds.Token, creds.Owner, creds.Repo)
	rel, err := client.GetOrCreateRelease(result.Tag)
	if err != nil {
		return fmt.Errorf("failed to get or create release for tag %s: %v", result.Tag, err)
	}
	for _, file := range result.Files {
		if err := client.UploadReleaseAsset(rel.ID, file); err != nil {
			return fmt.Errorf("failed to upload %s: %v", file, err)
		}
		fmt.Printf("Uploaded %s\n", file)
	}
	return nil
}
