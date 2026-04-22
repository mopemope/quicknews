# tui Layer Guide

## まず見る場所
- 状態の中心: `tui/model.go`
- 入力処理: `tui/update_handlers.go`
- リスト表示: `tui/feedlist.go`, `tui/articlelist.go`
- 詳細表示: `tui/summaryview.go`
- 小さな部品: `tui/components/`

## この層で守ること
- 画面遷移と state 更新を優先して追う。見た目だけの変更で `cmd/` や `models/` に広く触らない。
- 変更は既存の Bubble Tea の流れに合わせる。新しい state を追加する場合は `model.go` と update handler の整合を最初に確認する。
- 大きい view ファイルを最初から全文読まず、対象の section を `rg` で絞って読む。

## 最小検証
- `go test ./tui/...`
- 影響がコマンド起動に及ぶなら `go test ./cmd/... ./tui/...`
