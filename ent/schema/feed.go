package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Feed holds the schema definition for the Feed entity.
type Feed struct {
	ent.Schema
}

// Fields of the Feed.
func (Feed) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("Unique identifier"),
		field.String("url").
			Unique().   // URLは一意である必要があります
			NotEmpty(). // URLは空であってはなりません
			Comment("URL of the RSS feed"),
		field.String("title").
			NotEmpty(). // タイトルは空であってはなりません
			Comment("Title of the RSS feed"),
		field.String("description").
			Optional(). // 説明は任意です
			Comment("Description of the RSS feed"),
		field.String("link").
			Optional(). // サイトへのリンクは任意です
			Comment("Link to the website"),
		field.Int("order").
			Default(1). // デフォルト値を1に設定
			Comment("Order of the feed"),
		field.Bool("is_bookmark").
			Default(false).
			Comment("Bookmark feed flag"),
		field.Time("updated_at").
			Default(time.Now).       // デフォルトで現在時刻を設定
			UpdateDefault(time.Now). // 更新時に現在時刻を設定
			Comment("Last updated time from the feed"),
		field.Time("created_at").
			Default(time.Now). // デフォルトで現在時刻を設定
			Immutable().       // 作成後は変更不可
			Comment("Time the feed was added"),
	}
}

// Edges of the Feed.
func (Feed) Edges() []ent.Edge {
	return []ent.Edge{
		// Feed から Article へのリレーション (一対多)
		edge.To("articles", Article.Type), // Article スキーマを参照
		edge.To("summaries", Summary.Type),
	}
}
