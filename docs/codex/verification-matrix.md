# quicknews Verification Matrix

## 基本コマンド
- `make test-fast`
- `make test-all`
- `make test-integration`
- `make smoke-config`

## 変更種別ごとの最小検証

### CLI / command
- `go test ./cmd/...`
- 設定読み込みに触れたら `make smoke-config`

### Repository / DB / schema
- `go test ./database ./models/...`
- command へ波及したら `go test ./cmd/...`

### Gemini / TTS / R2
- `go test ./gemini ./tts ./storage`
- publish を触ったら `go test ./cmd/... ./storage`

### TUI
- `go test ./tui/...`
- command 境界まで触ったら `go test ./cmd/... ./tui/...`

## 補足
- `go test ./...` は最終確認。
- 外部 API 依存 test は環境変数未設定時に skip される前提を維持する。
