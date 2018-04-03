package prattle

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojekfarm/prattle/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrattleWithSingleNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discovery := config.Discovery{
		Name:               "test-prattle-service-02",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     fmt.Sprintf("%s/_healthz", testService.URL),
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, err := NewPrattle(consul, 9000, discovery)
	require.NoError(t, err)
	prattle.Shutdown()
}

func TestPrattleWithMoreThanOneNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	testServiceOne := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discoveryOne := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     testServiceOne.URL,
		HealthPingInterval: "1s",
		TTL:                "10s",
	}
	testServiceTwo := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discoveryTwo := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9001",
		HealthEndpoint:     testServiceTwo.URL,
		HealthPingInterval: "1s",
		TTL:                "10s",
	}
	prattleOne, errOne := NewPrattle(consul, 9000, discoveryOne)
	time.Sleep(1 * time.Second)
	prattleTwo, errTwo := NewPrattle(consul, 9001, discoveryTwo)
	time.Sleep(1 * time.Second)
	require.NoError(t, errOne)
	require.NoError(t, errTwo)
	defer prattleOne.Shutdown()
	defer prattleTwo.Shutdown()
	assert.Equal(t, prattleOne.Members(), prattleTwo.Members())
	assert.Equal(t, int(prattleOne.members.LocalNode().Port), 9000)
	assert.Equal(t, int(prattleTwo.members.LocalNode().Port), 9001)
	assert.Equal(t, 2, prattleOne.members.NumMembers())
	assert.Equal(t, 2, prattleOne.broadcasts.NumNodes())
}

func TestPrattleWhenMemberAddressIsAlreadyInUse(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discovery := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     testService.URL,
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, errOne := NewPrattle(consul, 9000, discovery)
	defer prattle.Shutdown()
	_, errTwo := NewPrattle(consul, 9000, discovery)
	require.NoError(t, errOne)
	require.Error(t, errTwo)
}

func TestGetWhenKeyIsNotFound(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discovery := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     testService.URL,
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, _ := NewPrattle(consul, 9000, discovery)
	value, found := prattle.Get("ping")
	assert.False(t, found)
	assert.Equal(t, value, nil)
	defer prattle.Shutdown()
}

func TestGetWhenKeyIsFound(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discovery := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     testService.URL,
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, _ := NewPrattle(consul, 9000, discovery)
	prattle.Set("ping", "pong")
	value, found := prattle.Get("ping")
	assert.True(t, found)
	assert.Equal(t, "pong", value)
	defer prattle.Shutdown()
}

func TestSetWhenKeyAlreadyExist(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	discovery := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		HealthEndpoint:     testService.URL,
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, _ := NewPrattle(consul, 9000, discovery)
	prattle.Set("ping", "pong")
	value, _ := prattle.Get("ping")
	assert.Equal(t, "pong", value)
	prattle.Set("ping", "pong2")
	newValue, _ := prattle.Get("ping")
	assert.Equal(t, "pong2", newValue)
	defer prattle.Shutdown()
}
