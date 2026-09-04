GOCACHE ?= /tmp/quicknews-go-build

gen:
	GOCACHE=$(GOCACHE) go generate ./ent

fmt:
	gofmt -w $(shell git ls-files --cached --others --exclude-standard '*.go')

GOLANGCI_LINT_VERSION ?= v2.13.2

lint:
	GOCACHE=$(GOCACHE) go vet ./...
	GOCACHE=$(GOCACHE) go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

check: lint test-fast

test-fast:
	GOCACHE=$(GOCACHE) go test ./config ./database ./gemini ./log ./models/... ./rss ./scraper ./storage ./tts ./tui/components ./tui/progress ./cmd ./cmd/fetch ./mcpserver

test-all:
	GOCACHE=$(GOCACHE) go test ./...

test-race:
	GOCACHE=$(GOCACHE) go test -race ./cmd/... ./tui/...

test-integration:
	GOCACHE=$(GOCACHE) go test ./gemini ./tts ./storage ./cmd/...

smoke-config:
	GOCACHE=$(GOCACHE) go run . --config ./config.example.toml --log /tmp/quicknews.log config --format json >/dev/null
