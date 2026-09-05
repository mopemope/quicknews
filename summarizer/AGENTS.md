# integration Layer Guide

## 対象
- 要約生成: `summarizer/` (gemini / openai の両プロバイダ)
- 音声合成: `tts/`
- 公開 / R2: `storage/`, `cmd/publish.go`

## この層で守ること
- 資格情報や外部 API 呼び出しは interface 越しに差し替えやすく保つ。プロバイダの差し替えは `summarizer.New` ファクトリと config の `summarize_provider` で行う。
- テストは env が無ければ skip できる形を維持する。OpenAI クライアントは `httptest` + `option.WithBaseURL` で実 API 無しで検証できる。
- 実 API の仕様確認が不要なら、fake 実装や pure function を先に追加してロジックを閉じ込める。
- レスポンス解析 (`parseResponse` 系) と出力ルール (`defaultInstructionBlock`) は両プロバイダ共通。プロバイダ固有なのはクライアント生成とプロンプトの入力手段 (Gemini は URL + search grounding、OpenAI は `scraper.GetPageContent` で取得した本文埋め込み)。
- リトライは `SummarizeWithRetry` に任せる。4xx (408/429 を除く) は non-retryable として即失敗する。個別に sleep/retry を書かない。

## ファイル構成
- `summarizer.go`: `Summarizer` interface / `New` ファクトリ
- `interface.go`: interface 定義
- `gemini.go`: Gemini 実装 + `promptFor` / `applyPromptTemplate` / `parseResponse` 系
- `openai.go`: OpenAI 実装 (`openai-go/v3` Responses API、`openai_base_url` 対応、`SummarizeContent` で取得済みページを再利用)
- `content.go`: `ContentAwareSummarizer` / `WithContent` / `SummarizeContentWithRetry` / `MinContentRunes` ガード
- `retry.go`: `SummarizeWithRetry` / `retryableError`

## リトライ分類
- 429 / 408 / 5xx と未知のエラーはリトライする。
- 4xx (408/429 除く)、context cancellation、`NonRetryable() bool` を実装するエラー (`scraper.PermanentError`、`summarizer.permanentError` など) は即失敗。
- OpenAI パスは本文が `MinContentRunes` (200 runes) 未満のページを permanent エラーにする (bot 拒否・JS-only ページのゴミ要約を防止)。

## 読みすぎ防止
- 外部連携の修正でも `README.md` 全体や unrelated package を読み直さない。
- config の shape は `config/config.go` と `config.example.toml` を見る。

## 最小検証
- `go test ./summarizer ./tts ./storage`
- publish 変更時は `go test ./cmd/... ./storage`
