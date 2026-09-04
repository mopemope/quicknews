---
name: quicknews-testing
description: Use when writing or updating quicknews Go tests and you need to pick the right pattern (enttest in-memory DB, hand-written fakes, table-driven, env-skip) per package type.
---

# quicknews testing

- パターン選択は [docs/codex/testing-conventions.md](../../../docs/codex/testing-conventions.md) が正本。まずそこを読む。
- repository / 永続化: `enttest.Open` + in-memory SQLite。DB を fake に置き換えない。
- command 層: 手書き fake + `var _ Interface = (*fake)(nil)`。分岐は pure function に切り出して table-driven で。
- 外部連携: credential 未設定時は `t.Skip`(CI が credential 無しで動く前提を維持)。
- assertion は testify。エラー検証は `require.ErrorContains`。
- 並行処理に触れたら `make test-race`。最終確認は `make test-all`。

既存の見本:
- fake + interface: `cmd/publish_test.go`
- enttest: `models/article/article_test.go`
- table-driven: `gemini/parse_test.go`, `mcpserver/server_test.go`
- env-skip: `gemini/gemini_test.go`
