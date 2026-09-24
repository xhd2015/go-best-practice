---
name: go-best-practice/cli/web-like-cli/id-allocator
description: >-
  Allocate every entity id from one persistent sequence per data directory:
  dot-pkgs idalloc, flock on a stable lock inode, atomic state replace, floor
  reseed from existing data, and POST /api/ids for a web page that needs an id
  before it renders. Use when ids are timestamp-based, per-entity counters,
  duplicated across types, or reused after a restart. Triggers: id allocator,
  sequential ids, id.json, id reuse, unique id, POST /api/ids.
---

# id-allocator — one sequence for every entity in the app

Use **one** id sequence per data directory, for **every** entity type:
targets, plans, spots, routes, notes, and images alike.

```go
import "github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc"

n, err := idalloc.New(filepath.Join(dataDir, "id.json")).Next(floor)
id := strconv.FormatInt(n, 10)
```

Ids are decimal **strings** on the wire and in JSON. They are opaque: no
caller may do arithmetic on them, and no caller may assume they are
contiguous.

## Why one sequence, not one per type

A per-type counter makes `7` mean "plan 7" and "image 7" at once. The moment
a route, a spot or an album row references an id, the reader must know the
type to resolve it — and every id becomes ambiguous in logs, URLs, and bug
reports. One sequence costs nothing and removes the question: **an id
identifies one record in the whole app, forever.**

Gaps are expected and fine. A caller that allocates and then fails leaves a
hole; nothing reuses it. Never "compact" ids.

## The package does the hard part

`idalloc` (`github.com/xhd2015/dot-pkgs/go-pkgs/file/idalloc`) is the SSOT.
Do not hand-roll the read-modify-write; the details are the whole difficulty:

| Property | How |
|----------|-----|
| Serialized across **processes** | `syscall.Flock(LOCK_EX)` on a sibling lock file, held only for one read-modify-write |
| Lock file is stable | Its name is inferred from the state file (`id.json` → `id.lock`; extensionless `ids` → `ids.lock`) and it is **never renamed** |
| State survives a crash | `{"last": N}` is written to a temp file, `fsync`ed, then `os.Rename`d over the state file — a reader sees the old or the new complete state, never a torn one |
| Permissions | state and lock `0600`, a directory the package creates `0700` |
| Corrupt state | a decode error says *"delete it to reseed from existing data"* rather than silently restarting at 1 |

```go
n, err := idalloc.New(stateFile).Next(min)  // min+1 at the floor; persists n
last, err := idalloc.New(stateFile).Last()  // read without allocating
```

## Floor reseed: make a lost `id.json` recoverable

`Next(min)` returns `max(persisted, min) + 1`. Pass **the largest id already
present in your data** as `min`, and deleting `id.json` degrades to "continue
after the highest id in use" instead of "hand out an id that already exists".

```go
// The floor must cover everything that holds an id — including records that
// live outside the main document.
func (s *Store) maxNumericID(doc *Snapshot) int64 {
	max := s.maxImageID() // images live in their own directories on disk
	for _, t := range doc.Targets {
		consider(t.ID) // strconv.ParseInt; ignore legacy hex ids
		for _, m := range doc.Milestones[t.ID] {
			consider(m.ID)
		}
	}
	return max
}
```

Two rules keep the floor honest:

- **Scan disk too.** Images (or any entity stored as a directory name) are
  not in the main document; leaving them out lets a lost `id.json` collide
  with an image id.
- **Ignore non-numeric legacy ids.** `idalloc.MaxNumeric([]string{...})`
  parses decimal and skips the rest, so a hex-id migration does not force the
  sequence to jump.

## The web needs an id before it can render a row

A page that adds a spot cannot wait for a save to learn the id: the id is
what the row's routes and its own images reference. So the client asks the
server:

```text
POST /api/ids   {"count": 2}   →   {"ids": ["714", "715"]}
```

- Allocate per **row the user is editing**, not in bulk. Cap the request
  (travel-map uses `maxAllocatedIDs = 100`) so one call cannot drain the
  sequence into a client that never uses it.
- An empty body means one id; a missing/`0`/negative `count` means one id.
- The unused ids leave gaps, which the sequence already accepts.

## Anti-patterns

| Don't | Why |
|-------|-----|
| `strconv.FormatInt(time.Now().UnixNano(), 10)` | Not a sequence: no ordering guarantee across processes, and it is not allocator-serialized |
| A counter per entity type | Ids collide across types; every reference becomes type-dependent |
| `max(id)+1` computed in app memory | Two processes compute the same `max` and hand out the same id |
| Reading the state file without the lock | You may read a value another process is about to replace; the atomic rename makes the *read* safe but the *decision* stale |
| Reusing an id freed by a delete | A cached page, log line, or backup still resolves it to the old record |
| Storing ids as JSON numbers | JS loses precision past 2^53, and the id stops being opaque |
| Locking a file you also rename | A second process opens a *new* inode and both hold "the" lock — hence the separate, never-renamed `id.lock` |

## Tests to write

- Sequential: three `Next(0)` calls return 1, 2, 3.
- Floor: `Next(42)` → 43 on an empty state; a persisted `44` then beats a
  lower floor.
- Restart: a fresh allocator on the same state file continues the sequence.
- Concurrency: N goroutines × M allocations (shared and separate allocator
  objects) produce N×M **distinct** ids.
- Lock identity: the lock file's inode is unchanged after several
  allocations, while the state file's inode changes (proves the atomic
  replace and that the lock was never renamed).
- Corrupt state: the error mentions reseeding.
- Permissions: `0600` files, `0700` created dir.
- App level: after deleting `id.json`, the next id is still greater than the
  largest id in the data (the floor).

## See also

- `unified-assets` — the image library that draws from this sequence
- `web-like-cli` — the parent: one URL space, and the layout both recipes share
- `time-string` — `created_at` fields on the records these ids name
