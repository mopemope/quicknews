# mcpserver Guide

## まず見る場所
- server と tool 定義: `server.go`
- CLI 配線: `cmd/mcp.go`(kong が bind した `*ent.Client` から article repository を組み立て `RunStdio` へ)

## この層で守ること
- repository 依存は `ArticleSearcher` interface で受ける。具象 `models/article` を直接参照しない(テストは fake で差し替え)。
- tool 追加の型: `mcp.AddTool(server, &mcp.Tool{...}, handler)`。Input/Output は struct + `jsonschema:` タグで記述する。
- 読み取り専用 tool は `ToolAnnotations` で `ReadOnlyHint: true` を宣言する。
- 入力の正規化(limit / offset の default と cap)は `normalizeSearchOptions` のような pure function に集める。既定値は `models/article` の `DefaultSearchLimit` / `MaxSearchLimit` を参照する。
- snippet や matching の判定も pure function(`containsAnyTerm`, `buildSnippet`)のまま保ち、table-driven test を書く。
- 出力のフィールド追加時は `SearchArticlesOutput` の struct tag と handler の組み立てを両方更新する。

## 読みすぎ防止
- `ent/` generated code は開かない。`WithFeed()` / `WithSummary()` の eager loading 有無は `models/article/article.go` の `Search` で確認する。

## 最小検証
- `go test ./mcpserver ./cmd/...`
- `models/article` の query を変えたら `go test ./models/article ./mcpserver`
