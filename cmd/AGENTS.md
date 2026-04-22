# cmd Layer Guide

## まず見る場所
- CLI 定義: `main.go`
- 各サブコマンド: `cmd/*.go`
- fetch 本体: `cmd/fetch/`
- config 出力: `cmd/config_inspect.go`

## この層で守ること
- コマンド層は orchestration に寄せる。永続化や外部 API の詳細は `models/`、`gemini/`、`tts/`、`storage/` に押し込む。
- 追加ロジックが分岐だらけになるなら pure function に切り出して test を先に足す。
- TTY 判定や date 解決のような条件分岐は単体テストしやすい形に保つ。

## 読みすぎ防止
- schema の都合で困るまで `ent/` generated code は開かない。
- TUI 修正でない限り `tui/` を広く読まない。

## 最小検証
- `go test ./cmd/...`
- 必要なら `go test ./cmd/... ./models/...`
