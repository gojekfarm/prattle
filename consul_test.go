package prattle

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojekfarm/prattle/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//:TODO Refactor and remove time.sleep
func TestRegisterWhenANodeIsHealthy(t *testing.T) {
	client, err := NewConsulClient("127.0.0.1:8500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	testServiceAddr := testService.Listener.Addr().String()
	testServiceName := "test-service-01"
	discovery := config.Discovery{
		Name:               testServiceName,
		Address:            testServiceAddr,
		HealthEndpoint:     testService.URL,
		HealthPingInterval: "1s",
		TTL:                "1s",
	}
	id, regErr := client.Register(discovery)
	require.NoError(t, regErr)
	time.Sleep(1 * time.Second)
	healthyServiceAddr, err := client.FetchHealthyNode(testServiceName)
	assert.NoError(t, err)
	assert.Equal(t, testServiceAddr, healthyServiceAddr)
	deregErr := client.consulClient.Agent().ServiceDeregister(id)
	require.NoError(t, deregErr)
}

func TestRegisterWhenANodeIsUnhealthy(t *testing.T) {
	client, err := NewConsulClient("127.0.0.1:8500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(500)
		}))
	testServiceAddr := testService.Listener.Addr().String()
	testServiceName := "test-service-02"
	service := config.Discovery{
		Name:               testServiceName,
		Address:            testServiceAddr,
		HealthEndpoint:     fmt.Sprintf("%s/_healthz", testService.URL),
		HealthPingInterval: "1s",
		TTL:                "1s",
	}
	id, regErr := client.Register(service)
	assert.NoError(t, regErr)
	time.Sleep(1 * time.Second)
	member, err := client.FetchHealthyNode(testServiceName)
	fmt.Println(member)
	assert.Equal(t, "", member)
	assert.NoError(t, err)
	deregErr := client.consulClient.Agent().ServiceDeregister(id)
	require.NoError(t, deregErr)
}

func TestTwoServicesRegistrationWhenOneIsUnhealthy(t *testing.T) {
	client, err := NewConsulClient("127.0.0.1:8500")
	assert.NoError(t, err)
	testServiceOne := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	testServiceTwo := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(500)
		}))
	testServiceAddrOne := testServiceOne.Listener.Addr().(*net.TCPAddr).String()
	testServiceAddrTwo := testServiceTwo.Listener.Addr().(*net.TCPAddr).String()

	testServiceNameOne := "test-service-01"
	serviceOne := config.Discovery{
		Name:               testServiceNameOne,
		Address:            testServiceAddrOne,
		HealthEndpoint:     "http://" + testServiceAddrOne + "/ping",
		HealthPingInterval: "1s",
		TTL:                "1s",
	}
	testServiceNameTwo := "test-service-02"
	serviceTwo := config.Discovery{
		Name:               testServiceNameTwo,
		Address:            testServiceAddrTwo,
		HealthEndpoint:     "http://" + testServiceAddrTwo + "/ping",
		HealthPingInterval: "1s",
		TTL:                "1s",
	}
	fmt.Println("health: " + testServiceAddrTwo)
	idOne, regErrOne := client.Register(serviceOne)
	idTwo, regErrTwo := client.Register(serviceTwo)
	time.Sleep(1 * time.Second)
	assert.NoError(t, regErrOne)
	assert.NoError(t, regErrTwo)
	member, err := client.FetchHealthyNode(testServiceNameOne)
	assert.Equal(t, serviceOne.Address, member)
	assert.NoError(t, err)
	deregErrOne := client.consulClient.Agent().ServiceDeregister(idOne)
	require.NoError(t, deregErrOne)
	deregErrTwo := client.consulClient.Agent().ServiceDeregister(idTwo)
	require.NoError(t, deregErrTwo)
}
