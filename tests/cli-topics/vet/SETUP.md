# Scenario

**Feature**: vet subcommand checks best-practice violations

```
user -> go-best-practice vet [flags] [dirs] -> violations or help
```

## Preconditions

- Binary on `req.Binary` from root Setup.

## Steps

1. Leaves under `help/` set vet argv.
