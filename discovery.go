package prattle

import (
	"github.com/hashicorp/consul/api"
	"log"
	"time"
)

func Register() {
	config := api.DefaultConfig()
	config.Address = "127.0.0.1"
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("Can not create client", err)
	}
	agent := client.Agent()
	check := api.AgentServiceCheck{
		HTTP:                           "127.0.0.1",
		Interval:                       string(time.Duration(10 * time.Second)),
		DeregisterCriticalServiceAfter: string(time.Duration(2 * time.Minute)),
	}
	service := api.AgentServiceRegistration{
		ID:      "string",
		Name:    "string",
		Port:    8080,
		Address: "string",
		Check:   &check,
	}
	err1 := agent.ServiceRegister(&service)
	if err1 != nil {
		log.Fatal("Can not create client", err)
	}
}
