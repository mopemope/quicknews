package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Article holds the schema definition for the Article entity.
type Article struct {
	ent.Schema
}

// Fields of the Article.
func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("Unique identifier"),
		field.String("title").
			NotEmpty().
			Comment("Title of the article"),
		field.String("url").
			Unique(). // URLは一意である必要があります
			NotEmpty().
			Comment("URL of the article"),
		field.String("description").
			Optional().
			Comment("Description or summary of the article"),
		field.String("content").
			Optional().
			Comment("Full content of the article"),
		field.Time("published_at").
			Optional(). // 公開日時はフィードに含まれていない場合がある
			Comment("Time the article was published"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Time the article was added to the database"),
	}
}

// Edges of the Article.
func (Article) Edges() []ent.Edge {
	return []ent.Edge{
		// Article から Feed へのリレーション (多対一)
		// Article は必ず一つの Feed に属する
		edge.From("feed", Feed.Type).
			Ref("articles"). // Feed スキーマの "articles" エッジと対応
			Unique().        // Article は一つの Feed にしか属さない
			Required(),      // Feed は必須
		edge.To("summary", Summary.Type).Unique(),
	}
}
