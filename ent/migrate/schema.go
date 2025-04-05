// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// ArticlesColumns holds the columns for the "articles" table.
	ArticlesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "title", Type: field.TypeString},
		{Name: "url", Type: field.TypeString, Unique: true},
		{Name: "description", Type: field.TypeString, Nullable: true},
		{Name: "content", Type: field.TypeString, Nullable: true},
		{Name: "published_at", Type: field.TypeTime, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "feed_articles", Type: field.TypeUUID},
	}
	// ArticlesTable holds the schema information for the "articles" table.
	ArticlesTable = &schema.Table{
		Name:       "articles",
		Columns:    ArticlesColumns,
		PrimaryKey: []*schema.Column{ArticlesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "articles_feeds_articles",
				Columns:    []*schema.Column{ArticlesColumns[7]},
				RefColumns: []*schema.Column{FeedsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// FeedsColumns holds the columns for the "feeds" table.
	FeedsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "url", Type: field.TypeString, Unique: true},
		{Name: "title", Type: field.TypeString},
		{Name: "description", Type: field.TypeString, Nullable: true},
		{Name: "link", Type: field.TypeString, Nullable: true},
		{Name: "order", Type: field.TypeInt, Default: 1},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "created_at", Type: field.TypeTime},
	}
	// FeedsTable holds the schema information for the "feeds" table.
	FeedsTable = &schema.Table{
		Name:       "feeds",
		Columns:    FeedsColumns,
		PrimaryKey: []*schema.Column{FeedsColumns[0]},
	}
	// SummariesColumns holds the columns for the "summaries" table.
	SummariesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "url", Type: field.TypeString, Unique: true},
		{Name: "title", Type: field.TypeString, Nullable: true},
		{Name: "summary", Type: field.TypeString, Nullable: true},
		{Name: "audio_data", Type: field.TypeBytes, Nullable: true},
		{Name: "readed", Type: field.TypeBool, Default: false},
		{Name: "listend", Type: field.TypeBool, Default: false},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "article_summary", Type: field.TypeUUID, Unique: true, Nullable: true},
		{Name: "feed_summaries", Type: field.TypeUUID},
	}
	// SummariesTable holds the schema information for the "summaries" table.
	SummariesTable = &schema.Table{
		Name:       "summaries",
		Columns:    SummariesColumns,
		PrimaryKey: []*schema.Column{SummariesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "summaries_articles_summary",
				Columns:    []*schema.Column{SummariesColumns[8]},
				RefColumns: []*schema.Column{ArticlesColumns[0]},
				OnDelete:   schema.SetNull,
			},
			{
				Symbol:     "summaries_feeds_summaries",
				Columns:    []*schema.Column{SummariesColumns[9]},
				RefColumns: []*schema.Column{FeedsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		ArticlesTable,
		FeedsTable,
		SummariesTable,
	}
)

func init() {
	ArticlesTable.ForeignKeys[0].RefTable = FeedsTable
	SummariesTable.ForeignKeys[0].RefTable = ArticlesTable
	SummariesTable.ForeignKeys[1].RefTable = FeedsTable
}
