package db

import (
	"time"

	"ariga.io/atlas/sql/postgres"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Mixin definition

// CommonMixin implements the ent.Mixin for sharing
// time fields with package schemas.
type CommonMixin struct {
	// We embed the `mixin.Schema` to avoid
	// implementing the rest of the methods.
	mixin.Schema
}

var TimeNowLocal = func() time.Time {
	return time.Now().In(time.Local)
}

func (CommonMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").SchemaType(map[string]string{
			dialect.Postgres: postgres.TypeBigSerial,
		}).Comment("自增ID"),
		field.Time("created_at").
			Immutable().
			Default(TimeNowLocal).Comment("创建时间"),
		field.Time("updated_at").
			Default(TimeNowLocal).
			UpdateDefault(TimeNowLocal).Comment("更新时间"),
	}
}
