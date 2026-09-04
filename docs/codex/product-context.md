# quicknews Product Context

## 概要
- `quicknews` は TUI ベースの RSS リーダー。
- 主目的は RSS フィードの購読、記事閲覧、要約生成、音声再生、Podcast 公開。
- 会話と運用の基本言語は日本語。

## 主要コマンド
- `add [URL...]`: RSS フィード追加
- `import [OPMLファイル]`: OPML から一括登録
- `fetch [-i INTERVAL]`: feed 更新、article 保存、必要なら要約・音声生成・Org export
- `read [--no-fetch]`: TUI 起動
- `play [--no-fetch] [--date YYYY-MM-DD]`: 未聴 summary の連続再生
- `bookmark [URL...]`: ブックマーク登録
- `export_audio`: 既存 summary の音声再生成
- `publish [YYYY-MM-DD]`: 複数日の音声を feed ごとに結合して R2 / podcast RSS へ公開
- `mcp`: stdio MCP server を起動。AI ツールから `search_articles` で記事検索が可能

## 技術スタック
- Go
- SQLite3
- Ent
- Bubble Tea / Bubbles
- Google Gemini
- Google TTS / Gemini TTS / VoiceVox
- Cloudflare R2

## 実装の見方
- CLI 入口: `main.go`, `cmd/`
- TUI: `tui/`
- 永続化: `models/`, `database/`, `ent/schema/`
- 外部連携: `gemini/`, `tts/`, `storage/`, `rss/`
- MCP: `cmd/mcp.go`, `mcpserver/`
