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
	"github.com/mopemope/quicknews/ent/summary"
)

// SummaryCreate is the builder for creating a Summary entity.
type SummaryCreate struct {
	config
	mutation *SummaryMutation
	hooks    []Hook
}

// SetURL sets the "url" field.
func (sc *SummaryCreate) SetURL(s string) *SummaryCreate {
	sc.mutation.SetURL(s)
	return sc
}

// SetTitle sets the "title" field.
func (sc *SummaryCreate) SetTitle(s string) *SummaryCreate {
	sc.mutation.SetTitle(s)
	return sc
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableTitle(s *string) *SummaryCreate {
	if s != nil {
		sc.SetTitle(*s)
	}
	return sc
}

// SetSummary sets the "summary" field.
func (sc *SummaryCreate) SetSummary(s string) *SummaryCreate {
	sc.mutation.SetSummary(s)
	return sc
}

// SetNillableSummary sets the "summary" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableSummary(s *string) *SummaryCreate {
	if s != nil {
		sc.SetSummary(*s)
	}
	return sc
}

// SetAudioData sets the "audio_data" field.
func (sc *SummaryCreate) SetAudioData(b []byte) *SummaryCreate {
	sc.mutation.SetAudioData(b)
	return sc
}

// SetReaded sets the "readed" field.
func (sc *SummaryCreate) SetReaded(b bool) *SummaryCreate {
	sc.mutation.SetReaded(b)
	return sc
}

// SetNillableReaded sets the "readed" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableReaded(b *bool) *SummaryCreate {
	if b != nil {
		sc.SetReaded(*b)
	}
	return sc
}

// SetListend sets the "listend" field.
func (sc *SummaryCreate) SetListend(b bool) *SummaryCreate {
	sc.mutation.SetListend(b)
	return sc
}

// SetNillableListend sets the "listend" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableListend(b *bool) *SummaryCreate {
	if b != nil {
		sc.SetListend(*b)
	}
	return sc
}

