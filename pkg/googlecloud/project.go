package googlecloud

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
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

func FetchProjects(ctx context.Context, bucketName string) ([]GoogleCloudProject, error) {
	var projects []GoogleCloudProject

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	rc, err := bucket.Object("projects.yml").NewReader(ctx)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}
