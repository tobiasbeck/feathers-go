package hooks

import (
	"reflect"
)

func NormalizeSlice(data interface{}) ([]map[string]interface{}, bool) {
	r := reflect.ValueOf(data)
	if r.Kind() == reflect.Slice {
		return data.([]map[string]interface{}), false
	} else {
		return []map[string]interface{}{data.(map[string]interface{})}, true
	}
}

func UnormalizeSlice(data []map[string]interface{}, normalized bool) interface{} {
	if normalized && len(data) > 0 {
		return data[0]
	}

	if !normalized {
		return data
	}

	return nil
}
