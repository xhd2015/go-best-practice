# Scenario

**Feature**: skill --list prints skill name and nested topic paths

```
# flat inventory from embed TreeFS
user -> go-best-practice skill --list -> go-best-practice + topic paths
```

## Preconditions

- Nested topics include `cli/output-alignment` after implementer port.

## Steps

1. Leaf sets `req.Args` to `skill --list`.
