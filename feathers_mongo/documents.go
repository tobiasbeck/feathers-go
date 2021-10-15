package feathers_mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IdDocument interface {
	GenerateID()
	IDIsZero() bool
}

type Document struct {
	ID primitive.ObjectID `bson:"_id" mapstructure:"_id,omitempty" json:"_id" ts_type:"string"`
}

func NewDocument() *Document {
	return &Document{
		ID: primitive.NewObjectID(),
	}
}

func (d *Document) GenerateID() {
	d.ID = primitive.NewObjectID()
}

func (d *Document) IDIsZero() bool {
	return d.ID.IsZero()
}

type Timestampable interface {
	SetCreatedAt()
	SetUpdatedAt()
}

type TimestampDoc struct {
	CreatedAt primitive.DateTime `bson:"createdAt" mapstructure:"createdAt" json:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" mapstructure:"updatedAt" json:"updatedAt"`
}

func (td *TimestampDoc) SetCreatedAt() {
	td.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
}

func (td *TimestampDoc) SetUpdatedAt() {
	td.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
}
