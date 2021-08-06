package event

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

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

func (i *ReactionAddedItem) FetchServiceAddr(ctx context.Context) (addr string, isLocalhost bool, err error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "ReactionAddedItem.FetchServiceAddr", sc)
	defer span.End()

	port, isSet := os.LookupEnv("GRPC_PORT")
	if !isSet {
		port = "443"
	}

	if util.IsCloudRun() {
		var projectID, url string
		projectID, err = util.FetchProjectID()
		if err != nil {
			return addr, isLocalhost, err
		}

		location, isSet := os.LookupEnv("CLOUD_RUN_LOCATION")
		if !isSet {
			location = "asia-northeast1"
		}

		url, err = util.FetchURLByServiceName(ctx, i.ServiceName, location, projectID)
		if err != nil {
			return addr, isLocalhost, err
		}

		addr = strings.Replace(url, "https://", "", 1) + ":" + port

		return addr, isLocalhost, nil
	}

	addr = "localhost:" + port
	isLocalhost = true

	return addr, isLocalhost, nil
}

func contain(target string, searchStrs []string) bool {
	for _, str := range searchStrs {
		if target == str {
			return true
		}
	}

	return false
}
