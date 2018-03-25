package consul

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string, httpClient *http.Client) *Client {
	return &Client{url: url, httpClient: httpClient}
}

func (c *Client) Register() error {
	var response *http.Response
	check := Check{DeregisterCriticalServiceAfter: "10m", HTTP: "", Interval: "1s"}
	service := Service{ID: "", Address: "", EnableTagOverride: false, Tags: []string{}, Name: "", Port: 1234, Check: check}
	serviceBytes, _ := json.Marshal(service)
	request, err := http.NewRequest("PUT", c.serviceRegistrationURL(), bytes.NewBuffer(serviceBytes))
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

func (c *Client) serviceRegistrationURL() string {
	return c.url + "v1/agent/service/register"
}
