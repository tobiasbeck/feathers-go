package feathers

import "errors"

func ToMapSlice(result interface{}, errs ...error) ([]map[string]interface{}, error) {
	if len(errs) > 0 {
		for _, err := range errs {
			if err != nil {
				return nil, err
			}
		}
	}
	if list, ok := result.([]map[string]interface{}); ok {
		return list, nil
	}
	return nil, errors.New("result is not a map slice")
}
