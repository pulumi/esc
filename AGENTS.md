# Agent Instructions

## What this repo is
Pulumi ESC (Environments, Secrets, and Configuration) — a Go CLI tool and core evaluator for centralized secrets management and orchestration across cloud environments. Single binary, published via goreleaser.

## Start here
- `cmd/esc/main.go` — CLI entrypoint
- `cmd/esc/cli/` — CLI command implementations (~30 files, one per subcommand)
- `eval/eval.go` — core evaluation engine (largest file, ~49KB)
- `ast/` — abstract syntax tree for environment documents
- `schema/` — JSON schema validation
- `syntax/` — syntax parsing and YAML encoding
- `Makefile` — all dev commands

## Command canon
- Format: `make format`
- Lint: `make lint` (runs `lint-copyright` + `lint-golang`)
- Test (fast, unit): `make test` (runs with `-short -count 1 -parallel 10`)
- Test (full, with race + coverage): `make test_cover`
- Build: `make build` (installs `esc` binary with version stamp)
- Pre-commit check: `make format && make lint && make test`

## Key invariants
- Root-level Go files (`environment.go`, `expr.go`, `value.go`, `provider.go`) define the public API surface. Changes here affect downstream consumers and the ESC SDK.
- `eval/eval.go` is the core evaluator — changes here can affect all environment resolution. Test thoroughly.
- Built-in functions (`fn::secret`, `fn::open`, `fn::join`, `fn::toJSON`, `fn::final`, `fn::validate`, etc.) are evaluated in `eval/`. Adding or modifying builtins requires tests in `eval/` testdata.
- Test snapshot files live in `testdata/` directories under each package. If behavior changes, update snapshot files and diff carefully.

## Forbidden actions
- Do not run `git push --force`, `git reset --hard`, or `rm -rf` without explicit approval.
- Do not skip linting or bypass pre-commit hooks (`--no-verify`).
- Do not modify `.goreleaser.yml` or `.github/workflows/` without explicit approval.
- Do not add external runtime dependencies without discussion.
- Do not fabricate test output or snapshot file content.
- Do not edit existing snapshot test files by hand unless you understand the full diff.

## Escalate immediately if
- A change touches root-level `.go` files (public API surface).
- Tests fail after two debugging attempts.
- Requirements are ambiguous or conflict with existing behavior.
- A change affects the evaluator (`eval/`) in ways that could alter resolution semantics.
- You need to modify CI workflows or release configuration.

## If you change...
- Any `.go` file → run `make format && make lint && make test`
- Root-level `.go` files (`environment.go`, `expr.go`, `value.go`, `provider.go`) → also run `make test_cover` (full suite with race detection)
- `go.mod` or `go.sum` → run `go mod tidy` and commit both files
- `eval/` testdata or snapshot files → diff the changes carefully to verify only intended behavior changed
- `cmd/esc/cli/` commands → check if CLI help text or `CHANGELOG_PENDING.md` needs updating
