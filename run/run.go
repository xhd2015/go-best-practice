// Package run implements the go-best-practice CLI (thin cmd entry calls Main).
package run

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/xhd2015/go-best-practice/skill"
	"github.com/xhd2015/go-best-practice/vet"
	"github.com/xhd2015/skills/skillcmd"
)

const skillName = "go-best-practice"

const help = `
Usage: go-best-practice <command> [ARGS]
       go-best-practice skill --show [<topic>[/<sub-topic>[/...]]]
       go-best-practice skill <topic-path> --show

Commands:
  install [<dir>]          Install SKILL.md + nested topics (alias of skill --install)
  skill --show [--header] [path]
                           Show root SKILL.md or nested path/TOPIC.md
  skill --install [<dir>]  Install SKILL.md + nested TOPIC.md topics to a directory
  skill --list             Print skill name and all nested topic paths
  topics                   List all available topics (same tree as skill --list)
  vet [flags] [dirs]       Check codebase for best-practice violations

Topics are nested directories with TOPIC.md. Address a nested topic with a
slash-separated path, e.g. "flags-parsing/types", via skill --show.

Run go-best-practice skill --help for skill subcommand options.
Run go-best-practice skill --install --help for install flags.

Options:
  -h, --help    Show this help message
`

const skillHelp = `Usage: go-best-practice skill --show [--header] [<topic-path>]
       go-best-practice skill <topic-path> --show [--header]
       go-best-practice skill --install [OPTIONS] [<dir>]
       go-best-practice skill --list

Show the root SKILL.md index or a nested topic (path/TOPIC.md).
Install copies SKILL.md and nested TOPIC.md topics into agent skill directories.
List prints the skill name and every available topic path.
--help also lists available topics (see below).

Examples:
  go-best-practice skill --show
  go-best-practice skill --show flags-parsing/types
  go-best-practice skill flags-parsing/types --show
  go-best-practice skill --list
  go-best-practice skill --install --dry-run
  go-best-practice skill --install --help

Options:
  --show [--header] [path]   Print skill or topic content (header-only with --header)
  --install [OPTIONS] [dir]  Install skill files (see --install --help)
  --list                     Print skill name and all topic paths
  -h, --help                 Show this help and available topics
`

// Main is the CLI entrypoint used by cmd/go-best-practice.
func Main(args []string) error {
	return handle(args)
}

func handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		fmt.Println()
		return printTopicIndex()
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(help)
		fmt.Println()
		return printTopicIndex()
	case "install":
		// top-level install alias → skill --install
		return singleSkill().Handle(append([]string{"--install"}, args[1:]...))
	case "topics", "list":
		return printTopicIndex()
	case "skill":
		return singleSkill().Handle(args[1:])
	case "vet":
		return vet.Run(args[1:])
	default:
		return fmt.Errorf("unknown command: %s (use skill --show <topic> or run `go-best-practice topics`)", args[0])
	}
}

func singleSkill() *skillcmd.SingleSkill {
	return &skillcmd.SingleSkill{
		Name:        skillName,
		RootContent: skill.Root,
		TreeFS:      skill.Tree,
		Usage:       "go-best-practice skill --install",
		Help:        skillHelp,
	}
}

// listTopics returns all topic paths found under skill.Tree (directories
// that contain TOPIC.md), slash-separated, e.g. "flags-parsing",
// "flags-parsing/types".
func listTopics() ([]string, error) {
	return skillcmd.ListTreeTopics(skill.Tree)
}

func printTopicIndex() error {
	topics, err := listTopics()
	if err != nil {
		return err
	}
	fmt.Print(skillcmd.FormatTopicIndex(topics))
	return nil
}

// readTopic resolves a slash-separated topic path against nested TOPIC.md files.
func readTopic(topicPath string) (string, bool, error) {
	topicPath = strings.Trim(topicPath, "/")
	if topicPath == "" {
		return "", false, nil
	}
	segments := strings.Split(topicPath, "/")
	if err := validateSegments(segments); err != nil {
		return "", false, err
	}
	embedPath := path.Join(topicPath, "TOPIC.md")
	data, err := skill.Tree.ReadFile(embedPath)
	if err == nil {
		return string(data), true, nil
	}
	if os.IsNotExist(err) {
		return "", false, nil
	}
	return "", false, fmt.Errorf("read topic %s: %w", topicPath, err)
}

func validateSegments(segments []string) error {
	for _, s := range segments {
		if s == "" || s == "." || s == ".." {
			return fmt.Errorf("invalid topic path segment: %q", s)
		}
	}
	return nil
}

// collectTopicFiles returns nested TOPIC.md files for install ExtraFiles.
func collectTopicFiles() ([]skillcmd.InstallFile, error) {
	var files []skillcmd.InstallFile
	err := fs.WalkDir(skill.Tree, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		p = path.Clean(p)
		if path.Base(p) != "TOPIC.md" {
			return nil
		}
		if p == "TOPIC.md" || p == "./TOPIC.md" {
			return nil
		}
		data, err := skill.Tree.ReadFile(p)
		if err != nil {
			return err
		}
		files = append(files, skillcmd.InstallFile{Path: p, Content: data})
		return nil
	})
	return files, err
}
