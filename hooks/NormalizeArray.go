package hooks

func NormalizeSlice(data interface{}) ([]interface{}, bool) {
	switch d := data.(type) {
	case []interface{}:
		return d, false
	default:
		return []interface{}{d}, true
	}
}

func UnormalizeSlice(data []interface{}, normalized bool) interface{} {
	if normalized && len(data) > 0 {
		return data[0]
	}

	if !normalized {
		return data
	}

	return nil
}
