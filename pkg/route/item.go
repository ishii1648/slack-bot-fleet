package route

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ishii1648/cloud-run-sdk/util"
	"go.opencensus.io/trace"
	"gopkg.in/yaml.v2"
)

type ReactionAddedItem struct {
	ServiceName  string   `yaml:"service_name"`
	Type         string   `yaml:"type"`
	Users        []string `yaml:"users"`
	Reactions    []string `yaml:"reactions"`
	ItemChannels []string `yaml:"item_channels"`
}

func ParseReactionAddedItem(ymlPath string) ([]ReactionAddedItem, error) {
	var items []ReactionAddedItem

	ymlFile, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(ymlFile, &items); err != nil {
		return nil, err
	}

	for _, item := range items {
		if len(item.Users) == 0 {
			return nil, fmt.Errorf("no set users on %s", ymlPath)
		}
		if len(item.Reactions) == 0 {
			return nil, fmt.Errorf("no set reactions on %s", ymlPath)
		}
		if len(item.ItemChannels) == 0 {
			return nil, fmt.Errorf("no set item_channels on %s", ymlPath)
		}
	}

	return items, nil
}

func (i *ReactionAddedItem) Match(userRealName, reaction, channelName string) bool {
	if len(i.Users) > 0 && !contain(userRealName, i.Users) {
		return false
	}

	if len(i.Reactions) > 0 && !contain(reaction, i.Reactions) {
		return false
	}

	if len(i.ItemChannels) > 0 && !contain(channelName, i.ItemChannels) {
		return false
	}

	return true
}

func (i *ReactionAddedItem) FetchServiceURL(ctx context.Context, projectID, locationID string) (string, error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "ReactionAddedItem.FetchServiceAddr", sc)
	defer span.End()

	url, err := util.FetchURLByServiceName(ctx, i.ServiceName, locationID, projectID)
	if err != nil {
		return "", err
	}

	return url, nil
}

func contain(target string, searchStrs []string) bool {
	for _, str := range searchStrs {
		if target == str {
			return true
		}
	}

	return false
}
