---
name: quicknews-fetch-pipeline
description: Use when changing the quicknews fetch pipeline (feed fetching, dedup, summarization, audio generation, org export) in cmd/fetch and you need the processing chain order and extension points.
---

# quicknews fetch pipeline

## 処理チェーン(この順序を壊さない)
1. `FeedProcessor.GetItems`: 全 feed 取得 → bookmark feed を skip → pond pool(並列度 5)で `gofeed` により feed fetch → feed 情報更新。
2. 各 item を `QueueItemWrapper` としてキューへ。
3. `ArticleProcessor.Process`: `GetFromURL` で重複チェック → 未保存なら記事保存。
4. summary が無い記事のみ: `gemini.SummarizeWithRetry` で要約 → summary 保存。
5. `config.SaveAudioData` が有効なら: 長さ上限(`summary.MaxAudioTextLength`)チェック → `summary.SaveAudioData` → `UpdateAudioFile`。
6. `org.ExportOrg` で export。

## 拡張の型
- 新処理の追加位置は chain のどの段かで決める。記事取得時のみなら `ArticleProcessor.Process`、summary 生成後なら `processSummary` の末尾。
- 外部サービス依存は interface(`gemini.Summarizer`)と関数フィールドで注入し、`cmd/fetch` 内に実 API 呼び出しを書かない。
- 非同期・並列処理に触れたら共有 slice への mutex 保護と `stderrors.Join` でのエラー集約を維持する。
- TTY / 非 TTY 両パスがある(`cmd/fetch.go`)。ロジックは processor 側に置き、UI 分岐を processor に持ち込まない。

## 注意
- 重複排除の正は `articleRepos.GetFromURL`。別 query で独自の存在チェックを書かない。
- リトライは `gemini.SummarizeWithRetry` に集約済み。
- 詳細は `cmd/fetch/AGENTS.md` を参照。

## 検証
- `go test ./cmd/...`
- 音声 / 要約の動線を変えたら `go test ./gemini ./tts ./models/summary`
- 並行処理を変えたら `make test-race`
