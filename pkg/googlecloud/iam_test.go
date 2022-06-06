package googlecloud

import (
	"os"
	"testing"
)

func TestSetIamPolicyWithFixedTime(t *testing.T) {
	projectID, isSet := os.LookupEnv("GOOGLE_PROJECT_ID")
	if !isSet {
		t.Skip("GOOGLE_PROJECT_ID is not set")
	}

	if err := SetIamPolicyWithFixedTime(projectID, "roles/container.admin"); err != nil {
		t.Fatal(err)
	}
}
