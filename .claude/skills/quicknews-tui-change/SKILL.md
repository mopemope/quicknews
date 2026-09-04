---
name: quicknews-tui-change
description: Use when changing the quicknews TUI (Bubble Tea) views, key bindings, message flow, or read/listen status, and you need the view state machine and key map without reading all view files.
---

# quicknews tui change

- State は `tui/model.go` の `model` に集約される。3 view(`feedList` / `articleList` / `summaryView`)はそれぞれ独立した model。
- 遷移はカスタム msg 駆動。msg の発行は各 view の Update、受信と view 切替は `tui/update_handlers.go` 側で行う。
- キーを追加するときは該当 view の `Update` 内 `switch msg.String()` に case を足し、[references/keymap.md](references/keymap.md) を更新する。
- 既読 / 聴取済みなど summary 状態の変更は `models/summary` の repository 経由で行う。TUI 内で ent query を直接書かない。
- 大きい view ファイル(`articlelist.go` など)は全文読まず、対象の `case` 部分だけ `rg` で絞って読む。
- 検証: `go test ./tui/...`。起動境界(`cmd/read.go`)に触れたら `go test ./cmd/... ./tui/...`。
