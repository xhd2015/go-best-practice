// Package skill holds the embedded SKILL.md index and nested TOPIC.md tree.
package skill

import "embed"

//go:embed SKILL.md
var Root string

// Tree is the nested topic directories (path/TOPIC.md layout).
//
//go:embed cli
//go:embed cmd-exec
//go:embed flags-parsing
//go:embed go-embed-assets
//go:embed go-embed-version
//go:embed kool-create
//go:embed time-string
var Tree embed.FS
