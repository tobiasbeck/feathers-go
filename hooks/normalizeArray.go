package hooks

import (
	"reflect"
)

func NormalizeSlice(data interface{}) (interface{}, bool) {
	r := reflect.ValueOf(data)
	if r.Kind() == reflect.Slice {
		return data, false
	} else {
		return []map[string]interface{}{data.(map[string]interface{})}, true
	}
}

func UnormalizeSlice(data interface{}, normalized bool) interface{} {
	r := reflect.ValueOf(data)
	if normalized && r.Len() > 0 {
		return r.Index(0).Interface()
	}

	if !normalized {
		return data
	}

	return nil
}
