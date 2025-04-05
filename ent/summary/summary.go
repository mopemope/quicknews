// Code generated by ent, DO NOT EDIT.

package summary

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the summary type in the database.
	Label = "summary"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldURL holds the string denoting the url field in the database.
	FieldURL = "url"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldSummary holds the string denoting the summary field in the database.
	FieldSummary = "summary"
	// FieldAudioData holds the string denoting the audio_data field in the database.
	FieldAudioData = "audio_data"
	// FieldReaded holds the string denoting the readed field in the database.
	FieldReaded = "readed"
	// FieldListend holds the string denoting the listend field in the database.
	FieldListend = "listend"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgeArticle holds the string denoting the article edge name in mutations.
	EdgeArticle = "article"
	// EdgeFeed holds the string denoting the feed edge name in mutations.
	EdgeFeed = "feed"
	// Table holds the table name of the summary in the database.
	Table = "summaries"
	// ArticleTable is the table that holds the article relation/edge.
	ArticleTable = "summaries"
	// ArticleInverseTable is the table name for the Article entity.
	// It exists in this package in order to avoid circular dependency with the "article" package.
	ArticleInverseTable = "articles"
	// ArticleColumn is the table column denoting the article relation/edge.
	ArticleColumn = "article_summary"
	// FeedTable is the table that holds the feed relation/edge.
	FeedTable = "summaries"
	// FeedInverseTable is the table name for the Feed entity.
	// It exists in this package in order to avoid circular dependency with the "feed" package.
	FeedInverseTable = "feeds"
	// FeedColumn is the table column denoting the feed relation/edge.
	FeedColumn = "feed_summaries"
)

// Columns holds all SQL columns for summary fields.
var Columns = []string{
	FieldID,
	FieldURL,
	FieldTitle,
	FieldSummary,
	FieldAudioData,
	FieldReaded,
	FieldListend,
	FieldCreatedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "summaries"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"article_summary",
	"feed_summaries",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// URLValidator is a validator for the "url" field. It is called by the builders before save.
	URLValidator func(string) error
	// DefaultReaded holds the default value on creation for the "readed" field.
	DefaultReaded bool
	// DefaultListend holds the default value on creation for the "listend" field.
	DefaultListend bool
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the Summary queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByURL orders the results by the url field.
func ByURL(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldURL, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// BySummary orders the results by the summary field.
func BySummary(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSummary, opts...).ToFunc()
}

// ByReaded orders the results by the readed field.
func ByReaded(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldReaded, opts...).ToFunc()
}

// ByListend orders the results by the listend field.
func ByListend(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldListend, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByArticleField orders the results by article field.
func ByArticleField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newArticleStep(), sql.OrderByField(field, opts...))
	}
}

// ByFeedField orders the results by feed field.
func ByFeedField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFeedStep(), sql.OrderByField(field, opts...))
	}
}
func newArticleStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ArticleInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, ArticleTable, ArticleColumn),
	)
}
func newFeedStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FeedInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, FeedTable, FeedColumn),
	)
}
