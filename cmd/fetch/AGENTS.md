# fetch pipeline Guide

## まず見る場所
- feed 収集と並列 fetch: `feed_processor.go`
- 記事単位の処理: `article_processor.go`
- orchestration と progress UI: `cmd/fetch.go`
- progress UI 本体: `tui/progress/`(single.go / parallel.go)

## 処理の流れ
- `FeedProcessor.GetItems`: 全 feed を取得(bookmark feed は skip)し、`pond.NewPool(5)` で並列 fetch。`items` と `errs` は mutex で保護し、`stderrors.Join` でまとめる。
- 各 feed の item は `QueueItemWrapper`(`progress.QueueItem` 実装)に包まれてキューに入る。
- `ArticleProcessor.Process`: `GetFromURL` で重複チェック → 未保存なら記事保存 → `Edges.Summary == nil` なら要約生成。
- 要約後のチェーン順序: summary 保存 → (`config.SaveAudioData` が有効なら) `summary.SaveAudioData` + `UpdateAudioFile` → `org.ExportOrg`。

## この層で守ること
- 重複排除は `GetFromURL` が正。新規保存の可否判断を別の query で増やさない。
- Gemini クライアントは `gemini.Summarizer` interface 越しに受け、`newSummarizer` 関数フィールドで差し替え可能に保つ(test はこれで fake を注入)。
- リトライは `gemini.SummarizeWithRetry` に任せる。個別に sleep/retry を書かない。
- 並列度を変える場合は `cmd/fetch.go` 側の progress UI(worker 数、`itemCount > 50` の分岐)との整合も確認する。
- エラーは個別 item で握りつぶさず `stderrors.Join` で上位に伝搬する。

## 読みすぎ防止
- `ent/` generated code は開かない。repository の挙動は `models/` の interface と実装で確認する。
- progress UI の見た目を変える以外の理由で `tui/progress/` 全体を読まない。

## 最小検証
- `go test ./cmd/...`(processor の test は `cmd/fetch/` 配下)
- 要約・音声の動線を変えたら `go test ./gemini ./tts ./models/summary`
