package prattle

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojekfarm/prattle/config"
	"github.com/stretchr/testify/assert"
)

func TestRegisterWhenANodeIsHealthy(t *testing.T) {
	client, err := NewConsulClient("127.0.0.1:18500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	testServiceAddr, ok := testService.Listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)
	err = client.Register(config.Discovery{
		Name:               "test-service-01",
		Address:            testServiceAddr.String(),
		HealthEndpoint:     fmt.Sprintf("%s/_healthz", testService.URL),
		HealthPingInterval: "10s",
		TTL:                "10s",
	})
	assert.NoError(t, err)
	healthyServiceAddr, err := client.FetchHealthyNode()
	assert.NoError(t, err)
	assert.Equal(t, testServiceAddr.String(), healthyServiceAddr)
}

func TestRegisterWhenANodeIsUnhealthy(t *testing.T) {
	client, err := NewConsulClient("127.0.0.1:18500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(500)
		}))
	testServiceAddr, ok := testService.Listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)
	err = client.Register(config.Discovery{
		Name:               "test-service-02",
		Address:            testServiceAddr.IP.String(),
		HealthEndpoint:     fmt.Sprintf("%s/_healthz", testService.URL),
		HealthPingInterval: "10s",
		TTL:                "10s",
	})
	assert.NoError(t, err)
	_, err = client.FetchHealthyNode()
	assert.NoError(t, err)
}
