# quicknews Task Routing

## 目的
Codex が最初の探索で読む範囲を絞るための導線。

## 変更内容ごとの入口
- サブコマンド追加、flag 変更、実行フロー変更
  - `main.go`
  - `cmd/<name>.go`
  - 必要なら `cmd/fetch/`
- fetch の並列処理、要約生成、保存フロー
  - `cmd/fetch/feed_processor.go`
  - `cmd/fetch/article_processor.go`
  - `models/article/`, `models/summary/`
  - `gemini/`, `tts/`
- TUI の一覧表示、入力操作、レイアウト
  - `tui/model.go`
  - `tui/update_handlers.go`
  - 対象 view file のみ
- DB schema や relation 変更
  - `ent/schema/`
  - `models/`
  - `database/db.go`
- podcast publish / R2
  - `cmd/publish.go`
  - `storage/r2.go`
  - `rss/`
- MCP server (`search_articles`)
  - `cmd/mcp.go`
  - `mcpserver/server.go`
  - `models/article/`

## 初手で避ける場所
- `ent/` generated code 全体
- `quicknews.log`
- `quicknews` バイナリ
- `.env`, `.envrc`

## よくある依存の見方
- CLI から永続化へ: `main.go` → `cmd/` → `models/` → `database/`
- fetch から外部連携へ: `cmd/fetch/` → `gemini/` / `tts/` / `org/`
- publish から配信へ: `cmd/publish.go` → `tts/` → `storage/` → `rss/`
