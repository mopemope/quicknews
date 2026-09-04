# ent / data Layer Guide

## まず見る場所
- schema 定義: `ent/schema/*.go`
- transaction helper: `database/db.go`
- repository 実装: `models/*/*.go`

## 強いルール
- `ent/schema/` 以外の `ent/` generated code は手編集しない。
- schema を変えたら `make gen`(= `go generate ./ent`)で再生成する。ent codegen は `go.mod` の `tool` ディレクティブで固定されており、`go.sum` に差分は出ない。
- relation や field を変えるときは、影響する repository と test を合わせて更新する。
- schema 変更時は generated code を読むより、`models/` 側の query と save path を確認する。

## 最小検証
- `go test ./database ./models/...`
- コマンドまで影響する場合は `go test ./cmd/...`
