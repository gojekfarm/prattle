package prattle

import (
	"errors"

	consulAPI "github.com/hashicorp/consul/api"

	"github.com/gojekfarm/prattle/config"
)

type Client struct {
	consulClient *consulAPI.Client
}

func NewConsulClient(consulAddress string) (*Client, error) {
	config := consulAPI.DefaultConfig()
	if consulAddress != "" {
		config.Address = consulAddress
	}
	consulClient, err := consulAPI.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		consulClient: consulClient,
	}, nil
}

func (client *Client) Register(discovery config.Discovery) error {
	check := consulAPI.AgentServiceCheck{
		HTTP:                           discovery.HealthEndpoint,
		Interval:                       discovery.HealthPingInterval,
		DeregisterCriticalServiceAfter: discovery.TTL,
	}
	serviceRegistration := consulAPI.AgentServiceRegistration{
		ID:                discovery.Name,
		Address:           discovery.Address,
		EnableTagOverride: false,
		Tags:              []string{},
		Name:              discovery.Name,
		Check:             &check,
	}
	return client.consulClient.Agent().ServiceRegister(&serviceRegistration)
}

func (client *Client) FetchHealthyNode() (string, error) {
	services, _ := client.consulClient.Agent().Services()
	if len(services) == 0 {
		return "", nil
	}
	for _, agentService := range services {
		return agentService.Address, nil
	}
	return "", errors.New("no healthy node found")
}
