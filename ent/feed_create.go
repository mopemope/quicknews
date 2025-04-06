// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/summary"
)

// FeedCreate is the builder for creating a Feed entity.
type FeedCreate struct {
	config
	mutation *FeedMutation
	hooks    []Hook
}

// SetURL sets the "url" field.
func (fc *FeedCreate) SetURL(s string) *FeedCreate {
	fc.mutation.SetURL(s)
	return fc
}

// SetTitle sets the "title" field.
func (fc *FeedCreate) SetTitle(s string) *FeedCreate {
	fc.mutation.SetTitle(s)
	return fc
}

// SetDescription sets the "description" field.
func (fc *FeedCreate) SetDescription(s string) *FeedCreate {
	fc.mutation.SetDescription(s)
	return fc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (fc *FeedCreate) SetNillableDescription(s *string) *FeedCreate {
	if s != nil {
		fc.SetDescription(*s)
	}
	return fc
}

// SetLink sets the "link" field.
func (fc *FeedCreate) SetLink(s string) *FeedCreate {
	fc.mutation.SetLink(s)
	return fc
}

// SetNillableLink sets the "link" field if the given value is not nil.
func (fc *FeedCreate) SetNillableLink(s *string) *FeedCreate {
	if s != nil {
		fc.SetLink(*s)
	}
	return fc
}

// SetOrder sets the "order" field.
func (fc *FeedCreate) SetOrder(i int) *FeedCreate {
	fc.mutation.SetOrder(i)
	return fc
}

// SetNillableOrder sets the "order" field if the given value is not nil.
func (fc *FeedCreate) SetNillableOrder(i *int) *FeedCreate {
	if i != nil {
		fc.SetOrder(*i)
	}
	return fc
}

// SetIsBookmark sets the "is_bookmark" field.
func (fc *FeedCreate) SetIsBookmark(b bool) *FeedCreate {
	fc.mutation.SetIsBookmark(b)
	return fc
}

// SetNillableIsBookmark sets the "is_bookmark" field if the given value is not nil.
func (fc *FeedCreate) SetNillableIsBookmark(b *bool) *FeedCreate {
	if b != nil {
		fc.SetIsBookmark(*b)
	}
	return fc
}

// SetLastCheckedAt sets the "last_checked_at" field.
func (fc *FeedCreate) SetLastCheckedAt(t time.Time) *FeedCreate {
	fc.mutation.SetLastCheckedAt(t)
	return fc
}

// SetNillableLastCheckedAt sets the "last_checked_at" field if the given value is not nil.
func (fc *FeedCreate) SetNillableLastCheckedAt(t *time.Time) *FeedCreate {
	if t != nil {
		fc.SetLastCheckedAt(*t)
	}
	return fc
}

// SetCreatedAt sets the "created_at" field.
func (fc *FeedCreate) SetCreatedAt(t time.Time) *FeedCreate {
	fc.mutation.SetCreatedAt(t)
	return fc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (fc *FeedCreate) SetNillableCreatedAt(t *time.Time) *FeedCreate {
	if t != nil {
		fc.SetCreatedAt(*t)
	}
	return fc
}

// SetUpdatedAt sets the "updated_at" field.
func (fc *FeedCreate) SetUpdatedAt(t time.Time) *FeedCreate {
	fc.mutation.SetUpdatedAt(t)
	return fc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (fc *FeedCreate) SetNillableUpdatedAt(t *time.Time) *FeedCreate {
	if t != nil {
		fc.SetUpdatedAt(*t)
	}
	return fc
}

// SetID sets the "id" field.
func (fc *FeedCreate) SetID(u uuid.UUID) *FeedCreate {
	fc.mutation.SetID(u)
	return fc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (fc *FeedCreate) SetNillableID(u *uuid.UUID) *FeedCreate {
	if u != nil {
		fc.SetID(*u)
	}
	return fc
}

// AddArticleIDs adds the "articles" edge to the Article entity by IDs.
func (fc *FeedCreate) AddArticleIDs(ids ...uuid.UUID) *FeedCreate {
	fc.mutation.AddArticleIDs(ids...)
	return fc
}

// AddArticles adds the "articles" edges to the Article entity.
func (fc *FeedCreate) AddArticles(a ...*Article) *FeedCreate {
	ids := make([]uuid.UUID, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return fc.AddArticleIDs(ids...)
}

// AddSummaryIDs adds the "summaries" edge to the Summary entity by IDs.
func (fc *FeedCreate) AddSummaryIDs(ids ...uuid.UUID) *FeedCreate {
	fc.mutation.AddSummaryIDs(ids...)
	return fc
}

// AddSummaries adds the "summaries" edges to the Summary entity.
func (fc *FeedCreate) AddSummaries(s ...*Summary) *FeedCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fc.AddSummaryIDs(ids...)
}

// Mutation returns the FeedMutation object of the builder.
func (fc *FeedCreate) Mutation() *FeedMutation {
	return fc.mutation
}

// Save creates the Feed in the database.
func (fc *FeedCreate) Save(ctx context.Context) (*Feed, error) {
	fc.defaults()
	return withHooks(ctx, fc.sqlSave, fc.mutation, fc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (fc *FeedCreate) SaveX(ctx context.Context) *Feed {
	v, err := fc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fc *FeedCreate) Exec(ctx context.Context) error {
	_, err := fc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fc *FeedCreate) ExecX(ctx context.Context) {
	if err := fc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (fc *FeedCreate) defaults() {
	if _, ok := fc.mutation.Order(); !ok {
		v := feed.DefaultOrder
		fc.mutation.SetOrder(v)
	}
	if _, ok := fc.mutation.IsBookmark(); !ok {
		v := feed.DefaultIsBookmark
		fc.mutation.SetIsBookmark(v)
	}
	if _, ok := fc.mutation.CreatedAt(); !ok {
		v := feed.DefaultCreatedAt()
		fc.mutation.SetCreatedAt(v)
	}
	if _, ok := fc.mutation.UpdatedAt(); !ok {
		v := feed.DefaultUpdatedAt()
		fc.mutation.SetUpdatedAt(v)
	}
	if _, ok := fc.mutation.ID(); !ok {
		v := feed.DefaultID()
		fc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fc *FeedCreate) check() error {
	if _, ok := fc.mutation.URL(); !ok {
		return &ValidationError{Name: "url", err: errors.New(`ent: missing required field "Feed.url"`)}
	}
	if v, ok := fc.mutation.URL(); ok {
		if err := feed.URLValidator(v); err != nil {
			return &ValidationError{Name: "url", err: fmt.Errorf(`ent: validator failed for field "Feed.url": %w`, err)}
		}
	}
	if _, ok := fc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`ent: missing required field "Feed.title"`)}
	}
	if v, ok := fc.mutation.Title(); ok {
		if err := feed.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`ent: validator failed for field "Feed.title": %w`, err)}
		}
	}
	if _, ok := fc.mutation.Order(); !ok {
		return &ValidationError{Name: "order", err: errors.New(`ent: missing required field "Feed.order"`)}
	}
	if _, ok := fc.mutation.IsBookmark(); !ok {
		return &ValidationError{Name: "is_bookmark", err: errors.New(`ent: missing required field "Feed.is_bookmark"`)}
	}
	if _, ok := fc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Feed.created_at"`)}
	}
	if _, ok := fc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Feed.updated_at"`)}
	}
	return nil
}

