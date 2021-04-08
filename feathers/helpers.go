package feathers

import (
	"errors"
	"fmt"
)

// Contains checks if a string slice contains another string
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

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

func ToMap(result interface{}, errs ...error) (map[string]interface{}, error) {
	if result == nil {
		return nil, nil
	}
	if element, ok := result.(map[string]interface{}); ok {
		return element, nil
	}
	return nil, errors.New(fmt.Sprintf("result is not a map (%T)", result))
}
