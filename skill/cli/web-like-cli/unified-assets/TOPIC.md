---
name: go-best-practice/cli/web-like-cli/unified-assets
description: >-
  One content-addressed image/asset library for the whole app: images/<id>/
  {image.<ext>, meta.json} in the shared data dir, md5 dedup so identical bytes
  reuse a record, magic-byte sniffing so a file name can never make a non-image
  a picture, meta.json written last as the commit marker, an audit that reports
  valid/problem/usages, and a delete guard driven by ONE container registry so
  no dangling image id is ever left behind. Use when adding image upload,
  attachments, or a media library to a Go+React app, or when images are stored
  per entity, a broken picture shows as a 200, or deleting an image breaks a
  page. Triggers: unified image storage, image library, asset store, upload
  dedup, md5 dedup, image audit, delete guard, dangling image id, not an image.
---

# unified-assets — one content-addressed library for every picture

Replace per-entity image directories with **one library**: every uploaded
file gets a record at `images/<id>/`, the id comes from the app-wide sequence
(see `id-allocator`), and domain documents reference it by id and
nothing else.

```text
<dataDir>/
├── id.json                    one sequence for the whole app
├── id.lock
├── <domain>.json              {"image_id": "7"} / {"image_ids": ["7","9"]}
└── images/
    └── 7/
        ├── image.jpg          the bytes; extension = the SNIFFED format
        └── meta.json          written LAST — its presence makes the record real
```

There is **no `images.json` catalog**. The directory listing is the index;
a catalog is a second source of truth that drifts from disk.

## How the two halves fit

The library never knows what a "page" is: the domain owns the container
registry, and the same registry drives both the audit and the detach.

```text
upload ──► sniff bytes ──► md5 already stored? ──yes──► reuse that record
                                 │no
                                 ▼
                        idalloc.Next(floor)  ──►  images/<id>/{image.<ext>, meta.json}
                                 │
                                 ▼
        domain document stores only {"image_id": "<id>"}   (never a path, never bytes)

delete ──► who references <id>?  ──none──► remove images/<id>/
                                 └─some──► refuse, naming the page
                                           (--force: detach every reference, then remove)
```

## Record shape

```go
type ImageMeta struct {
	ID               string `json:"id"`                // from the shared sequence
	MD5              string `json:"md5,omitempty"`     // dedup key
	Name             string `json:"name,omitempty"`
	Source           string `json:"source,omitempty"`  // where it came from, for external records
	OriginalFilename string `json:"original_filename,omitempty"`
	Ext              string `json:"ext,omitempty"`     // "" = external: lives at Source, no local bytes
	Mime             string `json:"mime,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"` // RFC3339 local offset; see `time-string`
	Caption          string `json:"caption,omitempty"`
	TakenAt          string `json:"taken_at,omitempty"`
}
```

`Ext == ""` is a **supported kind**, not a defect: the record is an external
reference that lives at its `Source` URL. Any code that assumes a local blob
must branch on `Ext` first.

## Write path

1. **Sniff the bytes** — decide format from magic bytes, never the file name.
2. **md5 the bytes** — if a record with the same md5 exists, **return it**.
   The upload is idempotent: same bytes, same id, same URL, no new record.
3. **Allocate the id** from the shared sequence (`idalloc.Next(floor)`).
4. **Write the blob** to `images/<id>/image.<ext>`.
5. **Write `meta.json` last.** A directory without it is a partial upload and
   is ignored by every reader — on failure, `RemoveAll` the directory.

```go
sum := md5Hex(data)
if existing, ok, _ := s.findImageIDByMD5Locked(sum); ok {
	return s.loadImageMetaLocked(existing) // dedup: no new id, no new bytes
}
ext, mime, err := detectImageExtMIME(in.OriginalFilename, in.Data, in.AllowInvalid)
id, err := idalloc.New(filepath.Join(s.Dir, "id.json")).Next(s.maxImageID())
// … os.MkdirAll, WriteFile(blob), then saveJSONFile(meta.json) LAST
```

Dedup is what makes the read path cacheable: **an id's bytes never change.**

### Bytes decide the format; the name only explains a rejection

```go
func detectImageExtMIME(filename string, data []byte, allowInvalid bool) (ext, mime string, err error) {
	if ext, mime, ok := sniffImageMagic(data); ok { return ext, mime, nil }        // JPEG/PNG/WEBP
	if ext, mime, ok := sniffContentTypeImage(data); ok { return ext, mime, nil }  // http.DetectContentType
	if allowInvalid { /* import escape hatch: keep bytes, report as not-an-image */ }
	return "", "", &ValidationError{Message: notAnImageMessage(filename, data)}
}
```

The name must **never** be able to make non-image bytes acceptable. A store
that fell back to the extension whenever sniffing failed kept a **465-byte
马蜂窝 XML error page as `image/jpeg` and served it with a 200** — the
download had failed, and the store recorded the failure page as a photo.

