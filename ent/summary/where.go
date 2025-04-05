// Code generated by ent, DO NOT EDIT.

package summary

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldID, id))
}

// URL applies equality check predicate on the "url" field. It's identical to URLEQ.
func URL(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldURL, v))
}

// Title applies equality check predicate on the "title" field. It's identical to TitleEQ.
func Title(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldTitle, v))
}

// Summary applies equality check predicate on the "summary" field. It's identical to SummaryEQ.
func Summary(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldSummary, v))
}

// AudioData applies equality check predicate on the "audio_data" field. It's identical to AudioDataEQ.
func AudioData(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldAudioData, v))
}

// Readed applies equality check predicate on the "readed" field. It's identical to ReadedEQ.
func Readed(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldReaded, v))
}

// Listend applies equality check predicate on the "listend" field. It's identical to ListendEQ.
func Listend(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldListend, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldCreatedAt, v))
}

// URLEQ applies the EQ predicate on the "url" field.
func URLEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldURL, v))
}

// URLNEQ applies the NEQ predicate on the "url" field.
func URLNEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldURL, v))
}

// URLIn applies the In predicate on the "url" field.
func URLIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldURL, vs...))
}

// URLNotIn applies the NotIn predicate on the "url" field.
func URLNotIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldURL, vs...))
}

// URLGT applies the GT predicate on the "url" field.
func URLGT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldURL, v))
}

// URLGTE applies the GTE predicate on the "url" field.
func URLGTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldURL, v))
}

// URLLT applies the LT predicate on the "url" field.
func URLLT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldURL, v))
}

// URLLTE applies the LTE predicate on the "url" field.
func URLLTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldURL, v))
}

// URLContains applies the Contains predicate on the "url" field.
func URLContains(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContains(FieldURL, v))
}

// URLHasPrefix applies the HasPrefix predicate on the "url" field.
func URLHasPrefix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasPrefix(FieldURL, v))
}

// URLHasSuffix applies the HasSuffix predicate on the "url" field.
func URLHasSuffix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasSuffix(FieldURL, v))
}

// URLEqualFold applies the EqualFold predicate on the "url" field.
func URLEqualFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEqualFold(FieldURL, v))
}

// URLContainsFold applies the ContainsFold predicate on the "url" field.
func URLContainsFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContainsFold(FieldURL, v))
}

// TitleEQ applies the EQ predicate on the "title" field.
func TitleEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldTitle, v))
}

// TitleNEQ applies the NEQ predicate on the "title" field.
func TitleNEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldTitle, v))
}

// TitleIn applies the In predicate on the "title" field.
func TitleIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldTitle, vs...))
}

// TitleNotIn applies the NotIn predicate on the "title" field.
func TitleNotIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldTitle, vs...))
}

// TitleGT applies the GT predicate on the "title" field.
func TitleGT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldTitle, v))
}

// TitleGTE applies the GTE predicate on the "title" field.
func TitleGTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldTitle, v))
}

// TitleLT applies the LT predicate on the "title" field.
func TitleLT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldTitle, v))
}

// TitleLTE applies the LTE predicate on the "title" field.
func TitleLTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldTitle, v))
}

// TitleContains applies the Contains predicate on the "title" field.
func TitleContains(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContains(FieldTitle, v))
}

// TitleHasPrefix applies the HasPrefix predicate on the "title" field.
func TitleHasPrefix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasPrefix(FieldTitle, v))
}

// TitleHasSuffix applies the HasSuffix predicate on the "title" field.
func TitleHasSuffix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasSuffix(FieldTitle, v))
}

// TitleIsNil applies the IsNil predicate on the "title" field.
func TitleIsNil() predicate.Summary {
	return predicate.Summary(sql.FieldIsNull(FieldTitle))
}

// TitleNotNil applies the NotNil predicate on the "title" field.
func TitleNotNil() predicate.Summary {
	return predicate.Summary(sql.FieldNotNull(FieldTitle))
}

// TitleEqualFold applies the EqualFold predicate on the "title" field.
func TitleEqualFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEqualFold(FieldTitle, v))
}

// TitleContainsFold applies the ContainsFold predicate on the "title" field.
func TitleContainsFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContainsFold(FieldTitle, v))
}

// SummaryEQ applies the EQ predicate on the "summary" field.
func SummaryEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldSummary, v))
}

// SummaryNEQ applies the NEQ predicate on the "summary" field.
func SummaryNEQ(v string) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldSummary, v))
}

// SummaryIn applies the In predicate on the "summary" field.
func SummaryIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldSummary, vs...))
}

