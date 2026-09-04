# quicknews Agent Guide

Codex / opencode / Claude Code 共通のガイド。

## 優先ルール
- チャットは日本語で行う。
- Python は使わない。確認、検証、補助は shell と Go で行う。
- 秘密情報を含む `.env`、`.envrc`、個人用 config、ログファイルは原則開かない。必要な設定形は `config.example.toml` と [docs/codex/config-and-secrets.md](docs/codex/config-and-secrets.md) を見る。
- 生成物は直接編集しない。`ent/` 配下は原則 generated code なので、schema 変更は `ent/schema/` から始め、`make gen` で再生成する。
- 変更前に関係する最小範囲だけ読む。最初から `ent/` 全体や大きい TUI ファイルを広く読まない。

## タスク別の最短参照先
- CLI コマンド追加・修正: `main.go` → `cmd/` → [cmd/AGENTS.md](cmd/AGENTS.md)
- repository / data access の修正: `models/` → [models/AGENTS.md](models/AGENTS.md)
- fetch / 要約 / 音声生成: `cmd/fetch/` → `gemini/` / `tts/` / `models/summary/`
- TUI 表示・キー操作: `tui/model.go`、`tui/update_handlers.go`、各 view ファイル → [tui/AGENTS.md](tui/AGENTS.md)
- DB schema / relation / migration の影響確認: `ent/schema/` → `models/` → [ent/AGENTS.md](ent/AGENTS.md)
- Gemini / TTS / R2 など外部連携: `gemini/`、`tts/`、`storage/` → [gemini/AGENTS.md](gemini/AGENTS.md)
- MCP server (`search_articles`): `cmd/mcp.go` → `mcpserver/server.go` → `models/article/`

## 見てはいけない場所
- `ent/article*.go`、`ent/feed*.go`、`ent/summary*.go` などの generated code を初手で読まない。
- `quicknews` バイナリ、`quicknews.log`、`.git/` 配下、個人用 secret file を読まない。
- 仕様確認のために README 全体を毎回読み直さない。必要箇所だけ拾う。

## 標準検証
- 最速確認: `make test-fast`
- 全体確認: `make test-all`
- 外部連携込みの確認: `make test-integration`
- config 読み込み smoke test: `make smoke-config`
- lint: `make lint`(未インストール環境では `go vet ./...` のみ実行される)
- formatter: `gofmt -w <files>`
- 単一テスト: `GOCACHE=/tmp/quicknews-go-build go test ./<pkg> -run <TestName>`

## 変更種別ごとの追加確認
- CLI 変更: 該当 package test と `make smoke-config`
- TUI 変更: `go test ./tui/... ./cmd/...`
- schema / repository 変更: `make gen` 後に `go test ./database ./models/... ./cmd/...`
- Gemini / TTS / storage 変更: `go test ./gemini ./tts ./storage`

## Skills (opencode / Claude Code)
- `.claude/skills/` 配下にタスク別 skill がある。該当タスクでは最初に読む。
  - `quicknews-routing`: ファイル導線の判断
  - `quicknews-schema-change`: schema 変更の手順と checklist
  - `quicknews-integration-test`: 外部連携の最小検証
  - `quicknews-cli-command`: サブコマンド追加の定型手順

## 追加資料
- [docs/codex/task-routing.md](docs/codex/task-routing.md)
- [docs/codex/verification-matrix.md](docs/codex/verification-matrix.md)
- [docs/codex/config-and-secrets.md](docs/codex/config-and-secrets.md)
- [docs/codex/product-context.md](docs/codex/product-context.md)
- [docs/codex/evals.md](docs/codex/evals.md)
