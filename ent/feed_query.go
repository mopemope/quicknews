// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/predicate"
	"github.com/mopemope/quicknews/ent/summary"
)

// FeedQuery is the builder for querying Feed entities.
type FeedQuery struct {
	config
	ctx           *QueryContext
	order         []feed.OrderOption
	inters        []Interceptor
	predicates    []predicate.Feed
	withArticles  *ArticleQuery
	withSummaries *SummaryQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the FeedQuery builder.
func (fq *FeedQuery) Where(ps ...predicate.Feed) *FeedQuery {
	fq.predicates = append(fq.predicates, ps...)
	return fq
}

// Limit the number of records to be returned by this query.
func (fq *FeedQuery) Limit(limit int) *FeedQuery {
	fq.ctx.Limit = &limit
	return fq
}

// Offset to start from.
func (fq *FeedQuery) Offset(offset int) *FeedQuery {
	fq.ctx.Offset = &offset
	return fq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (fq *FeedQuery) Unique(unique bool) *FeedQuery {
	fq.ctx.Unique = &unique
	return fq
}

// Order specifies how the records should be ordered.
func (fq *FeedQuery) Order(o ...feed.OrderOption) *FeedQuery {
	fq.order = append(fq.order, o...)
	return fq
}

// QueryArticles chains the current query on the "articles" edge.
func (fq *FeedQuery) QueryArticles() *ArticleQuery {
	query := (&ArticleClient{config: fq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := fq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := fq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(feed.Table, feed.FieldID, selector),
			sqlgraph.To(article.Table, article.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, feed.ArticlesTable, feed.ArticlesColumn),
		)
		fromU = sqlgraph.SetNeighbors(fq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QuerySummaries chains the current query on the "summaries" edge.
func (fq *FeedQuery) QuerySummaries() *SummaryQuery {
	query := (&SummaryClient{config: fq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := fq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := fq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(feed.Table, feed.FieldID, selector),
			sqlgraph.To(summary.Table, summary.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, feed.SummariesTable, feed.SummariesColumn),
		)
		fromU = sqlgraph.SetNeighbors(fq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Feed entity from the query.
// Returns a *NotFoundError when no Feed was found.
func (fq *FeedQuery) First(ctx context.Context) (*Feed, error) {
	nodes, err := fq.Limit(1).All(setContextOp(ctx, fq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{feed.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (fq *FeedQuery) FirstX(ctx context.Context) *Feed {
	node, err := fq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Feed ID from the query.
// Returns a *NotFoundError when no Feed ID was found.
func (fq *FeedQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = fq.Limit(1).IDs(setContextOp(ctx, fq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{feed.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (fq *FeedQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := fq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Feed entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Feed entity is found.
// Returns a *NotFoundError when no Feed entities are found.
func (fq *FeedQuery) Only(ctx context.Context) (*Feed, error) {
	nodes, err := fq.Limit(2).All(setContextOp(ctx, fq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{feed.Label}
	default:
		return nil, &NotSingularError{feed.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (fq *FeedQuery) OnlyX(ctx context.Context) *Feed {
	node, err := fq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Feed ID in the query.
// Returns a *NotSingularError when more than one Feed ID is found.
// Returns a *NotFoundError when no entities are found.
func (fq *FeedQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = fq.Limit(2).IDs(setContextOp(ctx, fq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{feed.Label}
	default:
		err = &NotSingularError{feed.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (fq *FeedQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := fq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Feeds.
func (fq *FeedQuery) All(ctx context.Context) ([]*Feed, error) {
	ctx = setContextOp(ctx, fq.ctx, ent.OpQueryAll)
	if err := fq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Feed, *FeedQuery]()
	return withInterceptors[[]*Feed](ctx, fq, qr, fq.inters)
}

// AllX is like All, but panics if an error occurs.
func (fq *FeedQuery) AllX(ctx context.Context) []*Feed {
	nodes, err := fq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Feed IDs.
func (fq *FeedQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if fq.ctx.Unique == nil && fq.path != nil {
		fq.Unique(true)
	}
	ctx = setContextOp(ctx, fq.ctx, ent.OpQueryIDs)
	if err = fq.Select(feed.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (fq *FeedQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := fq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (fq *FeedQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, fq.ctx, ent.OpQueryCount)
	if err := fq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, fq, querierCount[*FeedQuery](), fq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (fq *FeedQuery) CountX(ctx context.Context) int {
	count, err := fq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (fq *FeedQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, fq.ctx, ent.OpQueryExist)
	switch _, err := fq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (fq *FeedQuery) ExistX(ctx context.Context) bool {
	exist, err := fq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the FeedQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (fq *FeedQuery) Clone() *FeedQuery {
	if fq == nil {
		return nil
	}
	return &FeedQuery{
		config:        fq.config,
		ctx:           fq.ctx.Clone(),
		order:         append([]feed.OrderOption{}, fq.order...),
		inters:        append([]Interceptor{}, fq.inters...),
		predicates:    append([]predicate.Feed{}, fq.predicates...),
		withArticles:  fq.withArticles.Clone(),
		withSummaries: fq.withSummaries.Clone(),
		// clone intermediate query.
		sql:  fq.sql.Clone(),
		path: fq.path,
	}
}

// WithArticles tells the query-builder to eager-load the nodes that are connected to
// the "articles" edge. The optional arguments are used to configure the query builder of the edge.
func (fq *FeedQuery) WithArticles(opts ...func(*ArticleQuery)) *FeedQuery {
	query := (&ArticleClient{config: fq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	fq.withArticles = query
	return fq
}

// WithSummaries tells the query-builder to eager-load the nodes that are connected to
// the "summaries" edge. The optional arguments are used to configure the query builder of the edge.
func (fq *FeedQuery) WithSummaries(opts ...func(*SummaryQuery)) *FeedQuery {
	query := (&SummaryClient{config: fq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	fq.withSummaries = query
	return fq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		URL string `json:"url,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Feed.Query().
//		GroupBy(feed.FieldURL).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (fq *FeedQuery) GroupBy(field string, fields ...string) *FeedGroupBy {
	fq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &FeedGroupBy{build: fq}
	grbuild.flds = &fq.ctx.Fields
	grbuild.label = feed.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		URL string `json:"url,omitempty"`
//	}
//
//	client.Feed.Query().
//		Select(feed.FieldURL).
//		Scan(ctx, &v)
func (fq *FeedQuery) Select(fields ...string) *FeedSelect {
	fq.ctx.Fields = append(fq.ctx.Fields, fields...)
	sbuild := &FeedSelect{FeedQuery: fq}
	sbuild.label = feed.Label
	sbuild.flds, sbuild.scan = &fq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a FeedSelect configured with the given aggregations.
func (fq *FeedQuery) Aggregate(fns ...AggregateFunc) *FeedSelect {
	return fq.Select().Aggregate(fns...)
}

func (fq *FeedQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range fq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, fq); err != nil {
				return err
			}
		}
	}
	for _, f := range fq.ctx.Fields {
		if !feed.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if fq.path != nil {
		prev, err := fq.path(ctx)
		if err != nil {
			return err
		}
		fq.sql = prev
	}
	return nil
}

func (fq *FeedQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Feed, error) {
	var (
		nodes       = []*Feed{}
		_spec       = fq.querySpec()
		loadedTypes = [2]bool{
			fq.withArticles != nil,
			fq.withSummaries != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Feed).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Feed{config: fq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, fq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := fq.withArticles; query != nil {
		if err := fq.loadArticles(ctx, query, nodes,
			func(n *Feed) { n.Edges.Articles = []*Article{} },
			func(n *Feed, e *Article) { n.Edges.Articles = append(n.Edges.Articles, e) }); err != nil {
			return nil, err
		}
	}
	if query := fq.withSummaries; query != nil {
		if err := fq.loadSummaries(ctx, query, nodes,
			func(n *Feed) { n.Edges.Summaries = []*Summary{} },
			func(n *Feed, e *Summary) { n.Edges.Summaries = append(n.Edges.Summaries, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (fq *FeedQuery) loadArticles(ctx context.Context, query *ArticleQuery, nodes []*Feed, init func(*Feed), assign func(*Feed, *Article)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*Feed)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Article(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(feed.ArticlesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.feed_articles
		if fk == nil {
			return fmt.Errorf(`foreign-key "feed_articles" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "feed_articles" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (fq *FeedQuery) loadSummaries(ctx context.Context, query *SummaryQuery, nodes []*Feed, init func(*Feed), assign func(*Feed, *Summary)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*Feed)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Summary(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(feed.SummariesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.feed_summaries
		if fk == nil {
			return fmt.Errorf(`foreign-key "feed_summaries" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "feed_summaries" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (fq *FeedQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := fq.querySpec()
	_spec.Node.Columns = fq.ctx.Fields
	if len(fq.ctx.Fields) > 0 {
		_spec.Unique = fq.ctx.Unique != nil && *fq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, fq.driver, _spec)
}

func (fq *FeedQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(feed.Table, feed.Columns, sqlgraph.NewFieldSpec(feed.FieldID, field.TypeUUID))
	_spec.From = fq.sql
	if unique := fq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if fq.path != nil {
		_spec.Unique = true
	}
	if fields := fq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, feed.FieldID)
		for i := range fields {
			if fields[i] != feed.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := fq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := fq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := fq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := fq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (fq *FeedQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(fq.driver.Dialect())
	t1 := builder.Table(feed.Table)
	columns := fq.ctx.Fields
	if len(columns) == 0 {
		columns = feed.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if fq.sql != nil {
		selector = fq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if fq.ctx.Unique != nil && *fq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range fq.predicates {
		p(selector)
	}
	for _, p := range fq.order {
		p(selector)
	}
	if offset := fq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := fq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// FeedGroupBy is the group-by builder for Feed entities.
type FeedGroupBy struct {
	selector
	build *FeedQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (fgb *FeedGroupBy) Aggregate(fns ...AggregateFunc) *FeedGroupBy {
	fgb.fns = append(fgb.fns, fns...)
	return fgb
}

// Scan applies the selector query and scans the result into the given value.
func (fgb *FeedGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, fgb.build.ctx, ent.OpQueryGroupBy)
	if err := fgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*FeedQuery, *FeedGroupBy](ctx, fgb.build, fgb, fgb.build.inters, v)
}

func (fgb *FeedGroupBy) sqlScan(ctx context.Context, root *FeedQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(fgb.fns))
	for _, fn := range fgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*fgb.flds)+len(fgb.fns))
		for _, f := range *fgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*fgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := fgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// FeedSelect is the builder for selecting fields of Feed entities.
type FeedSelect struct {
	*FeedQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (fs *FeedSelect) Aggregate(fns ...AggregateFunc) *FeedSelect {
	fs.fns = append(fs.fns, fns...)
	return fs
}

// Scan applies the selector query and scans the result into the given value.
func (fs *FeedSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, fs.ctx, ent.OpQuerySelect)
	if err := fs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*FeedQuery, *FeedSelect](ctx, fs.FeedQuery, fs, fs.inters, v)
}

func (fs *FeedSelect) sqlScan(ctx context.Context, root *FeedQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(fs.fns))
	for _, fn := range fs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*fs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := fs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
