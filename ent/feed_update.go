// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/predicate"
	"github.com/mopemope/quicknews/ent/summary"
)

// FeedUpdate is the builder for updating Feed entities.
type FeedUpdate struct {
	config
	hooks    []Hook
	mutation *FeedMutation
}

// Where appends a list predicates to the FeedUpdate builder.
func (fu *FeedUpdate) Where(ps ...predicate.Feed) *FeedUpdate {
	fu.mutation.Where(ps...)
	return fu
}

// SetURL sets the "url" field.
func (fu *FeedUpdate) SetURL(s string) *FeedUpdate {
	fu.mutation.SetURL(s)
	return fu
}

// SetNillableURL sets the "url" field if the given value is not nil.
func (fu *FeedUpdate) SetNillableURL(s *string) *FeedUpdate {
	if s != nil {
		fu.SetURL(*s)
	}
	return fu
}

// SetTitle sets the "title" field.
func (fu *FeedUpdate) SetTitle(s string) *FeedUpdate {
	fu.mutation.SetTitle(s)
	return fu
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (fu *FeedUpdate) SetNillableTitle(s *string) *FeedUpdate {
	if s != nil {
		fu.SetTitle(*s)
	}
	return fu
}

// SetDescription sets the "description" field.
func (fu *FeedUpdate) SetDescription(s string) *FeedUpdate {
	fu.mutation.SetDescription(s)
	return fu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (fu *FeedUpdate) SetNillableDescription(s *string) *FeedUpdate {
	if s != nil {
		fu.SetDescription(*s)
	}
	return fu
}

// ClearDescription clears the value of the "description" field.
func (fu *FeedUpdate) ClearDescription() *FeedUpdate {
	fu.mutation.ClearDescription()
	return fu
}

// SetLink sets the "link" field.
func (fu *FeedUpdate) SetLink(s string) *FeedUpdate {
	fu.mutation.SetLink(s)
	return fu
}

// SetNillableLink sets the "link" field if the given value is not nil.
func (fu *FeedUpdate) SetNillableLink(s *string) *FeedUpdate {
	if s != nil {
		fu.SetLink(*s)
	}
	return fu
}

// ClearLink clears the value of the "link" field.
func (fu *FeedUpdate) ClearLink() *FeedUpdate {
	fu.mutation.ClearLink()
	return fu
}

// SetOrder sets the "order" field.
func (fu *FeedUpdate) SetOrder(i int) *FeedUpdate {
	fu.mutation.ResetOrder()
	fu.mutation.SetOrder(i)
	return fu
}

// SetNillableOrder sets the "order" field if the given value is not nil.
func (fu *FeedUpdate) SetNillableOrder(i *int) *FeedUpdate {
	if i != nil {
		fu.SetOrder(*i)
	}
	return fu
}

// AddOrder adds i to the "order" field.
func (fu *FeedUpdate) AddOrder(i int) *FeedUpdate {
	fu.mutation.AddOrder(i)
	return fu
}

// SetUpdatedAt sets the "updated_at" field.
func (fu *FeedUpdate) SetUpdatedAt(t time.Time) *FeedUpdate {
	fu.mutation.SetUpdatedAt(t)
	return fu
}

// AddArticleIDs adds the "articles" edge to the Article entity by IDs.
func (fu *FeedUpdate) AddArticleIDs(ids ...uuid.UUID) *FeedUpdate {
	fu.mutation.AddArticleIDs(ids...)
	return fu
}

// AddArticles adds the "articles" edges to the Article entity.
func (fu *FeedUpdate) AddArticles(a ...*Article) *FeedUpdate {
	ids := make([]uuid.UUID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return fu.AddArticleIDs(ids...)
}

// AddSummaryIDs adds the "summaries" edge to the Summary entity by IDs.
func (fu *FeedUpdate) AddSummaryIDs(ids ...uuid.UUID) *FeedUpdate {
	fu.mutation.AddSummaryIDs(ids...)
	return fu
}

// AddSummaries adds the "summaries" edges to the Summary entity.
func (fu *FeedUpdate) AddSummaries(s ...*Summary) *FeedUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fu.AddSummaryIDs(ids...)
}

// Mutation returns the FeedMutation object of the builder.
func (fu *FeedUpdate) Mutation() *FeedMutation {
	return fu.mutation
}

// ClearArticles clears all "articles" edges to the Article entity.
func (fu *FeedUpdate) ClearArticles() *FeedUpdate {
	fu.mutation.ClearArticles()
	return fu
}

// RemoveArticleIDs removes the "articles" edge to Article entities by IDs.
func (fu *FeedUpdate) RemoveArticleIDs(ids ...uuid.UUID) *FeedUpdate {
	fu.mutation.RemoveArticleIDs(ids...)
	return fu
}

// RemoveArticles removes "articles" edges to Article entities.
func (fu *FeedUpdate) RemoveArticles(a ...*Article) *FeedUpdate {
	ids := make([]uuid.UUID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return fu.RemoveArticleIDs(ids...)
}

// ClearSummaries clears all "summaries" edges to the Summary entity.
func (fu *FeedUpdate) ClearSummaries() *FeedUpdate {
	fu.mutation.ClearSummaries()
	return fu
}

// RemoveSummaryIDs removes the "summaries" edge to Summary entities by IDs.
func (fu *FeedUpdate) RemoveSummaryIDs(ids ...uuid.UUID) *FeedUpdate {
	fu.mutation.RemoveSummaryIDs(ids...)
	return fu
}

// RemoveSummaries removes "summaries" edges to Summary entities.
func (fu *FeedUpdate) RemoveSummaries(s ...*Summary) *FeedUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fu.RemoveSummaryIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (fu *FeedUpdate) Save(ctx context.Context) (int, error) {
	fu.defaults()
	return withHooks(ctx, fu.sqlSave, fu.mutation, fu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fu *FeedUpdate) SaveX(ctx context.Context) int {
	affected, err := fu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (fu *FeedUpdate) Exec(ctx context.Context) error {
	_, err := fu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fu *FeedUpdate) ExecX(ctx context.Context) {
	if err := fu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (fu *FeedUpdate) defaults() {
	if _, ok := fu.mutation.UpdatedAt(); !ok {
		v := feed.UpdateDefaultUpdatedAt()
		fu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fu *FeedUpdate) check() error {
	if v, ok := fu.mutation.URL(); ok {
		if err := feed.URLValidator(v); err != nil {
			return &ValidationError{Name: "url", err: fmt.Errorf(`ent: validator failed for field "Feed.url": %w`, err)}
		}
	}
	if v, ok := fu.mutation.Title(); ok {
		if err := feed.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`ent: validator failed for field "Feed.title": %w`, err)}
		}
	}
	return nil
}

func (fu *FeedUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := fu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(feed.Table, feed.Columns, sqlgraph.NewFieldSpec(feed.FieldID, field.TypeUUID))
	if ps := fu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fu.mutation.URL(); ok {
		_spec.SetField(feed.FieldURL, field.TypeString, value)
	}
	if value, ok := fu.mutation.Title(); ok {
		_spec.SetField(feed.FieldTitle, field.TypeString, value)
	}
	if value, ok := fu.mutation.Description(); ok {
		_spec.SetField(feed.FieldDescription, field.TypeString, value)
	}
	if fu.mutation.DescriptionCleared() {
		_spec.ClearField(feed.FieldDescription, field.TypeString)
	}
	if value, ok := fu.mutation.Link(); ok {
		_spec.SetField(feed.FieldLink, field.TypeString, value)
	}
	if fu.mutation.LinkCleared() {
		_spec.ClearField(feed.FieldLink, field.TypeString)
	}
	if value, ok := fu.mutation.Order(); ok {
		_spec.SetField(feed.FieldOrder, field.TypeInt, value)
	}
	if value, ok := fu.mutation.AddedOrder(); ok {
		_spec.AddField(feed.FieldOrder, field.TypeInt, value)
	}
	if value, ok := fu.mutation.UpdatedAt(); ok {
		_spec.SetField(feed.FieldUpdatedAt, field.TypeTime, value)
	}
	if fu.mutation.ArticlesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.RemovedArticlesIDs(); len(nodes) > 0 && !fu.mutation.ArticlesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.ArticlesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if fu.mutation.SummariesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.RemovedSummariesIDs(); len(nodes) > 0 && !fu.mutation.SummariesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.SummariesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, fu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{feed.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	fu.mutation.done = true
	return n, nil
}

// FeedUpdateOne is the builder for updating a single Feed entity.
type FeedUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *FeedMutation
}

// SetURL sets the "url" field.
func (fuo *FeedUpdateOne) SetURL(s string) *FeedUpdateOne {
	fuo.mutation.SetURL(s)
	return fuo
}

// SetNillableURL sets the "url" field if the given value is not nil.
func (fuo *FeedUpdateOne) SetNillableURL(s *string) *FeedUpdateOne {
	if s != nil {
		fuo.SetURL(*s)
	}
	return fuo
}

// SetTitle sets the "title" field.
func (fuo *FeedUpdateOne) SetTitle(s string) *FeedUpdateOne {
	fuo.mutation.SetTitle(s)
	return fuo
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (fuo *FeedUpdateOne) SetNillableTitle(s *string) *FeedUpdateOne {
	if s != nil {
		fuo.SetTitle(*s)
	}
	return fuo
}

// SetDescription sets the "description" field.
func (fuo *FeedUpdateOne) SetDescription(s string) *FeedUpdateOne {
	fuo.mutation.SetDescription(s)
	return fuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (fuo *FeedUpdateOne) SetNillableDescription(s *string) *FeedUpdateOne {
	if s != nil {
		fuo.SetDescription(*s)
	}
	return fuo
}

// ClearDescription clears the value of the "description" field.
func (fuo *FeedUpdateOne) ClearDescription() *FeedUpdateOne {
	fuo.mutation.ClearDescription()
	return fuo
}

// SetLink sets the "link" field.
func (fuo *FeedUpdateOne) SetLink(s string) *FeedUpdateOne {
	fuo.mutation.SetLink(s)
	return fuo
}

// SetNillableLink sets the "link" field if the given value is not nil.
func (fuo *FeedUpdateOne) SetNillableLink(s *string) *FeedUpdateOne {
	if s != nil {
		fuo.SetLink(*s)
	}
	return fuo
}

// ClearLink clears the value of the "link" field.
func (fuo *FeedUpdateOne) ClearLink() *FeedUpdateOne {
	fuo.mutation.ClearLink()
	return fuo
}

// SetOrder sets the "order" field.
func (fuo *FeedUpdateOne) SetOrder(i int) *FeedUpdateOne {
	fuo.mutation.ResetOrder()
	fuo.mutation.SetOrder(i)
	return fuo
}

// SetNillableOrder sets the "order" field if the given value is not nil.
func (fuo *FeedUpdateOne) SetNillableOrder(i *int) *FeedUpdateOne {
	if i != nil {
		fuo.SetOrder(*i)
	}
	return fuo
}

// AddOrder adds i to the "order" field.
func (fuo *FeedUpdateOne) AddOrder(i int) *FeedUpdateOne {
	fuo.mutation.AddOrder(i)
	return fuo
}

// SetUpdatedAt sets the "updated_at" field.
func (fuo *FeedUpdateOne) SetUpdatedAt(t time.Time) *FeedUpdateOne {
	fuo.mutation.SetUpdatedAt(t)
	return fuo
}

// AddArticleIDs adds the "articles" edge to the Article entity by IDs.
func (fuo *FeedUpdateOne) AddArticleIDs(ids ...uuid.UUID) *FeedUpdateOne {
	fuo.mutation.AddArticleIDs(ids...)
	return fuo
}

// AddArticles adds the "articles" edges to the Article entity.
func (fuo *FeedUpdateOne) AddArticles(a ...*Article) *FeedUpdateOne {
	ids := make([]uuid.UUID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return fuo.AddArticleIDs(ids...)
}

// AddSummaryIDs adds the "summaries" edge to the Summary entity by IDs.
func (fuo *FeedUpdateOne) AddSummaryIDs(ids ...uuid.UUID) *FeedUpdateOne {
	fuo.mutation.AddSummaryIDs(ids...)
	return fuo
}

// AddSummaries adds the "summaries" edges to the Summary entity.
func (fuo *FeedUpdateOne) AddSummaries(s ...*Summary) *FeedUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fuo.AddSummaryIDs(ids...)
}

// Mutation returns the FeedMutation object of the builder.
func (fuo *FeedUpdateOne) Mutation() *FeedMutation {
	return fuo.mutation
}

// ClearArticles clears all "articles" edges to the Article entity.
func (fuo *FeedUpdateOne) ClearArticles() *FeedUpdateOne {
	fuo.mutation.ClearArticles()
	return fuo
}

// RemoveArticleIDs removes the "articles" edge to Article entities by IDs.
func (fuo *FeedUpdateOne) RemoveArticleIDs(ids ...uuid.UUID) *FeedUpdateOne {
	fuo.mutation.RemoveArticleIDs(ids...)
	return fuo
}

// RemoveArticles removes "articles" edges to Article entities.
func (fuo *FeedUpdateOne) RemoveArticles(a ...*Article) *FeedUpdateOne {
	ids := make([]uuid.UUID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return fuo.RemoveArticleIDs(ids...)
}

// ClearSummaries clears all "summaries" edges to the Summary entity.
func (fuo *FeedUpdateOne) ClearSummaries() *FeedUpdateOne {
	fuo.mutation.ClearSummaries()
	return fuo
}

// RemoveSummaryIDs removes the "summaries" edge to Summary entities by IDs.
func (fuo *FeedUpdateOne) RemoveSummaryIDs(ids ...uuid.UUID) *FeedUpdateOne {
	fuo.mutation.RemoveSummaryIDs(ids...)
	return fuo
}

// RemoveSummaries removes "summaries" edges to Summary entities.
func (fuo *FeedUpdateOne) RemoveSummaries(s ...*Summary) *FeedUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fuo.RemoveSummaryIDs(ids...)
}

// Where appends a list predicates to the FeedUpdate builder.
func (fuo *FeedUpdateOne) Where(ps ...predicate.Feed) *FeedUpdateOne {
	fuo.mutation.Where(ps...)
	return fuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (fuo *FeedUpdateOne) Select(field string, fields ...string) *FeedUpdateOne {
	fuo.fields = append([]string{field}, fields...)
	return fuo
}

// Save executes the query and returns the updated Feed entity.
func (fuo *FeedUpdateOne) Save(ctx context.Context) (*Feed, error) {
	fuo.defaults()
	return withHooks(ctx, fuo.sqlSave, fuo.mutation, fuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fuo *FeedUpdateOne) SaveX(ctx context.Context) *Feed {
	node, err := fuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (fuo *FeedUpdateOne) Exec(ctx context.Context) error {
	_, err := fuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fuo *FeedUpdateOne) ExecX(ctx context.Context) {
	if err := fuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (fuo *FeedUpdateOne) defaults() {
	if _, ok := fuo.mutation.UpdatedAt(); !ok {
		v := feed.UpdateDefaultUpdatedAt()
		fuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fuo *FeedUpdateOne) check() error {
	if v, ok := fuo.mutation.URL(); ok {
		if err := feed.URLValidator(v); err != nil {
			return &ValidationError{Name: "url", err: fmt.Errorf(`ent: validator failed for field "Feed.url": %w`, err)}
		}
	}
	if v, ok := fuo.mutation.Title(); ok {
		if err := feed.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`ent: validator failed for field "Feed.title": %w`, err)}
		}
	}
	return nil
}

func (fuo *FeedUpdateOne) sqlSave(ctx context.Context) (_node *Feed, err error) {
	if err := fuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(feed.Table, feed.Columns, sqlgraph.NewFieldSpec(feed.FieldID, field.TypeUUID))
	id, ok := fuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Feed.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := fuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, feed.FieldID)
		for _, f := range fields {
			if !feed.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != feed.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := fuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fuo.mutation.URL(); ok {
		_spec.SetField(feed.FieldURL, field.TypeString, value)
	}
	if value, ok := fuo.mutation.Title(); ok {
		_spec.SetField(feed.FieldTitle, field.TypeString, value)
	}
	if value, ok := fuo.mutation.Description(); ok {
		_spec.SetField(feed.FieldDescription, field.TypeString, value)
	}
	if fuo.mutation.DescriptionCleared() {
		_spec.ClearField(feed.FieldDescription, field.TypeString)
	}
	if value, ok := fuo.mutation.Link(); ok {
		_spec.SetField(feed.FieldLink, field.TypeString, value)
	}
	if fuo.mutation.LinkCleared() {
		_spec.ClearField(feed.FieldLink, field.TypeString)
	}
	if value, ok := fuo.mutation.Order(); ok {
		_spec.SetField(feed.FieldOrder, field.TypeInt, value)
	}
	if value, ok := fuo.mutation.AddedOrder(); ok {
		_spec.AddField(feed.FieldOrder, field.TypeInt, value)
	}
	if value, ok := fuo.mutation.UpdatedAt(); ok {
		_spec.SetField(feed.FieldUpdatedAt, field.TypeTime, value)
	}
	if fuo.mutation.ArticlesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.RemovedArticlesIDs(); len(nodes) > 0 && !fuo.mutation.ArticlesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.ArticlesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.ArticlesTable,
			Columns: []string{feed.ArticlesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if fuo.mutation.SummariesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.RemovedSummariesIDs(); len(nodes) > 0 && !fuo.mutation.SummariesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.SummariesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   feed.SummariesTable,
			Columns: []string{feed.SummariesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Feed{config: fuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, fuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{feed.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	fuo.mutation.done = true
	return _node, nil
}
