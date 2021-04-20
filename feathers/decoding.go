package feathers

import (
	"fmt"

	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
)

var decodeHooks []mapstructure.DecodeHookFunc = make([]mapstructure.DecodeHookFunc, 0)
var decodeHookFunc mapstructure.DecodeHookFunc

func AddStructDecodeHookFunc(f mapstructure.DecodeHookFunc) {
	decodeHooks = append(decodeHooks, f)
	decodeHookFunc = mapstructure.ComposeDecodeHookFunc(decodeHooks...)
}

func newDecoder(target interface{}) (*mapstructure.Decoder, error) {
	config := &mapstructure.DecoderConfig{
		DecodeHook: decodeHookFunc,
		Result:     target,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		fmt.Printf("DECODER ERROR: %s\n", err)
		return nil, err
	}

	return decoder, nil
}

// MapToStruct maps service data into a struct (passed by pointer)
/*
Example:
````
model := Model{}
err := MapToStruct(data, &model)
````
*/
func MapToStruct(data map[string]interface{}, target interface{}) error {
	decoder, err := newDecoder(target)
	if err != nil {
		return err
	}
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	defaults.SetDefaults(target)
	return nil
}
