---
name: go-best-practice/cli/skill-cli
description: >-
  Skill CLI shapes: single-skill, multi-skill host, topic discovery.
---

# skill-cli — skill CLI shapes and templates

How to ship a CLI that embeds one or more `SKILL.md` definitions and supports
skill **actions as flags** (`--show`, `--install`), plus list/update for multi-skill hosts.

Prefer `github.com/xhd2015/skills/skillcmd` (install/update, header helpers,
`SingleSkill` / `Registry`, and parse). Deprecated shims still exist at
`skills/install` and `skills/skill_file`.

## Choose a shape

| Shape | When | Command sketch |
|-------|------|----------------|
| **1. Single-skill CLI** | The binary *is* one skill | `<cli> skill --show` / `skill --install` |
| **2. Multi-skill host** | One binary registers many skills | `<cli> skills` · `skill --show <name>` · `skills update` |
| **3. Topic discovery** | One skill, large tree of nested `TOPIC.md` | Shape 1 + `skill --show <topic>[/<sub>…]` |

Shape 3 extends Shape 1 (same install surface; nested skill dirs under the package root).

---

## Shared: action flags (not subcommands)

There are **no** `show` / `install` subcommands. Actions are flags only:

| Action | Flag |
|--------|------|
| Print skill content | `--show` |
| Install skill files | `--install` |
| List skill name(s) and, for multi-topic trees, all topic paths | `--list` / `-l` (Shape 1/2/3; also `skills` on Shape 2) |

**Both argument orders are valid** (flag before or after positionals):

```text
<cli> skill --show
<cli> skill --show [--header]
<cli> skill --install [OPTIONS] [<dir>]

# multi-skill / topic path — both orders
<cli> skill --show <name-or-path>
<cli> skill <name-or-path> --show
<cli> skill --install <name> [OPTIONS] [<dir>]
<cli> skill <name> --install [OPTIONS] [<dir>]
```

### Parsing rules (all shapes)

1. After `skill`, scan args for action/mode flags; remaining non-flag args are name, topic path, and/or install dir.
2. Exactly one of `--show` | `--install` | `--list` (unless using Shape 2 `skills` / `skills update`).
3. Reject combining `--show` and `--install`.
4. `--header` only with `--show` (YAML frontmatter only via `skillcmd.FormatHeaderWithDelimiters`).
5. Install package flags (`--global`, `--cursor`, `--codex`, `--opencode`, `--general-agents`, `--dry-run`, `--no-override`, `--force`) are passed through to `skillcmd.HandleInstall` / update handlers.
6. **`-h` / `--help` at every level** (required) — see below.

**Recommended strongly**

- `skill --show --header` (and `<path> --show --header` / `--show <path> --header`)
- `skill --list` / `-l`
- On Shape 1/3, top-level `install …` as alias of `skill --install …` is fine if you want a short path; document it in `--help` only, not in `SKILL.md`

---

## Shared: every command level needs `--help`

Users discover nested CLIs with help. **Do not** only implement root
`<cli> --help`. Every dispatch level must respond to `-h` / `--help` with
**that level's** usage.

| Level | Example | Shows |
|-------|---------|--------|
| Binary root | `<cli> --help` | top commands (`skill`, domain cmds, …) |
| `skill` entry | `<cli> skill --help` | `--show` / `--install` / `--list` (skill surface) |
| Install action | `<cli> skill --install --help` | install targets (`--global`, `--cursor`, …) |
| Shape 2 `skills` | `<cli> skills --help` | list vs `skills update` |
| Shape 2 update | `<cli> skills update --help` | update flags |

### Rules (skillcmd)

1. `skill -h` / `skill --help` (no action) → **skill-level** help  
   (`SingleSkill.Help` or `DefaultSingleSkillHelp`; multi-skill: `Registry.Help`).
2. `skill --show --help` / `skill --list --help` → same **skill-level** help  
   (explore without requiring a path/name).
