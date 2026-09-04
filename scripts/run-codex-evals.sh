#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

GOCACHE_DIR="${GOCACHE:-/tmp/quicknews-go-build}"

run_eval() {
	case "$1" in
		cli-flag)
			GOCACHE="$GOCACHE_DIR" go test ./cmd/...
			;;
		tui-list)
			GOCACHE="$GOCACHE_DIR" go test ./tui/...
			;;
		schema-change)
			GOCACHE="$GOCACHE_DIR" go test ./database ./models/... ./cmd/...
			;;
		gemini-prompt)
			GOCACHE="$GOCACHE_DIR" go test ./gemini ./config
			;;
		publish-r2)
			GOCACHE="$GOCACHE_DIR" go test ./cmd/... ./storage
			;;
		mcp-search)
			GOCACHE="$GOCACHE_DIR" go test ./mcpserver ./cmd/...
			;;
		*)
			echo "unknown eval: $1" >&2
			exit 1
			;;
	esac
}

main() {
	local target="${1:-all}"
	if [[ "$target" == "all" ]]; then
		for eval_name in cli-flag tui-list schema-change gemini-prompt publish-r2 mcp-search; do
			echo "==> $eval_name"
			run_eval "$eval_name"
		done
		return
	fi
	run_eval "$target"
}

main "$@"
