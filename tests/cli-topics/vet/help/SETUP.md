# Scenario

**Feature**: vet -h / --help prints usage and exits 0

```
user -> go-best-practice vet -h -> usage (vet / best-practice violations)
```

## Preconditions

- less-flags Help prints usage then exits 0 (no HelpNoExit).

## Steps

1. Leaf sets Args to `vet -h` (or equivalent help flag).
