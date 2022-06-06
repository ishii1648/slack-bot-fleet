package googlecloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"cloud.google.com/go/storage"
	"gopkg.in/yaml.v2"
)

type Role struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Resource string `yaml:"resource"`
}

func LoadProjectRoles(path string) ([]Role, error) {
	var roles, projectRoles []Role
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &roles); err != nil {
		return nil, err
	}

	for _, role := range roles {
		if role.Resource == "project" {
			projectRoles = append(projectRoles, role)
		}
	}

	return projectRoles, nil
}

func FetchProjectRoles(ctx context.Context, bucketName string) ([]Role, error) {
	var roles, projectRoles []Role

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	rc, err := bucket.Object("roles.yml").NewReader(ctx)
	if err != nil {
		return nil, err
	}

	yamlFile, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(yamlFile, &roles); err != nil {
		return nil, err
	}

	for _, role := range roles {
		if role.Resource == "project" {
			projectRoles = append(projectRoles, role)
		}
	}

	return projectRoles, nil
}

// I dare to use gcloud command, because projects.setIamPolicy of Resource Manager is a Danger Method.
// see detail. https://stackoverflow.com/questions/59420705/cant-add-user-for-role-editor-user-via-api-in-google-cloud-platform
func SetIamPolicyWithFixedTime(member string, projectID string, roleID string) error {
	condition := fmt.Sprintf("expression=request.time < timestamp(\"%s\"),title=tmp_%s", time.Now().Add(time.Minute*30).Format(time.RFC3339), time.Now().Format(time.RFC3339))

	if _, err := exec.Command("gcloud", "projects", "add-iam-policy-binding", projectID,
		"--member="+member,
		"--condition="+condition,
		"--role="+roleID,
	).Output(); err != nil {
		return err
	}

	return nil
}
