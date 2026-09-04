---
name: quicknews-routing
description: Use when working in the quicknews repository and you need the shortest file-routing path for CLI, TUI, fetch, schema, or publish tasks without wasting tokens on broad exploration.
---

# quicknews routing

- Start with `AGENTS.md` at the repo root, then open only the closest subdirectory `AGENTS.md` if the task is clearly scoped.
- Use [references/task-map.md](references/task-map.md) to choose the first 2-4 files to inspect.
- Avoid opening generated `ent/` files unless the task is specifically about generated output debugging.
- Avoid secrets and local noise: `.env`, `.envrc`, `quicknews.log`, built binaries.
- Prefer `rg` to jump to the exact symbol or command you need before opening large files.
