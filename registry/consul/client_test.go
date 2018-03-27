package consul

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"github.com/divya2661/prattle/config"
)

func TestThatItRegistersSuccessfullyWhenRegistrationResponseIsOK(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(200)
	}))
	consulURL := testserver.URL + "/"
	discovery := config.Discovery{
		TTL:                "10s",
		HealthEndpoint:     "http://localhost:3000/",
		HealthPingInterval: "10s",
		Address:            "http://localhost",
		Name:               "Test",
		Port:               1000,
		ConsulURL:          "http://localhost:8500/",
	}
	err := NewClient(consulURL, &http.Client{}).Register(discovery)
	assert.NoError(t, err)
}

func TestThatItDoesntRegisterTheServiceSuccessfullyWhenResponseIsNotOK(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(400)
	}))
	consulURL := testserver.URL + "/"
	discovery := config.Discovery{
		TTL:                "10s",
		HealthEndpoint:     "http://localhost:3000/",
		HealthPingInterval: "10s",
		Address:            "http://localhost",
		Name:               "Test",
		Port:               1000,
		ConsulURL:          "http://localhost:8500/",
	}
	err := NewClient(consulURL, &http.Client{}).Register(discovery)
	assert.Error(t, err)
}

func TestThatIfConsulGivesAllHealthyNodesInCluster(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		fileContents, err := ioutil.ReadFile("../../fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		responseWriter.Write(fileContents)
		responseWriter.WriteHeader(200)
	}))
	consulURL := testserver.URL + "/"
	healthyNode, err := NewClient(consulURL, &http.Client{}).FetchHealthyNode()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", healthyNode)
}

func TestThatItReturnsErrorWHenConsulGivesNoMembers(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Write([]byte("[]"))
		responseWriter.WriteHeader(200)
	}))
	consulURL := testserver.URL + "/"
	healthyNode, err := NewClient(consulURL, &http.Client{}).FetchHealthyNode()
	require.Error(t, err)
	assert.Equal(t, "", healthyNode)
}