func (fc *FeedCreate) sqlSave(ctx context.Context) (*Feed, error) {
	if err := fc.check(); err != nil {
		return nil, err
	}
	_node, _spec := fc.createSpec()
	if err := sqlgraph.CreateNode(ctx, fc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	fc.mutation.id = &_node.ID
	fc.mutation.done = true
	return _node, nil
}

func (fc *FeedCreate) createSpec() (*Feed, *sqlgraph.CreateSpec) {
	var (
		_node = &Feed{config: fc.config}
		_spec = sqlgraph.NewCreateSpec(feed.Table, sqlgraph.NewFieldSpec(feed.FieldID, field.TypeUUID))
	)
	if id, ok := fc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := fc.mutation.URL(); ok {
		_spec.SetField(feed.FieldURL, field.TypeString, value)
		_node.URL = value
	}
	if value, ok := fc.mutation.Title(); ok {
		_spec.SetField(feed.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := fc.mutation.Description(); ok {
		_spec.SetField(feed.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := fc.mutation.Link(); ok {
		_spec.SetField(feed.FieldLink, field.TypeString, value)
		_node.Link = value
	}
	if value, ok := fc.mutation.Order(); ok {
		_spec.SetField(feed.FieldOrder, field.TypeInt, value)
		_node.Order = value
	}
	if value, ok := fc.mutation.IsBookmark(); ok {
		_spec.SetField(feed.FieldIsBookmark, field.TypeBool, value)
		_node.IsBookmark = value
	}
	if value, ok := fc.mutation.LastCheckedAt(); ok {
		_spec.SetField(feed.FieldLastCheckedAt, field.TypeTime, value)
		_node.LastCheckedAt = value
	}
	if value, ok := fc.mutation.CreatedAt(); ok {
		_spec.SetField(feed.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := fc.mutation.UpdatedAt(); ok {
		_spec.SetField(feed.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := fc.mutation.ArticlesIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.SummariesIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// FeedCreateBulk is the builder for creating many Feed entities in bulk.
type FeedCreateBulk struct {
	config
	err      error
	builders []*FeedCreate
}

// Save creates the Feed entities in the database.
func (fcb *FeedCreateBulk) Save(ctx context.Context) ([]*Feed, error) {
	if fcb.err != nil {
		return nil, fcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(fcb.builders))
	nodes := make([]*Feed, len(fcb.builders))
	mutators := make([]Mutator, len(fcb.builders))
	for i := range fcb.builders {
		func(i int, root context.Context) {
			builder := fcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*FeedMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, fcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, fcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, fcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (fcb *FeedCreateBulk) SaveX(ctx context.Context) []*Feed {
	v, err := fcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fcb *FeedCreateBulk) Exec(ctx context.Context) error {
	_, err := fcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fcb *FeedCreateBulk) ExecX(ctx context.Context) {
	if err := fcb.Exec(ctx); err != nil {
		panic(err)
	}
}
