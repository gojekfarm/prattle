package prattle

import (
	"errors"
	"log"

	"github.com/gojekfarm/prattle/config"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
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

func (client *Client) Register(discovery config.Discovery) (string, error) {
	check := consulAPI.AgentServiceCheck{
		DeregisterCriticalServiceAfter: discovery.TTL,
		TTL:                            discovery.TTL,
	}
	serviceId := uuid.NewV4().String()
	serviceRegistration := consulAPI.AgentServiceRegistration{
		ID:                serviceId,
		Address:           discovery.Address,
		EnableTagOverride: false,
		Tags:              []string{},
		Name:              discovery.Name,
		Check:             &check,
	}
	return serviceId, client.consulClient.Agent().ServiceRegister(&serviceRegistration)
}

func (client *Client) FetchHealthyNode(serviceName string) (string, error) {
	queryOptions := &consulAPI.QueryOptions{}
	services, _, err := client.consulClient.Health().Service(serviceName, "", true, queryOptions)
	if err != nil {
		log.Println(err)
		log.Fatal("Can not fetch service")
	}
	if len(services) == 0 {
		return "", nil
	}
	for _, agentService := range services {
		log.Println("member: " + agentService.Service.Address)
		return agentService.Service.Address, nil
	}
	return "", errors.New("no healthy node found")
}

func (client *Client) Ping(checkID string) error {
	return client.consulClient.Agent().UpdateTTL(checkID, "", consulAPI.HealthPassing)
}