3. `skill --install --help` → **install-level** help (pass `--help` through to
   `HandleInstall`; do not swallow it as skill-level only).
4. `skills -h` / `skills --help` → **skills-level** help (`Registry.SkillsHelp`).
5. Root help text should mention:  
   `Run <cli> skill --help` and `Run <cli> skill --install --help`.
6. Errors that require an action should hint `(try --help)`.
7. **Multi-topic / Shape 3:** `skill --list` and `skill --help` must surface
   **all available topic paths** (not only the root skill name).  
   - `--list`: print skill name, then each nested path (`flags-parsing`,  
     `flags-parsing/types`, …), one per line.  
   - `--help`: after usage text, append an `Available topics:` index  
     (via `ListTreeTopics` / `FormatTopicIndex` on `TreeFS`).  
   Flat single-skill CLIs (no tree) still list only the skill name.

Same principle applies to **any** word sub-command outside skillcmd (see
`flags-parsing/subcommand`): each `case "cmd":` handler wires its own
`lessflags.Help("-h,--help", …)`.

### Commands (help exploration)

```text
<cli> --help
<cli> skill --help
<cli> skill -h
<cli> skill --install --help
# Shape 2
<cli> skills --help
<cli> skills update --help
```

---

## Shared: SKILL.md rules

`SKILL.md` is what the agent follows after install or `--show`.

```yaml
---
name: my-skill
description: >-
  What it does, plus trigger phrases (e.g. Use when the user runs /my-skill).
---

# My Skill

Imperative workflow for the agent…
```

**Required frontmatter:** `name`, `description` (include triggers / when-to-use).

**Do not put in `SKILL.md`:**

- `skill --install` usage or examples
- Install target flags as product docs (`--cursor`, `--global`, …)
- Those belong in CLI `--help` and project `README.md`

**Do put in `SKILL.md`:** domain workflow, decision rules, and (Shape 3 only)
**topic retrieve examples** using `--show` (both orders optional in examples).

---

## Shared: install targets and flags

`skillcmd.HandleInstall` writes:

```text
.<tool>/skills/<SkillDirName>/
├── SKILL.md
└── <ExtraFiles…>          # optional (nested TOPIC.md for Shape 3)
```

| Selection | Path (cwd-relative) | With `--global` |
|-----------|---------------------|-----------------|
| **default** (no flag, no dir) | `.agents/skills/<name>/` | `~/.agents/skills/<name>/` |
| `--general-agents` | same as default | under `$HOME` |
| `--cursor` | `.cursor/skills/<name>/` | `~/.cursor/skills/<name>/` |
| `--codex` | `.codex/skills/<name>/` | `~/.codex/skills/<name>/` |
| `--opencode` | `.opencode/skills/<name>/` | `~/.opencode/skills/<name>/` |
| positional `<dir>` | that path | home-joined if relative |

**Update** (Shape 2): refresh only directories that already have `SKILL.md`; skip missing installs. Use `install.HandleUpdate` / `HandleUpdateMany`.

---

## Shape 1 — Single-skill CLI

The binary owns exactly one skill (fixed `SkillDirName`, usually the CLI name).

### Commands

```text
<cli> skill --help
<cli> skill --show [--header]
<cli> skill --install [OPTIONS] [<dir>]
<cli> skill --install --help

# recommended strongly
<cli> skill --list
```

### Layout

```text
my-skill/
├── SKILL.md
└── main.go
```

### Brief main.go

