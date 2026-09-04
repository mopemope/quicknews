---
name: quicknews-cli-command
description: Use when adding or modifying a quicknews subcommand (kong CLI) and you need the standard wiring order for main.go, cmd structs, repository access, and the minimal tests to run.
---

# quicknews cli command

- CLI is kong-based. Register the command struct in `main.go` (`CLI`), then implement it in `cmd/<name>.go`.
- Follow `cmd/add.go` as the minimal template: command struct with kong tags, `Run(client *ent.Client) error`, `ctx := RunContext()`.
- Keep `cmd/` as orchestration only. Push persistence into `models/`, external APIs into `gemini/`, `tts/`, `storage/`.
- Dependencies arrive via kong binding (`kctx.Bind` in `main.go`): `*ent.Client`, `*config.Config`. Do not construct them inside commands.
- Read [references/workflow.md](references/workflow.md) for the step-by-step checklist before editing.
- Validate with `go test ./cmd/...` and `make smoke-config` when config parsing is touched.
