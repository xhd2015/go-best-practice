# go-best-practice

Standalone Go CLI and skill index of best-practice recipes (CLI UX,
flag parsing, embed assets, and more). Extracted from
[`github.com/xhd2015/skills`](https://github.com/xhd2015/skills).

## Install

```bash
go install github.com/xhd2015/go-best-practice/cmd/go-best-practice@latest
```

## Usage

```bash
go-best-practice skill --list
go-best-practice skill --show
go-best-practice skill --show cli/output-alignment
go-best-practice skill cli/color --show
go-best-practice vet -h
```

## Module layout

```
github.com/xhd2015/go-best-practice
├── cmd/go-best-practice/   # thin main only (go install …/cmd/go-best-practice)
├── run/                    # CLI orchestration (skill, topics, vet dispatch)
├── vet/                    # domain package
├── skill/                  # SKILL.md + nested TOPIC.md (//go:embed)
├── script/                 # build, install, bundle, github/release
├── install.sh              # curl installer for GitHub release binaries
└── tests/
```

## Developer install / release

```bash
go run ./script/install          # build then go install ./cmd/go-best-practice
go run ./script/build            # go build -o bin/go-best-practice
go run ./script/bundle           # host OS/arch binary
go run ./script/bundle/for-linux # linux/amd64 binary
go run ./script/github/release --dry-run
```

Depends on `github.com/xhd2015/skills/skillcmd` for skill `--list` /
`--show` / `--install` hosting. See `skill --show cli/project-layout`.
