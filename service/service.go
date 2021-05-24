package service

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Name             string `yaml:"name"`
	ForwardCondition `yaml:"forward_condition"`
}

type ForwardCondition struct {
	Channel  string   `yaml:"channel"`
	Reaction string   `yaml:"reaction"`
	Users    []string `yaml:"users"`
}

func ParseServicesYml(ymlPath string) ([]Service, error) {
	var services []Service

	ymlFile, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(ymlFile, &services); err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *Service) getService(channelName, reaction, user string) *Service {
	if channelName != svc.Channel || reaction != svc.Reaction {
		return nil
	}

	if !isUser(user, svc.Users) {
		return nil
	}

	return svc
}

func isUser(searchUser string, users []string) bool {
	for _, user := range users {
		if searchUser == user {
			return true
		}
	}
	return false
}

var microServiceFactories = make(map[microServiceName]microServiceFactory)

type microServiceName string

func (msn microServiceName) isService(serviceName string) bool {
	return msn == microServiceName(serviceName)
}

type microServiceFactory func() MicroService

type MicroService interface {
	Rpc() error
}

func registerMicroService(name microServiceName, factory microServiceFactory) {
	microServiceFactories[name] = factory
}

func GetMicroService(services []Service, channelName, reaction, user string) (MicroService, error) {
	var svc *Service
	for _, s := range services {
		svc = s.getService(channelName, reaction, user)
	}

	for msName, factory := range microServiceFactories {
		if msName.isService(svc.Name) {
			return factory(), nil
		}
	}

	return nil, errors.New("micro service is missing")
}
