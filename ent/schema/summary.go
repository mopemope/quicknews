package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Summary struct {
	ent.Schema
}

func (Summary) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("Unique identifier"),
		field.String("url").
			Unique().
			NotEmpty().
			Comment("URL"),
		field.String("title").
			Optional().
			Comment("Summary title"),
		field.String("summary").
			Optional().
			Comment("Summary text"),
		field.Bool("readed").
			Default(false).
			Comment("Read status"),
		field.Bool("listend").
			Default(false).
			Comment("Listened status"),

		field.Time("created_at").
			Default(time.Now). // デフォルトで現在時刻を設定
			Immutable().       // 作成後は変更不可
			Comment("Time the feed was added"),
	}
}

func (Summary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("article", Article.Type).
			Ref("summary").
			Unique(),
		edge.From("feed", Feed.Type).
			Ref("summaries").
			Unique().
			Required(),
	}
}
