# integration Layer Guide

## 対象
- 要約生成: `gemini/`
- 音声合成: `tts/`
- 公開 / R2: `storage/`, `cmd/publish.go`

## この層で守ること
- 資格情報や外部 API 呼び出しは interface 越しに差し替えやすく保つ。
- テストは env が無ければ skip できる形を維持する。
- 実 API の仕様確認が不要なら、fake 実装や pure function を先に追加してロジックを閉じ込める。

## 読みすぎ防止
- 外部連携の修正でも `README.md` 全体や unrelated package を読み直さない。
- config の shape は `config/config.go` と `config.example.toml` を見る。

## 最小検証
- `go test ./gemini ./tts ./storage`
- publish 変更時は `go test ./cmd/... ./storage`