`allowInvalid` is the one deliberate escape hatch, for bulk imports and
legacy resolution: it keeps unresolvable bytes so one bad file cannot abort a
migration, and the audit reports them (`problem_kind: not-an-image`).

Also decide what is *not* an allowed upload. SVG is normally excluded — an
uploaded SVG served from the app origin executes script:

```go
var allowedImageTypes = map[string]string{"image/png": ".png", "image/jpeg": ".jpg",
	"image/gif": ".gif", "image/webp": ".webp"} // no image/svg+xml
```

Always cap the size, and reject an empty body.

## Read path — one address per picture

Serve the bytes from the data dir through one route, and derive the public
URL from the record:

```go
// GET /api/data/images/<id>/image.<ext>   ← http.FileServer(http.Dir(dataDir))
func (s *GlobalStore) ImageURL(meta ImageMeta) string {
	if id, ext := strings.TrimSpace(meta.ID), strings.TrimSpace(meta.Ext); id != "" && ext != "" {
		return s.dataURLPrefix() + "images/" + id + "/" + "image." + ext
	}
	return strings.TrimSpace(meta.Source) // external: no local bytes
}
```

Because identical bytes reuse a record, the bytes at one id can never change:

```go
w.Header().Set("Content-Type", img.MimeType)
w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
w.Header().Set("X-Content-Type-Options", "nosniff")
```

**Return the stored path, so a caller never has to guess it.** The API
record carries the absolute path of the materialized blob (`path`, omitted
for external records); a CLI prints that path as its first line so an agent
can open the file. Never inline binary content in JSON or on a terminal, and
never join an origin onto an absolute URL:

```go
// wrong: external records hold an absolute Source, so this yields
// http://localhost:8080https://example.com/a.jpg
fmt.Fprintln(stdout, normalizeServer(server)+img.URL)
// right: print the server-reported path when there is one, else the URL as-is
```

## Audit — a 200 is not proof of a picture

`ListImages(verify)` returns every record plus the two facts an operator
needs: is it really a picture, and **who references it**.

```go
type LibraryImage struct {
	ImageMeta
	URL         string       `json:"url"`
	Path        string       `json:"path,omitempty"` // absolute blob path; "" for external
	Bytes       int64        `json:"bytes"`
	Valid       bool         `json:"valid"`
	Problem     string       `json:"problem,omitempty"`      // human sentence
	ProblemKind string       `json:"problem_kind,omitempty"` // machine-readable
	External    bool         `json:"external,omitempty"`
	Usages      []ImageUsage `json:"usages"`
}
```

| `problem_kind` | Meaning |
|----------------|---------|
| `not-an-image` | The bytes are not a picture (say what they *are*: `detected text/xml`) |
| `missing-blob` | `Ext` is set but the file is gone |
| `missing-source` | External record whose local source path no longer exists |
| `no-source` | No local file and no source at all |

`Valid`/`Problem` are **derived on every read**, never stored — a persisted
flag eventually claims a corrupt file is fine. `verify=false` skips reading
the blobs (the only reason to turn it off) but still walks references, so
unused images stay visible.

`Usages` is why deleting needs a guard, and it is built by walking the data
tree for the `image_id` / `image_ids` keys — matching by **key name** rather
than by N typed loaders, so a new container cannot silently escape the audit:

```go
for _, ref := range collectImageRefs(decodedDoc, nil) { // walks maps/arrays for the two keys
	out[ref.id] = appendUsage(out[ref.id], ImageUsage{
		Kind: container.kind, PagePath: owner.pagePath, Label: owner.label,
		Detachable: container.detach != nil,
	})
}
```

Skip `images/` itself while walking (`meta.json` describes an image, it does
not use one), and let an unreadable subtree be skipped rather than fail the
whole audit.

Report it in the CLI as a table **plus a summary line** (counts are what make
a problem visible):

```text
$ demo get /api/images
ID   Size      Format  Used by                  State
1    182.4 KB  jpg     Gallery (/api/gallery)   ok
2    1.2 KB    jpg     -                        NOT AN IMAGE (detected text/xml)
4    0 B       -       -                        unused

3 images · 1 not an image · 1 unused
warning: 1 of 3 library images is not a picture; run 'demo get /api/images' to list them
```

## Delete guard — one container registry, or dangling ids

Deleting a referenced image leaves the page that shows it with a broken
picture. So a plain delete is **refused**, naming the page and the way out;
`--force` detaches every reference and then removes the record.

The registry is the whole design:

```go
type imageContainer struct {
	file    string // base name that marks the container, e.g. "gallery.json"
	kind    string // usage kind reported for a reference found here
	retired bool   // data the product no longer reads: skipped entirely
	detach  func(s *Store, path, id string) (bool, error)
}

// Every file type carrying image_id / image_ids must appear here.
var imageContainers = []imageContainer{
	{file: "gallery.json", kind: "gallery", detach: detachGalleryFile},
}
```

