# Scenario

**Feature**: nested cli/output-alignment topic via skill --show (both flag orders)

```
# flag before path
user -> go-best-practice skill --show cli/output-alignment -> nested TOPIC.md

# path before flag
user -> go-best-practice skill cli/output-alignment --show -> same nested body
```

## Preconditions

- Embed includes `cli/output-alignment/TOPIC.md` with frontmatter name
  `go-best-practice/cli/output-alignment` and alignment guidance body.

## Steps

1. Leaves set Args for each flag order.
