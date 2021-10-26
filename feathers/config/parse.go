package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/tobiasbeck/feathers-go/feathers/yaml"
)

// LoadFileMap loads a yaml config file which has a hierachical structure
func LoadFileMap(path string) (map[string]interface{}, error) {
	var configs interface{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("FILE: '%s'", path)))
		return nil, err
	}
	mapConfig, ok := configs.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("config is of type %T and not map[string]interface{}", configs)
	}
	return mapConfig, nil
}

// LoadFileSlice loads a yaml config file which is a list
func LoadFile(path string) (interface{}, error) {
	var configs interface{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("FILE: '%s'", path)))
		return nil, err
	}

	return configs, nil
}

// LoadFileSlice loads a yaml config file which is a list
func LoadFileSlice(path string) ([]interface{}, error) {
	var configs interface{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &configs); err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("FILE: '%s'", path)))
		return nil, err
	}
	sliceConfig, ok := configs.([]interface{})
	if !ok {
		return nil, fmt.Errorf("config is of type %T and not []interface{}", configs)
	}
	return sliceConfig, nil
}

func fileEnding(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func fileName(name string) string {
	parts := strings.Split(name, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

// LoadDirectoryMap loads a whole directory of .yaml config files.
func LoadDirectoryMap(dirPath string) (map[string]map[string]interface{}, error) {
	configs := map[string]map[string]interface{}{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ending := fileEnding(file.Name())
		if ending != "yaml" {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", dirPath, file.Name())
		data, err := LoadFileMap(filePath)
		key := fileName(file.Name())
		if err != nil {
			if strings.Contains(err.Error(), "map[string]interface{}") {
				continue
			}
			return nil, errors.Wrap(err, "file: "+file.Name())
		}
		configs[key] = data
	}
	return configs, nil
}

// LoadDirectorySlice loads a whole directory of .yaml config files.
func LoadDirectorySlice(dirPath string) (map[string][]interface{}, error) {
	configs := map[string][]interface{}{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ending := fileEnding(file.Name())
		if ending != "yaml" {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", dirPath, file.Name())
		data, err := LoadFileSlice(filePath)
		key := fileName(file.Name())
		if err != nil {
			if strings.Contains(err.Error(), "[]interface{}") {
				continue
			}
			return nil, errors.Wrap(err, "file: "+file.Name())
		}
		configs[key] = data
	}
	return configs, nil
}

// LoadDirectory loads a whole directory of .yaml config files.
func LoadDirectory(dirPath string) (map[string]interface{}, error) {
	configs := map[string]interface{}{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ending := fileEnding(file.Name())
		if ending != "yaml" {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", dirPath, file.Name())
		data, err := LoadFile(filePath)
		key := fileName(file.Name())
		if err != nil {
			return nil, errors.Wrap(err, "file: "+file.Name())
		}
		configs[key] = data
	}
	return configs, nil
}
