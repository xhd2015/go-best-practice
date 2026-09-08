---
name: go-best-practice/cli/dry-run
description: >-
  Dry-run as a side-effect gate on one pipeline: same probes as live,
  bypass only the mutate call. Avoid a separate dry-run function.
  Do not prefix lines with [dry-run] — the flag is already on the argv.
---

# dry-run — one path, observe, gate side effects

`--dry-run` should answer: **what would *this* run do against current
state?** It is the **live function** with mutate calls gated — not a
second planner, and not a static script of would-lines from flags alone.

Keep **one control flow**. Dry-run **runs the same preflight / read-only
checks as live**. It **must not** perform stateful mutations (writes,
uploads, destructive remote commands, package installs). The same steps
resolve inputs and compute the plan in both modes; only the mutate step
is gated.

`--dry-run` is already on the command line. **Do not** prefix output
lines with `[dry-run]` — that is noisy and redundant. Use `would:` /
`skip:` / `notice:` / `warning:` (staged CLIs: kind-aligned under
`[n/total]`; see `cli/output/staged-markers`).

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
3. **Same probes as live** — dry-run **runs** the read-only checks live
   uses (`stat` / `test -f`, BatchMode SSH reads, API GET, `spl … run --`
   non-mutating probes). Optional verbose: `notice:` / `probing (read-only):`.
4. **Gate only the mutate call** — after observation, either print the
   **exact** mutate command live would run (`would:`), or an
   evidence-backed skip. Do not paraphrase (`would revoke somehow`).
5. **Exact command fidelity** — `would:` lines use the same argv /
   script / paths the live `apply` / `runOrDry` path would execute.
6. **No-op honesty** — if probes show nothing to change, print
   `skip: … (already absent|already done)` with evidence.
   Do not print a fake would-mutate. Do not stamp `skip (dry-run)`.
7. **Error policy** — live hard-fails required steps. Dry-run soft-fails
   only when the **preflight itself** is unavailable (e.g. network
   down): `warning:` on stderr, then best-effort exact plan from known
   paths/comments. Do not soft-skip observation when a cheap probe
   exists. Do not invent a parallel algorithm to “make dry-run work.”
8. **Output** — `--dry-run` is on the argv; do not repeat `[dry-run]` on
   every line. Planned mutate: `would:`. No-op: `skip:`. Soft-fail:
   `warning:` on stderr. Exit 0 when planning succeeded (including
   soft-failed probes that still produced a plan).

## Allowed vs forbidden under `--dry-run`

| Allowed (preflight) | Forbidden (mutation) |
| ------------------- | -------------------- |
| `test -f`, `ls`, `stat`, read files | `rm`, write files, `chmod` as a “fix write” |
| SSH BatchMode **read** (`grep`, `cat`, `echo`) | SSH that appends/rewrites `authorized_keys` |
| Remote `command -v`, `test -f` via task run | upload, install packages, remote `rm` |
| API GET / status | API create/update/delete |

## Anti-pattern → preferred

**Anti-pattern** — separate dry-run function, or **flag-only** dry-run
(print static would-lines without observing target state):

```go
func handle() error {
    var dryRun bool
    // parse flags...
    if dryRun {
        // Wrong: no probe; paraphrased plan; redundant [dry-run] tag
        fmt.Println("[dry-run] would clean remote keys")
        return nil
    }
    // live-only path...
    return nil
}
```

**Preferred** — same steps; observe; gate mutate:

```go
func handleTeardown(dryRun bool) error {
    // Same path resolution live and dry-run.
    present, err := probeRemoteKey() // read-only; runs even if dryRun
    if err != nil {
        if !dryRun {
            return err
        }
        fmt.Fprintf(os.Stderr, "warning: probe: %v\n", err)
    }
    cmd := exactRemoveCmd() // same argv live would use
    if !present {
        fmt.Printf("skip: remote key already absent\n")
        return nil
    }
    if dryRun {
        fmt.Printf("would: %s\n", cmd)
        return nil
    }
    return run(cmd)
}
```

**Preferred** — plan-then-apply with shared `plan` that embeds probe results:

```go
actions, err := plan(dir) // probes inside plan
if err != nil {
    return err
}
if dryRun {
    printPlan(actions) // would: / skip: from observed state
    return nil
}
return apply(actions)
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
    printPlan(actions) // would: / skip:
    return nil
}
return apply(actions)
```

Same `plan` for both modes. Dry-run never calls `apply`.

### Inline gate + preflight

Best for multi-step CLIs (release, sync, teardown):

```go
exists, err := probeExists(path) // always
if err != nil { /* live hard-fail; dry-run warn */ }
if !exists {
    fmt.Printf("skip: %s already absent\n", path)
    return nil
}
if dryRun {
    fmt.Printf("would: rm -f %s\n", path)
    return nil
}
return os.Remove(path)
```

### Effect interface (advanced)

When side effects are many or need unit tests, inject them:

```go
type Effects interface {
    Probe(path string) (bool, error) // read-only; used in dry-run
    Remove(path string) error        // gated in dry-run
}
```

Orchestration stays identical; dry-run calls `Probe` and prints would/skip
instead of `Remove`.

## Soft-fail vs hard-fail

| Mode | Required for the real run | Preflight unavailable | Optional display enrichment |
| ---- | ------------------------- | --------------------- | --------------------------- |
| Live | Hard-fail | Hard-fail | Hard-fail or omit |
| Dry-run | Prefer hard-fail if live would abort **and** no alternate exact plan exists | Soft-warn + best-effort exact would from known constants | Soft-warn + default OK |

Example: missing upload credentials — live must fail; dry-run may warn
and print a default `owner/repo` so the rest of the plan is still
visible.

Example: local pubkey missing for revoke — prefer probe remote by known
comment and print the **exact** comment-based revoke script (or skip if
absent). Do not stop at a paraphrased “would revoke path”.

## Output convention

`--dry-run` is on the argv. Do not prefix lines with `[dry-run]`.

```text
$ mytool teardown --dry-run
probing (read-only): ssh host 'grep … authorized_keys'
skip: no managed lines for comment tool@host
probing (read-only): spl … -- test -f /tmp/key
would: spl … -- rm -f /tmp/key
```

```text
$ mytool release --dry-run
warning: open .upload-credentials.json: no such file or directory
tag: v1.2.3
would: build mytool-v1.2.3-linux-amd64
```

- Probing lines → **stdout** (or kind-aligned `notice:` when staged / `-v`)
- Planned mutate → **stdout**, `would:`
- No-op → **stdout**, `skip:`
- Soft-fail → **stderr**, `warning:`
- Exit **0** when planning succeeded

Staged pipelines: same `[n/total]` spine as live; kind-aligned `would:` /
`skip:` under the open stage (`cli/output/staged-markers`). Never
`skip (dry-run)`.

## When a separate preview is OK

Almost never for multi-step CLIs. A static help example or a pure
documentation string that cannot drift is fine. If the preview can
diverge from live behavior, it is not dry-run—merge it into the real
path.

## See also

- `cli/output/staged-markers` — one `[n/total]` per stage; kind-aligned `would:` under mutate stages
- `cli/output/streaming` — print planned units as you go; do not buffer the
  whole dry-run report unless you need a table
- `flags-parsing` — wire `--dry-run` with less-flags
- `cmd-exec` — run preflight always (or under dry-run); print mutate
  command lines in dry-run instead of executing them
