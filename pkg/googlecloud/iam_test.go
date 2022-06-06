package googlecloud

import (
	"context"
	"os"
	"testing"
)

func TestSetIamPolicyWithFixedTime(t *testing.T) {
	projectID, isSet := os.LookupEnv("GOOGLE_PROJECT_ID")
	if !isSet {
		t.Skip("GOOGLE_PROJECT_ID is not set")
	}

	member, isSet := os.LookupEnv("IAM_MEMBER")
	if !isSet {
		t.Skip("IAM_MEMBER is not set")
	}

	if err := SetIamPolicyWithFixedTime(member, projectID, "roles/container.admin"); err != nil {
		t.Fatal(err)
	}
}

func TestFetchProjectRoles(t *testing.T) {
	ctx := context.Background()

	bucketName, isSet := os.LookupEnv("GOOGLE_BUCKET_NAME")
	if !isSet {
		t.Skip("GOOGLE_BUCKET_NAME is not set")
	}

	projectRoles, err := FetchProjectRoles(ctx, bucketName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(projectRoles)
}
