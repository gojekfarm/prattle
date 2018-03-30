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

var discovery = config.Discovery{
	TTL:                "10s",
	HealthEndpoint:     "http://localhost:3000/",
	HealthPingInterval: "10s",
	Address:            "http://localhost",
	Name:               "Test",
	Port:               1000,
	ConsulURL:          "http://localhost:8500",
}

func TestNewPrattleWithSingleNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	testServiceAddr, ok := testService.Listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)
	discovery := config.Discovery{
		Name:               "test-prattle-service-02",
		Address:            testServiceAddr.IP.String(),
		Port:               testServiceAddr.Port,
		HealthEndpoint:     fmt.Sprintf("%s/_healthz", testService.URL),
		HealthPingInterval: "10s",
		TTL:                "10s",
	}
	prattle, err := NewPrattle(consul, testServiceAddr.Port, discovery)
	assert.Nil(t, err)
	defer func() {
		if prattle != nil {
			prattle.Shutdown()
		}
	}()
}

func TestPrattleWithMoreThanOneNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.NoError(t, err)
	prattleOne, errOne := NewPrattle(consul, 8080, discovery)
	prattleTwo, errTwo := NewPrattle(consul, 8081, discovery)
	assert.Nil(t, errOne)
	assert.Nil(t, errTwo)
	assert.Equal(t, prattleOne.Members(), prattleTwo.Members())
	assert.Equal(t, int(prattleOne.members.LocalNode().Port), 8080)
	assert.Equal(t, int(prattleTwo.members.LocalNode().Port), 8081)
	assert.Equal(t, 2, prattleOne.members.NumMembers())
	assert.Equal(t, 2, prattleOne.broadcasts.NumNodes())
	assert.Equal(t, 3, prattleOne.broadcasts.RetransmitMult)
	defer prattleOne.Shutdown()
	defer prattleTwo.Shutdown()
}

func TestNewPrattleWhenMemberAddressIsNotInUse(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	prattle, err := NewPrattle(consul, 8080, discovery)
	defer prattle.Shutdown()
	assert.Nil(t, err)
	assert.Equal(t, 1, prattle.members.NumMembers())
	assert.Equal(t, 1, prattle.broadcasts.NumNodes())
	assert.NotNil(t, prattle.database.connection)
}

func TestPrattleWhenMemberAddressIsAlreadyInUse(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.Nil(t, err)
	prattle, errOne := NewPrattle(consul, 8080, discovery)
	defer prattle.Shutdown()
	_, errTwo := NewPrattle(consul, 8080, discovery)
	assert.Nil(t, errOne)
	assert.NotNil(t, errTwo)
}

func TestGetWhenKeyIsNotFound(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.Nil(t, err)
	prattle, _ := NewPrattle(consul, 8080, discovery)
	value, found := prattle.Get("ping")
	assert.False(t, found)
	assert.Equal(t, value, nil)
	defer prattle.Shutdown()
}

func TestGetWhenKeyIsFound(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.Nil(t, err)
	prattle, _ := NewPrattle(consul, 8080, discovery)
	prattle.Set("ping", "pong")
	value, found := prattle.Get("ping")
	assert.True(t, found)
	assert.Equal(t, "pong", value)
	defer prattle.Shutdown()
}

func TestSetWhenKeyAlreadyExist(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:18500")
	assert.Nil(t, err)
	prattle, _ := NewPrattle(consul, 8080, discovery)
	prattle.Set("ping", "pong")
	value, _ := prattle.Get("ping")
	assert.Equal(t, "pong", value)
	prattle.Set("ping", "pong2")
	newValue, _ := prattle.Get("ping")
	assert.Equal(t, "pong2", newValue)
	defer prattle.Shutdown()
}