```go
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/skills/install"
	"github.com/xhd2015/skills/skill_file"
)

//go:embed SKILL.md
var skillContent string

const skillName = "my-skill"

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		return nil
	}
	// recommended strongly: top-level install alias
	if args[0] == "install" {
		return handleInstall(args[1:])
	}
	if args[0] != "skill" {
		return fmt.Errorf("unknown command: %s", args[0])
	}
	return handleSkill(args[1:])
}

func handleSkill(args []string) error {
	show, installMode, list, header, rest := parseSkillFlags(args)
	n := 0
	if show {
		n++
	}
	if installMode {
		n++
	}
	if list {
		n++
	}
	if n != 1 {
		return fmt.Errorf("expected exactly one of --show, --install, --list")
	}
	if list {
		fmt.Println(skillName)
		return nil
	}
	if show {
		if len(rest) > 0 {
			return fmt.Errorf("unexpected arguments: %v", rest)
		}
		if header {
			out, err := skill_file.FormatHeaderWithDelimiters(skillContent)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		}
		fmt.Print(skillContent)
		return nil
	}
	return handleInstall(rest)
}

// parseSkillFlags accepts --show/--install/--list/--header in any order.
// Other args (including install flags like --global) stay in rest.
func parseSkillFlags(args []string) (show, installMode, list, header bool, rest []string) {
	for _, a := range args {
		switch a {
		case "--show":
			show = true
		case "--install":
			installMode = true
		case "--list", "-l":
			list = true
		case "--header":
			header = true
		default:
			rest = append(rest, a)
		}
	}
	return
}

func handleInstall(args []string) error {
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: skillName,
		SkillContent: skillContent,
		Usage:        skillName + " skill --install",
	}, args)
}

const help = `Usage: my-skill skill --show [--header]
       my-skill skill --install [OPTIONS] [<dir>]
       my-skill skill --list
`
```

---

## Shape 2 — Multi-skill host CLI

One binary embeds a **registry** of skills. List, show/install by name, and
batch-update already-installed copies.

### Commands

```text
# help at each level
<cli> skill --help
<cli> skills --help
<cli> skills update --help

# list — aliases
<cli> skills
<cli> skill --list
<cli> skill -l

# per skill — both orders
<cli> skill --show <name>
<cli> skill <name> --show

<cli> skill --install <name> [OPTIONS] [<dir>]
<cli> skill <name> --install [OPTIONS] [<dir>]
<cli> skill --install <name> --help

# refresh installs that already have SKILL.md (local and/or --global)
<cli> skills update [OPTIONS] [<dir>]
```

### Layout

```text
agents/
  foo/run/SKILL.md      # + small Go file with //go:embed
  bar/run/SKILL.md
cmd/my-host/
  main.go
```

Each skill package exports embedded content, e.g. `var SkillFile string`.

### Brief main.go (registry)

