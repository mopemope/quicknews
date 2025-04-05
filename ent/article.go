// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent/article"
	"github.com/mopemope/quicknews/ent/feed"
	"github.com/mopemope/quicknews/ent/summary"
)

// Article is the model entity for the Article schema.
type Article struct {
	config `json:"-"`
	// ID of the ent.
	// Unique identifier
	ID uuid.UUID `json:"id,omitempty"`
	// Title of the article
	Title string `json:"title,omitempty"`
	// URL of the article
	URL string `json:"url,omitempty"`
	// Description or summary of the article
	Description string `json:"description,omitempty"`
	// Full content of the article
	Content string `json:"content,omitempty"`
	// Time the article was published
	PublishedAt time.Time `json:"published_at,omitempty"`
	// Time the article was added to the database
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the ArticleQuery when eager-loading is set.
	Edges         ArticleEdges `json:"edges"`
	feed_articles *uuid.UUID
	selectValues  sql.SelectValues
}

// ArticleEdges holds the relations/edges for other nodes in the graph.
type ArticleEdges struct {
	// Feed holds the value of the feed edge.
	Feed *Feed `json:"feed,omitempty"`
	// Summary holds the value of the summary edge.
	Summary *Summary `json:"summary,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// FeedOrErr returns the Feed value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ArticleEdges) FeedOrErr() (*Feed, error) {
	if e.Feed != nil {
		return e.Feed, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: feed.Label}
	}
	return nil, &NotLoadedError{edge: "feed"}
}

// SummaryOrErr returns the Summary value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ArticleEdges) SummaryOrErr() (*Summary, error) {
	if e.Summary != nil {
		return e.Summary, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: summary.Label}
	}
	return nil, &NotLoadedError{edge: "summary"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Article) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case article.FieldTitle, article.FieldURL, article.FieldDescription, article.FieldContent:
			values[i] = new(sql.NullString)
		case article.FieldPublishedAt, article.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case article.FieldID:
			values[i] = new(uuid.UUID)
		case article.ForeignKeys[0]: // feed_articles
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Article fields.
func (a *Article) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case article.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				a.ID = *value
			}
		case article.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				a.Title = value.String
			}
		case article.FieldURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field url", values[i])
			} else if value.Valid {
				a.URL = value.String
			}
		case article.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				a.Description = value.String
			}
		case article.FieldContent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field content", values[i])
			} else if value.Valid {
				a.Content = value.String
			}
		case article.FieldPublishedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field published_at", values[i])
			} else if value.Valid {
				a.PublishedAt = value.Time
			}
		case article.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				a.CreatedAt = value.Time
			}
		case article.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field feed_articles", values[i])
			} else if value.Valid {
				a.feed_articles = new(uuid.UUID)
				*a.feed_articles = *value.S.(*uuid.UUID)
			}
		default:
			a.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Article.
// This includes values selected through modifiers, order, etc.
func (a *Article) Value(name string) (ent.Value, error) {
	return a.selectValues.Get(name)
}

// QueryFeed queries the "feed" edge of the Article entity.
func (a *Article) QueryFeed() *FeedQuery {
	return NewArticleClient(a.config).QueryFeed(a)
}

// QuerySummary queries the "summary" edge of the Article entity.
func (a *Article) QuerySummary() *SummaryQuery {
	return NewArticleClient(a.config).QuerySummary(a)
}

// Update returns a builder for updating this Article.
// Note that you need to call Article.Unwrap() before calling this method if this Article
// was returned from a transaction, and the transaction was committed or rolled back.
func (a *Article) Update() *ArticleUpdateOne {
	return NewArticleClient(a.config).UpdateOne(a)
}

// Unwrap unwraps the Article entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (a *Article) Unwrap() *Article {
	_tx, ok := a.config.driver.(*txDriver)
	if !ok {
		panic("ent: Article is not a transactional entity")
	}
	a.config.driver = _tx.drv
	return a
}

// String implements the fmt.Stringer.
func (a *Article) String() string {
	var builder strings.Builder
	builder.WriteString("Article(")
	builder.WriteString(fmt.Sprintf("id=%v, ", a.ID))
	builder.WriteString("title=")
	builder.WriteString(a.Title)
	builder.WriteString(", ")
	builder.WriteString("url=")
	builder.WriteString(a.URL)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(a.Description)
	builder.WriteString(", ")
	builder.WriteString("content=")
	builder.WriteString(a.Content)
	builder.WriteString(", ")
	builder.WriteString("published_at=")
	builder.WriteString(a.PublishedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(a.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Articles is a parsable slice of Article.
type Articles []*Article
