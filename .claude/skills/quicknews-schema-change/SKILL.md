---
name: quicknews-schema-change
description: Use when changing quicknews database fields, relations, repository queries, or Ent-backed persistence and you need to avoid editing generated code or missing affected repositories and tests.
---

# quicknews schema change

- Start at `ent/schema/`, not generated `ent/*.go`.
- After choosing the schema file, inspect the matching repository under `models/` and transaction helper usage in `database/db.go`.
- Read [references/checklist.md](references/checklist.md) before editing.
- Do not hand-edit generated Ent files.
- Validate with the smallest command that covers the touched repositories before running `go test ./...`.
