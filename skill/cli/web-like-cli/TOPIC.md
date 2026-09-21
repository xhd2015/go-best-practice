---
name: go-best-practice/cli/web-like-cli
description: >-
  One URL space for page, API and terminal: a CLI that speaks the web app's
  routes. Page paths render whole page documents, collection paths print rows,
  writes answer with the resource and its page URL, and pasting the browser
  URL just works. Use when adding CLI access to a Go+React app, when CLI
  paths and web routes drift, or when an agent must convert URLs before
  calling the CLI. Triggers: web-like CLI, CLI aligned with web, page
  document, one URL space, CLI paths match web routes.
---

# web-like CLI — one URL space for page, API and terminal

The CLI and the web page are two views of the same data. Keep them on **one
URL space**, so the page a human opens in the browser and the path an agent
types in the terminal are the same address:

- page URL ≡ CLI path ≡ server API route
- the CLI and the page must **always agree** about what a path means

When you add a web page or route, add its CLI mapping in the **same change**.
A CLI that lags the web forces every agent (and human) to convert URLs by
hand — the exact drift this recipe prevents.

## Path grammar

| Path kind | Shape | GET prints | Writes |
|-----------|-------|------------|--------|
| page path | `/users/1`, `/stories/268` | the **page document** | usually none (page) |
| collection path | `/stories/268/uis` | the rows | full-list replace |
| one-record path | `/plans/12`, `/stories?id=3` | one record / row | minimal fields only |
| tree / list path | `/stories`, `/prds` | list rows | create (`post`) |

Accepted input forms — all parse to the same value:

```text
<cli> get /stories/268                                # bare path
<cli> get /<app>/stories/268                          # web-prefixed
<cli> get 'http://127.0.0.1:8080/<app>/stories/268'   # full URL: the origin
                                                      # pins the server, the
                                                      # query is preserved
```

Full URL = strip the origin (remember it as the server), keep the query
string, strip the web prefix. A pasted browser URL needs **zero conversion**.

## Page documents

GET on a page path renders the page the way the web shows it: breadcrumb,
title, meta, then **every section in card order**, each with the page's own
wording and empty states — never raw JSON.

The mental model: `get <page>` is a **human viewing the web page**. The
renderer performs the same orchestration the React page performs — every API
call the page makes, correctly ordered and called (a card that needs a
resolved id fetches the catalog first, a card fed by another card's data
waits for it), then rendered as text. Only the extra HTML/UI markup is
stripped: what remains is the page's content, structure and empty states.

Corollary: when the web page gains a card or a new data call, the page
renderer gains it in the same change. A page document that skips one of the
page's calls is a page the agent sees only in part — that is drift, not
alignment.

The replication is deliberate. The renderer is a super-simplified yet still
complete replica of the frontend behavior: simplified in markup (plain text,
no layout), complete in data, orchestration and states. Output is optimized
for agents — info-dense: every line carries data; chrome, filler and layout
noise are exactly the markup we strip.

```text
$ <cli> get 'http://127.0.0.1:8421/ai-workshop/user-stories/268'
Workshop / Off-plat Checkout → SPL / Off-plat PC rule matching (试算)
title: Off-plat PC rule matching (试算)
meta: story · kb · entry · AS checkout · SPL checks the PC rule quota …
url:  http://127.0.0.1:8421/ai-workshop/user-stories/268

## Flow
requirement: Request journey spine · highlight main · DB/Redis above …
No flow yet. Add spine nodes (User, Page, services), mark main, attach DB/Redis.

## UIs
No UIs yet

## Sequences
- Typical Path
  at: /ai-workshop/user-stories/268/sequences/274
```

**One home for page words.** The React page and the CLI renderer import the
same text catalog (one JSON part per card), so a title, hint or empty state
cannot drift between page and CLI. Do not retype page wording in Go strings;
add the word to the catalog and read it from both sides.

Sections mirror the web cards **in web order**; a section without content
prints the page's own empty state, never silence and never invented data.

## Images and attachments

An agent cannot see an `<img>` tag, so every image or attachment a page shows
renders as an **address**:

