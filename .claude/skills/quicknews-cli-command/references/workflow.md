# subcommand workflow

1. Read `main.go` and find the `CLI` struct. Add the new command field with kong tags
   (`cmd:""`, optional `aliases:""`, `help:"..."`).
2. Create `cmd/<name>.go`:
   - Exported struct `NameCmd` with kong tags for flags/args.
   - `func (c *NameCmd) Run(client *ent.Client) error` (add `*config.Config` if needed; it is bound too).
   - Use `ctx := RunContext()` for cancellation, `log/slog` for logging, wrap errors with `%w`.
3. If the command needs new queries, add repository methods under `models/` first,
   with tests in the same package. Do not put SQL or ent query logic in `cmd/`.
4. Branch-heavy logic: extract pure functions so they are unit-testable (see `cmd/publish.go`).
5. Tests:
   - Add table tests in `cmd/<name>_test.go` for parsing/date/pure logic.
   - Run `go test ./cmd/...`.
6. If global flags or config parsing changed, run `make smoke-config` and check
   `config/config.go` + `config.example.toml` stay in sync.
7. Update `docs/codex/product-context.md` main command list and README subcommand
   section if the user-facing surface changed.
