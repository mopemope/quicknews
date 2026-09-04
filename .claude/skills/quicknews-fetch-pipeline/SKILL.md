---
name: quicknews-fetch-pipeline
description: Use when changing the quicknews fetch pipeline (feed fetching, dedup, summarization, audio generation, org export) in cmd/fetch and you need the processing chain order and extension points.
---

# quicknews fetch pipeline

- 処理チェーンの正本(chain 順序、重複排除、リトライ方針)は [cmd/fetch/AGENTS.md](../../../cmd/fetch/AGENTS.md)。まずそこを読む。この skill ではそれを複製しない。

## 拡張の型(AGENTS.md に書かれていない補足)
- 新処理の追加位置は chain のどの段かで決める。記事取得時のみなら `ArticleProcessor.Process`、summary 生成後なら `processSummary` の末尾。
- 外部サービス依存は interface(`gemini.Summarizer`)と関数フィールドで注入し、`cmd/fetch` 内に実 API 呼び出しを書かない。
- TTY / 非 TTY 両パスがある(`cmd/fetch.go`)。ロジックは processor 側に置き、UI 分岐を processor に持ち込まない。
- 共有 slice への mutex 保護と `stderrors.Join` でのエラー集約を維持する。並列度変更は `cmd/fetch.go` 側の progress UI との整合も確認する。

## 検証
- `go test ./cmd/...`
- 音声 / 要約の動線を変えたら `go test ./gemini ./tts ./models/summary`
- 並行処理を変えたら `make test-race`
