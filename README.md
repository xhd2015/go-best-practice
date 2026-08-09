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

## Module

```
github.com/xhd2015/go-best-practice
```

Binary package: `cmd/go-best-practice`. Depends on
`github.com/xhd2015/skills/skillcmd` for skill `--list` / `--show` /
`--install` hosting.
