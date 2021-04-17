package feathers

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

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
			configWithEnv := parseEnvVariables(defaultConfig)
			return configWithEnv, nil
		}
		err = mergo.Merge(&defaultConfig, envConfig, mergo.WithOverride)
		if err != nil {
			return nil, err
		}
		configWithEnv := parseEnvVariables(defaultConfig)
		return configWithEnv, nil
	}
	configWithEnv := parseEnvVariables(defaultConfig)
	return configWithEnv, nil
}

func parseEnvVariables(config map[string]interface{}) map[string]interface{} {
	for key, value := range config {
		switch v := value.(type) {
		case map[string]interface{}:
			config[key] = parseEnvVariables(v)
		case string:
			if strings.HasPrefix(v, "$") {
				val, ok := os.LookupEnv(strings.TrimLeft(v, "$"))
				if ok {
					config[key] = val
				}
			}
		}
	}
	return config
}
