package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	consul "github.com/hashicorp/consul/api"

	"github.com/gojekfarm/prattle/config"
)

// consul.Client makes it easy to communicate with the consul API

type Check struct {
	DeregisterCriticalServiceAfter string
	HTTP                           string
	Interval                       string
}

type Service struct {
	ID                string
	Name              string
	Address           string
	Tags              []string
	Port              int
	EnableTagOverride bool
	Check             Check
}

type Instance struct {
	Service Service
}

type Client struct {
	url          string
	httpClient   *http.Client
	consulClient *consul.Client
	discovery    config.Discovery
}

func NewClient(url string, httpClient *http.Client, discovery config.Discovery) *Client {
	consulClient, _ := consul.NewClient(consul.DefaultConfig())
	return &Client{
		url:          url,
		httpClient:   httpClient,
		consulClient: consulClient,
		discovery:    discovery,
	}
}

func (c *Client) Register() error {
	var response *http.Response
	check := Check{DeregisterCriticalServiceAfter: c.discovery.TTL, HTTP: c.discovery.HealthEndpoint, Interval: c.discovery.HealthPingInterval}
	service := Service{
		ID:                c.discovery.Name,
		Address:           c.discovery.Address,
		EnableTagOverride: false,
		Tags:              []string{},
		Name:              c.discovery.Name,
		Port:              c.discovery.Port,
		Check:             check}
	serviceBytes, _ := json.Marshal(service)
	request, err := http.NewRequest(http.MethodPut, c.serviceRegistrationURL(), bytes.NewBuffer(serviceBytes))
	if err != nil {
		return err
	}
	response, err = c.httpClient.Do(request)
	fmt.Println(response.StatusCode)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("failed to register")
	}
	return nil
}

//TODO: Refactor this, remove business logic
func (c *Client) FetchHealthyNode() (string, error) {
	services, _ := c.consulClient.Agent().Services()
	for _, agentService := range services {
		return fmt.Sprintf(
			"%s:%d",
			agentService.Address,
			agentService.Port), nil
	}
	return "", nil
}

func (c *Client) serviceRegistrationURL() string {
	return c.url + "v1/agent/service/register"
}

func (c *Client) healtyNodesUrl() string {
	return c.url + "v1/agent/services"
}
