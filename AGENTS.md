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
- fetch pipeline (feed fetch / 要約 / 音声 / org export): `cmd/fetch/` → [cmd/fetch/AGENTS.md](cmd/fetch/AGENTS.md)
- TUI 表示・キー操作: `tui/model.go`、`tui/update_handlers.go`、各 view ファイル → [tui/AGENTS.md](tui/AGENTS.md)
- DB schema / relation / migration の影響確認: `ent/schema/` → `models/` → [ent/AGENTS.md](ent/AGENTS.md)
- Gemini / TTS / R2 など外部連携: `gemini/`、`tts/`、`storage/` → [gemini/AGENTS.md](gemini/AGENTS.md)
- MCP server (`search_articles`): `cmd/mcp.go` → [mcpserver/AGENTS.md](mcpserver/AGENTS.md)

## 見てはいけない場所
- `ent/article*.go`、`ent/feed*.go`、`ent/summary*.go` などの generated code を初手で読まない。
- `quicknews` バイナリ、`quicknews.log`、`.git/` 配下、個人用 secret file を読まない。
- 仕様確認のために README 全体を毎回読み直さない。必要箇所だけ拾う。

## 標準検証
- 最速確認: `make test-fast`(config / database / gemini / log / models / rss / scraper / storage / tts / tui(components, progress) / cmd / cmd/fetch / mcpserver を回す)
- 全体確認: `make test-all`
- 外部連携込みの確認: `make test-integration`
- 並行処理 (fetch pipeline / TUI) に触れたら: `make test-race`
- config 読み込み smoke test: `make smoke-config`
- lint: `make lint`(go vet + golangci-lint。lint は CI と同じバージョンを `go run ...@v2.13.2` で実行する)
- commit 前: `make check`(lint + test-fast)
- formatter: `make fmt`(全 `.go` への gofmt)
- 単一テスト: `GOCACHE=/tmp/quicknews-go-build go test ./<pkg> -run <TestName>`

## 変更種別ごとの追加確認
- CLI 変更: 該当 package test と `make smoke-config`
- TUI 変更: `go test ./tui/... ./cmd/...`
- schema / repository 変更: `make gen` 後に `go test ./database ./models/... ./cmd/...`
- Gemini / TTS / storage 変更: `go test ./gemini ./tts ./storage`

## テストの書き方
- パッケージ種別ごとのパターンは [docs/codex/testing-conventions.md](docs/codex/testing-conventions.md) を参照。
- repository は `enttest` + in-memory SQLite、command 層は手書き fake、外部連携は env 未設定時 skip。

## コミット規約
- Conventional Commits 形式: `feat(scope): ...` / `fix: ...` / `chore: ...` / `refactor(scope): ...` / `test: ...`。
- scope は対象領域(`tui`, `fetch`, `gemini` など)。既存履歴に合わせる。

## ドキュメント・検証のメンテ規則
コード変更に応じて以下を更新する。エージェントが次回同じ drift を起こさないための規則。
- 新パッケージ追加: `Makefile` の `test-fast`(該当するなら `test-race` / `test-integration`)に追加し、[docs/codex/task-routing.md](docs/codex/task-routing.md) と [docs/codex/verification-matrix.md](docs/codex/verification-matrix.md) に導線を足す。
- 新サブコマンド追加: [cmd/AGENTS.md](cmd/AGENTS.md) の参照が必要なら更新し、[docs/codex/product-context.md](docs/codex/product-context.md) の主要コマンド一覧に足す。
- fetch 処理チェーン変更: 正本は [cmd/fetch/AGENTS.md](cmd/fetch/AGENTS.md)。そちらのみ更新する(skill は参照しない)。
- 検証コマンドの追加・変更: [docs/codex/verification-matrix.md](docs/codex/verification-matrix.md) を同期する。
- MCP tool の入出力変更: [docs/codex/mcp-usage.md](docs/codex/mcp-usage.md) を同期する。

## Skills (opencode / Claude Code)
- `.claude/skills/` 配下にタスク別 skill がある。該当タスクでは最初に読む。
  - `quicknews-routing`: ファイル導線の判断
  - `quicknews-schema-change`: schema 変更の手順と checklist
  - `quicknews-integration-test`: 外部連携の最小検証
  - `quicknews-cli-command`: サブコマンド追加の定型手順
  - `quicknews-tui-change`: TUI の view 遷移・キー追加
  - `quicknews-fetch-pipeline`: fetch 処理チェーンの拡張
  - `quicknews-testing`: テストパターンの選び方

## MCP dogfooding
- 開発中はエージェントから本プロダクト自身の MCP server (`search_articles`) を使える。`opencode.json` / `.mcp.json` に `go run . --config ./config.example.toml mcp` として登録済み。
- 開発タスクの記事検索はこの tool 経由にする(対象は `config.example.toml` の dev DB)。ユーザーの実 DB や実 config を直接開かない。
- 引数・返り値の詳細は [docs/codex/mcp-usage.md](docs/codex/mcp-usage.md) を参照。

## 追加資料
- [docs/codex/task-routing.md](docs/codex/task-routing.md)
- [docs/codex/verification-matrix.md](docs/codex/verification-matrix.md)
- [docs/codex/config-and-secrets.md](docs/codex/config-and-secrets.md)
- [docs/codex/testing-conventions.md](docs/codex/testing-conventions.md)
- [docs/codex/product-context.md](docs/codex/product-context.md)
- [docs/codex/mcp-usage.md](docs/codex/mcp-usage.md)
- [docs/codex/evals.md](docs/codex/evals.md)
