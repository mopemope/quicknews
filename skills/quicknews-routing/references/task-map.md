# quicknews task map

## CLI
- `main.go`
- `cmd/*.go`

## fetch / summary / audio
- `cmd/fetch/feed_processor.go`
- `cmd/fetch/article_processor.go`
- `models/article/article.go`
- `models/summary/summary.go`
- `gemini/gemini.go`
- `tts/tts.go`

## TUI
- `tui/model.go`
- `tui/update_handlers.go`
- target view file only

## schema / data
- `ent/schema/*.go`
- `models/*/*.go`
- `database/db.go`

## publish / podcast
- `cmd/publish.go`
- `storage/r2.go`
- `rss/rss.go`