```go
package main

import (
	"fmt"
	"os"
	"strings"

	foo_run "example.com/my-host/agents/foo/run"
	bar_run "example.com/my-host/agents/bar/run"
	"github.com/xhd2015/skills/install"
	"github.com/xhd2015/skills/skill_file"
)

type skillInfo struct {
	Name        string
	Description string
	Content     string
}

var knownSkills = map[string]skillInfo{
	"foo": {Name: "foo", Description: extractDescription(foo_run.SkillFile), Content: foo_run.SkillFile},
	"bar": {Name: "bar", Description: extractDescription(bar_run.SkillFile), Content: bar_run.SkillFile},
}

func knownSkillNames() []string { return []string{"bar", "foo"} } // stable order

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		return nil
	}
	switch args[0] {
	case "skills":
		return handleSkills(args[1:])
	case "skill":
		return handleSkill(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func handleSkills(args []string) error {
	if len(args) == 0 {
		return listSkills()
	}
	if args[0] == "update" {
		return handleUpdate(args[1:])
	}
	// "skills <name> --show" same as "skill <name> --show"
	return handleSkill(args)
}

func handleSkill(args []string) error {
	show, installMode, list, header, rest := parseSkillFlags(args)
	if list {
		if show || installMode || len(rest) > 0 {
			return fmt.Errorf("--list does not take a skill name")
		}
		return listSkills()
	}
	if show == installMode { // both false or both true
		if show {
			return fmt.Errorf("use only one of --show, --install")
		}
		return fmt.Errorf("expected --show, --install, or --list")
	}
	if len(rest) == 0 {
		return fmt.Errorf("expected skill name")
	}
	name := rest[0]
	sk, ok := knownSkills[name]
	if !ok {
		return fmt.Errorf("unknown skill: %s", name)
	}
	if show {
		if header {
			out, err := skill_file.FormatHeaderWithDelimiters(sk.Content)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		}
		fmt.Print(sk.Content)
		return nil
	}
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: sk.Name,
		SkillContent: sk.Content,
		Usage:        "my-host skill --install " + sk.Name,
	}, rest[1:])
}

func parseSkillFlags(args []string) (show, installMode, list, header bool, rest []string) {
	for _, a := range args {
		switch a {
		case "--show":
			show = true
		case "--install":
			installMode = true
		case "--list", "-l":
			list = true
		case "--header":
			header = true
		default:
			rest = append(rest, a)
		}
	}
	return
}

func listSkills() error {
	fmt.Println("Available skills:")
	for _, name := range knownSkillNames() {
		sk := knownSkills[name]
		fmt.Printf("  %-12s %s\n", name, sk.Description)
	}
	return nil
}

func handleUpdate(args []string) error {
	var skills []install.UpdateSkill
	for _, name := range knownSkillNames() {
		sk := knownSkills[name]
		skills = append(skills, install.UpdateSkill{
			InstallOptions: install.InstallOptions{
				SkillDirName: sk.Name,
				SkillContent: sk.Content,
				Usage:        "my-host skills update",
			},
			Name: name,
		})
	}
	return install.HandleUpdateMany(skills, args)
}

func extractDescription(skillMD string) string {
	if !strings.HasPrefix(skillMD, "---\n") {
		return ""
	}
	rest := skillMD[4:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return ""
	}
	lines := strings.Split(rest[:end], "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "description:") {
			continue
		}
		desc := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		if desc == ">-" || desc == ">" {
			for j := i + 1; j < len(lines); j++ {
				if lines[j] == "" || (lines[j][0] != ' ' && lines[j][0] != '\t') {
					break
				}
				return strings.TrimSpace(lines[j])
			}
			return ""
		}
		return desc
	}
	return ""
}

const help = `Usage: my-host skills
       my-host skill --list
       my-host skill --show <name>
       my-host skill <name> --show
       my-host skill --install <name> [OPTIONS]
       my-host skills update
`
```

Embed helper per skill (`agents/foo/run/foo.go`):

```go
package run

import _ "embed"

//go:embed SKILL.md
var SkillFile string
```

---

## Shape 3 — Single skill with nested topic `TOPIC.md`s

Same as Shape 1 for install/list/header, plus a **directory tree of topics**.
Each nested node is a directory containing `TOPIC.md` (not nested `SKILL.md`,
not `topics/<name>.md`). Nested `TOPIC.md` files are **not** agent-discovered
skills after install.

- Root `SKILL.md` is the **index** (topic list + `--show` retrieve commands).
- Sub-topic path `a/b` maps to file `a/b/TOPIC.md`.
- Frontmatter `name` follows **`{root}/{sub-path}`** (slash-separated).

### Commands

```text
<cli> skill --help                            # usage + Available topics: index
<cli> skill --show [--header]                 # root index
<cli> skill --show <topic>[/<sub>…]           # nested TOPIC.md
<cli> skill <topic>[/<sub>…] --show           # same (both orders)
<cli> skill --install [OPTIONS] [<dir>]       # root SKILL.md + all nested TOPIC.md
<cli> skill --install --help

# required for multi-topic: skill name + every nested path
<cli> skill --list
# e.g.
#   my-skill
#   cli
#   cli/color
#   cli/skill-cli
#   flags-parsing
#   flags-parsing/types
```

### Layout