**The audit and the forced delete must be driven by the same table.** When
they were two hand-written walks, the guard detected references in five
containers while the detach rewrote two — so `delete --force` on an image
used by a spot deleted the picture and left the id dangling, the opposite of
what its own refusal message promised. One table makes that asymmetry
unrepresentable.

Three further rules:

- **Undetachable ⇒ refuse.** A reference in a file this build cannot rewrite
  blocks the delete; removing the picture would leave a page pointing at an
  id that no longer exists.
- **Retired containers are skipped.** Refusing a delete because of a file
  nothing renders is misleading.
- **Never leave a dangling id.** Detach then delete; the refusal message must
  name the exact page(s) so the operator can act.

```text
$ demo delete /api/images/7
Error: image 7 is used by Gallery (/api/gallery) · gallery
  detach it first, or pass --force to delete and detach

$ demo delete /api/images/7 --force
deleted image 7 · detached 1 reference
```

Detach goes through the **typed loader the app uses** (`LoadGallery` /
`SaveGallery`), so normalization and file formatting match the web; only the
reference is dropped, never the containing record.

## HTTP API

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/images` | multipart `file` (+`name`,`caption`) → record; md5 dedup returns the existing one |
| GET | `/api/images` | audit list; `?verify=0` skips byte reads, keeps usages |
| GET | `/api/images/<id>` | one record including `path` and `usages` |
| DELETE | `/api/images/<id>[?force=1]` | guarded delete; force detaches first |
| GET | `/api/data/images/<id>/image.<ext>` | the bytes |
| POST | `/api/ids` | `{"count":n}` → `{"ids":[…]}` (see `id-allocator`) |

Reject the upload **before** storing it, and answer with the reason:

```text
Error: not an image: detected text/xml; the file name says .jpg
```

Serialization: one RWMutex per library directory, taken for the whole
dedup-then-allocate sequence (otherwise two concurrent uploads of the same
bytes both miss the md5 lookup and store two records). Never hold the image
lock while taking another lock — resolve labels (place/page names) before.

## Retrofitting an existing app

Migrate per-entity image directories into the library once, idempotently:

1. Move each entity's files into `images/<id>/` (allocate through the shared
   sequence), writing `meta.json` last.
2. Rewrite every JSON reference from a path to the new **id**.
3. Import remaining loose image files (flattening legacy directories).
4. Delete the now-empty legacy image directories.
5. **Delete any legacy catalog** (e.g. `images.json`) so it cannot become a
   second source of truth.

Make it re-runnable and safe to call on every server start; a half-migrated
data dir must still serve.

## Anti-patterns

| Don't | Why |
|-------|-----|
| Store bytes per entity (`places/3/images/…`) | The same picture is stored N times; no app-wide "list images"; deleting the owner orphans bytes |
| Key a record by its original filename | Names collide; a rename rewrites every reference |
| Trust `Content-Type` or the extension | The 马蜂窝 XML error page became an `image/jpeg` served with a 200 |
| Store a `valid` flag | Goes stale; derive it from the bytes on every read |
| Keep an `images.json` catalog | A second source of truth that drifts from `images/` |
| Two walks: one to refuse, one to detach | They disagree; `--force` leaves dangling ids (use one registry) |
| Reuse an id for different bytes | Breaks `immutable` caching; the whole dedup design assumes an id's bytes never change |
| Refuse only for containers you happen to know | A new `image_id` field escapes the guard; walk by key name |
| Report problems only in `--json` | The summary line and the stderr warning are how a human notices |

## Tests to write

Anchor each safety property to a test (travel-map's CLI doctests are a good
model: `tests/cli/against-server/image-{library-audit,delete-guard,upload-rejects-non-image}`):

- Upload real bytes → record on disk, `valid`, blob + `meta.json` present.
- Upload the **same** bytes → same id, no second directory (dedup).
- Upload XML named `.jpg` → rejected, and nothing written into the library.
- Plant a non-picture with a hand-written `meta.json` → audit says
  `not-an-image` and names what the bytes are (`text/xml`), **not** "ok".
- Audit names the page a healthy image is used by.
- `verify=0` → no byte verdict, usages still listed.
- Delete a referenced image → refused (non-zero exit), message names the page,
  image still present.
- Delete `--force` → detached **and** removed; the container no longer carries
  the id (no dangling reference), and the library no longer lists it.
- Delete an unreferenced image → succeeds without `--force`.
- External record (no `Ext`) → reported `external`, not broken; a dead local
  source → `missing-source`.
- Read one image → the reported `path` exists on disk.
- Missing `id.json` → the next id is still greater than every id in use.

## See also

- `id-allocator` — where these ids come from
- `web-like-cli` — the parent recipe: one URL space, and `get` on an image path
  prints an address, never bytes
- `time-string` — `created_at` / `taken_at` formatting
- `kool-create` — the `go-react-agent-cli` scaffold ships this library
