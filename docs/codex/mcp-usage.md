# quicknews MCP Usage

## 概要
- 本プロダクト自身の MCP server を開発タスクで dogfooding する。
- 対象 DB は `config.example.toml` の `db = "/tmp/quicknews.db"`(dev DB)。ユーザーの実 DB や実 config を開かない。
- tool は `search_articles` のみ(読み取り専用)。

## 起動方法
- opencode / Claude Code からは `opencode.json` / `.mcp.json` に登録済みなので、自動で起動する。
- 手動確認: `go run . --config ./config.example.toml --log /tmp/quicknews-mcp.log mcp`

## search_articles の入力
| 引数 | 型 | 既定 | 説明 |
|---|---|---|---|
| `query` | string | 必須 | キーワード検索。複数語は AND 条件(部分一致)。 |
| `limit` | int | 10 | 返す記事数の上限。50 で cap(`models/article` の `DefaultSearchLimit` / `MaxSearchLimit`)。 |
| `offset` | int | 0 | 先頭からスキップする件数。 |

## search_articles の出力
- `query` / `count` / `limit` / `offset` / `results[]`
- 各 result: `id`(UUID) / `title` / `url` / `feed_title` / `published_at`(RFC3339) / `snippet`(240 rune 以下) / `matched_fields`(`article.title` / `article.description` / `article.content` / `summary.title` / `summary.summary` のうち一致した field) / `readed` / `listened`

## 挙動の詳細を確認するとき
- tool 定義と正規化: `mcpserver/server.go`(`normalizeSearchOptions` / `buildSnippet`)
- 検索 query の実体: `models/article/article.go` の `Search`(eager loading は `WithFeed()` / `WithSummary()`)
- テスト: `go test ./mcpserver`

## 変更時
- 入出力の field を変えたら `mcpserver/AGENTS.md` と本ドキュメントを同期する。
