package googlecloud

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type GoogleCloudProject struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

func LoadProjects(path string) ([]GoogleCloudProject, error) {
	var projects []GoogleCloudProject
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}
