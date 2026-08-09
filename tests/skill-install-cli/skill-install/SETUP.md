# Scenario

**Feature**: `go-best-practice skill --install` resolves targets and nested extras

```
# dry-run previews install without writing files
user -> go-best-practice skill --install --dry-run -> [dry-run] stdout

# real install writes SKILL.md and nested cli/skill-cli/TOPIC.md
user -> go-best-practice skill --install -> .agents/skills/go-best-practice/
```

## Preconditions

- `req.Binary` and `req.SkillName` are set by root Setup.
- Default local target is `.agents/skills/go-best-practice` relative to the
  isolated work dir (`cmd.Dir`).

## Steps

1. Each leaf sets `req.Args` for local, global, or real install behavior.
2. Local leaves set `req.UseWorkDir`; global leaf sets `req.UseGlobalHome`.
