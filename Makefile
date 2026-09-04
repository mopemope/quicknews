GOCACHE ?= /tmp/quicknews-go-build

gen:
	GOCACHE=$(GOCACHE) go generate ./ent

fmt:
	gofmt -w $(shell git ls-files --cached --others --exclude-standard '*.go')

lint:
	GOCACHE=$(GOCACHE) go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		GOCACHE=$(GOCACHE) golangci-lint run; \
	else \
		echo "golangci-lint not installed; ran go vet only"; \
	fi

check: lint test-fast

test-fast:
	GOCACHE=$(GOCACHE) go test ./config ./database ./gemini ./log ./models/... ./rss ./scraper ./storage ./tts ./tui/components

test-all:
	GOCACHE=$(GOCACHE) go test ./...

test-race:
	GOCACHE=$(GOCACHE) go test -race ./cmd/... ./tui/...

test-integration:
	GOCACHE=$(GOCACHE) go test ./gemini ./tts ./storage ./cmd/...

smoke-config:
	GOCACHE=$(GOCACHE) go run . --config ./config.example.toml --log /tmp/quicknews.log config --format json >/dev/null