// SummaryNotIn applies the NotIn predicate on the "summary" field.
func SummaryNotIn(vs ...string) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldSummary, vs...))
}

// SummaryGT applies the GT predicate on the "summary" field.
func SummaryGT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldSummary, v))
}

// SummaryGTE applies the GTE predicate on the "summary" field.
func SummaryGTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldSummary, v))
}

// SummaryLT applies the LT predicate on the "summary" field.
func SummaryLT(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldSummary, v))
}

// SummaryLTE applies the LTE predicate on the "summary" field.
func SummaryLTE(v string) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldSummary, v))
}

// SummaryContains applies the Contains predicate on the "summary" field.
func SummaryContains(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContains(FieldSummary, v))
}

// SummaryHasPrefix applies the HasPrefix predicate on the "summary" field.
func SummaryHasPrefix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasPrefix(FieldSummary, v))
}

// SummaryHasSuffix applies the HasSuffix predicate on the "summary" field.
func SummaryHasSuffix(v string) predicate.Summary {
	return predicate.Summary(sql.FieldHasSuffix(FieldSummary, v))
}

// SummaryIsNil applies the IsNil predicate on the "summary" field.
func SummaryIsNil() predicate.Summary {
	return predicate.Summary(sql.FieldIsNull(FieldSummary))
}

// SummaryNotNil applies the NotNil predicate on the "summary" field.
func SummaryNotNil() predicate.Summary {
	return predicate.Summary(sql.FieldNotNull(FieldSummary))
}

// SummaryEqualFold applies the EqualFold predicate on the "summary" field.
func SummaryEqualFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldEqualFold(FieldSummary, v))
}

// SummaryContainsFold applies the ContainsFold predicate on the "summary" field.
func SummaryContainsFold(v string) predicate.Summary {
	return predicate.Summary(sql.FieldContainsFold(FieldSummary, v))
}

// AudioDataEQ applies the EQ predicate on the "audio_data" field.
func AudioDataEQ(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldAudioData, v))
}

// AudioDataNEQ applies the NEQ predicate on the "audio_data" field.
func AudioDataNEQ(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldAudioData, v))
}

// AudioDataIn applies the In predicate on the "audio_data" field.
func AudioDataIn(vs ...[]byte) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldAudioData, vs...))
}

// AudioDataNotIn applies the NotIn predicate on the "audio_data" field.
func AudioDataNotIn(vs ...[]byte) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldAudioData, vs...))
}

// AudioDataGT applies the GT predicate on the "audio_data" field.
func AudioDataGT(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldAudioData, v))
}

// AudioDataGTE applies the GTE predicate on the "audio_data" field.
func AudioDataGTE(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldAudioData, v))
}

// AudioDataLT applies the LT predicate on the "audio_data" field.
func AudioDataLT(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldAudioData, v))
}

// AudioDataLTE applies the LTE predicate on the "audio_data" field.
func AudioDataLTE(v []byte) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldAudioData, v))
}

// AudioDataIsNil applies the IsNil predicate on the "audio_data" field.
func AudioDataIsNil() predicate.Summary {
	return predicate.Summary(sql.FieldIsNull(FieldAudioData))
}

// AudioDataNotNil applies the NotNil predicate on the "audio_data" field.
func AudioDataNotNil() predicate.Summary {
	return predicate.Summary(sql.FieldNotNull(FieldAudioData))
}

// ReadedEQ applies the EQ predicate on the "readed" field.
func ReadedEQ(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldReaded, v))
}

// ReadedNEQ applies the NEQ predicate on the "readed" field.
func ReadedNEQ(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldReaded, v))
}

// ListendEQ applies the EQ predicate on the "listend" field.
func ListendEQ(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldListend, v))
}

// ListendNEQ applies the NEQ predicate on the "listend" field.
func ListendNEQ(v bool) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldListend, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.Summary {
	return predicate.Summary(sql.FieldLTE(FieldCreatedAt, v))
}

// HasArticle applies the HasEdge predicate on the "article" edge.
func HasArticle() predicate.Summary {
	return predicate.Summary(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, ArticleTable, ArticleColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasArticleWith applies the HasEdge predicate on the "article" edge with a given conditions (other predicates).
func HasArticleWith(preds ...predicate.Article) predicate.Summary {
	return predicate.Summary(func(s *sql.Selector) {
		step := newArticleStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Summary) predicate.Summary {
	return predicate.Summary(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Summary) predicate.Summary {
	return predicate.Summary(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Summary) predicate.Summary {
	return predicate.Summary(sql.NotPredicates(p))
}
