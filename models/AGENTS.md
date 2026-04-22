# models Layer Guide

## まず見る場所
- repository interface と query 実装: `models/*/*.go`
- transaction 境界: `database/db.go`
- schema の根: `ent/schema/*.go`

## この層で守ること
- query の意図と eager loading を先に確認する。`WithFeed()` / `WithSummary()` の有無を見落とさない。
- repository 変更で済むなら command や TUI へ不用意に広げない。
- schema 変更が絡む場合だけ `ent/schema/` を見て、generated `ent/` は手編集しない。

## 最小検証
- `go test ./models/... ./database`
- command まで波及したら `go test ./cmd/... ./models/...`