```text
my-skill/
├── SKILL.md                          # name: my-skill
├── main.go
├── cli/
│   ├── TOPIC.md                      # name: my-skill/cli
│   ├── color/
│   │   └── TOPIC.md                  # name: my-skill/cli/color
│   └── skill-cli/
│       └── TOPIC.md                  # name: my-skill/cli/skill-cli
└── flags-parsing/
    ├── TOPIC.md                      # name: my-skill/flags-parsing
    ├── types/
    │   └── TOPIC.md                  # name: my-skill/flags-parsing/types
    └── subcommand/
        └── TOPIC.md                  # name: my-skill/flags-parsing/subcommand
```

After install:

```text
.agents/skills/my-skill/
├── SKILL.md
├── cli/TOPIC.md
├── cli/color/TOPIC.md
├── cli/skill-cli/TOPIC.md
├── flags-parsing/TOPIC.md
├── flags-parsing/types/TOPIC.md
└── flags-parsing/subcommand/TOPIC.md
```

### Index SKILL.md (required pattern)

List topics and show **retrieve** examples (domain commands — allowed here):

    ---
    name: my-skill
    description: >-
      Index of recipes. Load a sub-topic with:
      my-skill skill --show <topic-path>
    ---

    # My Skill

    This skill is an **index**. Load content with:

        my-skill skill --show
        my-skill skill --show cli
        my-skill skill --show cli/color
        my-skill skill --show flags-parsing
        my-skill skill --show flags-parsing/types
        my-skill skill flags-parsing/types --show

    ## Topics

    - `cli` — CLI UX and skill CLI packaging
      - `color` — ANSI color policy
      - `skill-cli` — skill CLI shapes
    - `flags-parsing` — CLI flag parsing
      - `types` — supported target types
      - `subcommand` — dispatcher patterns

### Sub-topic TOPIC.md example

    ---
    name: my-skill/flags-parsing/types
    description: >-
      Supported flag target types for less-flags.
    ---

    # Flag types

    …

### Brief main.go (path → `…/TOPIC.md`)

```go
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/xhd2015/skills/install"
	"github.com/xhd2015/skills/skill_file"
)

//go:embed SKILL.md
var skillContent string

// Embed every top-level skill directory (add more as you add topics).
//
//go:embed cli
//go:embed flags-parsing
var skillTreeFS embed.FS

const skillName = "my-skill"

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(help)
		return nil
	}
	if args[0] == "install" {
		return handleInstall(args[1:])
	}
	if args[0] != "skill" {
		return fmt.Errorf("unknown command: %s", args[0])
	}
	return handleSkill(args[1:])
}

func handleSkill(args []string) error {
	show, installMode, list, header, rest := parseSkillFlags(args)
	n := 0
	if show {
		n++
	}
	if installMode {
		n++
	}
	if list {
		n++
	}
	if n != 1 {
		return fmt.Errorf("expected exactly one of --show, --install, --list")
	}
	if list {
		fmt.Println(skillName)
		return nil
	}
	if installMode {
		return handleInstall(rest)
	}
	// --show: rest is optional topic path (0 or 1 path; join if you allow only one)
	topicPath := ""
	if len(rest) > 0 {
		topicPath = rest[0]
		if len(rest) > 1 {
			return fmt.Errorf("unexpected arguments: %v", rest[1:])
		}
	}
	content, err := loadSkill(topicPath)
	if err != nil {
		return err
	}
	if header {
		out, err := skill_file.FormatHeaderWithDelimiters(content)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}
	fmt.Print(content)
	return nil
}

func parseSkillFlags(args []string) (show, installMode, list, header bool, rest []string) {
	for _, a := range args {
		switch a {
		case "--show":
			show = true
		case "--install":
			installMode = true
		case "--list", "-l":
			list = true
		case "--header":
			header = true
		default:
			rest = append(rest, a)
		}
	}
	return
}

// loadSkill maps "" → root skillContent; "a/b" → a/b/TOPIC.md in skillTreeFS.
func loadSkill(topicPath string) (string, error) {
	topicPath = strings.Trim(topicPath, "/")
	if topicPath == "" {
		return skillContent, nil
	}
	for _, s := range strings.Split(topicPath, "/") {
		if s == "" || s == "." || s == ".." {
			return "", fmt.Errorf("invalid topic path segment: %q", s)
		}
	}
	embedPath := path.Join(topicPath, "TOPIC.md")
	data, err := skillTreeFS.ReadFile(embedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("unknown topic: %s", topicPath)
		}
		return "", err
	}
	return string(data), nil
}

func handleInstall(args []string) error {
	files, err := collectNestedSkillFiles()
	if err != nil {
		return err
	}
	return install.HandleInstall(install.InstallOptions{
		SkillDirName: skillName,
		SkillContent: skillContent,
		ExtraFiles:   files,
		Usage:        skillName + " skill --install",
	}, args)
}

func collectNestedSkillFiles() ([]install.InstallFile, error) {
	var files []install.InstallFile
	err := fs.WalkDir(skillTreeFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		// Nested topics use TOPIC.md so agent loaders do not treat them as skills.
		if path.Base(p) != "TOPIC.md" {
			return nil
		}
		if p == "TOPIC.md" {
			return nil
		}
		data, err := skillTreeFS.ReadFile(p)
		if err != nil {
			return err
		}
		files = append(files, install.InstallFile{Path: p, Content: data})
		return nil
	})
	return files, err
}

const help = `Usage: my-skill skill --show [--header]
       my-skill skill --show <topic>[/<sub>...]
       my-skill skill <topic>[/<sub>...] --show
       my-skill skill --install [OPTIONS] [<dir>]
