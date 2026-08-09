# Scenario

**Feature**: skill subcommand surfaces list and show for embedded topics

```
# skill flag actions via skillcmd
user -> go-best-practice skill --list|--show … -> SingleSkill.Handle
```

## Preconditions

- Binary on `req.Binary` from root Setup (this module build only).

## Steps

1. Leaves under `list/` and `show/` set concrete skill argv.
