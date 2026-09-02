package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

const (
	FileObjectStatusPending  = "pending"
	FileObjectStatusUploaded = "uploaded"
	FileObjectStatusFailed   = "failed"
	FileObjectStatusExpired  = "expired"
)

type FileObject struct {
	ent.Schema
}

func (FileObject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "file_objects"},
	}
}

func (FileObject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (FileObject) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("owner_user_id"),
		field.Int64("api_key_id").
			Optional().
			Nillable(),
		field.String("purpose").
			MaxLen(64).
			Default("vision_input"),
		field.String("storage_provider").
			MaxLen(32).
			Default("s3"),
		field.String("bucket").
			MaxLen(255).
			NotEmpty(),
		field.String("object_key").
			MaxLen(1024).
			NotEmpty(),
		field.String("original_filename").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("mime_type").
			MaxLen(255).
			NotEmpty(),
		field.Int64("size_bytes").
			Default(0),
		field.String("sha256").
			MaxLen(64).
			Optional().
			Nillable(),
		field.Enum("status").
			Values(
				FileObjectStatusPending,
				FileObjectStatusUploaded,
				FileObjectStatusFailed,
				FileObjectStatusExpired,
			).
			Default(FileObjectStatusPending),
		field.JSON("metadata", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("uploaded_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (FileObject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner_user", User.Type).
			Ref("file_objects").
			Field("owner_user_id").
			Required().
			Unique(),
		edge.From("api_key", APIKey.Type).
			Ref("file_objects").
			Field("api_key_id").
			Unique(),
	}
}

func (FileObject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_user_id"),
		index.Fields("api_key_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("bucket", "object_key").Unique(),
		index.Fields("deleted_at"),
	}
}
