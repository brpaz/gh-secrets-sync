# AGENTS.md

Guidelines for agentic coding agents operating in this repository.

## Quick Reference

| Task | Command |
|------|--------|
| Run all tests | `task test` or `go test ./...` |
| Run single test (filter) | `go test ./internal/commands/root -run TestNew -v` |
| Run single package | `go test ./internal/config -v` |
| Run lint | `task lint` |
| Auto-fix lint | `task lint-fix` |
| Build binary | `go build -o gh-secrets-sync ./cmd/gh-secrets-sync` |
| List all tasks | `task -l` |

## Project Structure

```
cmd/gh-secrets-sync/    # Binary entry point (thin main.go)
internal/
  app/              # Composition root – wires all dependencies
  commands/          # One sub-package per CLI sub-command
    root/           # Root command (name, version, Before hook)
    add/            # gh secrets-sync add
    configeditor/     # gh secrets-sync config
    delete/         # gh secrets-sync delete
    list/           # gh secrets-sync list
    sync/           # gh secrets-sync sync
    update/         # gh secrets-sync update
  config/           # YAML config file I/O and domain types
  gh/              # gh CLI wrapper (Client, Executor interface)
  cmdutil/         # Shared CLI helpers
  testutils/        # Test helpers (SetupConfig)
```

## Code Style Guidelines

### Formatting & Imports

- **Formatter**: gofumpt (stricter than `go fmt`)
- **Import sorter**: gci with section order: standard → default → blank → prefix(github.com/brpaz/gh-secrets-sync) → blank → dot → alias → localmodule
- Run `golangci-lint run --fix` to auto-format.

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, short | `config`, `cmdutil`, `gh` |
| Exported types | PascalCase | `Config`, `Secret`, `Client` |
| Unexported | camelCase | `ensureConfigExists`, `splitRepos` |
| Constants | PascalCase | `AppName = "gh-secrets-sync"` |
| Interfaces | PascalCase + -er | `Executor`, `Reader` |
| Test files | `*_test.go` external | `package foo_test` |

### Types & Functions

- Use explicit receiver names: `func (cfg *Config)` not `func (c *Config)`.
- Use functional options: `func WithX(opt) Option`.
- Max 3-4 parameters; use options struct otherwise.
- Keep functions under 30 lines, max 20 cyclomatic complexity.

### Error Handling

- Return errors directly: `if err != nil { return err }`.
- Wrap with context: `fmt.Errorf("failed to X: %w", err)`.
- Do NOT log and return nil – let caller decide.
- `fmt.Fprintf`/`fmt.Fprintln` excluded from errcheck.

## Testing Guidelines

### Test Structure

Use **external test packages** (`package foo_test`) with **parallel subtests**:

```go
func TestConfig_AddSecret(t *testing.T) {
    t.Parallel()

    t.Run("adds new secret", func(t *testing.T) {
        t.Parallel()
        cfg := &config.Config{}
        err := cfg.AddSecret(config.Secret{Name: "FOO", Value: "bar"}, false)
        require.NoError(t, err)
        assert.Len(t, cfg.Secrets, 1)
    })

    t.Run("errors on duplicate without force", func(t *testing.T) {
        t.Parallel()
        cfg := &config.Config{Secrets: []config.Secret{{Name: "FOO", Value: "x"}}}
        err := cfg.AddSecret(config.Secret{Name: "FOO", Value: "y"}, false)
        require.Error(t, err)
    })
}
```

### Conventions

- Call `t.Parallel()` in parent and each subtest.
- Use `testify`: `assert` for assertions, `require` for fatal failures.
- Use `testutils.SetupConfig` for config file setup.
- Use **Executor interface** in `internal/gh` to mock gh CLI.
- Run `testutils.SetupConfig(t, cfg)` creates temp config file.

### Single Test Commands

```bash
go test ./internal/config -v
go test ./internal/commands/root -run TestNew -v
go test ./internal/commands/root -run TestNew/defaults -v
go test -coverprofile=coverage.out -coverpkg=./internal/commands/root ./internal/commands/root -v
```

## Code Quality

Linters in `.golangci.yml`: gofumpt, gci, errcheck, gocyclo, govet, staticcheck.

Run `task lint` before committing. Run `task lint-fix` to auto-fix.

## Conventional Commits

```
<type>([scope]): <description>
Types: feat, fix, docs, ci, refactor, test, chore
```
Examples:
```
feat(config): add UpdateSecret method
fix(gh): handle missing gh CLI gracefully
test(root): add subtests for New function
```

## Dependencies

- Go 1.25
- `github.com/urfave/cli/v3` – CLI framework
- `github.com/cli/go-gh/v2` – gh CLI wrapper
- `github.com/stretchr/testify` – Testing
- `gopkg.in/yaml.v3` – Config serialization
- `github.com/AlecAivazis/survey/v2` – Interactive prompts

## Coverage

Include only production packages:
```bash
go test -coverprofile=coverage.out -coverpkg=./internal/commands/...,./internal/config,./internal/app ./...
```

Or filter after:
```bash
go test -coverprofile=raw.out ./...
grep -v '/testutils/' raw.out > coverage.out
go tool cover -func=coverage.out
```

## Getting Help

- Taskfile: `task -l`
- Devenv: `devenv shell`