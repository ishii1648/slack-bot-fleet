package sdk

import (
	"context"
	"fmt"

	"google.golang.org/api/run/v1"
)

func FetchURLByServiceName(ctx context.Context, name, region string) (string, error) {
	c, err := run.NewService(ctx)
	if err != nil {
		return "", err
	}
	c.BasePath = fmt.Sprintf("https://%s-run.googleapis.com/", region)

	service, err := c.Namespaces.Services.Get(fmt.Sprintf("namespaces/%s/services/%s", ProjectID, name)).Do()
	if err != nil {
		return "", err
	}

	return service.Status.Url, nil
}
