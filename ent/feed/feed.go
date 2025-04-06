// Code generated by ent, DO NOT EDIT.

package feed

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the feed type in the database.
	Label = "feed"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldURL holds the string denoting the url field in the database.
	FieldURL = "url"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldLink holds the string denoting the link field in the database.
	FieldLink = "link"
	// FieldOrder holds the string denoting the order field in the database.
	FieldOrder = "order"
	// FieldIsBookmark holds the string denoting the is_bookmark field in the database.
	FieldIsBookmark = "is_bookmark"
	// FieldLastCheckedAt holds the string denoting the last_checked_at field in the database.
	FieldLastCheckedAt = "last_checked_at"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeArticles holds the string denoting the articles edge name in mutations.
	EdgeArticles = "articles"
	// EdgeSummaries holds the string denoting the summaries edge name in mutations.
	EdgeSummaries = "summaries"
	// Table holds the table name of the feed in the database.
	Table = "feeds"
	// ArticlesTable is the table that holds the articles relation/edge.
	ArticlesTable = "articles"
	// ArticlesInverseTable is the table name for the Article entity.
	// It exists in this package in order to avoid circular dependency with the "article" package.
	ArticlesInverseTable = "articles"
	// ArticlesColumn is the table column denoting the articles relation/edge.
	ArticlesColumn = "feed_articles"
	// SummariesTable is the table that holds the summaries relation/edge.
	SummariesTable = "summaries"
	// SummariesInverseTable is the table name for the Summary entity.
	// It exists in this package in order to avoid circular dependency with the "summary" package.
	SummariesInverseTable = "summaries"
	// SummariesColumn is the table column denoting the summaries relation/edge.
	SummariesColumn = "feed_summaries"
)

// Columns holds all SQL columns for feed fields.
var Columns = []string{
	FieldID,
	FieldURL,
	FieldTitle,
	FieldDescription,
	FieldLink,
	FieldOrder,
	FieldIsBookmark,
	FieldLastCheckedAt,
	FieldCreatedAt,
	FieldUpdatedAt,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// URLValidator is a validator for the "url" field. It is called by the builders before save.
	URLValidator func(string) error
	// TitleValidator is a validator for the "title" field. It is called by the builders before save.
	TitleValidator func(string) error
	// DefaultOrder holds the default value on creation for the "order" field.
	DefaultOrder int
	// DefaultIsBookmark holds the default value on creation for the "is_bookmark" field.
	DefaultIsBookmark bool
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the Feed queries.
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

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByLink orders the results by the link field.
func ByLink(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLink, opts...).ToFunc()
}

// ByOrder orders the results by the order field.
func ByOrder(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOrder, opts...).ToFunc()
}

// ByIsBookmark orders the results by the is_bookmark field.
func ByIsBookmark(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsBookmark, opts...).ToFunc()
}

// ByLastCheckedAt orders the results by the last_checked_at field.
func ByLastCheckedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLastCheckedAt, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByArticlesCount orders the results by articles count.
func ByArticlesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newArticlesStep(), opts...)
	}
}

// ByArticles orders the results by articles terms.
func ByArticles(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newArticlesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySummariesCount orders the results by summaries count.
func BySummariesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSummariesStep(), opts...)
	}
}

// BySummaries orders the results by summaries terms.
func BySummaries(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSummariesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newArticlesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ArticlesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ArticlesTable, ArticlesColumn),
	)
}
func newSummariesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SummariesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, SummariesTable, SummariesColumn),
	)
}
