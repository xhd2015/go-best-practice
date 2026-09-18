---
name: go-best-practice/cli/project-layout/script/dev
description: >-
  Development scripts using github.com/xhd2015/dot-pkgs/go-pkgs/dev/server
  for Go/Air hot reload, optional Vite HMR, occupied-port handling,
  invocation-specific readiness, and owned-process cleanup.
---

# script/dev — shared development runner

Provide a thin `script/dev` entry calling application configuration in an
importable package. Use `github.com/xhd2015/dot-pkgs/go-pkgs/dev/server`
for supervision; it uses `dev/hot_reload/air` internally. The runner targets
macOS/Linux.

## Configure the application

```go
import devserver "github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"

root, err := devserver.FindRoot("frontend/package.json")
if err != nil {
    return err
}
return devserver.Main(args, devserver.Config{
    Name: "My App", Root: root, BuildPackage: "./script/dev",
    BackendPort: 8080, FrontendPort: 5173, BrowserPath: "/",
    WatchDirs: []string{"internal", "server", "script/dev"},
    Frontend: &devserver.Vite{
        Dir: "frontend", ConfigFile: "vite.config.ts",
        Install: []string{"pnpm", "install", "--frozen-lockfile"},
        Command: []string{"pnpm", "exec", "vite"},
    },
    BackendHandler: newDevHandler,
})
```

`BackendHandler(context.Context, devserver.Backend) (http.Handler, error)`
constructs the application's routes and proxies frontend requests to
`Backend.FrontendURL`. It must not start listeners, Air, Vite, or a browser.
Keep application storage choices in this adapter. Optional `OnListen`
registers application discovery and returns its cleanup callback.

Use the project's declared package manager; the example is pnpm, not a
universal requirement. Set `Frontend: nil` for backend-only projects.
The Vite adapter wraps the existing config, preserving its plugins and config
function, and adds a development identity endpoint. No project plugin is
required. Ensure the configured Vite command accepts appended CLI arguments.

The development build target must compile without production frontend assets.
Use a dev-only entry or the placeholder strategy from `go-embed-assets`.
Include backend source directories and exclude generated frontend assets from
the application's watch scope.

## Shared behavior — do not duplicate in scripts

- Air runs by default. Vite remains alive across Go rebuilds.
- Existing IPv4/IPv6 listeners are checked before bind probes. Occupied default
  ports are skipped with a warning; occupied explicit ports fail. A probe is
  not a reservation: the actual bind still must succeed.
- Backend and Vite expose `/__dev/ready` with a random per-invocation identity.
  Only a matching identity satisfies readiness. Generic HTTP 200, a TCP
  connection, or a fixed sleep cannot prove this invocation is ready.
- Readiness bypasses environment proxies and refuses redirects. Browser opening
  and the ready message occur once, after verified startup.
- The runner owns cancellation, early-exit handling, startup deadlines, and
  cleanup of its process groups. It never kills arbitrary port listeners.
  Before stopping Air, it requests cleanup from its identified backend through
  a development-only shutdown endpoint; other invocation IDs are rejected.
- Temporary binaries, the generated Vite wrapper, and Air diagnostics live in
  a unique ignored `tmp/dev-*` directory; application data stays separate.
- Missing Air is installed by the helper via `brew install go-air`, falling
  back to `go install github.com/air-verse/air@latest`. Missing frontend
  dependencies use the configured install command.

## CLI contract

`Main` provides `--port`, `--vite-port`, `--no-open`, `--use-air`,
`--no-use-air`, internal `--serve-only`, and `-h`/`--help`.

```text
$ go run ./script/dev
Starting frontend on 5173...
Building backend on 8081...
My App ready: http://127.0.0.1:8081/

# stderr; non-fatal, normal shutdown exits 0
warning: backend port 8080 is occupied; using 8081

$ go run ./script/dev --port 8080
Error: requested backend port 8080 is unavailable
# stderr; non-zero exit

$ go run ./script/dev --help
Usage: go run ./script/dev [options]
...
```

## Verify a consumer

Run the shared package's tests and the application's affected tests. Then check
fresh development startup, occupied wildcard ports, actual Go rebuilds with a
stable frontend PID, browser HMR, compile-error recovery, and Ctrl-C releasing
owned listeners. Verify application-specific discovery and data directories.
Use existing APIs to check behavior without modifying real application data.

For local cross-repository work, use an ignored `go.work` selecting the
consumer and the dot-pkgs `go-pkgs` module. Document the wiring; update the
published dependency version when releasing. Do not publish a machine-specific
module replacement.

## Lower-level use

Use `dev/hot_reload/air.Run` or `Start` directly only when the shared server
runner does not fit (for example, a non-HTTP process). The helper handles Air
installation and graceful cancellation; `Start` callers still observe
`Exited` and defer `Stop()`. Do not rebuild a parallel web supervisor in each
project.

Retrieve with `go-best-practice skill --show cli/project-layout/script/dev`.
