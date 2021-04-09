package feathers_mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Document struct {
	ID primitive.ObjectID `bson:"_id" mapstructure:"_id"`
}

type Timestampable interface {
	SetCreatedAt()
	SetUpdatedAt()
}

type TimestampDoc struct {
	CreatedAt time.Time `bson:"createdAt" mapstructure:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" mapstructure:"updatedAt"`
}

func (td *TimestampDoc) SetCreatedAt() {
	td.CreatedAt = time.Now()
}

func (td *TimestampDoc) SetUpdatedAt() {
	td.UpdatedAt = time.Now()
}