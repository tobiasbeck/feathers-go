package feathers

import (
	"os"
	"path"
	"strings"

	"github.com/imdario/mergo"
	"github.com/tobiasbeck/feathers-go/feathers/config"
)

// LoadConfigFile loads a yaml config file
// DEPRECATED! use config.LoadConfigFile instead
func LoadConfigFile(path string) (map[string]interface{}, error) {
	return config.LoadFileMap(path)
}

// LoadConfigDirectory loads a whole directory of .yaml config files.
// DEPRECATED! use config.LoadConfigMap instead
func LoadConfigDirectory(dirPath string) (map[string]map[string]interface{}, error) {
	return config.LoadDirectoryMap(dirPath)
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
