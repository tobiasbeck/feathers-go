package service_test

type TestModel struct {
	Text string `validate:"required" mapstructure:"text" bson:"text"`
}

func NewTestModel() interface{} {
	return &TestModel{}
}
