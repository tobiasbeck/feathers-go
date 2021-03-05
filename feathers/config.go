package feathers

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/imdario/mergo"
	"github.com/tobiasbeck/feathers-go/feathers/yaml"
)

// LoadConfigFile loads a yaml config file
func LoadConfigFile(path string) (map[string]interface{}, error) {
	var configs interface{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return configs.(map[string]interface{}), nil
}

func loadConfig(configPath string) (map[string]interface{}, error) {
	defaultConfig, err := LoadConfigFile(path.Join(configPath, "default.yaml"))
	if err != nil {
		return nil, err
	}
	if env, ok := os.LookupEnv("APP_ENV"); ok {
		envConfig, err := LoadConfigFile(path.Join(configPath, env+".yaml"))
		if err != nil {
			return defaultConfig, nil
		}
		err = mergo.Merge(&defaultConfig, envConfig, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}
	return defaultConfig, nil
}
