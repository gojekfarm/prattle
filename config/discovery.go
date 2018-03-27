package config

type Discovery struct {
	TTL                string
	HealthEndpoint     string
	HealthPingInterval string
	Address            string
	Name               string
	Port               int
	ConsulURL          string
}
