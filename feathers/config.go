package feathers

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

func loadConfigFile(path string) (map[string]interface{}, error) {
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

func loadConfig(configPath string) (map[string]interface{}, error) {
	defaultConfig, err := loadConfigFile(path.Join(configPath, "default.yaml"))
	if err != nil {
		return nil, err
	}
	if env, ok := os.LookupEnv("APP_ENV"); ok {
		envConfig, err := loadConfigFile(path.Join(configPath, env+".yaml"))
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