`
```

### Name convention checklist (Shape 3)

| File | `name` field |
|------|----------------|
| `./SKILL.md` | `my-skill` |
| `flags-parsing/TOPIC.md` | `my-skill/flags-parsing` |
| `flags-parsing/types/TOPIC.md` | `my-skill/flags-parsing/types` |

Pattern: **`{root-skill-name}/{slash-separated-directory-path}`**.

Intermediate directories that have children still ship their own `TOPIC.md` so
`skill --show flags-parsing` always resolves.

---

## Checklist

**All shapes**

1. `SKILL.md` has `name` + trigger-rich `description`
2. No install plumbing in skill body (Shape 3 may document `--show` retrieve commands)
3. Actions are **`--show` / `--install` / `--list` only** (no `show`/`install` subcommands)
4. Both flag orders work (flag before or after name/path)
5. Default install dir is `.agents/skills/<name>/`
6. **`--help` at every level:** root, `skill`, `skill --install`, and (Shape 2) `skills` / `skills update`
7. Set `SingleSkill.Help` / `Registry.Help` (or accept defaults); install help via `HandleInstall`
8. **Multi-topic:** `--list` and `--help` include full topic path inventory from `TreeFS`
9. **Recommended strongly:** `--header` with `--show`, `--list`

**Shape 2**

10. Stable `knownSkillNames()` for list/update order  
11. `skills` ≡ `skill --list` / `-l`  
12. `skills update` via `HandleUpdateMany`  
13. Register every skill’s embed + help text lists  
14. `skills --help` and `skills update --help` work  

**Shape 3**

15. Every nested node is `<path>/TOPIC.md` (not nested `SKILL.md`, not `topics/<path>.md`)  
16. Frontmatter `name` is `root/sub/…` matching the directory path  
17. `skill --show a/b` reads `a/b/TOPIC.md`  
18. Install passes nested `TOPIC.md` files as `ExtraFiles`  
19. Reject empty / `.` / `..` path segments  
20. Index root lists topics and `--show` examples  
21. `skill --list` / `skill --help` enumerate full topic inventory (`ListTreeTopics`)  

---

## See also

- `cli/color` — terminal ANSI color policy for skill CLIs
- `cli/streaming` — stream CLI output as work proceeds
- `flags-parsing/subcommand` — sub-command dispatch + **every level needs `--help`**
- `flags-parsing` — less-flags `Help(...)` options

Reveal with:

```bash
go-best-practice skill --show cli/skill-cli
```