- a **local path** when the file is already materialized (imported or
  downloaded into the app's data dir), e.g.
  `image: data/ui-images/ui1/preview.png`;
- otherwise the **source URL**, and `get` accepts that URL itself:

```text
$ <cli> get 'https://confluence.example/download/attachments/123/spec.pdf'
downloaded to /tmp/myapp-attachments/spec.pdf
```

The download lands in a local tmp dir (stable name per URL; a repeat `get`
reuses the cached file) and the path is printed — the agent opens or reads
the file in its next step. Never render a bare image marker with no address,
and never inline base64 into the page document.

## Sub-sections and shownPage

Address a card as a path segment under the page:

```text
<cli> get  /stories/268/uis          # the UIs card's rows
<cli> put  /stories/268/uis @uis.json
```

A sub-path is JSON-only data that the page displays in a card, so writes and
reads **read back as the page** (`shownPage`): the `saved … at <url>` line
names the page URL where the user will see the result, not the raw API path.

Writes confirm with the resource and its page URL:

```text
$ <cli> post '/stories?parentId=267' '{"kind":"story","name":"Receive instruction"}'
created story 278 · Receive instruction
url  http://127.0.0.1:8421/ai-workshop/user-stories/278
```

## Writes mirror the web

- A collection `put` is a **full list replace** — read the collection first,
  edit the records, send the whole list back.
- A one-record `put` takes **minimal fields** (`{"status":"done"}` only);
  do not smuggle siblings into a record path.
- `delete` names its record in the path/query and carries **no body**; a 204
  reads as `deleted <page-url>`.
- Warnings are non-fatal: `warning:` on **stderr**, exit 0, so scripts carry
  on. Errors are `Error:` on stderr, non-zero.
- `--dry-run` runs the **same pipeline** with the side effect gated — never a
  separate dry-run code path.

## Server discovery

- The command finds the running server by itself (a recorded server-info
  file written by `serve`); `--port` and `--url` override.
- A full URL in the path **wins** — the pasted origin pins the server.
- Never hardcode a port in docs or examples beyond the documented default.
- `--json` prints the page document (or the raw records for collections);
  machine-readable output stays clean on stdout — notices go to stderr.

## Help and the path table

`-h/--help` at **every** level (root, verb, action). The root help
enumerates the **whole path space**, mirroring the web routes — the path
table is the contract between page and CLI:

```text
Paths (get / put / post / delete):
  /stories                       the story tree (post ?parentId=, delete ?id=)
  /stories/<id>                  the story page (flow, uis, apis, sequences)
  /stories/<id>/uis | apis       one story card (get, or full-replace put)
  /stories/<id>/sequences/<seq>  one sequence page
```

## Wrong → correct

| Wrong | Correct |
|-------|---------|
| CLI paths invented per command (`list-stories`, `show-ui 268`) | The web route, restated: `/stories`, `/stories/268/uis` |
| `get /stories/268` dumps raw JSON | Page document: sections in card order, the page's own words |
| Page renderer fetches one or two APIs and calls it a page document | Every call the page makes, in the page's order; only markup stripped |
| Images render as `[image]` with no address, or base64 inline | Local path when materialized; else a URL `get` downloads to tmp and prints |
| Page wording retyped in Go strings | Shared text catalog imported by page and renderer |
| Write prints `ok` or nothing | `created story 278 · <name>` + the page URL |
| Hardcoded `localhost:8081` in examples | Discovery by default; `--port/--url`; URL form wins |
| Help lists verbs but not paths | Root help enumerates the whole path space |
| URL pasted by the user needs hand-conversion | Parser strips origin + web prefix, keeps the query |

## Reference implementation

The spl repo's `ai-workshop` package is the working exemplar:

- `pagepath.go` — `parsePagePath` (bare / prefixed / full URL + query) and
  `resolveEndpoint` (the path grammar table)
- `pageresolve.go` — `resolvePageJob` / `resolveWriteEndpoint` (dynamic
  resolution incl. catalog lookup) falling back to the static table
- `pagetext/` — the shared text catalog (`parts/*.json`) the React page
  imports via a vite alias and the CLI reads via `//go:embed`
- `pageview_story.go` — `buildUserStoryPage` / `buildSequencePage` (page
  documents mirroring the web cards, in web order)
- story UIs (`user_story_uis.go` + `user-story-ui-image` endpoints) —
  imported images materialize under the story dir
  (`ui-images/<id>/preview.png`): the materialized tier of the image rule
- `cli.go` — `handleWrite` (dry-run, warnings, `created`/`saved` + page URL)

Scaffold: `kool create go-react-agent-cli` ships the transport half
(`run/client.go`: verbs, full-URL acceptance, `--json`/`--dry-run`,
per-verb help) and carries the binding rule list in its `AGENTS.md`.

## See also

- `cli/skill-cli` — packaging the skill surface of such a CLI
- `cli/dry-run` — side-effect gates in one pipeline
- `cli/output/streaming` / `staged-markers` — stdout/stderr and `[n/total]` spines
- `flags-parsing/subcommand` — `-h/--help` at every dispatch level
- `kool-create` — the `go-react-agent-cli` scaffold
