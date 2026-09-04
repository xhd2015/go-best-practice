---
name: go-best-practice/time-string
description: >-
  Persist instants as RFC3339 strings with local numeric offset
  (e.g. +08:00), not UTC Z. Seconds precision; pin offset in tests.
---

# time-string — RFC3339 with local offset

Store timestamps as **`string`**, formatted with `time.RFC3339`
(`2006-01-02T15:04:05Z07:00`) in the **local** location. Beijing
wall time is `2026-09-04T15:21:00+08:00`, not `…Z`.

`Z` is not “has a timezone”. Go prints **`Z` only when the `time.Time`
location is UTC**. `.UTC()` before `Format` throws the offset away.

## Policy

| Persist | Do |
| ------- | -- |
| JSON / jsonl field | `string`, not `time.Time` |
| Format | `time.RFC3339` (seconds) |
| Location | local (`time.Now()` already is) |
| Parse | `time.Parse(time.RFC3339, s)` — accepts `Z` and `+08:00` |
| Tests | pin a fixed offset string; do not assert the host’s real zone |

Do **not** store IANA names (`Asia/Shanghai`) in the same field — that
is not RFC3339. Do **not** rewrite historical `…Z` rows.

## Write

```go
// Preferred — local offset (Beijing → +08:00; UTC host → Z, correctly).
s := time.Now().Format(time.RFC3339)

// If `t` is already UTC and you want the operator's offset:
s = t.In(time.Local).Format(time.RFC3339)
```

## Anti-pattern

```go
// Wrong: always Z; local offset is gone.
s := time.Now().UTC().Format(time.RFC3339)

// Wrong: encoding/json uses RFC3339Nano; nanos + location drift.
type Event struct {
    TS time.Time `json:"ts"`
}
```

```go
// Preferred — explicit string, seconds, local offset.
type Event struct {
    TS string `json:"ts"`
}

func nowTS() string {
    return time.Now().Format(time.RFC3339)
}
```

## Read mixed history

Old `…Z` and new `…+08:00` lines coexist. Parse RFC3339; display the
stored string. Compare as `time.Time` when you need ordering.

## Tests

Do not assert `time.Now()`’s offset — CI hosts differ. When a date env
pins the calendar day (e.g. `TSK_DATE=2026-07-09`), emit a **fixed**
offset such as `+08:00`:

```go
func NowLocalTimestamp() string {
    if date := os.Getenv("TSK_DATE"); date != "" {
        return date + "T12:00:00+08:00"
    }
    return time.Now().Format(time.RFC3339)
}
```

Live path still uses real local offset. Only the pinned branch is stable.

## Out of scope

Unix seconds, IANA zone names in `ts`, RFC3339Nano unless you truly
need subseconds, migrating old `Z` rows.

## See also

- `cli/output/streaming` — jsonl as the on-disk log shape
- `cli/config` — other small JSON files under a tool home
