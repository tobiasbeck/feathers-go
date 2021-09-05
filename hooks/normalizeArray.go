package hooks

import (
	"reflect"
)

func NormalizeSlice(data interface{}) (interface{}, bool) {
	r := reflect.ValueOf(data)
	// fmt.Printf("R: %#v\n", r)
	// fmt.Printf("DATA: %#v\n", data)
	// if !r.IsValid() {
	// 	return map[string]interface{}{}, true
	// }
	if r.Kind() == reflect.Slice {
		return data, false
	} else if !r.IsNil() {
		return []map[string]interface{}{data.(map[string]interface{})}, true
	}

	return map[string]interface{}{}, true
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
