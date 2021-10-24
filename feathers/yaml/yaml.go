// Copyright (c) 2015-2016 Michael Persson
// Copyright (c) 2012–2015 Elasticsearch <http://www.elastic.co>
//
// Originally distributed as part of "beats" repository (https://github.com/elastic/beats).
// Modified specifically for "iodatafmt" package.
//
// Distributed underneath "Apache License, Version 2.0" which is compatible with the LICENSE for this package.

package yaml

import (
	// Base packages.
	"fmt"

	// Third party packages.
	"gopkg.in/yaml.v2"
)

// Unmarshal YAML to map[string]interface{} instead of map[interface{}]interface{} and []interface{} to []map[string]interface{}
func Unmarshal(in []byte, out interface{}) error {
	var res interface{}

	if err := yaml.Unmarshal(in, &res); err != nil {
		return err
	}
	switch v := res.(type) {
	case map[interface{}]interface{}:
		*out.(*interface{}) = cleanupMapValue(v)
	case []interface{}:
		*out.(*interface{}) = cleanupInterfaceArray(v)
	default:
		return fmt.Errorf("could not Unmarshall config of type %T", res)
	}
	return nil
}

// Marshal YAML wrapper function.
func Marshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		switch vt := v.(type) {
		case map[interface{}]interface{}:
			res[i] = cleanupMapValue(vt)
		case []interface{}:
			res[i] = cleanupInterfaceArray(vt)
		default:
			res[i] = v
		}
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return v
	}
}
