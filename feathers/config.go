package feathers

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func loadConfig(path string) (map[string]interface{}, error) {
	configs := make(map[string]interface{})
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return configs, nil
}
