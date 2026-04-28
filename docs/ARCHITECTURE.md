# Architecture

This document describes the technical structure of `gh-secrets-sync` and how its components interact.

## Overview

`gh-secrets-sync` is a GitHub CLI extension written in Go. It allows users to manage a local YAML secrets configuration file and push those secrets to one or more GitHub repositories via the `gh` CLI.

## Entry Point

```
cmd/gh-secrets-sync/main.go
```

`main.go` constructs an `App` with build-time version metadata and calls `App.Run`. It is intentionally thin — all wiring lives in `internal/app`.

## Package Structure

```
cmd/
  gh-secrets-sync/        # Binary entry point (main package)

internal/
  app/                    # Composition root — wires all dependencies and sub-commands
  commands/               # One sub-package per CLI sub-command
    root/                 # Root command definition (name, version, Before hook)
    add/                  # gh secrets-sync add
    configeditor/         # gh secrets-sync config  (opens editor)
    delete/               # gh secrets-sync delete
    list/                 # gh secrets-sync list
    sync/                 # gh secrets-sync sync
    update/               # gh secrets-sync update
  config/                 # YAML config file I/O and domain types (Config, Secret)
  gh/                     # Thin wrapper around the gh CLI (Client, Executor interface)
  cmdutil/                # Shared CLI helpers (e.g. SplitRepos)
  testutils/              # Test helpers (e.g. SetupConfig)

docs/                     # Project documentation
```

## Key Components

### `internal/app`

The composition root. `App` holds top-level dependencies (`VersionInfo`, `gh.Client`) and exposes a single `Run(ctx, args)` method. On startup the `onInit` Before hook creates the config file if it does not yet exist and prints a getting-started hint.

### `internal/commands/root`

Defines the root `*cli.Command` using the [urfave/cli v3](https://github.com/urfave/cli) framework. Accepts functional options (`WithVersion`, `WithCommand`, `WithOnInit`) so that `app` can compose it without coupling.

### `internal/commands/*`

Each sub-command package exposes a single `New() *cli.Command` constructor that is self-contained: it loads the config file from the default path, performs its operation, and saves the result. Sub-commands do **not** receive shared state through the context — they rely on the config file as the single source of truth.

### `internal/config`

Owns the `Config` and `Secret` types and all YAML file operations:

| Function | Description |
|---|---|
| `DefaultConfigPath()` | Returns config path, checking `GH_SECRETS_SYNC_CONFIG_FILE` env var first |
| `EnsureConfigExists()` | Creates the file with a skeleton YAML if it does not exist |
| `Load()` | Reads and unmarshals the YAML config |
| `Config.Save()` | Marshals and writes the config back to disk |
| `Config.AddSecret()` | Adds a secret, with optional force-overwrite |
| `Config.UpdateSecret()` | Patches value and/or repositories of an existing secret |
| `Config.DeleteSecret()` | Removes a secret by name |

### `internal/gh`

Wraps the `gh` CLI binary via [cli/go-gh](https://github.com/cli/go-gh). The `Client` embeds an `Executor` interface so the real subprocess can be swapped out in tests. `NewClient` verifies that the `gh` binary is available before returning.

### `internal/cmdutil`

Small, stateless helpers shared across command packages. Currently provides `SplitRepos`, which normalises comma-separated and repeated `--repos` flag values into a clean `[]string`.

## Data Flow

```
User runs: gh secrets-sync sync

main.go
  └─ app.App.Run(ctx, args)
       ├─ root.New(...)                   # build *cli.Command tree
       │    └─ onInit Before hook         # ensure config file exists
       └─ sync.New(ghClient).Run(...)
            ├─ config.Load(path)          # read secrets.yaml
            └─ for each secret/repo
                 └─ gh.Client            # call `gh secret set` via gh CLI
```

## Testing Approach

- Every package has a `_test.go` file using the external test package convention (`package foo_test`).
- Tests use [testify](https://github.com/stretchr/testify) (`assert` / `require`) and Go subtests (`t.Run`) with `t.Parallel()`.
- `internal/testutils` provides `SetupConfig` which writes a temporary config file and sets `GH_SECRETS_SYNC_CONFIG` so commands pick it up without touching real user state.
- The `gh.Executor` interface allows `gh.Client` to be exercised in unit tests without a real `gh` binary.
