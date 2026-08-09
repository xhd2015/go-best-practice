---
name: go-best-practice/cli/dry-run
description: >-
  Dry-run as a side-effect gate on one pipeline: plan-then-apply and
  inline gates. Avoid a separate dry-run function that duplicates logic.
---

# dry-run — one path, gate side effects

`--dry-run` should answer: **what would *this* run do?** It is not a
second program that reimplements discovery, planning, or naming.

Keep **one control flow**. Dry-run only skips irreversible or expensive
side effects (writes, network mutates, multi-platform builds). The same
steps resolve inputs and compute the plan in both modes.

## Why one path

| Concern | Separate `handleDryRun()` | Single path + gate |
| ------- | ------------------------- | ------------------ |
| Drift | Live and dry-run diverge silently | Same plan, same names |
| Trust | Dry-run cannot validate the real path | Dry-run exercises the same steps |
| Cost | Duplicate code to maintain | One place to change |
| Purpose | Answers a simplified story | Answers what this run would do |

## Policy

1. **One function / pipeline** — pass `dryRun bool` (or an options
   struct). Do not branch to a sibling that reimplements the flow.
2. **Same discovery and plan** — tag, inventory, specs, target paths,
   and artifact names come from the same helpers in both modes.
3. **Gate only side effects** — `if dryRun { print plan; return }` or
   `if !dryRun { apply }` after the plan exists. Expensive builds may
   be skipped when dry-run; still print the **same** planned targets.
4. **Error policy** — live hard-fails required steps. Dry-run may
   **soft-fail** optional enrichment (e.g. missing credentials used only
   to print an upload target): warn on stderr, substitute a default, and
   continue. Do not invent a parallel algorithm to “make dry-run work.”
5. **Output** — prefix planned lines with `[dry-run]`; warnings on
   stderr; exit 0 when the plan itself succeeded.

## Anti-pattern → preferred

**Anti-pattern** — separate dry-run function:

```go
func handle() error {
    var dryRun bool
    // parse flags...
    if dryRun {
        return handleDryRun() // reimplements tag, names, targets
    }
    // live-only path: different helpers, easy to drift
    result, err := release.BuildRelease(name, nil, release.DefaultSpecs)
    // ...
    return nil
}

func handleDryRun() error {
    tag, _ := release.GetTag() // soft-fail, different story
    for _, spec := range release.DefaultSpecs {
        fmt.Printf("[dry-run] would build: %s-%s-%s-%s\n", name, tag, spec.OS, spec.Arch)
    }
    return nil
}
```

**Preferred** — same steps; gate at the end (or per effect):

```go
func handle() error {
    var dryRun bool
    // parse flags...

    tag, err := release.GetTag()
    if err != nil {
        if !dryRun {
            return err
        }
        fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
        tag = "(unknown)"
    }

    creds, err := release.LoadCredentials(".upload-credentials.json")
    if err != nil {
        if !dryRun {
            return err
        }
        fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
        creds = &release.Credentials{Owner: "owner", Repo: "repo"}
    }

    // Same naming formula live BuildRelease would use.
    var planned []string
    for _, spec := range release.DefaultSpecs {
        planned = append(planned, fmt.Sprintf("%s-%s-%s-%s", name, tag, spec.OS, spec.Arch))
    }

    if dryRun {
        fmt.Printf("[dry-run] tag: %s\n", tag)
        for _, f := range planned {
            fmt.Printf("[dry-run] would build: %s\n", f)
        }
        fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n",
            creds.Owner, creds.Repo)
        return nil
    }

    result, err := release.BuildRelease(name, nil, release.DefaultSpecs)
    if err != nil {
        return err
    }
    // upload result.Files ...
    return nil
}
```

## Recipes

### Plan then apply

Best when the work is “compute inventory / actions, then mutate”:

```go
actions, err := plan(dir)
if err != nil {
    return err
}
if dryRun {
    printPlan(actions) // [dry-run] lines
    return nil
}
return apply(actions)
```

Same `plan` for both modes. Dry-run never calls `apply`.

### Inline gate

Best for multi-step CLIs (release, generate, sync): walk the same
steps; skip writes / uploads / heavy builds when `dryRun`.

```go
for _, need := range needs {
    if dryRun {
        fmt.Printf("[dry-run] would generate: %s\n", need.Path)
        continue
    }
    if err := generateAndWrite(need); err != nil {
        return err
    }
}
```

### Effect interface (advanced)

When side effects are many or need unit tests, inject them:

```go
type Effects interface {
    Build(name, tag string, specs []*Spec) ([]string, error)
    Upload(owner, repo, tag string, files []string) error
}

// RealEffects does builds and API calls.
// DryRunEffects only prints [dry-run] lines; Build may return planned names.
```

Orchestration stays identical; only the effect implementation changes.
Prefer plan-then-apply or inline gates for small scripts.

## Soft-fail vs hard-fail

| Mode | Required for the real run | Optional enrichment for display |
| ---- | ------------------------- | ------------------------------- |
| Live | Hard-fail | Hard-fail or omit |
| Dry-run | Prefer hard-fail if the run would abort | Soft-warn + default OK |

Example: missing upload credentials — live must fail; dry-run may warn
and print a default `owner/repo` so the rest of the plan is still
visible.

## Output convention

```text
$ mytool release --dry-run
[dry-run] tag: v1.2.3
[dry-run] would build: mytool-v1.2.3-linux-amd64
[dry-run] would upload to acme/mytool release (creates if 404)
```

```text
$ mytool release --dry-run
[dry-run] warning: open .upload-credentials.json: no such file or directory
[dry-run] tag: v1.2.3
...
```

- Planned / would-do lines → **stdout**, `[dry-run]` prefix  
- Soft-fail warnings → **stderr**, `[dry-run] warning:`  
- Exit **0** when planning succeeded (including soft-failed enrichment)

## When a separate preview is OK

Almost never for multi-step CLIs. A static help example or a pure
documentation string that cannot drift is fine. If the preview can
diverge from live behavior, it is not dry-run—merge it into the real
path.

## See also

- `cli/streaming` — print planned units as you go; do not buffer the
  whole dry-run report unless you need a table
- `flags-parsing` — wire `--dry-run` with less-flags
- `cmd-exec` — external commands still run only when not dry-run (or
  print the command line in dry-run)
