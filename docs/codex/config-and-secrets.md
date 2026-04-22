# quicknews Config And Secrets

## source of truth
- config shape: `config/config.go`
- 共有サンプル: `config.example.toml`
- 実行時確認: `quicknews config --format table`

## 読み取りルール
- `.env`、`.envrc`、個人の `~/.config/quicknews/config.toml` は原則読まない。
- secret 値そのものは不要。必要なのは key 名と fallback 順序だけ。
- 既定値や optional 設定は `config.example.toml` を参照する。

## 主な設定
- DB: `db`
- Gemini: `gemini_api_key`, `gemini_model`
- TTS: `use_gemini_tts`, `google_application_credentials`, `voicevox.*`
- export: `export_org`, `audio`
- publish: `cloudflare.*`, `podcast.*`

## 検証
- config 周りを触ったら `make smoke-config`
- 表示整形を触ったら `go test ./cmd/... ./config`
