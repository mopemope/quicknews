# schema change checklist

- Identify the owning schema in `ent/schema/`.
- Find repository methods in `models/` that read or write the field.
- Check whether command-layer date, filter, or display logic depends on it.
- Regenerate with `make gen` (do not hand-edit generated `ent/*.go`). The ent codegen is pinned via the `tool` directive in `go.mod`, so no `go.sum` churn is expected.
- Add or update tests in the nearest `*_test.go`.
- Run `go test ./database ./models/...` at minimum.
