package feathers_auth

type Model struct {
	Strategy string                 `mapstructure:"strategy" validate:"required"`
	Params   map[string]interface{} `mapstructure:",remain"`
}

func NewModel() interface{} {
	return &Model{}
}
