---
name: quicknews-integration-test
description: Use when changing Gemini, TTS, publish, Cloudflare R2, or config-driven integrations in quicknews and you need a minimal verification path that works with optional credentials.
---

# quicknews integration test

- Start with `config/config.go` and `config.example.toml` to understand the config shape without reading secrets.
- For Gemini changes, inspect `gemini/` and any caller in `cmd/fetch/`.
- For TTS changes, inspect `tts/` and `models/summary/summary.go`.
- For publish or R2 changes, inspect `cmd/publish.go`, `storage/r2.go`, and `rss/`.
- Read [references/verification.md](references/verification.md) to pick the smallest useful test command.
- Keep env-dependent tests skippable when credentials are missing.
