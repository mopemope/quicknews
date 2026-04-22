GOCACHE ?= /tmp/quicknews-go-build

test-fast:
	GOCACHE=$(GOCACHE) go test ./config ./database ./gemini ./log ./models/... ./rss ./scraper ./storage ./tts ./tui/components

test-all:
	GOCACHE=$(GOCACHE) go test ./...

test-integration:
	GOCACHE=$(GOCACHE) go test ./gemini ./tts ./storage ./cmd/...

smoke-config:
	GOCACHE=$(GOCACHE) go run . --config ./config.example.toml --log /tmp/quicknews.log config --format json >/dev/null
