# schema change checklist

- Identify the owning schema in `ent/schema/`.
- Find repository methods in `models/` that read or write the field.
- Check whether command-layer date, filter, or display logic depends on it.
- Add or update tests in the nearest `*_test.go`.
- Run `go test ./database ./models/...` at minimum.
