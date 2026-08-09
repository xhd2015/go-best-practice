# Scenario

**Feature**: skill --show prints root SKILL.md or nested path/TOPIC.md

```
# root index
user -> go-best-practice skill --show -> SKILL.md (go-best-practice + Topics)

# nested path (both flag orders)
user -> go-best-practice skill --show <path> -> path/TOPIC.md
user -> go-best-practice skill <path> --show -> same nested body
```

## Preconditions

- Root index and nested topics are embedded in the binary.

## Steps

1. Leaves under `root/` and `nested-output-alignment/` set show argv.