// SetCreatedAt sets the "created_at" field.
func (sc *SummaryCreate) SetCreatedAt(t time.Time) *SummaryCreate {
	sc.mutation.SetCreatedAt(t)
	return sc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableCreatedAt(t *time.Time) *SummaryCreate {
	if t != nil {
		sc.SetCreatedAt(*t)
	}
	return sc
}

// SetID sets the "id" field.
func (sc *SummaryCreate) SetID(u uuid.UUID) *SummaryCreate {
	sc.mutation.SetID(u)
	return sc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (sc *SummaryCreate) SetNillableID(u *uuid.UUID) *SummaryCreate {
	if u != nil {
		sc.SetID(*u)
	}
	return sc
}

// SetArticleID sets the "article" edge to the Article entity by ID.
func (sc *SummaryCreate) SetArticleID(id uuid.UUID) *SummaryCreate {
	sc.mutation.SetArticleID(id)
	return sc
}

// SetNillableArticleID sets the "article" edge to the Article entity by ID if the given value is not nil.
func (sc *SummaryCreate) SetNillableArticleID(id *uuid.UUID) *SummaryCreate {
	if id != nil {
		sc = sc.SetArticleID(*id)
	}
	return sc
}

// SetArticle sets the "article" edge to the Article entity.
func (sc *SummaryCreate) SetArticle(a *Article) *SummaryCreate {
	return sc.SetArticleID(a.ID)
}

// Mutation returns the SummaryMutation object of the builder.
func (sc *SummaryCreate) Mutation() *SummaryMutation {
	return sc.mutation
}

// Save creates the Summary in the database.
func (sc *SummaryCreate) Save(ctx context.Context) (*Summary, error) {
	sc.defaults()
	return withHooks(ctx, sc.sqlSave, sc.mutation, sc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (sc *SummaryCreate) SaveX(ctx context.Context) *Summary {
	v, err := sc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (sc *SummaryCreate) Exec(ctx context.Context) error {
	_, err := sc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sc *SummaryCreate) ExecX(ctx context.Context) {
	if err := sc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (sc *SummaryCreate) defaults() {
	if _, ok := sc.mutation.Readed(); !ok {
		v := summary.DefaultReaded
		sc.mutation.SetReaded(v)
	}
	if _, ok := sc.mutation.Listend(); !ok {
		v := summary.DefaultListend
		sc.mutation.SetListend(v)
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		v := summary.DefaultCreatedAt()
		sc.mutation.SetCreatedAt(v)
	}
	if _, ok := sc.mutation.ID(); !ok {
		v := summary.DefaultID()
		sc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sc *SummaryCreate) check() error {
	if _, ok := sc.mutation.URL(); !ok {
		return &ValidationError{Name: "url", err: errors.New(`ent: missing required field "Summary.url"`)}
	}
	if v, ok := sc.mutation.URL(); ok {
		if err := summary.URLValidator(v); err != nil {
			return &ValidationError{Name: "url", err: fmt.Errorf(`ent: validator failed for field "Summary.url": %w`, err)}
		}
	}
	if _, ok := sc.mutation.Readed(); !ok {
		return &ValidationError{Name: "readed", err: errors.New(`ent: missing required field "Summary.readed"`)}
	}
	if _, ok := sc.mutation.Listend(); !ok {
		return &ValidationError{Name: "listend", err: errors.New(`ent: missing required field "Summary.listend"`)}
	}
	if _, ok := sc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Summary.created_at"`)}
	}
	return nil
}

func (sc *SummaryCreate) sqlSave(ctx context.Context) (*Summary, error) {
	if err := sc.check(); err != nil {
		return nil, err
	}
	_node, _spec := sc.createSpec()
	if err := sqlgraph.CreateNode(ctx, sc.driver, _spec); err != nil {
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
	sc.mutation.id = &_node.ID
	sc.mutation.done = true
	return _node, nil
}

func (sc *SummaryCreate) createSpec() (*Summary, *sqlgraph.CreateSpec) {
	var (
		_node = &Summary{config: sc.config}
		_spec = sqlgraph.NewCreateSpec(summary.Table, sqlgraph.NewFieldSpec(summary.FieldID, field.TypeUUID))
	)
	if id, ok := sc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := sc.mutation.URL(); ok {
		_spec.SetField(summary.FieldURL, field.TypeString, value)
		_node.URL = value
	}
	if value, ok := sc.mutation.Title(); ok {
		_spec.SetField(summary.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := sc.mutation.Summary(); ok {
		_spec.SetField(summary.FieldSummary, field.TypeString, value)
		_node.Summary = value
	}
	if value, ok := sc.mutation.AudioData(); ok {
		_spec.SetField(summary.FieldAudioData, field.TypeBytes, value)
		_node.AudioData = value
	}
	if value, ok := sc.mutation.Readed(); ok {
		_spec.SetField(summary.FieldReaded, field.TypeBool, value)
		_node.Readed = value
	}
	if value, ok := sc.mutation.Listend(); ok {
		_spec.SetField(summary.FieldListend, field.TypeBool, value)
		_node.Listend = value
	}
	if value, ok := sc.mutation.CreatedAt(); ok {
		_spec.SetField(summary.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := sc.mutation.ArticleIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   summary.ArticleTable,
			Columns: []string{summary.ArticleColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(article.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.article_summary = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SummaryCreateBulk is the builder for creating many Summary entities in bulk.
type SummaryCreateBulk struct {
	config
	err      error
	builders []*SummaryCreate
}

// Save creates the Summary entities in the database.
func (scb *SummaryCreateBulk) Save(ctx context.Context) ([]*Summary, error) {
	if scb.err != nil {
		return nil, scb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(scb.builders))
	nodes := make([]*Summary, len(scb.builders))
	mutators := make([]Mutator, len(scb.builders))
	for i := range scb.builders {
		func(i int, root context.Context) {
			builder := scb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SummaryMutation)
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
					_, err = mutators[i+1].Mutate(root, scb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, scb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, scb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (scb *SummaryCreateBulk) SaveX(ctx context.Context) []*Summary {
	v, err := scb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (scb *SummaryCreateBulk) Exec(ctx context.Context) error {
	_, err := scb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scb *SummaryCreateBulk) ExecX(ctx context.Context) {
	if err := scb.Exec(ctx); err != nil {
		panic(err)
	}
}
