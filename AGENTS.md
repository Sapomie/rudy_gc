# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` contains entrypoints like `cmd/api` (main service) and `cmd/db_migrate`.
- `internal/` holds core application code (config, domain, repo, transport, observability).
- `pkg/` hosts shared packages for reuse outside `internal/`.
- `data/` includes generated models and local data assets (see `data/modelx`).
- `ui/` contains server-rendered templates and static assets (`ui/templates`, `ui/static`).
- `deploy/` has Docker build assets; `make/` holds modular Make targets.
- `oldmodel/` and `z_text/` are legacy/notes; avoid new work here unless required.

## Build, Test, and Development Commands
- `go run ./cmd/api` starts the API service locally.
- `make -f make/dev.mk run` runs the same entrypoint via Make.
- `make -f make/model.mk gen-model` generates Go models from MySQL via `goctl`.
- `make -f make/migrate.mk auto-migrate` runs database migrations.
- `make -f make/docker.mk docker-build` builds the API image using `deploy/docker/api.Dockerfile`.
- `make -f make/cache.mk clear-cache` clears Redis keys matching `cache:rudyGc:*`.

## Coding Style & Naming Conventions
- Go code follows standard `gofmt` formatting; run `gofmt` on modified `.go` files.
- Package names are short, lowercase, and aligned to folder names.
- Exported identifiers use `PascalCase`; unexported use `camelCase`.
- Templates live in `ui/templates/pages` with shared fragments in `ui/templates/partials`.
- Static assets are organized by type under `ui/static/` (e.g., `css`, `js`, `image`).

## Testing Guidelines
- No `_test.go` files are present yet; add tests alongside the package under test.
- Use `go test ./...` for package-level tests when you add them.
- Name tests `TestXxx` and use table-driven patterns for multi-case coverage.

## Commit & Pull Request Guidelines
- Recent commits use conventional prefixes like `feat`, `refactor`, and scoped forms like `feat(ui)`; some are simple `add` messages.
- Prefer `type(scope): concise summary` with imperative phrasing (e.g., `feat(api): add movie cache warmup`).
- PRs should include a clear description, link to any issue, and screenshots for UI changes.
- Mention any special setup (DB/Redis) and the commands you ran to validate.

## Configuration Notes
- Default DB/Redis settings live in `make/common.mk` (e.g., `DB_URL`, `REDIS_HOST`).
- Override with environment variables or `make` arguments: `make -f make/model.mk gen-model DB_URL=...`.
