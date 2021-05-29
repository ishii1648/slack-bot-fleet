package service

import (
	"context"
	"errors"
	"io/ioutil"

	sdk "github.com/ishii1648/slack-bot-fleet/pkg/cloud-run-sdk"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type ReactionEventCriteria struct {
	ServiceName  string   `yaml:"service_name"`
	Type         string   `yaml:"type"`
	Users        []string `yaml:"users"`
	Reactions    []string `yaml:"reactions"`
	ItemChannels []string `yaml:"item_channels"`
}

func ParseReactionEventCriteriaYml(ymlPath string) ([]ReactionEventCriteria, error) {
	var rawCriterias, criterias []ReactionEventCriteria

	ymlFile, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(ymlFile, &rawCriterias); err != nil {
		return nil, err
	}

	for _, criteria := range rawCriterias {
		if criteria.Type == "reaction_added" {
			criterias = append(criterias, criteria)
		}
	}

	return criterias, nil
}

func (t *ReactionEventCriteria) Match(user, reaction, itemChannel string) bool {
	if len(t.Users) > 0 && !contain(user, t.Users) {
		return false
	}

	if len(t.Reactions) > 0 && !contain(reaction, t.Reactions) {
		return false
	}

	if len(t.ItemChannels) > 0 && !contain(itemChannel, t.ItemChannels) {
		return false
	}

	return true
}

func (t *ReactionEventCriteria) GetService(ctx context.Context, l *zerolog.Logger, serviceFactories map[serviceName]reactionEventServiceFactory, localForwardFlag bool) (ReactionEventService, error) {
	for msName, factory := range serviceFactories {
		if msName.isService(t.ServiceName) {
			l.Info().Msgf("match service : %s", msName)

			if localForwardFlag {
				return factory(l, "localhost:8080"), nil
			}

			addr, err := sdk.FetchURLByServiceName(ctx, string(msName), "asia-east1")
			if err != nil {
				return nil, err
			}

			return factory(l, addr), nil
		}
	}

	return nil, errors.New("service is missing")
}

func contain(target string, searchStrs []string) bool {
	for _, str := range searchStrs {
		if target == str {
			return true
		}
	}

	return false
}
