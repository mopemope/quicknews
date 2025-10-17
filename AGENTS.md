# Repository Guidelines

## Project Structure & Module Organization
- `main.go` wires Kong CLI parsing, config loading, Ent migrations, and binds shared dependencies before dispatching to subcommands.
- `cmd/` holds the CLI verbs (`add`, `fetch`, `read`, `play`, `publish`, etc.); look here first when adjusting runtime behaviour.
- `tui/` contains Bubble Tea models and views for the terminal reader; UI tweaks belong here.
- `ent/`, `models/`, and `database/` cover data access (Ent schemas, repositories, migrations) backed by SQLite; keep schema updates coordinated across these folders.
- `gemini/`, `tts/`, `scraper/`, `rss/`, and `storage/` encapsulate integrations (LLM summarisation, TTS engines, scraping, feed parsing, and persistence helpers).

## Build, Test, and Development Commands
- `go build -o quicknews .` produces the CLI binary with all subcommands.
- `go run . fetch --debug` is the quickest way to exercise live fetching with verbose logs.
- `go test ./...` runs unit tests; network-bound Gemini tests auto-skip without credentials.
- `go test ./gemini -run TestNewClient_NoApiKey` verifies failure paths without hitting external services.

## Coding Style & Naming Conventions
- Follow idiomatic Go: tabs for indentation, exported identifiers use `CamelCase`, internal helpers use `lowerCamelCase`.
- Always run `gofmt -w` or `go fmt ./...` before submitting; imports should stay alphabetised (use `goimports` if available).
- Configuration structs (`config.Config`, `cmd.*Cmd`) mirror TOML/env keys—keep field names stable and prefer explicit types over interface{}.

## Testing Guidelines
- Primary tests live in `gemini/` and `tts/`; integration cases require `GEMINI_API_KEY` and may produce TTS audio.
- Export `GEMINI_API_KEY=...` (and set any Google credentials paths) before running tests that hit external APIs.
- Prefer table-driven tests for new logic; name files `*_test.go` beside tested code and functions `TestFeature_Scenario` for clarity.
- Record expected side effects (e.g., generated audio files) in `/tmp` and clean them up within tests to keep the workspace tidy.

## Commit & Pull Request Guidelines
- Commits follow Conventional Commits (`feat:`, `fix:`, `chore:`); scope components when useful (e.g. `feat(cmd/publish): ...`).
- Keep commit messages imperative and focused on one change; split sweeping refactors from behavioural updates.
- Pull requests should describe the user impact, note required config/env changes, and include screenshots or TUI recordings when UI behaviour changes.

## Configuration & Secrets
- Default paths assume `~/.config/quicknews/config.toml` and `~/quicknews.db`; document overrides in PRs.
- Never commit API keys or service JSON; rely on environment variables (`GEMINI_API_KEY`, `GOOGLE_APPLICATION_CREDENTIALS`) and sample values in docs only.
