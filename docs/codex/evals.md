# quicknews Codex Eval Tasks

## 使い方
以下の課題を Codex に投げ、初手の探索と検証が局所化できるかを見る。
検証コマンドだけ回したい場合は `./scripts/run-codex-evals.sh <eval-name|all>` を使う。

## Eval 1: CLI flag 追加 (`cli-flag`)
- 課題: `fetch` に dry-run flag を追加し、対象 feed 数だけ表示する。
- 最初に見るべき場所: `main.go`, `cmd/fetch.go`, `cmd/fetch/feed_processor.go`
- 見てはいけない場所: `ent/` generated code 全体
- 期待する検証: `go test ./cmd/...`

## Eval 2: TUI の一覧表示変更 (`tui-list`)
- 課題: article list に published date を追加する。
- 最初に見るべき場所: `tui/articlelist.go`, `tui/model.go`
- 見てはいけない場所: `storage/`, `gemini/`
- 期待する検証: `go test ./tui/...`

## Eval 3: Ent schema 変更 (`schema-change`)
- 課題: feed に表示用カテゴリ文字列を追加する。
- 最初に見るべき場所: `ent/schema/feed.go`, `models/feed/feed.go`
- 見てはいけない場所: `ent/feed_create.go` など generated code の手編集
- 期待する検証: `go test ./database ./models/... ./cmd/...`

## Eval 4: Gemini 要約挙動変更 (`gemini-prompt`)
- 課題: summarize prompt の既定文言を調整し、custom prompt 優先を維持する。
- 最初に見るべき場所: `gemini/gemini.go`, `config/config.go`
- 見てはいけない場所: `.env`, README 全文
- 期待する検証: `go test ./gemini ./config`

## Eval 5: publish / R2 変更 (`publish-r2`)
- 課題: publish 対象日リストの生成を関数化して unit test を追加する。
- 最初に見るべき場所: `cmd/publish.go`, `storage/r2.go`
- 見てはいけない場所: `tui/`
- 期待する検証: `go test ./cmd/... ./storage`
